package issuerepo

import (
	"github.com/monitoring-system/backend/internal/domain/issue"
	"gorm.io/gorm"
)

type habitRepository struct {
	db *gorm.DB
}

func NewHabitRepository(db *gorm.DB) issue.HabitRepository {
	return &habitRepository{db: db}
}

func (r *habitRepository) FindAll(page, limit int, search string) ([]issue.Habit, int64, error) {
	var habits []issue.Habit
	var total int64

	query := r.db.Model(&issue.Habit{})
	if search != "" {
		s := "%" + search + "%"
		query = query.Where("\"HabitName\" ILIKE ? OR \"HabitCode\" ILIKE ? OR \"HabitCategory\" ILIKE ?", s, s, s)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("\"HabitCreatedAt\" DESC").Offset(offset).Limit(limit).Find(&habits).Error; err != nil {
		return nil, 0, err
	}

	return habits, total, nil
}

func (r *habitRepository) FindByID(id string) (*issue.Habit, error) {
	var h issue.Habit
	if err := r.db.Where("\"HabitID\" = ?", id).First(&h).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *habitRepository) Create(h *issue.Habit) error {
	return r.db.Create(h).Error
}

func (r *habitRepository) Update(h *issue.Habit) error {
	return r.db.Save(h).Error
}

func (r *habitRepository) Delete(id string) error {
	return r.db.Where("\"HabitID\" = ?", id).Delete(&issue.Habit{}).Error
}
