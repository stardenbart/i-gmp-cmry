package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

type InspectionHeader struct {
	InspectionID string `gorm:"column:InspectionID;primaryKey"`
	AreaID       string `gorm:"column:AreaID"`
}

func (InspectionHeader) TableName() string { return "Inspection_Header" }

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

	var header InspectionHeader
	err = db.Where("\"InspectionID\" = ?", "INSP-20260721-003").First(&header).Error
	if err != nil {
		fmt.Println("Header not found:", err)
		os.Exit(1)
	}

	fmt.Printf("Header found: %+v\n", header)

	// Fetch all results
	var results []InspectionResult
	db.Where("\"InspectionID\" = ?", "INSP-20260721-003").Find(&results)
	fmt.Printf("Found %d results in DB for INSP-20260721-003\n", len(results))
	for _, r := range results {
		fmt.Printf("  UraianID: %s, Checking: %s\n", r.UraianID, r.Checking)
	}
}
