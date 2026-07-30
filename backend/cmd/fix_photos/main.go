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

	// Delete proof photos for issues that are currently Rejected
	err = db.Exec(`DELETE FROM "Issue_Photo" WHERE "PhotoType" IN ('WOWR', 'FollowUp') AND "IssueID" IN (SELECT "IssueID" FROM "Issue" WHERE "WOWRStatus" = 'Rejected')`).Error
	if err != nil {
		log.Fatalf("Failed to delete rejected photos: %v", err)
	}

	log.Println("Data has been successfully restored!")
}
