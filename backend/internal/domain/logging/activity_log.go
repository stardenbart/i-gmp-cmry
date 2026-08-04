package logging

import (
	"context"
	"time"
)

// ActivityLog represents the Activity_Log table.
// Full audit trail for all CRUD actions across the system.
type ActivityLog struct {
	ActivityLogID       string    `gorm:"column:ActivityLogID;primaryKey" json:"activity_log_id"`
	UserID              string    `gorm:"column:UserID;not null" json:"user_id"`
	ModuleID            *string   `gorm:"column:ModuleID" json:"module_id,omitempty"`
	PermissionID        *string   `gorm:"column:PermissionID" json:"permission_id,omitempty"`
	ActivityAction      string    `gorm:"column:ActivityAction;size:50;not null" json:"activity_action"`
	TableAffected       string    `gorm:"column:TableAffected;size:100" json:"table_affected"`
	RecordID            string    `gorm:"column:RecordID;size:50" json:"record_id"`
	OldValue            string    `gorm:"column:OldValue;type:text" json:"old_value,omitempty"`
	NewValue            string    `gorm:"column:NewValue;type:text" json:"new_value,omitempty"`
	ActivityDescription string    `gorm:"column:ActivityDescription;size:255" json:"activity_description"`
	IPAddress           string    `gorm:"column:IPAddress;size:50" json:"ip_address"`
	ActivityCreatedAt   time.Time `gorm:"column:ActivityCreatedAt;autoCreateTime" json:"created_at"`
}

func (ActivityLog) TableName() string { return "Activity_Log" }

// ─── DTOs ──────────────────────────────────────────────────────────────────

// CreateActivityLogRequest is used internally by the logger middleware.
type CreateActivityLogRequest struct {
	UserID              string `json:"user_id"`
	ModuleID            string `json:"module_id"`
	PermissionID        string `json:"permission_id"`
	ActivityAction      string `json:"activity_action"`
	TableAffected       string `json:"table_affected"`
	RecordID            string `json:"record_id"`
	OldValue            string `json:"old_value"`
	NewValue            string `json:"new_value"`
	ActivityDescription string `json:"activity_description"`
	IPAddress           string `json:"ip_address"`
}

// ─── Repository Interface ──────────────────────────────────────────────────

type ActivityLogRepository interface {
	FindAll(page, limit int, userID, moduleID, action, plantID string) ([]ActivityLog, int64, error)
	FindByID(id string) (*ActivityLog, error)
	Create(a *ActivityLog) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type ActivityLogUseCase interface {
	GetAll(page, limit int, userID, moduleID, action, plantID string) ([]ActivityLog, int64, error)
	GetByID(id string) (*ActivityLog, error)
	Record(ctx context.Context, req *CreateActivityLogRequest) error
}
