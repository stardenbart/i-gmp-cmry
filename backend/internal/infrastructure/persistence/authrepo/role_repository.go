package authrepo

import (
	"strings"

	"github.com/monitoring-system/backend/internal/domain/auth"
	"gorm.io/gorm"
)

type roleRepo struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) auth.RoleRepository {
	return &roleRepo{db: db}
}

func (r *roleRepo) FindAll(page, limit int, search string) ([]auth.Role, int64, error) {
	var roles []auth.Role
	var total int64

	query := r.db.Model(&auth.Role{})

	if search != "" {
		search = "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(RoleName) LIKE ? OR LOWER(RoleDescription) LIKE ?", search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("RoleCreatedAt DESC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

func (r *roleRepo) FindByID(id string) (*auth.Role, error) {
	var role auth.Role
	if err := r.db.Where("RoleID = ?", id).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepo) Create(role *auth.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepo) Update(role *auth.Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepo) Delete(id string) error {
	return r.db.Delete(&auth.Role{}, "RoleID = ?", id).Error
}
