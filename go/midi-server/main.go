// midi-server: A Go server that connects MIDI input/output with WebSocket clients.
// It sends and receives MIDI note messages and allows scene-based cue broadcasts.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/midi/writer"
	"gitlab.com/gomidi/portmididrv"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"

	idleTimeout   = 5 * time.Minute
	sceneInterval = 5 * time.Second
)

func logServer(format string, args ...interface{}) {
	log.Printf(colorGreen+"[SERVER] "+format+colorReset, args...)
}

func logMIDI(format string, args ...interface{}) {
	log.Printf(colorBlue+"[MIDI] "+format+colorReset, args...)
}

func logWS(format string, args ...interface{}) {
	log.Printf(colorCyan+"[WS] "+format+colorReset, args...)
}

func logTimeout(format string, args ...interface{}) {
	log.Printf(colorYellow+"[TIMEOUT] "+format+colorReset, args...)
}

func logError(format string, args ...interface{}) {
	log.Printf(colorRed+"[ERROR] "+format+colorReset, args...)
}

// MIDIManager encapsulates MIDI setup, listening, and note handling
type MIDIManager struct {
	driverCloser io.Closer
	writer       *writer.Writer
	out          midi.Out
	in           midi.In
}

func (m *MIDIManager) Setup() error {
	var err error
	d, err := portmididrv.New()
	if err != nil {
		return err
	}
	m.driverCloser = d

	outs, err := d.Outs()
	if err != nil {
		return err
	}

	found := false
	for _, o := range outs {
		if o.String() == "IAC Driver Bus 1" {
			if err := o.Open(); err != nil {
				return err
			}
			m.out = o
			m.writer = writer.New(m.out)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("IAC Driver not found")
	}

	ins, err := d.Ins()
	if err != nil {
		return err
	}
	if len(ins) == 0 {
		return fmt.Errorf("No MIDI input devices found.")
	}

	m.in = ins[0]
	if err := m.in.Open(); err != nil {
		return err
	}

	return nil
}

func (m *MIDIManager) Listen() {
	rdr := reader.New(
		reader.NoteOn(func(pos *reader.Position, channel, key, velocity uint8) {
			logMIDI("NoteOn: Channel %d, Key %d, Velocity %d", channel, key, velocity)
			hub.Broadcast <- MIDIMessage{Type: "note", Note: key, Velocity: velocity}
		}),
	)

	logMIDI("Listening for MIDI input: %s", m.in.String())
	rdr.ListenTo(m.in)
}

func (m *MIDIManager) Close() {
	if m.in != nil {
		m.in.Close()
	}
	if m.out != nil {
		m.out.Close()
	}
	if m.driverCloser != nil {
		m.driverCloser.Close()
	}
}

func (m *MIDIManager) NoteOn(note uint8, velocity uint8) error {
	if m.writer == nil {
		return fmt.Errorf("MIDI writer not initialized")
	}
	return writer.NoteOn(m.writer, note, velocity)
}

func (m *MIDIManager) NoteOff(note uint8) error {
	if m.writer == nil {
		return fmt.Errorf("MIDI writer not initialized")
	}
	return writer.NoteOff(m.writer, note)
}

// FlushAllNotes sends "All Notes Off" (CC#123) on all MIDI channels.
func (m *MIDIManager) FlushAllNotes() {
	if m.writer == nil {
		return
	}
	for ch := uint8(0); ch < 16; ch++ {
		// m.writer = writer.Channel(m.writer, ch)
		writer.ControlChange(m.writer, 123, 0) // CC#123 All Notes Off
	}
}

// --- Global Variables ---

type WebSocketClient struct {
	Conn  *websocket.Conn
	Send  chan interface{}
	Timer *time.Timer
	Once  sync.Once
}

func (c *WebSocketClient) Close() {
	c.Once.Do(func() {
		c.Timer.Stop()
		close(c.Send)
		c.Conn.Close()
		hub.Unregister(c)
	})
}

type Hub struct {
	Clients   map[*websocket.Conn]*WebSocketClient
	Broadcast chan interface{}
	Shutdown  chan struct{}
}

func (h *Hub) Register(client *WebSocketClient) {
	h.Clients[client.Conn] = client
}

func (h *Hub) Unregister(client *WebSocketClient) {
	delete(h.Clients, client.Conn)
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case msg := <-h.Broadcast:

			if m, ok := msg.(MIDIMessage); ok {
				logMIDI("Broadcast Note: %d Velocity: %d", m.Note, m.Velocity)

				atomic.AddInt64(&noteEventsThisPeriod, 1)

				if m.Type == "note" {
					noteStatusMutex.Lock()
					if noteStatus[m.Note] {
						noteStatusMutex.Unlock()
						continue
					}
					noteStatus[m.Note] = true
					noteStatusMutex.Unlock()

					if midiManager != nil {
						err := midiManager.NoteOn(m.Note, m.Velocity)
						if err != nil {
							logError("MIDI out error: %v", err)
						}
						go func(note uint8) {
							time.Sleep(200 * time.Millisecond)
							err := midiManager.NoteOff(note)
							if err != nil {
								if err.Error() != fmt.Sprintf("can't write channel.NoteOff channel 0 key %d. note is not running.", note) {
									logError("MIDI out NoteOff error: %v", err)
								}
							}
							noteStatusMutex.Lock()
							noteStatus[note] = false
							noteStatusMutex.Unlock()
						}(m.Note)
					}
				}
			}

			for _, client := range h.Clients {
				select {
				case client.Send <- msg:
				default:
					client.Close()
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

type IncomingMessage struct {
	Type     string `json:"type"`
	Note     *uint8 `json:"note,omitempty"`
	Velocity *uint8 `json:"velocity,omitempty"`
}

var hub = &Hub{
	Clients:   make(map[*websocket.Conn]*WebSocketClient),
	Broadcast: make(chan interface{}),
	Shutdown:  make(chan struct{}),
}

var noteStatus = make(map[uint8]bool) // track active notes
var noteStatusMutex sync.Mutex

var midiManager *MIDIManager
var upgrader = websocket.Upgrader{}

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

type FullSceneMessage struct {
	Type   string           `json:"type"`
	Cue    string           `json:"cue"`
	Labels map[uint8]string `json:"labels"`
}

var noteEventsThisPeriod int64
var newConnectionsThisPeriod int64

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
		Cue: "😎 Possible Applications",
		Labels: map[uint8]string{
			60: "Music Performance", 62: "Lecture", 64: "Trivia Night", 65: "Workshop", 67: "Art Installation", 69: "Parties", 71: "VJ", 72: "Audience Interaction",
		},
	},
	{
		Cue: "🌐 Powered by Go and Web Technologies",
		Labels: map[uint8]string{
			60: "GoLang", 62: "HTTP", 64: "WebSocket", 65: "MIDI I/O", 67: "Frontend", 69: "Backend", 71: "PortMidi", 72: "Bootstrap",
		},
	},
	{
		Cue: "🎶 Designed for Live Interaction",
		Labels: map[uint8]string{
			60: "Touch", 62: "Buttons", 64: "LED", 65: "Cue", 67: "Velocity", 69: "Scenes", 71: "Next", 72: "Advance",
		},
	},
	{
		Cue: "🚀 Prototype for Future Expansion",
		Labels: map[uint8]string{
			60: "Mobile", 62: "Performance", 64: "DAW", 65: "Ableton", 67: "OSC", 69: "MIDI 2.0", 71: "Cloud", 72: "Realtime",
		},
	},
}

var currentScene int

// --- Scene Broadcasting ---

func broadcastScene() {
	scene := scenes[currentScene%len(scenes)]
	currentScene++

	logServer("Broadcasting scene: %s", scene.Cue)

	// Broadcast full scene message (cue + labels) to all clients
	fullScene := map[string]interface{}{
		"type":   "cue",
		"text":   scene.Cue,
		"labels": scene.Labels,
	}
	for _, client := range hub.Clients {
		select {
		case client.Send <- fullScene:
		default:
			// If send channel is blocked, close and remove client
			client.Close()
		}
	}
	atomic.StoreInt64(&noteEventsThisPeriod, 0)
	atomic.StoreInt64(&newConnectionsThisPeriod, 0)
}

// --- Main Server Setup ---

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal catching
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)

	// Goroutine to catch signals and cancel context
	go func() {
		<-sigchan
		logServer("Shutdown signal received. Closing connections...")
		cancel()
	}()

	midiManager = &MIDIManager{}
	err := midiManager.Setup()
	if err != nil {
		logError("%v", err)
	}

	midiManager.FlushAllNotes()

	// Start Hub broadcaster goroutine
	go hub.Run(ctx)

	// Setup HTTP server and WebSocket handling
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/stats", statsHandler)

	server := &http.Server{Addr: ":8080"}

	// Start listening for incoming MIDI input
	go midiManager.Listen()

	// Periodically broadcast scenes every sceneInterval
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(sceneInterval):
				broadcastScene()
			}
		}
	}()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logError("ListenAndServe error: %v", err)
		}
	}()

	<-ctx.Done()
	logServer("Shutting down HTTP server...")
	server.Shutdown(context.Background())
	midiManager.FlushAllNotes()
	midiManager.Close()
	for _, client := range hub.Clients {
		client.Close()
	}
	logServer("Server shutdown complete.")
}

