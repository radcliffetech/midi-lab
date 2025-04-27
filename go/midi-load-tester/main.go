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

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

var (
	serverAddress   = flag.String("server", "localhost:8080", "WebSocket server address")
	connectionCount = flag.Int("clients", 500, "Number of fake clients")
	maxNotes        = flag.Int64("max-notes", 0, "Optional maximum number of notes to send before stopping (0 = unlimited)")
)

var (
	clientsConnected int64
	notesSent        int64
)

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
		if id%10 == 0 {
			cueChance = 0.3 // scene masters send more cues
		}
		if rand.Float64() < cueChance {
			msg = map[string]interface{}{
				"type": "cue",
				"text": "Scene Change",
			}
		} else {
			note := rand.Intn(20) + 60
			velocity := rand.Intn(100) + 1
			msg = map[string]interface{}{
				"type":     "note",
				"note":     note,
				"velocity": velocity,
			}
			atomic.AddInt64(&notesSent, 1)

			if *maxNotes > 0 && atomic.LoadInt64(&notesSent) >= *maxNotes {
				log.Printf(colorYellow+"[CLIENT %d] Max notes limit reached (%d). Closing connection."+colorReset, id, *maxNotes)
				_ = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Max notes reached"))
				time.Sleep(1 * time.Second)
				return
			}
		}

		err := c.WriteJSON(msg)
		if err != nil {
			log.Printf(colorRed+"[ERROR] Client %d: %v"+colorReset, id, err)
			return
		}
	}
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup

	go func() {
		for {
			log.Printf(colorGreen+"[STATS] Clients: %d, Notes Sent: %d"+colorReset, atomic.LoadInt64(&clientsConnected), atomic.LoadInt64(&notesSent))
			time.Sleep(5 * time.Second)
		}
	}()

	for i := 0; i < *connectionCount; i++ {
		wg.Add(1)
		go simulateClient(i, &wg)

		// Optional: small pause to avoid connection storms
		time.Sleep(5 * time.Millisecond)
	}

	wg.Wait()

	log.Println(colorGreen + "[TEST] ✅ Load test complete." + colorReset)
	log.Printf(colorGreen+"[STATS] Final Stats - Clients Connected: %d, Notes Sent: %d"+colorReset, atomic.LoadInt64(&clientsConnected), atomic.LoadInt64(&notesSent))
	time.Sleep(1 * time.Second)
}
