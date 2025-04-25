package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/midi/writer"
	"gitlab.com/gomidi/portmididrv"
)

var clients = make(map[*websocket.Conn]chan interface{}) // Connected clients with send channels
var broadcast = make(chan MIDIMessage)                   // MIDI messages to broadcast

var outOutPort midi.Out
var midiWriter *writer.Writer

var upgrader = websocket.Upgrader{} // Use default options

type MIDIMessage struct {
	Type     string `json:"type"`
	Note     uint8  `json:"note"`
	Velocity uint8  `json:"velocity"`
}

type CueMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type UpdateLabelMessage struct {
	Type  string `json:"type"`
	Note  uint8  `json:"note"`
	Label string `json:"label"`
}

type BulkLabelUpdateMessage struct {
	Type   string           `json:"type"`
	Labels map[uint8]string `json:"labels"`
}

type Scene struct {
	Cue    string
	Labels map[uint8]string
}

var scenes = []Scene{
	{
		Cue: "Welcome to the MIDI Controller!",
		Labels: map[uint8]string{
			60: "C$", 62: "D4", 64: "E4", 65: "F4", 67: "G4", 69: "G4", 71: "A4", 72: "B4", 74: "C5",
		},
	},
	{
		Cue: "🎵 Scene 1: Warmup",
		Labels: map[uint8]string{
			60: "Kick", 62: "Snare", 64: "Hat", 65: "Tom", 67: "Bass", 69: "Synth", 71: "Pad", 72: "Lead", 74: "FX",
		},
	},
	{
		Cue: "🔥 Scene 2: Chorus",
		Labels: map[uint8]string{
			60: "Boom", 62: "Clap", 64: "Shaker", 65: "808", 67: "Sub", 69: "Keys", 71: "Strings", 72: "Vox", 74: "Noise",
		},
	},
	{
		Cue: "🚀 Scene 3: Drop!",
		Labels: map[uint8]string{
			60: "Kick+", 62: "Snare+", 64: "Hat+", 65: "Synth+", 67: "Bass+", 69: "FX+", 71: "Riser", 72: "Drop", 74: "Impact",
		},
	},
	{
		Cue: "🌌 Scene 4: Outro",
		Labels: map[uint8]string{
			60: "Fade", 62: "Echo", 64: "Reverb", 65: "Chill", 67: "Ambient", 69: "Vibe", 71: "Outro", 72: "Goodbye", 74: "Silence",
		},
	},
}

var currentScene int

func broadcastScene() {
	scene := scenes[currentScene%len(scenes)]
	currentScene++

	log.Printf("Broadcasting scene: %s\n", scene.Cue)
	// Broadcast cue
	cue := CueMessage{Type: "cue", Text: scene.Cue}
	for client, send := range clients {
		select {
		case send <- cue:
		default:
			close(send)
			delete(clients, client)
		}
	}

	// // Broadcast labels
	// update := BulkLabelUpdateMessage{Type: "bulkUpdateLabels", Labels: scene.Labels}

	// for client, send := range clients {
	// 	select {
	// 	case send <- update:
	// 	default:
	// 		close(send)
	// 		delete(clients, client)
	// 	}
	// }
}

func main() {
	// Initialize the MIDI driver
	drv, err := portmididrv.New()
	if err != nil {
		log.Fatal(err)
	}

	outs, err := drv.Outs()
	if err != nil {
		log.Fatal(err)
	}

	// Find and open the IAC Driver
	found := false
	for _, o := range outs {
		if o.String() == "IAC Driver Bus 1" {
			if err := o.Open(); err != nil {
				log.Fatal(err)
			}
			outOutPort = o
			midiWriter = writer.New(outOutPort)
			found = true
			break
		}
	}
	if !found {
		log.Fatal("IAC Driver not found")
	}

	// Start WebSocket broadcaster
	go handleMessages()

	// Setup HTTP server
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming MIDI
	go listenToMIDI()

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true } // Allow all origins
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	send := make(chan interface{})
	clients[ws] = send

	go func() {
		for msg := range send {
			err := ws.WriteJSON(msg)
			if err != nil {
				log.Println("WebSocket write error:", err)
				ws.Close()
				delete(clients, ws)
				break
			}
		}
	}()

	for {
		var rawMsg map[string]interface{}
		err := ws.ReadJSON(&rawMsg)
		if err != nil {
			log.Println("WebSocket read error:", err)
			close(send)
			delete(clients, ws)
			break
		}
		if msgType, ok := rawMsg["type"].(string); ok && msgType == "nextScene" {
			log.Println("Received nextScene request from client.")
			broadcastScene()
			continue
		} else if msgType == "note" {
			// Reconstruct a MIDIMessage and send to broadcast
			note := uint8(rawMsg["note"].(float64))
			velocity := uint8(rawMsg["velocity"].(float64))
			broadcast <- MIDIMessage{Type: "note", Note: note, Velocity: velocity}
		}
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		log.Printf("Broadcasting Note: %d Velocity: %d\n", msg.Note, msg.Velocity)

		if msg.Type == "note" {
			// Send to IAC Driver
			if midiWriter != nil {
				err := writer.NoteOn(midiWriter, msg.Note, msg.Velocity)
				if err != nil {
					log.Println("MIDI out error:", err)
				}

				// Automatically send a NoteOff after a short delay
				go func(note uint8) {
					time.Sleep(200 * time.Millisecond)
					err := writer.NoteOff(midiWriter, note)
					if err != nil {
						log.Println("MIDI out NoteOff error:", err)
					}
				}(msg.Note)
			}
		}

		for client, send := range clients {
			select {
			case send <- msg:
			default:
				close(send)
				delete(clients, client)
			}
		}
	}
}

func listenToMIDI() {
	drv, err := portmididrv.New()
	if err != nil {
		log.Fatal(err)
	}
	defer drv.Close()

	ins, err := drv.Ins()
	if err != nil {
		log.Fatal(err)
	}

	if len(ins) == 0 {
		log.Fatal("No MIDI input devices found.")
	}

	in := ins[0]
	if err := in.Open(); err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	rdr := reader.New(
		reader.NoteOn(func(pos *reader.Position, channel, key, velocity uint8) {
			log.Printf("NoteOn: Channel %d, Key %d, Velocity %d", channel, key, velocity)
			broadcast <- MIDIMessage{Type: "note", Note: key, Velocity: velocity}
		}),
	)

	fmt.Println("Listening for MIDI input:", in.String())
	rdr.ListenTo(in)
}
