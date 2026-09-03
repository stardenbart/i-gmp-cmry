package authrepo

import (
	"time"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"gorm.io/gorm"
)

type refreshTokenRepository struct{ db *gorm.DB }

func NewRefreshTokenRepository(db *gorm.DB) authdomain.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(rt *authdomain.RefreshToken) error {
	return r.db.Create(rt).Error
}

func (r *refreshTokenRepository) FindByHash(hash string) (*authdomain.RefreshToken, error) {
	var rt authdomain.RefreshToken
	err := r.db.Where("\"TokenHash\" = ?", hash).First(&rt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) RevokeByID(id string, replacedBy *string) error {
	updates := map[string]interface{}{"RevokedAt": time.Now()}
	if replacedBy != nil {
		updates["ReplacedBy"] = *replacedBy
	}
	return r.db.Model(&authdomain.RefreshToken{}).
		Where("\"ID\" = ?", id).
		Updates(updates).Error
}

func (r *refreshTokenRepository) RevokeFamily(familyID string) error {
	return r.db.Model(&authdomain.RefreshToken{}).
		Where("\"FamilyID\" = ? AND \"RevokedAt\" IS NULL", familyID).
		Update("RevokedAt", time.Now()).Error
}

func (r *refreshTokenRepository) RevokeAllForUser(userID string) error {
	return r.db.Model(&authdomain.RefreshToken{}).
		Where("\"UserID\" = ? AND \"RevokedAt\" IS NULL", userID).
		Update("RevokedAt", time.Now()).Error
}
