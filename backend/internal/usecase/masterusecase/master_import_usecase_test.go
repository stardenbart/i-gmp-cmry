package masterusecase

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/monitoring-system/backend/internal/domain/master"
	redis "github.com/redis/go-redis/v9"
)

func TestResolveImportPlant(t *testing.T) {
	tests := []struct {
		name      string
		kind      string
		actor     master.ImportActor
		requested string
		want      string
		status    int
	}{
		{name: "plant user uses own scope", kind: "aspek", actor: master.ImportActor{PlantID: "PLT-1"}, want: "PLT-1"},
		{name: "plant user cannot cross scope", kind: "detail", actor: master.ImportActor{PlantID: "PLT-1"}, requested: "PLT-2", status: http.StatusForbidden},
		{name: "superadmin selects plant", kind: "uraian", actor: master.ImportActor{RoleID: "ROLE-000"}, requested: "PLT-2", want: "PLT-2"},
		{name: "superadmin must select plant", kind: "aspek", actor: master.ImportActor{RoleID: "ROLE-000"}, status: http.StatusBadRequest},
		{name: "hei remains global", kind: "hei", actor: master.ImportActor{PlantID: "PLT-1"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveImportPlant(tt.kind, tt.actor, tt.requested)
			if tt.status == 0 {
				if err != nil || got != tt.want {
					t.Fatalf("resolveImportPlant() = %q, %v; want %q", got, err, tt.want)
				}
				return
			}
			importError, ok := err.(*MasterImportError)
			if !ok || importError.Status != tt.status {
				t.Fatalf("expected status %d, got %v", tt.status, err)
			}
		})
	}
}

func TestValidationTokenIsSingleUseWithRedis(t *testing.T) {
	address := os.Getenv("MASTER_IMPORT_TEST_REDIS_ADDR")
	if address == "" {
		t.Skip("MASTER_IMPORT_TEST_REDIS_ADDR is not set")
	}
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: address})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("redis ping: %v", err)
	}
	uc := NewMasterImportUseCase(nil, rdb, nil)
	state := master.ImportTokenState{
		ImportID: "MIMP-TEST", UserID: "USR-TEST", PlantID: "PLT-TEST",
		ImportType: master.ImportTypeAspek, FileHash: "hash", TemplateVersion: master.ImportTemplateV1,
		RowCount: 1, Status: "validated", ExpiresAt: time.Now().Add(master.ImportTokenTTL),
	}
	token, err := uc.storeValidationToken(ctx, state)
	if err != nil {
		t.Fatalf("storeValidationToken: %v", err)
	}
	acquired, key, err := uc.acquireValidationToken(ctx, token)
	if err != nil || acquired.Status != "committing" {
		t.Fatalf("first acquire = %+v, %v", acquired, err)
	}
	defer rdb.Del(ctx, key)
	_, _, err = uc.acquireValidationToken(ctx, token)
	importError, ok := err.(*MasterImportError)
	if !ok || importError.Status != http.StatusGone {
		t.Fatalf("replayed token must be rejected with 410, got %v", err)
	}
}

func TestImportRateLimitWithRedis(t *testing.T) {
	address := os.Getenv("MASTER_IMPORT_TEST_REDIS_ADDR")
	if address == "" {
		t.Skip("MASTER_IMPORT_TEST_REDIS_ADDR is not set")
	}
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: address})
	defer rdb.Close()
	uc := NewMasterImportUseCase(nil, rdb, nil)
	userID := "USR-RATE-" + time.Now().Format("150405.000000000")
	defer rdb.Del(ctx, "master-import:rate:"+userID)
	for i := 0; i < 5; i++ {
		if err := uc.CheckRateLimit(ctx, userID); err != nil {
			t.Fatalf("attempt %d unexpectedly rejected: %v", i+1, err)
		}
	}
	err := uc.CheckRateLimit(ctx, userID)
	importError, ok := err.(*MasterImportError)
	if !ok || importError.Status != http.StatusTooManyRequests {
		t.Fatalf("sixth attempt must be rate limited, got %v", err)
	}
}

func TestValidateFieldValues(t *testing.T) {
	row := master.ImportPreviewRow{
		NormalizedData: map[string]any{"detail_id": "DET-001", "uraian_text": "Lantai bersih", "standard_score": "2"},
		Errors:         []master.ImportFieldError{},
	}
	validateFieldValues(master.ImportTypeUraian, &row)
	if len(row.Errors) != 0 || row.NormalizedData["standard_score"] != 2 {
		t.Fatalf("expected a valid normalized score, got %+v", row)
	}

	row.NormalizedData["standard_score"] = "0"
	validateFieldValues(master.ImportTypeUraian, &row)
	if !hasErrorCode(row, "INVALID_VALUE") {
		t.Fatalf("expected invalid score error, got %+v", row.Errors)
	}
}

func TestHEIDuplicateKeysIncludeNameAndCode(t *testing.T) {
	raw := map[string]string{"category_name": " Habit ", "hei_name": "APD", "hei_code": " h-1 "}
	normalized := map[string]normalizedValue{
		" Habit ": {Normalized: "habit"},
		"APD":     {Normalized: "apd"},
		" h-1 ":   {Code: "H-1"},
	}
	keys := duplicateKeys(master.ImportTypeHEI, map[string]any{}, normalized, raw)
	if len(keys) != 2 || keys[0] != "name:habit\x00apd" || keys[1] != "code:habit\x00H-1" {
		t.Fatalf("unexpected HEI duplicate keys: %#v", keys)
	}
}

func TestValidateParentsAllowsNewHEICategory(t *testing.T) {
	uc := NewMasterImportUseCase(nil, nil, nil)
	rows := []master.ImportPreviewRow{{
		NormalizedData: map[string]any{
			"category_name": "Food Safety Culture",
			"hei_name":      "Pelaporan kondisi tidak aman",
		},
	}}

	if err := uc.validateParents(context.Background(), nil, master.ImportTypeHEI, "", rows, nil, nil); err != nil {
		t.Fatalf("new HEI categories must not require an existing reference: %v", err)
	}
}
