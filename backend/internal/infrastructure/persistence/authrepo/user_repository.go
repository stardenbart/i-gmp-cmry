package authrepo

import (
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"gorm.io/gorm"
)

type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) authdomain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindAll(page, limit int, search, roleID, deptID, plantID string) ([]authdomain.User, int64, error) {
	var users []authdomain.User
	var total int64
	q := r.db.Model(&authdomain.User{})
	if search != "" {
		q = q.Where("\"Username\" LIKE ? OR \"FullName\" LIKE ? OR \"Email\" LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if roleID != "" {
		q = q.Where("\"RoleID\" = ?", roleID)
	}
	if deptID != "" {
		q = q.Where("\"DepartmentID\" = ?", deptID)
	}
	if plantID != "" {
		if plantID == "GLOBAL" || plantID == "NULL" {
			q = q.Where("\"PlantID\" IS NULL OR \"PlantID\" = ''")
		} else {
			q = q.Where("\"PlantID\" = ?", plantID)
		}
	}
	q.Count(&total)
	err := q.Offset((page - 1) * limit).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *userRepository) FindByID(id string) (*authdomain.User, error) {
	var user authdomain.User
	err := r.db.Preload("Role").Preload("Department").Preload("PICMappings").Where("\"UserID\" = ?", id).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByUsername(username string) (*authdomain.User, error) {
	var user authdomain.User
	err := r.db.Where("\"Username\" = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*authdomain.User, error) {
	var user authdomain.User
	err := r.db.Where("\"Email\" = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepository) Create(u *authdomain.User) error {
	return r.db.Create(u).Error
}

func (r *userRepository) Update(u *authdomain.User) error {
	return r.db.Save(u).Error
}

func (r *userRepository) Delete(id string) error {
	return r.db.Where("\"UserID\" = ?", id).Delete(&authdomain.User{}).Error
}
