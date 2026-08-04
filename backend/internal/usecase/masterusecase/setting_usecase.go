package masterusecase

import (
	"context"
	"fmt"

	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/crypto"
	pkgstorage "github.com/monitoring-system/backend/pkg/storage"
)

type settingUseCase struct {
	repo   master.SettingRepository
	crypto *crypto.Service
	minio  *pkgstorage.MinioStorage
}

func NewSettingUseCase(
	repo master.SettingRepository,
	crypto *crypto.Service,
	minio *pkgstorage.MinioStorage,
) master.SettingUseCase {
	return &settingUseCase{repo: repo, crypto: crypto, minio: minio}
}

func (uc *settingUseCase) GetAll(plantID string) ([]master.SettingResponse, error) {
	items, err := uc.repo.FindAll(plantID)
	if err != nil {
		return nil, err
	}
	result := make([]master.SettingResponse, 0, len(items))
	for _, item := range items {
		result = append(result, uc.toResponse(&item))
	}
	return result, nil
}

func (uc *settingUseCase) GetByKey(key, plantID string) (*master.SettingResponse, error) {
	item, err := uc.repo.FindByKey(key, plantID)
	if err != nil {
		return nil, fmt.Errorf("setting not found: %w", err)
	}
	res := uc.toResponse(item)
	return &res, nil
}

func (uc *settingUseCase) Update(key string, req *master.UpdateSettingRequest, updatedBy, plantID string) (*master.SettingResponse, error) {
	targetPlantID := plantID
	if req.PlantID != "" {
		targetPlantID = req.PlantID
	}

	// Check the setting exists (or fallback to global)
	existing, err := uc.repo.FindByKey(key, targetPlantID)
	if err != nil {
		return nil, fmt.Errorf("setting not found: %w", err)
	}

	valueToStore := req.SettingValue

	// If this setting is marked as encrypted, encrypt before persisting
	if existing.IsEncrypted && uc.crypto != nil {
		encrypted, err := uc.crypto.Encrypt(req.SettingValue)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt value: %w", err)
		}
		valueToStore = encrypted
	}

	// Persist new value
	if err := uc.repo.Update(key, valueToStore, updatedBy, targetPlantID); err != nil {
		return nil, err
	}

	// Side effect: If updating MinIO IPs, reload policy in real-time
	if key == master.SettingKeyMinioAllowedIPs && uc.minio != nil {
		if err := uc.minio.UpdateIPWhitelistPolicy(context.Background(), req.SettingValue); err != nil {
			// Non-fatal: log but do not fail the API call
			// The DB is already updated; the next restart will also re-sync
			return nil, fmt.Errorf("setting saved but failed to sync MinIO policy: %w", err)
		}
	}

	// Return updated record
	return uc.GetByKey(key, targetPlantID)
}

// toResponse converts a Setting DB record to a safe SettingResponse.
// Encrypted values are masked from the API response.
func (uc *settingUseCase) toResponse(s *master.Setting) master.SettingResponse {
	value := s.SettingValue
	if s.IsEncrypted {
		value = "***" // never expose encrypted values in API responses
	}
	return master.SettingResponse{
		SettingKey:   s.SettingKey,
		PlantID:      s.PlantID,
		SettingValue: value,
		IsEncrypted:  s.IsEncrypted,
		Description:  s.Description,
		UpdatedAt:    s.UpdatedAt,
		UpdatedBy:    s.UpdatedBy,
	}
}
