package broadcaster

import (
	"encoding/json"
	"log"
	"sync"
)

type Broadcaster struct {
	mu      sync.RWMutex
	clients map[chan string]struct{}
}

func New() *Broadcaster {
	return &Broadcaster{
		clients: make(map[chan string]struct{}),
	}
}

func (b *Broadcaster) Subscribe() chan string {
	ch := make(chan string, 8)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *Broadcaster) Unsubscribe(ch chan string) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
	close(ch)
}

type EventType string

const (
	OrderCreated EventType = "orderCreated"
)

type Event struct {
	Type EventType `json:"type"`
	Data any       `json:"data"`
}

func NewEvent(eventType EventType, data any) Event {
	return Event{
		Type: eventType,
		Data: data,
	}
}

func (b *Broadcaster) Publish(e Event) {
	data, err := json.Marshal(e)
	if err != nil {
		log.Printf("broadcaster: failed to marshal event: %v", err)
		return
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- string(data):
		default:
		}
	}
}
