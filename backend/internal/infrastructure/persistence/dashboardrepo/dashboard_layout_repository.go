package dashboardrepo

import (
	"errors"

	dashboarddomain "github.com/monitoring-system/backend/internal/domain/dashboard"
	"gorm.io/gorm"
)

type dashboardLayoutRepository struct {
	db *gorm.DB
}

func NewDashboardLayoutRepository(db *gorm.DB) dashboarddomain.DashboardLayoutRepository {
	_ = db.AutoMigrate(&dashboarddomain.UserDashboardLayout{})
	return &dashboardLayoutRepository{db: db}
}

func (r *dashboardLayoutRepository) FindByUserID(userID string) (*dashboarddomain.UserDashboardLayout, error) {
	var item dashboarddomain.UserDashboardLayout
	err := r.db.Where(`"UserID" = ?`, userID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *dashboardLayoutRepository) Upsert(userID string, layoutJSON string) error {
	var existing dashboarddomain.UserDashboardLayout
	err := r.db.Where(`"UserID" = ?`, userID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(&dashboarddomain.UserDashboardLayout{
			UserID:     userID,
			LayoutJSON: layoutJSON,
		}).Error
	}
	if err != nil {
		return err
	}
	return r.db.Model(&dashboarddomain.UserDashboardLayout{}).
		Where(`"UserID" = ?`, userID).
		Update("LayoutJSON", layoutJSON).Error
}
