// Package main provides a utility to generate a cryptographically secure
// base64-encoded 32-byte AES-256 key for use as SETTING_ENCRYPTION_KEY.
//
// Run: go run ./cmd/genkey/main.go
package main

import (
	"fmt"
	"log"

	"github.com/monitoring-system/backend/pkg/crypto"
)

func main() {
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("failed to generate key: %v", err)
	}
	fmt.Println("Generated AES-256 Encryption Key (store this in SETTING_ENCRYPTION_KEY):")
	fmt.Println(key)
}
