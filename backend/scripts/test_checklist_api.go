package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"

	"github.com/monitoring-system/backend/internal/infrastructure/persistence/inspectionrepo"
)

func main() {
	dsn := "host=127.0.0.1 user=postgres password=secret dbname=monitoring_audit port=5434 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("failed to connect database", err)
		os.Exit(1)
	}

	repo := inspectionrepo.NewInspectionHeaderRepository(db)
	
	// INSP-20260721-003 belongs to Area A002
	checklist, err := repo.GetFullChecklist("A002", "INSP-20260721-003")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	bytes, _ := json.MarshalIndent(checklist, "", "  ")
	fmt.Println(string(bytes))
}
