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

var clients = make(map[*websocket.Conn]bool) // Connected clients
var broadcast = make(chan MIDIMessage)       // MIDI messages to broadcast

var outOutPort midi.Out
var midiWriter *writer.Writer

var upgrader = websocket.Upgrader{} // Use default options

type MIDIMessage struct {
	Note     uint8 `json:"note"`
	Velocity uint8 `json:"velocity"`
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
	defer ws.Close()

	clients[ws] = true

	for {
		var msg MIDIMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println("WebSocket read error:", err)
			delete(clients, ws)
			break
		}
		// Broadcast the received message
		broadcast <- msg
	}
}
func handleMessages() {
	for {
		msg := <-broadcast
		log.Printf("Broadcasting Note: %d Velocity: %d\n", msg.Note, msg.Velocity)

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

		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
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
			broadcast <- MIDIMessage{Note: key, Velocity: velocity}
		}),
	)

	fmt.Println("Listening for MIDI input:", in.String())
	rdr.ListenTo(in)
}
