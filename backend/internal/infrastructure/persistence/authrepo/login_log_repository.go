package authrepo

import (
	"time"

	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"gorm.io/gorm"
)

type loginLogRepository struct{ db *gorm.DB }

func NewLoginLogRepository(db *gorm.DB) logdomain.LoginLogRepository {
	return &loginLogRepository{db: db}
}

func (r *loginLogRepository) FindAll(page, limit int, userID string) ([]logdomain.LoginLog, int64, error) {
	var logs []logdomain.LoginLog
	var total int64
	q := r.db.Model(&logdomain.LoginLog{})
	if userID != "" {
		q = q.Where("UserID = ?", userID)
	}
	q.Count(&total)
	err := q.Order("LoginAt DESC").Offset((page-1)*limit).Limit(limit).Find(&logs).Error
	return logs, total, err
}

func (r *loginLogRepository) FindByID(id string) (*logdomain.LoginLog, error) {
	var l logdomain.LoginLog
	err := r.db.Where("LoginLogID = ?", id).First(&l).Error
	return &l, err
}

func (r *loginLogRepository) Create(l *logdomain.LoginLog) error {
	return r.db.Create(l).Error
}

func (r *loginLogRepository) UpdateLogoutAt(id string, logoutAt time.Time) error {
	return r.db.Model(&logdomain.LoginLog{}).
		Where("LoginLogID = ?", id).
		Update("LogoutAt", logoutAt).Error
}
