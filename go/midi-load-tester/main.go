package main

import (
	"flag"
	"log"
	"math/rand"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// ========================
// Constants and Variables
// ========================

// C major pentatonic scale: C, D, E, G, A (MIDI 60, 62, 64, 67, 69), two octaves
var pentatonicC = []int{60, 62, 64, 67, 69}

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

// ========================
// Command-line Flags
// ========================

var (
	serverAddress   = flag.String("server", "localhost:8080", "WebSocket server address")
	connectionCount = flag.Int("clients", 500, "Number of fake clients")
)

// ========================
// Global Atomic Counters
// ========================

var (
	clientsConnected int64
	notesSent        int64
)

// ========================
// Helper Functions
// ========================

// generateNoteMessage creates and returns a note message map to send over the websocket connection.
// It randomly decides whether to send a chord or a single note, with some bass note probability.
func generateNoteMessage(c *websocket.Conn) map[string]interface{} {
	chordChance := 0.1
	if rand.Float64() < chordChance {
		// Send a chord (multiple notes)
		chordSize := rand.Intn(2) + 2
		for range chordSize {
			baseNote := pentatonicC[rand.Intn(len(pentatonicC))]
			octaveShift := rand.Intn(2) * 12
			note := baseNote + octaveShift
			velocity := rand.Intn(100) + 1
			msg := map[string]interface{}{
				"type":     "note",
				"note":     note,
				"velocity": velocity,
			}
			err := c.WriteJSON(msg)
			if err != nil {
				log.Printf(colorRed+"[ERROR] Client chord write error: %v"+colorReset, err)
			}
			atomic.AddInt64(&notesSent, 1)
			time.Sleep(30 * time.Millisecond)
		}
		return nil
	}

	// Send a single note, possibly a bass note
	bassChance := 0.1
	var note int
	if rand.Float64() < bassChance {
		bassNotes := []int{60, 67}
		baseNote := bassNotes[rand.Intn(len(bassNotes))]
		note = baseNote - 12
	} else {
		baseNote := pentatonicC[rand.Intn(len(pentatonicC))]
		octaveShift := rand.Intn(2) * 12
		note = baseNote + octaveShift
	}
	velocity := rand.Intn(100) + 1
	return map[string]interface{}{
		"type":     "note",
		"note":     note,
		"velocity": velocity,
	}
}

// ========================
// Client Simulator
// ========================

// simulateClient establishes a websocket connection and simulates sending note and cue messages.
// It handles connection retries and logs connection status and errors.
func simulateClient(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	u := url.URL{Scheme: "ws", Host: *serverAddress, Path: "/ws"}

	var c *websocket.Conn
	var err error
	for retries := 0; retries < 3; retries++ {
		c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err == nil {
			break
		}
		log.Printf(colorYellow+"[CLIENT %d] Connection failed: %v (retrying...)"+colorReset, id, err)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Printf(colorRed+"[ERROR] Client %d connection failed permanently: %v"+colorReset, id, err)
		return
	}
	defer c.Close()
	atomic.AddInt64(&clientsConnected, 1)

	log.Printf(colorCyan+"[CLIENT %d] Connected"+colorReset, id)

	for {
		burstMode := rand.Float64() < 0.7
		if burstMode {
			time.Sleep(time.Duration(rand.Intn(200)+100) * time.Millisecond)
		} else {
			time.Sleep(time.Duration(rand.Intn(4000)+1000) * time.Millisecond)
		}

		var msg map[string]interface{}
		cueChance := 0.1
		if rand.Float64() < cueChance {
			// Send a cue message
			msg = map[string]interface{}{
				"type": "cue",
				"text": "Scene Change",
			}
		} else {
			// Generate and send a note message
			msg = generateNoteMessage(c)
			if msg == nil {
				continue
			}
			atomic.AddInt64(&notesSent, 1)
		}

		err := c.WriteJSON(msg)
		if err != nil {
			log.Printf(colorRed+"[ERROR] Client %d: %v"+colorReset, id, err)
			return
		}
	}
}

// ========================
// Main Entry Point
// ========================

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup

	// Periodically log stats about clients and notes sent
	go func() {
		for {
			log.Printf(colorGreen+"[STATS] Clients: %d, Notes Sent: %d"+colorReset, atomic.LoadInt64(&clientsConnected), atomic.LoadInt64(&notesSent))
			time.Sleep(5 * time.Second)
		}
	}()

	// Launch simulated clients with a small pause to reduce connection storms
	for i := 0; i < *connectionCount; i++ {
		wg.Add(1)
		go simulateClient(i, &wg)
		time.Sleep(5 * time.Millisecond)
	}

	wg.Wait()

	log.Println(colorGreen + "[TEST] ✅ Load test complete." + colorReset)
	log.Printf(colorGreen+"[STATS] Final Stats - Clients Connected: %d, Notes Sent: %d"+colorReset, atomic.LoadInt64(&clientsConnected), atomic.LoadInt64(&notesSent))
	time.Sleep(1 * time.Second)
}
