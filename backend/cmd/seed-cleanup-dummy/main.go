// cmd/seed-cleanup-dummy/main.go
// Jalankan dengan: go run ./cmd/seed-cleanup-dummy
//
// Kebalikan dari cmd/seed-dummy: menghapus semua data dummy (ISSUE-*,
// HIST-*, INS-*, A0xx/K0xx/DK0xx/ASP0xx/DET0xx/UR0xx, dst — lihat
// seeds.CleanupDummyData untuk pola LIKE lengkapnya) sambil MEMPERTAHANKAN
// data "REAL" (prefix *-REAL-*/AREA-*/KWS-*/dst). Aman dijalankan berulang
// (semua query DELETE idempotent).
package main

import (
	"log"

	"github.com/monitoring-system/backend/config"
	"github.com/monitoring-system/backend/seeds"
)

func main() {
	cfg := config.Load()
	db, err := config.OpenDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	seeds.CleanupDummyData(db)
}
