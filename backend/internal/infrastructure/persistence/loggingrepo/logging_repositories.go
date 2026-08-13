package loggingrepo

import (
	"time"

	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"gorm.io/gorm"
)

// ── LoginLog ──────────────────────────────────────────────────────────────

type loginLogRepository struct{ db *gorm.DB }

func NewLoginLogRepository(db *gorm.DB) logdomain.LoginLogRepository {
	return &loginLogRepository{db: db}
}

func (r *loginLogRepository) FindAll(page, limit int, userID, plantID string) ([]logdomain.LoginLog, int64, error) {
	var items []logdomain.LoginLog
	var total int64
	q := r.db.Model(&logdomain.LoginLog{})
	if userID != "" {
		q = q.Where("\"Login_Log\".\"UserID\" = ?", userID)
	}
	if plantID != "" {
		q = q.Joins("JOIN \"Users\" ON \"Login_Log\".\"UserID\" = \"Users\".\"UserID\"")
		if plantID == "GLOBAL" || plantID == "NULL" {
			q = q.Where("\"Users\".\"PlantID\" IS NULL OR \"Users\".\"PlantID\" = ''")
		} else {
			q = q.Where("\"Users\".\"PlantID\" = ?", plantID)
		}
	}
	q.Count(&total)
	err := q.Order(`"LoginAt" DESC`).Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *loginLogRepository) FindByID(id string) (*logdomain.LoginLog, error) {
	var item logdomain.LoginLog
	err := r.db.Where("\"LoginLogID\" = ?", id).First(&item).Error
	return &item, err
}

func (r *loginLogRepository) Create(l *logdomain.LoginLog) error { return r.db.Create(l).Error }
func (r *loginLogRepository) UpdateLogoutAt(id string, logoutAt time.Time) error {
	return r.db.Model(&logdomain.LoginLog{}).Where("\"LoginLogID\" = ?", id).Update("\"LogoutAt\"", logoutAt).Error
}

// ── ActivityLog ───────────────────────────────────────────────────────────

type activityLogRepository struct{ db *gorm.DB }

func NewActivityLogRepository(db *gorm.DB) logdomain.ActivityLogRepository {
	return &activityLogRepository{db: db}
}

func (r *activityLogRepository) FindAll(page, limit int, userID, moduleID, action, plantID string) ([]logdomain.ActivityLog, int64, error) {
	var items []logdomain.ActivityLog
	var total int64
	q := r.db.Model(&logdomain.ActivityLog{})
	if userID != "" {
		q = q.Where("\"Activity_Log\".\"UserID\" = ?", userID)
	}
	if moduleID != "" {
		q = q.Where("\"Activity_Log\".\"ModuleID\" = ?", moduleID)
	}
	if plantID != "" {
		q = q.Joins("JOIN \"Users\" ON \"Activity_Log\".\"UserID\" = \"Users\".\"UserID\"")
		if plantID == "GLOBAL" || plantID == "NULL" {
			q = q.Where("\"Users\".\"PlantID\" IS NULL OR \"Users\".\"PlantID\" = ''")
		} else {
			q = q.Where("\"Users\".\"PlantID\" = ?", plantID)
		}
	}
	if action != "" {
		pattern := "%" + action + "%"
		q = q.Where(`"Activity_Log"."ActivityAction" ILIKE ? OR "Activity_Log"."ActivityDescription" ILIKE ? OR "Activity_Log"."TableAffected" ILIKE ? OR "Activity_Log"."RecordID" ILIKE ? OR "Activity_Log"."OldValue" ILIKE ? OR "Activity_Log"."NewValue" ILIKE ?`, pattern, pattern, pattern, pattern, pattern, pattern)
	}
	q.Count(&total)
	err := q.Order(`"ActivityCreatedAt" DESC`).Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *activityLogRepository) FindByID(id string) (*logdomain.ActivityLog, error) {
	var item logdomain.ActivityLog
	err := r.db.Where("\"ActivityLogID\" = ?", id).First(&item).Error
	return &item, err
}

func (r *activityLogRepository) Create(a *logdomain.ActivityLog) error {
	if a == nil {
		return nil
	}
	logCopy := *a
	go func(item logdomain.ActivityLog) {
		_ = r.db.Create(&item).Error
	}(logCopy)
	return nil
}
