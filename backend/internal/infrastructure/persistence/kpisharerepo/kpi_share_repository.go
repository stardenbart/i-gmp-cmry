package kpisharerepo

import (
	"errors"
	"time"

	"github.com/monitoring-system/backend/internal/domain/kpishare"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type repository struct{ db *gorm.DB }

func New(db *gorm.DB) kpishare.Repository { return &repository{db: db} }

func (r *repository) Create(share *kpishare.PublicShare) error { return r.db.Create(share).Error }

func (r *repository) ListByOwner(ownerUserID string) ([]kpishare.PublicShare, error) {
	var shares []kpishare.PublicShare
	err := r.db.Where(`"OwnerUserID" = ?`, ownerUserID).Order(`"CreatedAt" DESC`).Find(&shares).Error
	return shares, err
}

func (r *repository) FindByID(shareID string) (*kpishare.PublicShare, error) {
	var share kpishare.PublicShare
	err := r.db.Where(`"ShareID" = ?`, shareID).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &share, err
}

func (r *repository) FindByTokenHash(tokenHash string) (*kpishare.PublicShare, error) {
	var share kpishare.PublicShare
	err := r.db.Where(`"TokenHash" = ?`, tokenHash).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &share, err
}

func (r *repository) Revoke(shareID, ownerUserID string, now time.Time) (bool, error) {
	result := r.db.Model(&kpishare.PublicShare{}).
		Where(`"ShareID" = ? AND "OwnerUserID" = ? AND "RevokedAt" IS NULL`, shareID, ownerUserID).
		Updates(map[string]interface{}{"RevokedAt": now, "UpdatedAt": now})
	return result.RowsAffected > 0, result.Error
}

func (r *repository) Rotate(shareID, ownerUserID, tokenHash, tokenPrefix string, now time.Time) (*kpishare.PublicShare, error) {
	var share kpishare.PublicShare
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(`"ShareID" = ? AND "OwnerUserID" = ?`, shareID, ownerUserID).First(&share).Error; err != nil {
			return err
		}
		if err := tx.Model(&share).Updates(map[string]interface{}{
			"TokenHash": tokenHash, "TokenPrefix": tokenPrefix, "RevokedAt": nil,
			"AccessCount": 0, "LastAccessedAt": nil, "UpdatedAt": now,
		}).Error; err != nil {
			return err
		}
		share.TokenHash, share.TokenPrefix, share.RevokedAt = tokenHash, tokenPrefix, nil
		share.AccessCount, share.LastAccessedAt, share.UpdatedAt = 0, nil, now
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &share, err
}

func (r *repository) TouchAccess(shareID string, now time.Time) error {
	// Throttle writes: one update per five minutes per link, while still
	// counting meaningful visits rather than every widget request.
	return r.db.Model(&kpishare.PublicShare{}).
		Where(`"ShareID" = ? AND ("LastAccessedAt" IS NULL OR "LastAccessedAt" < ?)`, shareID, now.Add(-5*time.Minute)).
		Updates(map[string]interface{}{"LastAccessedAt": now, "AccessCount": gorm.Expr(`"AccessCount" + 1`)}).Error
}

func (r *repository) PlantName(plantID string) (string, error) {
	var name string
	err := r.db.Table(`"Plant_Master"`).Select(`"PlantName"`).Where(`"PlantID" = ?`, plantID).Scan(&name).Error
	if err == nil && name == "" {
		return "", gorm.ErrRecordNotFound
	}
	return name, err
}
