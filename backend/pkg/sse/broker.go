package sse

import (
	"log"
	"sync"
)

// Broker handles all active SSE client connections and broadcasts messages to them.
type Broker struct {
	clients        map[chan string]bool
	newClients     chan chan string
	defunctClients chan chan string
	messages       chan string
	mu             sync.Mutex
}

// NewBroker creates a new SSE Broker.
func NewBroker() *Broker {
	return &Broker{
		clients:        make(map[chan string]bool),
		newClients:     make(chan chan string),
		defunctClients: make(chan chan string),
		messages:       make(chan string),
	}
}

// Start begins the listening loop for the broker. It handles adding and removing clients,
// as well as broadcasting messages.
func (b *Broker) Start() {
	for {
		select {
		case s := <-b.newClients:
			b.mu.Lock()
			b.clients[s] = true
			b.mu.Unlock()
			log.Printf("SSE Client connected. Total clients: %d", len(b.clients))

		case s := <-b.defunctClients:
			b.mu.Lock()
			if _, ok := b.clients[s]; ok {
				delete(b.clients, s)
				close(s)
				log.Printf("SSE Client disconnected. Total clients: %d", len(b.clients))
			}
			b.mu.Unlock()

		case msg := <-b.messages:
			b.mu.Lock()
			for s := range b.clients {
				// Use a non-blocking send to prevent hanging on a slow client
				select {
				case s <- msg:
				default:
					// If the channel is full/blocked, we consider the client defunct
					// Cannot write to defunctClients inside the mutex easily, so we just close and delete
					delete(b.clients, s)
					close(s)
					log.Printf("SSE Client slow/unresponsive. Disconnected.")
				}
			}
			b.mu.Unlock()
		}
	}
}

// AddClient returns a channel that will receive broadcasted messages.
func (b *Broker) AddClient() chan string {
	clientChan := make(chan string, 10)
	b.newClients <- clientChan
	return clientChan
}

// RemoveClient removes a client channel.
func (b *Broker) RemoveClient(clientChan chan string) {
	b.defunctClients <- clientChan
}

// Broadcast sends a message to all connected clients.
func (b *Broker) Broadcast(msg string) {
	b.messages <- msg
}
