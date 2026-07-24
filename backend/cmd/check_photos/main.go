package main

import (
	"encoding/json"
	"fmt"
	"github.com/monitoring-system/backend/config"
	"github.com/monitoring-system/backend/internal/domain/issue"
	"log"
)

func main() {
	cfg := config.Load()
	db, err := config.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	var photos []issue.IssuePhoto
	if err := db.Find(&photos).Error; err != nil {
		log.Fatalf("Failed to query photos: %v", err)
	}

	b, _ := json.MarshalIndent(photos, "", "  ")
	fmt.Println(string(b))
}
