package apikeyusecase

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/monitoring-system/backend/internal/domain/apikey"
)

type useCase struct {
	repo apikey.Repository
}

func NewAPIKeyUseCase(repo apikey.Repository) apikey.UseCase {
	return &useCase{repo: repo}
}

func (uc *useCase) CreateKey(req *apikey.CreateAPIKeyRequest, createdByID string) (*apikey.CreateAPIKeyResponse, error) {
	if req.Name == "" {
		return nil, errors.New("key name is required")
	}

	isSingleUse := false
	if req.IsSingleUse != nil {
		isSingleUse = *req.IsSingleUse
	}

	// 1. Generate random token
	randomBytes := make([]byte, 24)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	randomHex := hex.EncodeToString(randomBytes)
	rawToken := fmt.Sprintf("MAKEY_%s", randomHex)
	prefix := rawToken[:12] // e.g. "MAKEY_a1b2c3"

	// 2. Hash token using SHA-256
	hash := sha256.Sum256([]byte(rawToken))
	keyHash := hex.EncodeToString(hash[:])

	// 3. Save to DB
	keyID := fmt.Sprintf("AKEY-%s", uuid.New().String()[:8])
	entity := &apikey.APIKey{
		KeyID:       keyID,
		Name:        req.Name,
		KeyHash:     keyHash,
		Prefix:      prefix,
		CreatedByID: createdByID,
		IsActive:    true,
		IsSingleUse: isSingleUse,
		CreatedAt:   time.Now(),
	}

	if err := uc.repo.Create(entity); err != nil {
		return nil, fmt.Errorf("failed to save api key: %w", err)
	}

	return &apikey.CreateAPIKeyResponse{
		KeyID:       keyID,
		Name:        entity.Name,
		RawToken:    rawToken,
		Prefix:      prefix,
		IsSingleUse: isSingleUse,
		CreatedAt:   entity.CreatedAt,
	}, nil
}

func (uc *useCase) ListKeys(createdBy string) ([]apikey.APIKeyResponse, error) {
	keys, err := uc.repo.FindAll(createdBy)
	if err != nil {
		return nil, err
	}

	res := make([]apikey.APIKeyResponse, len(keys))
	for i, k := range keys {
		res[i] = apikey.APIKeyResponse{
			KeyID:       k.KeyID,
			Name:        k.Name,
			Prefix:      k.Prefix,
			CreatedByID: k.CreatedByID,
			IsActive:    k.IsActive,
			IsSingleUse: k.IsSingleUse,
			UsedAt:      k.UsedAt,
			ExpiresAt:   k.ExpiresAt,
			CreatedAt:   k.CreatedAt,
		}
	}
	return res, nil
}

func (uc *useCase) RevokeKey(keyID string, requestedBy string) error {
	existing, err := uc.repo.FindByID(keyID)
	if err != nil || existing == nil {
		return errors.New("api key not found")
	}
	return uc.repo.Revoke(keyID)
}

func (uc *useCase) ValidateAndConsumeKey(rawToken string) (*apikey.APIKey, error) {
	if rawToken == "" {
		return nil, errors.New("empty api key token")
	}

	hash := sha256.Sum256([]byte(rawToken))
	keyHash := hex.EncodeToString(hash[:])

	k, err := uc.repo.FindByHash(keyHash)
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	if k == nil {
		return nil, errors.New("invalid API key")
	}

	if !k.IsActive {
		if k.UsedAt != nil {
			return nil, fmt.Errorf("API key has already been used on %s", k.UsedAt.Format(time.RFC3339))
		}
		return nil, errors.New("API key has been revoked")
	}

	if k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt) {
		return nil, errors.New("API key has expired")
	}

	// Mark as used if single-use
	if k.IsSingleUse {
		_ = uc.repo.MarkAsUsed(k.KeyID)
	}

	return k, nil
}
