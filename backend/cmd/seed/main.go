package main

import (
	"log"

	"github.com/monitoring-system/backend/config"
	"github.com/monitoring-system/backend/seeds"
)

func main() {
	cfg := config.Load()
	db, err := config.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	seeds.Run(db)
}
