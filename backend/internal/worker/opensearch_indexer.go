package worker

import (
	"context"
	"encoding/json"

	"github.com/monitoring-system/backend/internal/domain/events"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/opensearch"
	"github.com/segmentio/kafka-go"
)

// OpenSearchIndexer consumes Kafka events and indexes them into OpenSearch
type OpenSearchIndexer struct {
	osClient *opensearch.Client
	log      *logger.Logger
}

func NewOpenSearchIndexer(osClient *opensearch.Client, log *logger.Logger) *OpenSearchIndexer {
	return &OpenSearchIndexer{
		osClient: osClient,
		log:      log,
	}
}

// HandleInspectionEvent processes events from audit.inspections
func (w *OpenSearchIndexer) HandleInspectionEvent(ctx context.Context, msg kafka.Message) error {
	var event events.InspectionEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err // malformed message
	}

	w.log.Info("Indexing inspection event", logger.String("eventID", event.EventID))
	return w.osClient.IndexDocument(ctx, "audit-inspections", event.EventID, event)
}

// HandleIssueEvent processes events from audit.issues
func (w *OpenSearchIndexer) HandleIssueEvent(ctx context.Context, msg kafka.Message) error {
	var event events.IssueEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}

	w.log.Info("Indexing issue event", logger.String("eventID", event.EventID))
	return w.osClient.IndexDocument(ctx, "audit-issues", event.EventID, event)
}

// HandleActivityLogEvent processes events from audit.activity-logs
func (w *OpenSearchIndexer) HandleActivityLogEvent(ctx context.Context, msg kafka.Message) error {
	var event events.ActivityLogEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}

	w.log.Info("Indexing activity log event", logger.String("eventID", event.EventID))
	return w.osClient.IndexDocument(ctx, "audit-activity-logs", event.EventID, event)
}
