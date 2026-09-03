package auth

import "time"

// RefreshToken represents one issued refresh token. Only its SHA-256 hash is
// ever stored — the raw value lives solely in the client's httpOnly cookie,
// so a compromised database backup can't be replayed as a session.
//
// FamilyID links every token produced by rotating from the same original
// login. On rotation the old row is marked RevokedAt + ReplacedBy and a new
// row is created in the same family; if a revoked token is ever presented
// again (RevokedAt already set), that's reuse of a stolen/rotated token, and
// the whole family is revoked immediately (see authusecase.Refresh).
type RefreshToken struct {
	ID         string     `gorm:"column:ID;primaryKey" json:"id"`
	UserID     string     `gorm:"column:UserID;not null;index" json:"user_id"`
	FamilyID   string     `gorm:"column:FamilyID;not null;index" json:"family_id"`
	TokenHash  string     `gorm:"column:TokenHash;not null;uniqueIndex" json:"-"`
	ExpiresAt  time.Time  `gorm:"column:ExpiresAt;not null" json:"expires_at"`
	RevokedAt  *time.Time `gorm:"column:RevokedAt" json:"revoked_at,omitempty"`
	ReplacedBy *string    `gorm:"column:ReplacedBy" json:"replaced_by,omitempty"`
	IPAddress  string     `gorm:"column:IPAddress;size:50" json:"ip_address"`
	DeviceInfo string     `gorm:"column:DeviceInfo;size:255" json:"device_info"`
	CreatedAt  time.Time  `gorm:"column:CreatedAt;autoCreateTime" json:"created_at"`
}

func (RefreshToken) TableName() string { return "Refresh_Token" }

// RefreshTokenRepository persists and revokes refresh tokens.
type RefreshTokenRepository interface {
	Create(rt *RefreshToken) error
	FindByHash(hash string) (*RefreshToken, error)
	// RevokeByID marks a single token revoked. replacedBy is the ID of the
	// token it was rotated into, or nil for a plain logout/manual revoke.
	RevokeByID(id string, replacedBy *string) error
	// RevokeFamily revokes every still-active token descended from the same
	// original login — used when a revoked token is replayed (theft signal).
	RevokeFamily(familyID string) error
	// RevokeAllForUser revokes every active session for a user (e.g. an
	// admin forcing logout, or a password change).
	RevokeAllForUser(userID string) error
}
