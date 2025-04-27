// midi-server: A Go server that connects MIDI input/output with WebSocket clients.
// It sends and receives MIDI note messages and allows scene-based cue broadcasts.

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

// --- Global Variables ---

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

// --- Scene Definitions ---

type Scene struct {
	Cue    string
	Labels map[uint8]string
}

var scenes = []Scene{
	{
		Cue: "🎛️ Welcome to the MIDI Control Prototype!",
		Labels: map[uint8]string{
			60: "WebSocket", 62: "MIDI", 64: "Realtime", 65: "Client", 67: "Server", 69: "Scene", 71: "Cue", 72: "Pad", 74: "Control",
		},
	},
	{
		Cue: "🌐 Powered by Go and Web Technologies",
		Labels: map[uint8]string{
			60: "GoLang", 62: "HTTP", 64: "WebSocket", 65: "MIDI I/O", 67: "Frontend", 69: "Backend", 71: "PortMidi", 72: "Bootstrap", 74: "Gorilla",
		},
	},
	{
		Cue: "🎶 Designed for Live Interaction",
		Labels: map[uint8]string{
			60: "Touch", 62: "Buttons", 64: "LED", 65: "Cue", 67: "Velocity", 69: "Scenes", 71: "Next", 72: "Advance", 74: "Update",
		},
	},
	{
		Cue: "🚀 Prototype for Future Expansion",
		Labels: map[uint8]string{
			60: "Mobile", 62: "Performance", 64: "DAW", 65: "Ableton", 67: "OSC", 69: "MIDI 2.0", 71: "Cloud", 72: "Realtime", 74: "AI",
		},
	},
	{
		Cue: "🚀 Possible Applications - Musical Performance with Audience Interaction, Trivia Night, Quizes",
		Labels: map[uint8]string{
			60: "Mobile", 62: "Performance", 64: "DAW", 65: "Ableton", 67: "OSC", 69: "MIDI 2.0", 71: "Cloud", 72: "Realtime", 74: "AI",
		},
	},
}

var currentScene int

// --- Scene Broadcasting ---

func broadcastScene() {
	scene := scenes[currentScene%len(scenes)]
	currentScene++

	log.Printf("Broadcasting scene: %s\n", scene.Cue)

	// Broadcast cue message to all clients
	cue := CueMessage{Type: "cue", Text: scene.Cue}
	for client, send := range clients {
		select {
		case send <- cue:
		default:
			// If send channel is blocked, close and remove client
			close(send)
			delete(clients, client)
		}
	}
}

// --- Main Server Setup ---

func main() {
	// Initialize MIDI output driver
	drv, err := portmididrv.New()
	if err != nil {
		log.Fatal(err)
	}

	outs, err := drv.Outs()
	if err != nil {
		log.Fatal(err)
	}

	// Find and open the IAC Driver output port
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

	// Start WebSocket broadcaster goroutine
	go handleMessages()

	// Setup HTTP server and WebSocket handling
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming MIDI input
	go listenToMIDI()

	// Periodically broadcast scenes every 5 seconds
	go func() {
		for {
			time.Sleep(5 * time.Second)
			broadcastScene()
		}
	}()

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// --- WebSocket Connection Handling ---

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Accept new WebSocket connections
	upgrader.CheckOrigin = func(r *http.Request) bool { return true } // Allow all origins
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	send := make(chan interface{})
	clients[ws] = send

	// Goroutine to write messages to WebSocket client
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

	// Read incoming WebSocket messages
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
			log.Println("Received 'note' message from client:", rawMsg)

			noteRaw, noteOk := rawMsg["note"]
			velocityRaw, velocityOk := rawMsg["velocity"]

			if !noteOk || !velocityOk {
				log.Println("Malformed 'note' message: missing fields")
				continue
			}

			note, noteCastOk := noteRaw.(float64)
			velocity, velocityCastOk := velocityRaw.(float64)

			if !noteCastOk || !velocityCastOk {
				log.Println("Malformed 'note' message: fields not numbers")
				continue
			}

			log.Printf("Parsed Note: %d, Velocity: %d\n", uint8(note), uint8(velocity))
			broadcast <- MIDIMessage{Type: "note", Note: uint8(note), Velocity: uint8(velocity)}
		}
	}

}

// --- Message Broadcasting to Clients ---

func handleMessages() {
	for {
		msg := <-broadcast
		log.Printf("Broadcasting Note: %d Velocity: %d\n", msg.Note, msg.Velocity)

		if msg.Type == "note" {
			// Send MIDI NoteOn to output device
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

		// Broadcast message to all connected clients
		for client, send := range clients {
			select {
			case send <- msg:
			default:
				// If send channel is blocked, close and remove client
				close(send)
				delete(clients, client)
			}
		}
	}
}

// --- Listening to Incoming MIDI Notes ---

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

	// Open first available MIDI input
	in := ins[0]
	if err := in.Open(); err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	// Setup reader to broadcast NoteOn events
	rdr := reader.New(
		reader.NoteOn(func(pos *reader.Position, channel, key, velocity uint8) {
			log.Printf("NoteOn: Channel %d, Key %d, Velocity %d", channel, key, velocity)
			broadcast <- MIDIMessage{Type: "note", Note: key, Velocity: velocity}
		}),
	)

	fmt.Println("Listening for MIDI input:", in.String())
	rdr.ListenTo(in)
}
