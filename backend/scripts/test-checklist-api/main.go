package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/monitoring-system/backend/config"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/inspectionrepo"
)

func main() {
	db, err := config.OpenDatabase(config.Load())
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
