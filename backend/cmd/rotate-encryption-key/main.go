package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/monitoring-system/backend/config"
	appcrypto "github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/keyrotation"
)

func main() {
	apply := flag.Bool("apply", false, "commit the rotation; without this flag only a dry-run is performed")
	flag.Parse()

	oldRaw := os.Getenv("OLD_SETTING_ENCRYPTION_KEY")
	newRaw := os.Getenv("NEW_SETTING_ENCRYPTION_KEY")
	if oldRaw == "" || newRaw == "" {
		log.Fatal("OLD_SETTING_ENCRYPTION_KEY and NEW_SETTING_ENCRYPTION_KEY are required")
	}
	if oldRaw == newRaw {
		log.Fatal("old and new encryption keys must differ")
	}
	oldKey, err := appcrypto.New(oldRaw)
	if err != nil {
		log.Fatalf("invalid old encryption key: %v", err)
	}
	newKey, err := appcrypto.New(newRaw)
	if err != nil {
		log.Fatalf("invalid new encryption key: %v", err)
	}

	cfg := config.Load()
	// Maintenance commands must not log SQL parameters or ciphertext.
	cfg.AppEnv = "production"
	db, err := config.OpenDatabase(cfg)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	stats, err := keyrotation.Rotate(db, oldKey, newKey, *apply)
	if err != nil {
		log.Fatalf("rotation aborted: %v", err)
	}
	mode := "DRY RUN"
	if *apply {
		mode = "APPLIED"
	}
	fmt.Printf("%s: settings=%d issues=%d photo_fields=%d already_current=%d plaintext_skipped=%d\n",
		mode, stats.SettingsRotated, stats.IssuesRotated, stats.PhotosRotated, stats.AlreadyCurrent, stats.PlaintextSkipped)
}
