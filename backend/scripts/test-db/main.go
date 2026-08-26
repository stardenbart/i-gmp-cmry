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

	type AuditeeStatusRow struct {
		AreaName   string `gorm:"column:name"`
		TotalCheck int64  `gorm:"column:total_check"`
		TotalOK    int64  `gorm:"column:total_ok"`
		OpenIssues int64  `gorm:"column:open_issues"`
	}
	var auditeeRows []AuditeeStatusRow
	err = db.Raw(`
		SELECT 
			am.AreaName as name,
			COUNT(DISTINCT ir.ResultID) as total_check,
			COUNT(DISTINCT CASE WHEN ir.Checking = 'OK' THEN ir.ResultID END) as total_ok,
			COUNT(DISTINCT CASE WHEN i.IssueStatus != 'Closed' AND i.IssueStatus != 'Verified' AND i.IssueStatus IS NOT NULL THEN i.IssueID END) as open_issues
		FROM Area_Master am
		LEFT JOIN Inspection_Header ih ON ih.AreaID = am.AreaID
		LEFT JOIN Inspection_Result ir ON ir.InspectionID = ih.InspectionID
		LEFT JOIN Issue i ON i.ResultID = ir.ResultID
		GROUP BY am.AreaID, am.AreaName
		ORDER BY open_issues DESC
		LIMIT 5
	`).Scan(&auditeeRows).Error

	if err != nil {
		fmt.Println("QUERY ERROR:", err)
	} else {
		fmt.Printf("ROWS RETURNED: %d\n", len(auditeeRows))
		for _, r := range auditeeRows {
			fmt.Printf("%+v\n", r)
		}
	}
}
