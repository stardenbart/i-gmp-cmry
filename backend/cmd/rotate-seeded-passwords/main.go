package main

import (
	"fmt"
	"log"

	"github.com/monitoring-system/backend/config"
	"github.com/monitoring-system/backend/seeds"
)

func main() {
	cfg := config.Load()
	// Password hashes are sensitive too; suppress SQL parameter logging.
	cfg.AppEnv = "production"
	db, err := config.OpenDatabase(cfg)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	updated, err := seeds.RotateDefaultUserPasswords(db)
	if err != nil {
		log.Fatalf("rotate seeded user passwords: %v", err)
	}
	fmt.Printf("rotated passwords for %d seeded accounts\n", updated)
}
