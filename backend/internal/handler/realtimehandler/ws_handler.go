package realtimehandler

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
	redis "github.com/redis/go-redis/v9"
)

type RealtimeMessage struct {
	Type      string      `json:"type"`
	KawasanID string      `json:"kawasan_id,omitempty"`
	AspekID   string      `json:"aspek_id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

type ClientConnection struct {
	Conn      *websocket.Conn
	UserID    string
	KawasanID string
	SendChan  chan []byte
}

type Hub struct {
	redisClient *redis.Client
	jwtManager  *jwt.Manager
	logger      *logger.Logger

	mu          sync.RWMutex
	clients     map[*ClientConnection]bool
	kawasanSubs map[string]map[*ClientConnection]bool
}

func NewHub(redisClient *redis.Client, jwtManager *jwt.Manager, log *logger.Logger) *Hub {
	h := &Hub{
		redisClient: redisClient,
		jwtManager:  jwtManager,
		logger:      log,
		clients:     make(map[*ClientConnection]bool),
		kawasanSubs: make(map[string]map[*ClientConnection]bool),
	}

	// Start global Redis Pub/Sub listener for broadcast
	if redisClient != nil {
		go h.listenRedisPubSub()
	}

	return h
}

func (h *Hub) listenRedisPubSub() {
	ctx := context.Background()
	pubsub := h.redisClient.PSubscribe(ctx, "channel:kawasan:*", "channel:global")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		// Extract channel name
		parts := strings.Split(msg.Channel, ":")
		kawasanID := ""
		if len(parts) >= 3 {
			kawasanID = parts[2]
		}

		h.BroadcastToKawasan(kawasanID, []byte(msg.Payload))
	}
}

func (h *Hub) Register(c *ClientConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[c] = true
	if c.KawasanID != "" {
		if _, exists := h.kawasanSubs[c.KawasanID]; !exists {
			h.kawasanSubs[c.KawasanID] = make(map[*ClientConnection]bool)
		}
		h.kawasanSubs[c.KawasanID][c] = true
	}
}

func (h *Hub) Unregister(c *ClientConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[c]; exists {
		delete(h.clients, c)
		close(c.SendChan)
	}

	if c.KawasanID != "" {
		if subs, exists := h.kawasanSubs[c.KawasanID]; exists {
			delete(subs, c)
			if len(subs) == 0 {
				delete(h.kawasanSubs, c.KawasanID)
			}
		}
	}
}

func (h *Hub) BroadcastToKawasan(kawasanID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if kawasanID == "" || kawasanID == "global" {
		for c := range h.clients {
			select {
			case c.SendChan <- message:
			default:
			}
		}
		return
	}

	if subs, exists := h.kawasanSubs[kawasanID]; exists {
		for c := range subs {
			select {
			case c.SendChan <- message:
			default:
			}
		}
	}
}

func (h *Hub) PublishEvent(ctx context.Context, kawasanID string, msg RealtimeMessage) {
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMilli()
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	if h.redisClient != nil {
		channel := "channel:global"
		if kawasanID != "" {
			channel = "channel:kawasan:" + kawasanID
		}
		h.redisClient.Publish(ctx, channel, string(bytes))
	} else {
		h.BroadcastToKawasan(kawasanID, bytes)
	}
}

// UpgradeHandler ensures connection is upgraded to WebSocket
func (h *Hub) UpgradeHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// WSHandler manages the live WebSocket connection loop
func (h *Hub) WSHandler() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		tokenStr := c.Query("token")
		kawasanID := c.Query("kawasan_id")

		userID := "anonymous"
		if tokenStr != "" && h.jwtManager != nil {
			if claims, err := h.jwtManager.Parse(tokenStr); err == nil {
				userID = claims.UserID
			}
		}

		client := &ClientConnection{
			Conn:      c,
			UserID:    userID,
			KawasanID: kawasanID,
			SendChan:  make(chan []byte, 256),
		}

		h.Register(client)
		defer h.Unregister(client)

		// Write Pump Goroutine
		go func() {
			for msg := range client.SendChan {
				if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
					break
				}
			}
		}()

		// Read Pump Loop
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				break
			}
			// Respond to ping
			if string(msg) == "ping" {
				pong, _ := json.Marshal(map[string]interface{}{
					"type":      "pong",
					"timestamp": time.Now().UnixMilli(),
				})
				client.SendChan <- pong
			}
		}
	})
}
