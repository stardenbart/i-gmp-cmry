package authrepo

import (
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"gorm.io/gorm"
)

type moduleRepository struct{ db *gorm.DB }

func NewModuleRepository(db *gorm.DB) authdomain.ModuleRepository {
	return &moduleRepository{db: db}
}

func (r *moduleRepository) FindAll() ([]authdomain.Module, error) {
	var modules []authdomain.Module
	err := r.db.Preload("Permissions").Find(&modules).Error
	return modules, err
}

func (r *moduleRepository) FindByID(id string) (*authdomain.Module, error) {
	var module authdomain.Module
	err := r.db.Preload("Permissions").Where("\"ModuleID\" = ?", id).First(&module).Error
	return &module, err
}

func (r *moduleRepository) Create(m *authdomain.Module) error {
	return r.db.Create(m).Error
}

func (r *moduleRepository) Update(m *authdomain.Module) error {
	return r.db.Save(m).Error
}

func (r *moduleRepository) Delete(id string) error {
	return r.db.Where("\"ModuleID\" = ?", id).Delete(&authdomain.Module{}).Error
}
