package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

type InspectionResult struct {
	ResultID     string `gorm:"column:ResultID;primaryKey"`
	InspectionID string `gorm:"column:InspectionID"`
	UraianID     string `gorm:"column:UraianID"`
	Checking     string `gorm:"column:Checking"`
	Nilai        int    `gorm:"column:Nilai"`
	Keterangan   string `gorm:"column:Keterangan"`
}

func (InspectionResult) TableName() string { return "Inspection_Result" }

func main() {
	dsn := "host=127.0.0.1 user=postgres password=secret dbname=monitoring_audit port=5434 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("failed to connect database", err)
		os.Exit(1)
	}

	var results []InspectionResult
	err = db.Find(&results).Error
	if err != nil {
		fmt.Println("Query error:", err)
		os.Exit(1)
	}

	fmt.Printf("Total results: %d\n", len(results))
	bytes, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(bytes))
}
