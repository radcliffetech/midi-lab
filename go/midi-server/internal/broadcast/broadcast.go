package broadcast

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ClientSender interface {
	SendChannel() chan interface{}
	Close()
}

type Broadcaster interface {
	Broadcast(clients map[*websocket.Conn]ClientSender, message interface{})
}

type DefaultBroadcaster struct{}

func (b *DefaultBroadcaster) Broadcast(clients map[*websocket.Conn]ClientSender, message interface{}) {
	for _, client := range clients {
		select {
		case client.SendChannel() <- message:
		default:
			client.Close()
		}
	}
}

type BufferedBroadcaster struct{}

func (b *BufferedBroadcaster) Broadcast(clients map[*websocket.Conn]ClientSender, message interface{}) {
	for _, client := range clients {
		select {
		case client.SendChannel() <- message:
		default:
			go func(c ClientSender) {
				select {
				case c.SendChannel() <- message:
				case <-time.After(50 * time.Millisecond):
					c.Close()
				}
			}(client)
		}
	}
}

type BatchBroadcaster struct{}

func (b *BatchBroadcaster) Broadcast(clients map[*websocket.Conn]ClientSender, message interface{}) {
	var wg sync.WaitGroup
	for _, client := range clients {
		wg.Add(1)
		go func(c ClientSender) {
			defer wg.Done()
			select {
			case c.SendChannel() <- message:
			default:
				c.Close()
			}
		}(client)
	}
	wg.Wait()
}

type LossyBroadcaster struct{}

func (b *LossyBroadcaster) Broadcast(clients map[*websocket.Conn]ClientSender, message interface{}) {
	for _, client := range clients {
		select {
		case client.SendChannel() <- message:
		default:
			// Skip slow clients but don't close them
		}
	}
}
