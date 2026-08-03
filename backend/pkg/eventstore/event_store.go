package eventstore

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// Event structs
type GlobalEvent struct {
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
}

type UserEvent struct {
	Type      string      `json:"type"`
	AspekID   string      `json:"aspek_id,omitempty"`
	Message   string      `json:"message,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

type KawasanEvent struct {
	Type      string      `json:"type"`
	KawasanID string      `json:"kawasan_id"`
	AspekID   string      `json:"aspek_id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
	UserName  string      `json:"user_name,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

type EventStore struct {
	rdb        *redis.Client
	mu         sync.RWMutex
	memGlobal  []GlobalEvent
	memUser    map[string][]UserEvent
	memKawasan map[string][]KawasanEvent
}

var (
	instance *EventStore
	once     sync.Once
)

func GetEventStore(rdb *redis.Client) *EventStore {
	once.Do(func() {
		instance = &EventStore{
			rdb:        rdb,
			memGlobal:  make([]GlobalEvent, 0),
			memUser:    make(map[string][]UserEvent),
			memKawasan: make(map[string][]KawasanEvent),
		}
	})
	if rdb != nil && instance.rdb == nil {
		instance.rdb = rdb
	}
	return instance
}

// Global Events (Issues, Dashboard, Inspections)
func (s *EventStore) PushGlobal(eventType string) {
	now := time.Now().UnixMilli()
	event := GlobalEvent{
		Type:      eventType,
		Timestamp: now,
	}

	if s.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		data, err := json.Marshal(event)
		if err == nil {
			key := "events:global"
			s.rdb.ZAdd(ctx, key, redis.Z{
				Score:  float64(now),
				Member: string(data),
			})
			// Clean up events older than 5 minutes (300,000 ms)
			s.rdb.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(now-300000, 10))
			s.rdb.Expire(ctx, key, 5*time.Minute)
			return
		}
	}

	// Fallback to In-Memory
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memGlobal = append(s.memGlobal, event)
	s.cleanMemoryGlobal(now)
}

func (s *EventStore) GetGlobalSince(sinceMs int64) []GlobalEvent {
	now := time.Now().UnixMilli()
	if sinceMs <= 0 {
		sinceMs = now - 5000 // default last 5 seconds if since not provided
	}

	if s.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		key := "events:global"
		vals, err := s.rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{
			Min: "(" + formatScore(sinceMs),
			Max: "+inf",
		}).Result()

		if err == nil {
			events := make([]GlobalEvent, 0, len(vals))
			for _, val := range vals {
				var ev GlobalEvent
				if err := json.Unmarshal([]byte(val), &ev); err == nil {
					events = append(events, ev)
				}
			}
			return events
		}
		log.Printf("[EventStore] Redis GetGlobalSince error: %v, falling back to memory", err)
	}

	// Fallback to In-Memory
	s.mu.RLock()
	defer s.mu.RUnlock()
	events := make([]GlobalEvent, 0)
	for _, ev := range s.memGlobal {
		if ev.Timestamp > sinceMs {
			events = append(events, ev)
		}
	}
	return events
}

// User Notifications (Inspeksi)
func (s *EventStore) PushUser(userID string, event UserEvent) {
	now := time.Now().UnixMilli()
	event.Timestamp = now

	if s.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		data, err := json.Marshal(event)
		if err == nil {
			key := "events:user:" + userID
			s.rdb.ZAdd(ctx, key, redis.Z{
				Score:  float64(now),
				Member: string(data),
			})
			s.rdb.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(now-300000, 10))
			s.rdb.Expire(ctx, key, 5*time.Minute)
			return
		}
	}

	// In-Memory Fallback
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memUser[userID] = append(s.memUser[userID], event)
	s.cleanMemoryUser(userID, now)
}

func (s *EventStore) GetUserSince(userID string, sinceMs int64) []UserEvent {
	now := time.Now().UnixMilli()
	if sinceMs <= 0 {
		sinceMs = now - 5000
	}

	if s.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		key := "events:user:" + userID
		vals, err := s.rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{
			Min: "(" + formatScore(sinceMs),
			Max: "+inf",
		}).Result()

		if err == nil {
			events := make([]UserEvent, 0, len(vals))
			for _, val := range vals {
				var ev UserEvent
				if err := json.Unmarshal([]byte(val), &ev); err == nil {
					events = append(events, ev)
				}
			}
			return events
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	events := make([]UserEvent, 0)
	for _, ev := range s.memUser[userID] {
		if ev.Timestamp > sinceMs {
			events = append(events, ev)
		}
	}
	return events
}

// Kawasan Events (WebSocket replacement for Kawasan Lock & Sync)
func (s *EventStore) PushKawasan(kawasanID string, event KawasanEvent) {
	now := time.Now().UnixMilli()
	event.Timestamp = now

	if s.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		data, err := json.Marshal(event)
		if err == nil {
			key := "events:kawasan:" + kawasanID
			s.rdb.ZAdd(ctx, key, redis.Z{
				Score:  float64(now),
				Member: string(data),
			})
			s.rdb.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(now-300000, 10))
			s.rdb.Expire(ctx, key, 5*time.Minute)
			return
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.memKawasan[kawasanID] = append(s.memKawasan[kawasanID], event)
	s.cleanMemoryKawasan(kawasanID, now)
}

func (s *EventStore) GetKawasanSince(kawasanID string, sinceMs int64) []KawasanEvent {
	now := time.Now().UnixMilli()
	if sinceMs <= 0 {
		sinceMs = now - 5000
	}

	if s.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		key := "events:kawasan:" + kawasanID
		vals, err := s.rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{
			Min: "(" + formatScore(sinceMs),
			Max: "+inf",
		}).Result()

		if err == nil {
			events := make([]KawasanEvent, 0, len(vals))
			for _, val := range vals {
				var ev KawasanEvent
				if err := json.Unmarshal([]byte(val), &ev); err == nil {
					events = append(events, ev)
				}
			}
			return events
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	events := make([]KawasanEvent, 0)
	for _, ev := range s.memKawasan[kawasanID] {
		if ev.Timestamp > sinceMs {
			events = append(events, ev)
		}
	}
	return events
}

// Memory Cleanup Helpers
func (s *EventStore) cleanMemoryGlobal(now int64) {
	cutoff := now - 300000 // 5 minutes
	n := 0
	for _, ev := range s.memGlobal {
		if ev.Timestamp >= cutoff {
			s.memGlobal[n] = ev
			n++
		}
	}
	s.memGlobal = s.memGlobal[:n]
}

func (s *EventStore) cleanMemoryUser(userID string, now int64) {
	cutoff := now - 300000
	events := s.memUser[userID]
	n := 0
	for _, ev := range events {
		if ev.Timestamp >= cutoff {
			events[n] = ev
			n++
		}
	}
	s.memUser[userID] = events[:n]
}

func (s *EventStore) cleanMemoryKawasan(kawasanID string, now int64) {
	cutoff := now - 300000
	events := s.memKawasan[kawasanID]
	n := 0
	for _, ev := range events {
		if ev.Timestamp >= cutoff {
			events[n] = ev
			n++
		}
	}
	s.memKawasan[kawasanID] = events[:n]
}

func formatScore(ms int64) string {
	return strconv.FormatInt(ms, 10)
}
