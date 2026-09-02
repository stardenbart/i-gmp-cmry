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
	// The DashboardKey column + composite primary key are introduced by
	// migration 043, which always runs before this constructor at startup —
	// AutoMigrate here only needs to keep other column additions in sync,
	// never the primary key itself (GORM won't safely rewrite an existing
	// PK constraint, and doesn't need to: the migration already did it).
	_ = db.AutoMigrate(&dashboarddomain.UserDashboardLayout{})
	return &dashboardLayoutRepository{db: db}
}

func (r *dashboardLayoutRepository) FindByUserID(userID, dashboardKey string) (*dashboarddomain.UserDashboardLayout, error) {
	var item dashboarddomain.UserDashboardLayout
	err := r.db.Where(`"UserID" = ? AND "DashboardKey" = ?`, userID, dashboardKey).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *dashboardLayoutRepository) Upsert(userID, dashboardKey string, layoutJSON string) error {
	var existing dashboarddomain.UserDashboardLayout
	err := r.db.Where(`"UserID" = ? AND "DashboardKey" = ?`, userID, dashboardKey).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(&dashboarddomain.UserDashboardLayout{
			UserID:       userID,
			DashboardKey: dashboardKey,
			LayoutJSON:   layoutJSON,
		}).Error
	}
	if err != nil {
		return err
	}
	return r.db.Model(&dashboarddomain.UserDashboardLayout{}).
		Where(`"UserID" = ? AND "DashboardKey" = ?`, userID, dashboardKey).
		Update("LayoutJSON", layoutJSON).Error
}
