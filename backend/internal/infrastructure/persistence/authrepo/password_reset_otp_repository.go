package authrepo

import (
	"errors"
	"time"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/pkg/idgen"
	"github.com/monitoring-system/backend/pkg/password"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type passwordResetOTPRepository struct {
	db *gorm.DB
}

func NewPasswordResetOTPRepository(db *gorm.DB) authdomain.PasswordResetOTPRepository {
	return &passwordResetOTPRepository{db: db}
}

func (r *passwordResetOTPRepository) ReplaceForUser(userID, otpHash string, expiresAt time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&authdomain.PasswordResetOTP{}).
			Where(`"UserID" = ? AND "UsedAt" IS NULL`, userID).
			Update("UsedAt", now).Error; err != nil {
			return err
		}

		challenge := authdomain.PasswordResetOTP{
			ResetID:   idgen.GenerateRandom("PROTP"),
			UserID:    userID,
			OTPHash:   otpHash,
			ExpiresAt: expiresAt,
		}
		if err := tx.Create(&challenge).Error; err != nil {
			return err
		}

		// Keep the table bounded without affecting active challenges.
		return tx.Where(`"ExpiresAt" < ?`, now.AddDate(0, 0, -30)).
			Delete(&authdomain.PasswordResetOTP{}).Error
	})
}

func (r *passwordResetOTPRepository) ConsumeValid(userID, otp string, now time.Time, maxAttempts int) (bool, error) {
	valid := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var challenge authdomain.PasswordResetOTP
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(`"UserID" = ? AND "UsedAt" IS NULL AND "ExpiresAt" > ? AND "Attempts" < ?`, userID, now, maxAttempts).
			Order(`"CreatedAt" DESC`).
			Take(&challenge).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}

		challenge.Attempts++
		if password.Compare(otp, challenge.OTPHash) {
			challenge.UsedAt = &now
			valid = true
		}
		return tx.Model(&challenge).Select("Attempts", "UsedAt").Updates(&challenge).Error
	})
	return valid, err
}
