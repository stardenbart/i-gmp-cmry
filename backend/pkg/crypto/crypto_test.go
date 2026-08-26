package crypto

import (
	"strings"
	"testing"
)

const testKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="
const otherTestKey = "YWJjZGVmMDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODk="

func TestEncryptUsesVersionedEnvelopeAndRoundTrips(t *testing.T) {
	svc, err := New(testKey)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := svc.Encrypt("sensitive value")
	if err != nil {
		t.Fatal(err)
	}
	if !IsVersionedCiphertext(ciphertext) {
		t.Fatalf("ciphertext is not versioned: %q", ciphertext)
	}
	plaintext, err := svc.Decrypt(ciphertext)
	if err != nil || plaintext != "sensitive value" {
		t.Fatalf("round trip failed: plaintext=%q err=%v", plaintext, err)
	}
}

func TestDecryptSupportsLegacyCiphertext(t *testing.T) {
	svc, _ := New(testKey)
	ciphertext, _ := svc.Encrypt("legacy")
	legacy := strings.TrimPrefix(ciphertext, ciphertextPrefix)
	plaintext, err := svc.Decrypt(legacy)
	if err != nil || plaintext != "legacy" {
		t.Fatalf("legacy decrypt failed: plaintext=%q err=%v", plaintext, err)
	}
}

func TestDecryptRejectsWrongKey(t *testing.T) {
	oldSvc, _ := New(testKey)
	newSvc, _ := New(otherTestKey)
	ciphertext, _ := oldSvc.Encrypt("secret")
	if _, err := newSvc.Decrypt(ciphertext); err == nil {
		t.Fatal("expected authentication failure with wrong key")
	}
}
