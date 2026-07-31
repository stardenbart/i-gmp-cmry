package realtime

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type WSEvent struct {
	Type        string      `json:"type"` // LOCK_ACQUIRED | LOCK_RELEASED | LOCK_EXPIRED | ASPEK_COMPLETED | KAWASAN_SYNCED
	KawasanID   string      `json:"kawasan_id"`
	AspekID     string      `json:"aspek_id,omitempty"`
	UserID      string      `json:"user_id,omitempty"`
	UserName    string      `json:"user_name,omitempty"`
	Payload     interface{} `json:"payload,omitempty"`
}

type WSHub struct {
	mu          sync.RWMutex
	connections map[string]map[*websocket.Conn]bool // kawasanID -> set of WebSocket connections
}

var globalWSHub *WSHub
var wsOnce sync.Once

func GetWSHub() *WSHub {
	wsOnce.Do(func() {
		globalWSHub = &WSHub{
			connections: make(map[string]map[*websocket.Conn]bool),
		}
	})
	return globalWSHub
}

func (h *WSHub) Register(kawasanID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.connections[kawasanID]; !ok {
		h.connections[kawasanID] = make(map[*websocket.Conn]bool)
	}
	h.connections[kawasanID][conn] = true
	log.Printf("🔌 [WebSocket] Client registered for Kawasan %s (Total: %d)", kawasanID, len(h.connections[kawasanID]))
}

func (h *WSHub) Unregister(kawasanID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.connections[kawasanID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.connections, kawasanID)
		}
	}
	conn.Close()
	log.Printf("🔌 [WebSocket] Client unregistered for Kawasan %s", kawasanID)
}

func (h *WSHub) BroadcastToKawasan(kawasanID string, event WSEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.connections[kawasanID]
	if !ok || len(conns) == 0 {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("❌ [WebSocket] Failed to marshal event: %v", err)
		return
	}

	for conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("⚠️ [WebSocket] Failed to send to conn: %v", err)
			conn.Close()
		}
	}
}
