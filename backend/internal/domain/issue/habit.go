package issue

import (
	"time"
)

// Habit represents the Habit_Master table.
type Habit struct {
	HabitID        string    `gorm:"column:HabitID;primaryKey" json:"habit_id"`
	HabitCode      string    `gorm:"column:HabitCode;unique" json:"habit_code,omitempty"`
	HabitName      string    `gorm:"column:HabitName;not null" json:"habit_name"`
	HabitCategory  string    `gorm:"column:HabitCategory" json:"habit_category,omitempty"`
	Description    string    `gorm:"column:Description" json:"description,omitempty"`
	HabitCreatedAt time.Time `gorm:"column:HabitCreatedAt;autoCreateTime" json:"created_at"`
	HabitUpdatedAt time.Time `gorm:"column:HabitUpdatedAt;autoUpdateTime" json:"updated_at"`
}

func (Habit) TableName() string { return "Habit_Master" }

// DTOs
type CreateHabitRequest struct {
	HabitCode     string `json:"habit_code"`
	HabitName     string `json:"habit_name" validate:"required,max=150"`
	HabitCategory string `json:"habit_category"`
	Description   string `json:"description"`
}

type UpdateHabitRequest struct {
	HabitCode     string `json:"habit_code"`
	HabitName     string `json:"habit_name" validate:"omitempty,max=150"`
	HabitCategory string `json:"habit_category"`
	Description   string `json:"description"`
}

// Interfaces
type HabitRepository interface {
	FindAll(page, limit int, search string) ([]Habit, int64, error)
	FindByID(id string) (*Habit, error)
	Create(h *Habit) error
	Update(h *Habit) error
	Delete(id string) error
}

type HabitUseCase interface {
	GetAll(page, limit int, search string) ([]Habit, int64, error)
	GetByID(id string) (*Habit, error)
	Create(req *CreateHabitRequest) (*Habit, error)
	Update(id string, req *UpdateHabitRequest) (*Habit, error)
	Delete(id string) error
}
