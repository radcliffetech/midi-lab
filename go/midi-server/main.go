package main

import (
	"context"
	"encoding/json"
	"flag"
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

	"github.com/radcliffetech/midi-lab/go/midi-server/internal/broadcast"
)

// --------------------
// Constants
// --------------------

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

// --------------------
// Logging Helpers
// --------------------

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

// --------------------
// MIDI Manager
// --------------------

type MIDIManager struct {
	driverCloser io.Closer
	writer       *writer.Writer
	out          midi.Out
	in           midi.In
}

func (m *MIDIManager) Setup() error {
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
		writer.ControlChange(m.writer, 123, 0) // CC#123 All Notes Off
	}
}

// --------------------
// Globals
// --------------------

var (
	hub                      *Hub
	noteStatus               = make(map[uint8]bool) // track active notes
	noteStatusMutex          sync.Mutex
	midiManager              *MIDIManager
	upgrader                 = websocket.Upgrader{}
	noteEventsThisPeriod     int64
	newConnectionsThisPeriod int64
	scenes                   []Scene
	currentScene             int
)

// --------------------
// Types
// --------------------

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

func (c *WebSocketClient) SendChannel() chan interface{} {
	return c.Send
}

type Hub struct {
	Clients     map[*websocket.Conn]*WebSocketClient
	Broadcast   chan interface{}
	Shutdown    chan struct{}
	Broadcaster broadcast.Broadcaster
}

type IncomingMessage struct {
	Type     string `json:"type"`
	Note     *uint8 `json:"note,omitempty"`
	Velocity *uint8 `json:"velocity,omitempty"`
}

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
	Type        string           `json:"type"`
	Cue         string           `json:"cue"`
	Labels      map[uint8]string `json:"labels"`
	NormalColor string           `json:"normalColor"`
	PressColor  string           `json:"pressColor"`
}

type Scene struct {
	Cue         string
	Labels      map[uint8]string
	NormalColor string
	PressColor  string
}

// --------------------
// Scene Handling
// --------------------

func loadScenesFromFile(path string) ([]Scene, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var loadedScenes []Scene
	err = json.NewDecoder(f).Decode(&loadedScenes)
	if err != nil {
		return nil, err
	}

	return loadedScenes, nil
}

func broadcastScene() {
	if len(scenes) == 0 {
		logServer("No scenes to broadcast")
		return
	}

	scene := scenes[currentScene%len(scenes)]
	currentScene++

	logServer("Broadcasting scene: %s", scene.Cue)

	fullScene := map[string]interface{}{
		"type":        "cue",
		"text":        scene.Cue,
		"labels":      scene.Labels,
		"normalColor": scene.NormalColor,
		"pressColor":  scene.PressColor,
	}

	for _, client := range hub.Clients {
		select {
		case client.Send <- fullScene:
		default:
			client.Close()
		}
	}

	atomic.StoreInt64(&noteEventsThisPeriod, 0)
	atomic.StoreInt64(&newConnectionsThisPeriod, 0)
}

// --------------------
// HTTP Handlers
// --------------------

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

	cue := ""
	if len(scenes) > 0 {
		cue = scenes[currentScene%len(scenes)].Cue
	}

	stats := Stats{
		ConnectedClients:     len(hub.Clients),
		ActiveNotes:          activeNotes,
		Cue:                  cue,
		NotesPerPeriod:       int(atomic.LoadInt64(&noteEventsThisPeriod)),
		ConnectionsPerPeriod: int(atomic.LoadInt64(&newConnectionsThisPeriod)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func reloadScenesHandler(w http.ResponseWriter, r *http.Request) {
	sc, err := loadScenesFromFile("scenes.json")
	if err != nil {
		http.Error(w, "Failed to reload scenes", http.StatusInternalServerError)
		logError("Failed to reload scenes: %v", err)
		return
	}
	scenes = sc
	logServer("Reloaded scenes from scenes.json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Scenes reloaded successfully"))
}

// --------------------
// WebSocket Handling
// --------------------

func handleConnections(w http.ResponseWriter, r *http.Request) {
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

	go func() {
		defer client.Close()
		for {
			select {
			case msg, ok := <-client.Send:
				if !ok {
					return
				}
				if err := ws.WriteJSON(msg); err != nil {
					logWS("WebSocket write error: %v", err)
					return
				}
			}
		}
	}()

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

// --------------------
// Hub Methods
// --------------------

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
							time.Sleep(500 * time.Millisecond)
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

			clients := make(map[*websocket.Conn]broadcast.ClientSender)
			for conn, client := range h.Clients {
				clients[conn] = client
			}
			h.Broadcaster.Broadcast(clients, msg)

		case <-ctx.Done():
			return
		}
	}
}

// --------------------
// Main
// --------------------

func main() {
	// Available broadcast modes:
	// - default  => Immediate send, close slow clients
	// - buffered => Allow 50ms grace for slow clients before closing (recommended)
	// - batch    => Parallel sending with goroutines
	// - lossy    => Skip slow clients without closing them
	var broadcastMode = flag.String("broadcast-mode", "", "Broadcast mode: default, buffered, batch, lossy")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)

	go func() {
		<-sigchan
		logServer("Shutdown signal received. Closing connections...")
		cancel()
	}()

	var err error
	scenes, err = loadScenesFromFile("scenes.json")
	if err != nil {
		log.Fatalf("Failed to load scenes: %v", err)
	}
	logServer("Loaded %d scenes", len(scenes))

	hub = &Hub{
		Clients:   make(map[*websocket.Conn]*WebSocketClient),
		Broadcast: make(chan interface{}),
		Shutdown:  make(chan struct{}),
	}

	switch *broadcastMode {
	case "", "buffered":
		hub.Broadcaster = &broadcast.BufferedBroadcaster{}
	case "default":
		hub.Broadcaster = &broadcast.DefaultBroadcaster{}
	case "batch":
		hub.Broadcaster = &broadcast.BatchBroadcaster{}
	case "lossy":
		hub.Broadcaster = &broadcast.LossyBroadcaster{}
	default:
		log.Fatalf("Unknown broadcast mode: %s", *broadcastMode)
	}

	midiManager = &MIDIManager{}
	err = midiManager.Setup()
	if err != nil {
		logError("%v", err)
	}

	midiManager.FlushAllNotes()

	go hub.Run(ctx)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/reload-scenes", reloadScenesHandler)

	server := &http.Server{Addr: ":8080"}

	go midiManager.Listen()

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
