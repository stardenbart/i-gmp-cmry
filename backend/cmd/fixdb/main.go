package main

import (
	"github.com/monitoring-system/backend/config"
	"log"
)

func main() {
	cfg := config.Load()
	db, err := config.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Cek apakah kolom WOWRStatus ada, jika tidak, tambahkan
	err = db.Exec(`ALTER TABLE "Issue" ADD COLUMN IF NOT EXISTS "WOWRStatus" VARCHAR(50) DEFAULT 'None';`).Error
	if err != nil {
		log.Fatalf("Failed to alter table: %v", err)
	}

	log.Println("Successfully added WOWRStatus column!")
}
