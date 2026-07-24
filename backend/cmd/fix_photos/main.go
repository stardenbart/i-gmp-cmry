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

	// Update corrupted photo types
	err = db.Exec(`UPDATE "Issue_Photo" SET "PhotoType" = 'WOWR' WHERE "PhotoType" = 'undefined'`).Error
	if err != nil {
		log.Fatalf("Failed to update photos: %v", err)
	}

	// Update issue status that was left behind
	err = db.Exec(`UPDATE "Issue" SET "WOWRStatus" = 'PendingValidation' WHERE "IssueID" = 'ISSUE-002'`).Error
	if err != nil {
		log.Fatalf("Failed to update issue status: %v", err)
	}

	log.Println("Data has been successfully restored!")
}
