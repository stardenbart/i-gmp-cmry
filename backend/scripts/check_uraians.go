package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

type Uraian struct {
	UraianID   string `gorm:"column:UraianID;primaryKey"`
	UraianText string `gorm:"column:UraianText"`
}

func (Uraian) TableName() string { return "Uraian_Master" }

func main() {
	dsn := "host=127.0.0.1 user=postgres password=secret dbname=monitoring_audit port=5434 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
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
