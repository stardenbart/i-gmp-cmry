package main

import (
	"log"

	"github.com/monitoring-system/backend/config"
	"github.com/monitoring-system/backend/pkg/migrator"
)

func main() {
	cfg := config.Load()
	db, err := config.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect database: %v", err)
	}

	if err := migrator.RunMigrations(db, "./migrations"); err != nil {
		log.Fatalf("❌ Migration execution failed: %v", err)
	}

	log.Println("🎉 All migrations applied successfully.")
}
