package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const ciphertextPrefix = "enc:v1:"

// Service handles AES-256-GCM encryption/decryption for sensitive settings values.
type Service struct {
	key []byte
}

// New creates an AES encryption service. key must be exactly 32 bytes (for AES-256).
func New(base64Key string) (*Service, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key (must be base64-encoded 32-byte key): %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must decode to exactly 32 bytes, got %d", len(key))
	}
	return &Service{key: key}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns base64-encoded ciphertext.
func (s *Service) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertextPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded AES-256-GCM ciphertext.
func (s *Service) Decrypt(cipherBase64 string) (string, error) {
	cipherBase64 = strings.TrimPrefix(cipherBase64, ciphertextPrefix)
	ciphertext, err := base64.StdEncoding.DecodeString(cipherBase64)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

// IsVersionedCiphertext reports whether a value uses the authenticated
// envelope emitted by current versions of Encrypt.
func IsVersionedCiphertext(value string) bool {
	return strings.HasPrefix(value, ciphertextPrefix)
}

// DecryptWithFallback attempts to decrypt the ciphertext. If it fails (e.g. not a valid base64 or not encrypted),
// it returns the original string. This is useful for backward compatibility with plaintext data.
func (s *Service) DecryptWithFallback(cipherBase64 string) string {
	if cipherBase64 == "" {
		return ""
	}
	plaintext, err := s.Decrypt(cipherBase64)
	if err != nil {
		return cipherBase64 // fallback to plaintext
	}
	return plaintext
}

// GenerateKey generates a new cryptographically secure 32-byte key, base64-encoded.
// Run once to get the key and store it in SETTING_ENCRYPTION_KEY env variable.
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
