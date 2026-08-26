package main

import (
	"fmt"
	"os"

	"github.com/monitoring-system/backend/config"
)

type Uraian struct {
	UraianID   string `gorm:"column:UraianID;primaryKey"`
	UraianText string `gorm:"column:UraianText"`
}

func (Uraian) TableName() string { return "Uraian_Master" }

func main() {
	db, err := config.OpenDatabase(config.Load())
	if err != nil {
		fmt.Println("failed to connect database", err)
		os.Exit(1)
	}

	var items []Uraian
	db.Find(&items)
	fmt.Printf("Total uraians: %d\n", len(items))
	for _, it := range items {
		fmt.Printf("ID: %s, Text: %s\n", it.UraianID, it.UraianText)
	}
}