// --- WebSocket Connection Handling ---

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Accept new WebSocket connections
	upgrader.CheckOrigin = func(r *http.Request) bool { return true } // Allow all origins
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logError("%v", err)
		return
	}
	logWS("New WebSocket connection established.")

	atomic.AddInt64(&newConnectionsThisPeriod, 1)

	client := &WebSocketClient{
		Conn:  ws,
		Send:  make(chan interface{}),
		Timer: time.NewTimer(idleTimeout),
	}
	hub.Register(client)

	go func() {
		<-client.Timer.C
		logTimeout("Idle timeout, closing WebSocket connection.")
		client.Close()
	}()

	// Goroutine to write messages to WebSocket client
	go func() {
		defer func() {
			client.Close()
		}()

		for {
			select {
			case msg, ok := <-client.Send:
				if !ok {
					// Send channel closed, exit writer goroutine
					return
				}
				if err := ws.WriteJSON(msg); err != nil {
					logWS("WebSocket write error: %v", err)
					return
				}
			}
		}
	}()

	// Read incoming WebSocket messages
	for {
		var incoming IncomingMessage
		err := ws.ReadJSON(&incoming)
		if err != nil {
			logWS("WebSocket read error: %v", err)
			client.Close()
			break
		}

		client.Timer.Reset(idleTimeout)

		switch incoming.Type {
		case "nextScene":
			logWS("Received nextScene request from client.")
			broadcastScene()
		case "note":
			if incoming.Note == nil || incoming.Velocity == nil {
				logError("Malformed 'note' message: missing fields")
				continue
			}
			logWS("Parsed Note: %d, Velocity: %d", *incoming.Note, *incoming.Velocity)
			hub.Broadcast <- MIDIMessage{Type: "note", Note: *incoming.Note, Velocity: *incoming.Velocity}
		}
	}

}

// --- Stats HTTP Handler ---
func statsHandler(w http.ResponseWriter, r *http.Request) {
	type Stats struct {
		ConnectedClients     int    `json:"connected_clients"`
		ActiveNotes          int    `json:"active_notes"`
		Cue                  string `json:"cue"`
		NotesPerPeriod       int    `json:"notes_per_period"`
		ConnectionsPerPeriod int    `json:"connections_per_period"`
	}

	noteStatusMutex.Lock()
	activeNotes := 0
	for _, active := range noteStatus {
		if active {
			activeNotes++
		}
	}
	noteStatusMutex.Unlock()

	stats := Stats{
		ConnectedClients:     len(hub.Clients),
		ActiveNotes:          activeNotes,
		Cue:                  scenes[currentScene%len(scenes)].Cue,
		NotesPerPeriod:       int(atomic.LoadInt64(&noteEventsThisPeriod)),
		ConnectionsPerPeriod: int(atomic.LoadInt64(&newConnectionsThisPeriod)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
