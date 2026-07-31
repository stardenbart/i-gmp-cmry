package realtime

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

type SSEEvent struct {
	Type    string      `json:"type"` // SAVE_SUCCESS | SAVE_FAILED | LOCK_EXPIRED | KAWASAN_DONE | YIELD_REQUEST
	AspekID string      `json:"aspek_id,omitempty"`
	Message string      `json:"message,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

type SSEClient struct {
	UserID  string
	Channel chan SSEEvent
}

type SSEBroker struct {
	mu      sync.RWMutex
	clients map[string]map[chan SSEEvent]bool // userID -> set of channels
}

var globalSSEBroker *SSEBroker
var sseOnce sync.Once

func GetSSEBroker() *SSEBroker {
	sseOnce.Do(func() {
		globalSSEBroker = &SSEBroker{
			clients: make(map[string]map[chan SSEEvent]bool),
		}
	})
	return globalSSEBroker
}

func (b *SSEBroker) Register(userID string) chan SSEEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan SSEEvent, 10)
	if _, ok := b.clients[userID]; !ok {
		b.clients[userID] = make(map[chan SSEEvent]bool)
	}
	b.clients[userID][ch] = true
	return ch
}

func (b *SSEBroker) Unregister(userID string, ch chan SSEEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if userClients, ok := b.clients[userID]; ok {
		delete(userClients, ch)
		close(ch)
		if len(userClients) == 0 {
			delete(b.clients, userID)
		}
	}
}

func (b *SSEBroker) SendToUser(userID string, event SSEEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	userClients, ok := b.clients[userID]
	if !ok || len(userClients) == 0 {
		return
	}

	for ch := range userClients {
		select {
		case ch <- event:
		default:
			// Non-blocking send, drop if buffer full
		}
	}
}

func FormatSSEData(event SSEEvent) (string, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("event: %s\ndata: %s\n\n", event.Type, string(data)), nil
}

func StreamToClient(w io.Writer, ch chan SSEEvent, stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			msg, err := FormatSSEData(event)
			if err == nil {
				w.Write([]byte(msg))
			}
		}
	}
}
