package main

import (
	"fmt"
	"os"

	"github.com/monitoring-system/backend/config"
)

func main() {
	db, err := config.OpenDatabase(config.Load())
	if err != nil {
		fmt.Println("failed to connect database", err)
		os.Exit(1)
	}

	type Result struct {
		ResultID     string
		InspectionID string
		Checking     string
	}
	var results []Result
	db.Raw(`SELECT "ResultID", "InspectionID", "Checking" FROM "Inspection_Result" ORDER BY "ResultID" DESC LIMIT 10`).Scan(&results)

	fmt.Println("Latest Inspection Results:")
	for _, r := range results {
		fmt.Printf("ResultID: %s, InspID: %s, Checking: %s\n", r.ResultID, r.InspectionID, r.Checking)

		var issueCount int64
		db.Table("Issue").Where("\"ResultID\" = ?", r.ResultID).Count(&issueCount)
		fmt.Printf("  -> Linked Issues: %d\n", issueCount)
	}
}
