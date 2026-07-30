package kafka

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/segmentio/kafka-go"
)

// MessageHandler is the signature for functions that process a Kafka message.
type MessageHandler func(ctx context.Context, msg kafka.Message) error

// EventConsumer represents a Kafka consumer group reader.
type EventConsumer struct {
	reader *kafka.Reader
	log    *logger.Logger
}

// NewConsumer creates a new Kafka consumer attached to a specific topic and consumer group.
func NewConsumer(brokers []string, groupID, topic string, log *logger.Logger) *EventConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &EventConsumer{
		reader: r,
		log:    log,
	}
}

// Start begins listening to the Kafka topic and calls the handler for each message.
// It runs continuously until the context is canceled.
func (c *EventConsumer) Start(ctx context.Context, handler MessageHandler) {
	c.log.Info("Starting Kafka consumer", logger.String("topic", c.reader.Config().Topic))

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.log.Info("Stopping Kafka consumer", logger.String("topic", c.reader.Config().Topic))
				if err := c.reader.Close(); err != nil {
					c.log.Error("Failed to close consumer", logger.Error(err))
				}
				return
			default:
				msg, err := c.reader.FetchMessage(ctx)
				if err != nil {
					// Don't log if it's just a context cancellation
					if ctx.Err() == nil {
						if errors.Is(err, io.EOF) {
							c.log.Warn("Kafka consumer connection EOF, retrying in 2 seconds...", logger.String("topic", c.reader.Config().Topic))
						} else {
							c.log.Error("Error fetching message", logger.Error(err))
						}
						time.Sleep(2 * time.Second)
					}
					continue
				}

				// Process message
				if err := handler(ctx, msg); err != nil {
					c.log.Error("Failed to process message (DLQ should handle this in production)",
						logger.Error(err),
						logger.String("topic", msg.Topic),
						logger.String("key", string(msg.Key)),
					)
					// Depending on strictness, we might skip committing offset if it fails completely.
					// For now, we log the error and commit to keep moving forward.
				}

				// Commit offset after successful (or skipped) processing
				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					c.log.Error("Failed to commit message offset", logger.Error(err))
				}
			}
		}
	}()
}
