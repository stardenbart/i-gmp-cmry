package ssehandler

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/pkg/sse"
	"github.com/valyala/fasthttp"
)

type SSEHandler struct {
	broker *sse.Broker
}

func NewSSEHandler(broker *sse.Broker) *SSEHandler {
	return &SSEHandler{
		broker: broker,
	}
}

// @Summary Subscribe to Server-Sent Events
// @Description Subscribe to real-time events for dashboard updates
// @Tags sse
// @Produce text/event-stream
// @Success 200 {string} string "Event stream"
// @Router /api/v1/sse/events [get]
// @Security BearerAuth
func (h *SSEHandler) StreamEvents(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	// Allow CORS for SSE (though handled by global middleware, sometimes helpful to be explicit)
	c.Set("Access-Control-Allow-Origin", "*")

	clientChan := h.broker.AddClient()

	// Notify when client connection is closed
	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		defer h.broker.RemoveClient(clientChan)

		// Send initial connection success event
		fmt.Fprintf(w, "event: connected\ndata: connected\n\n")
		w.Flush()

		// Keep connection alive with periodic pings or wait for messages
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case msg, ok := <-clientChan:
				if !ok {
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", msg)
				if err := w.Flush(); err != nil {
					return // Client disconnected
				}
			case <-ticker.C:
				// Send ping to keep connection alive
				fmt.Fprintf(w, "event: ping\ndata: ping\n\n")
				if err := w.Flush(); err != nil {
					return // Client disconnected
				}
			}
		}
	}))

	return nil
}
