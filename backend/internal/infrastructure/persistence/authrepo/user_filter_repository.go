package authrepo

import (
	"time"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"gorm.io/gorm"
)

// ── User Filter Repository ────────────────────────────────────────────────

type userFilterRepository struct{ db *gorm.DB }

// NewUserFilterRepository creates a new UserFilterRepository.
func NewUserFilterRepository(db *gorm.DB) authdomain.UserFilterRepository {
	return &userFilterRepository{db: db}
}

func (r *userFilterRepository) FindFiltered(f *authdomain.UserFilter) ([]authdomain.User, int64, error) {
	var items []authdomain.User
	var total int64

	base := r.db.Model(&authdomain.User{})
	base = f.ApplyTo(base)
	base.Count(&total)

	err := f.ApplySort(base).
		Preload("Role").
		Preload("PICMappings").
		Offset((f.Page - 1) * f.Limit).
		Limit(f.Limit).
		Find(&items).Error

	return items, total, err
}

func (r *userFilterRepository) FindFacets(f *authdomain.UserFilter) (authdomain.UserFacets, error) {
	facets := authdomain.UserFacets{
		RoleID:       make(map[string]int64),
		DepartmentID: make(map[string]int64),
		UserStatus:   make(map[string]int64),
	}
	facets.DateRange.Field = "created_at"

	type kv struct {
		Key   string
		Count int64
	}

	// Helper: build base query excluding a specific field
	baseWithout := func(excludeRole, excludeDept, excludeStatus, excludeDate bool) *gorm.DB {
		tmp := &authdomain.UserFilter{
			Q: f.Q,
		}
		if !excludeRole {
			tmp.RoleID = f.RoleID
			tmp.RoleIDIn = f.RoleIDIn
		}
		if !excludeDept {
			tmp.DepartmentID = f.DepartmentID
		}
		if !excludeStatus {
			tmp.UserStatus = f.UserStatus
			tmp.UserStatusIn = f.UserStatusIn
		}
		if !excludeDate {
			tmp.DateFrom = f.DateFrom
			tmp.DateTo = f.DateTo
		}
		return tmp.ApplyTo(r.db.Model(&authdomain.User{}))
	}

	// RoleID facet
	var roleRows []kv
	if err := baseWithout(true, false, false, false).
		Select(`"RoleID" AS key, COUNT(*) AS count`).
		Group(`"RoleID"`).
		Scan(&roleRows).Error; err == nil {
		for _, row := range roleRows {
			facets.RoleID[row.Key] = row.Count
		}
	}

	// DepartmentID facet
	var deptRows []kv
	if err := baseWithout(false, true, false, false).
		Select(`"DepartmentID" AS key, COUNT(*) AS count`).
		Group(`"DepartmentID"`).
		Scan(&deptRows).Error; err == nil {
		for _, row := range deptRows {
			facets.DepartmentID[row.Key] = row.Count
		}
	}

	// UserStatus facet
	var statusRows []kv
	if err := baseWithout(false, false, true, false).
		Select(`"UserStatus" AS key, COUNT(*) AS count`).
		Group(`"UserStatus"`).
		Scan(&statusRows).Error; err == nil {
		for _, row := range statusRows {
			facets.UserStatus[row.Key] = row.Count
		}
	}

	// Date range
	type dateResult struct {
		Min *time.Time
		Max *time.Time
	}
	var dr dateResult
	if err := baseWithout(false, false, false, true).
		Select(`MIN("UserCreatedAt") AS min, MAX("UserCreatedAt") AS max`).
		Scan(&dr).Error; err == nil {
		facets.DateRange.Min = dr.Min
		facets.DateRange.Max = dr.Max
	}

	return facets, nil
}
