package events

import "time"

// ── Kafka Topic Names ─────────────────────────────────────────────────────

const (
	TopicAuditInspections = "audit.inspections"
	TopicAuditIssues      = "audit.issues"
	TopicActivityLogs     = "audit.activity-logs"
)

// ── Event Types ───────────────────────────────────────────────────────────

const (
	EventTypeCreated   = "CREATED"
	EventTypeUpdated   = "UPDATED"
	EventTypeDeleted   = "DELETED"
	EventTypeSubmitted = "SUBMITTED"
	EventTypeResolved  = "RESOLVED"
	EventTypeClosed    = "CLOSED"
	EventTypeConfirmed = "CONFIRMED"
)

// ── Event Payloads (Domain Events) ────────────────────────────────────────

// BaseEvent holds common fields for all events
type BaseEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	ActorID   string    `json:"actor_id"` // User who triggered the event
}

// InspectionEvent represents changes to an inspection record
type InspectionEvent struct {
	BaseEvent
	InspectionID string  `json:"inspection_id"`
	PICUserID    string  `json:"pic_user_id"`
	Status       string  `json:"status"`
	AreaID       string  `json:"area_id"`
	Score        float64 `json:"score,omitempty"`
}

// IssueEvent represents changes to an issue record
type IssueEvent struct {
	BaseEvent
	IssueID        string     `json:"issue_id"`
	ResultID       string     `json:"result_id"`
	IssuePICUserID string     `json:"issue_pic_user_id"`
	Status         string     `json:"status"`
	DueDate        *time.Time `json:"due_date,omitempty"`
}

// ActivityLogEvent represents an explicit audit log trail for user actions
type ActivityLogEvent struct {
	BaseEvent
	Action    string `json:"action"`
	Entity    string `json:"entity"`
	EntityID  string `json:"entity_id"`
	Details   string `json:"details"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}
