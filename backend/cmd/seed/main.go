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

	if err := seeds.Run(db); err != nil {
		log.Fatalf("❌ Failed to seed database: %v", err)
	}
}
