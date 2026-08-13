package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// EventProducer defines the interface to publish events to Kafka.
type EventProducer interface {
	PublishEvent(ctx context.Context, topic string, key string, event interface{}) error
	Close() error
}

type producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer.
func NewProducer(brokers []string) EventProducer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Balancer:               &kafka.Hash{},
		RequiredAcks:           kafka.RequireOne,  // Leader-only ack: ~10x faster than RequireAll
		MaxAttempts:            2,
		AllowAutoTopicCreation: true,
		BatchTimeout:           5 * time.Millisecond,
		WriteTimeout:           2 * time.Second,   // Batas atas latency per write
		ReadTimeout:            2 * time.Second,
	}

	return &producer{writer: w}
}

func (p *producer) PublishEvent(ctx context.Context, topic string, key string, event interface{}) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", topic, err)
	}
	return nil
}

func (p *producer) Close() error {
	return p.writer.Close()
}
