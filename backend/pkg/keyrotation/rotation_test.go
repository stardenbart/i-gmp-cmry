package keyrotation

import (
	"strings"
	"testing"

	appcrypto "github.com/monitoring-system/backend/pkg/crypto"
)

const oldKeyRaw = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="
const newKeyRaw = "YWJjZGVmMDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODk="

func services(t *testing.T) (*appcrypto.Service, *appcrypto.Service) {
	t.Helper()
	oldKey, err := appcrypto.New(oldKeyRaw)
	if err != nil {
		t.Fatal(err)
	}
	newKey, err := appcrypto.New(newKeyRaw)
	if err != nil {
		t.Fatal(err)
	}
	return oldKey, newKey
}

func TestRotateValueMigratesLegacyCiphertext(t *testing.T) {
	oldKey, newKey := services(t)
	legacy, _ := oldKey.Encrypt("audit note")
	legacy = strings.TrimPrefix(legacy, "enc:v1:")
	rotated, state, err := rotateValue(legacy, false, oldKey, newKey)
	if err != nil || state != "rotated" || !appcrypto.IsVersionedCiphertext(rotated) {
		t.Fatalf("rotation failed: state=%s err=%v", state, err)
	}
	plaintext, err := newKey.Decrypt(rotated)
	if err != nil || plaintext != "audit note" {
		t.Fatalf("new key cannot decrypt rotated value: %q %v", plaintext, err)
	}
}

func TestRotateValueSkipsPlaintextAndRecognizesCurrent(t *testing.T) {
	oldKey, newKey := services(t)
	if value, state, err := rotateValue("YWJjZA==", false, oldKey, newKey); err != nil || state != "plaintext" || value != "YWJjZA==" {
		t.Fatalf("base64-like plaintext changed: value=%q state=%q err=%v", value, state, err)
	}
	current, _ := newKey.Encrypt("current")
	if _, state, err := rotateValue(current, false, oldKey, newKey); err != nil || state != "current" {
		t.Fatalf("current ciphertext not recognized: state=%q err=%v", state, err)
	}
}

func TestRotateValueRejectsUnreadableRequiredSetting(t *testing.T) {
	oldKey, newKey := services(t)
	if _, _, err := rotateValue("not encrypted", true, oldKey, newKey); err == nil {
		t.Fatal("expected required encrypted setting to fail")
	}
}
