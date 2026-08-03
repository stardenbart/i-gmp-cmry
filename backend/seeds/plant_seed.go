package seeds

import (
	"log"

	masterdomain "github.com/monitoring-system/backend/internal/domain/master"
	"gorm.io/gorm"
)

var defaultPlants = []masterdomain.Plant{
	{PlantID: "PLT-SENTUL", PlantCode: "PLT-001", PlantName: "Plant Sentul Utama", Address: "Jl. Industri Sentul No. 1, Bogor"},
	{PlantID: "PLT-CICURUG", PlantCode: "PLT-002", PlantName: "Plant Cicurug", Address: "Jl. Raya Sukabumi No. 45, Sukabumi"},
	{PlantID: "PLT-PASURUAN", PlantCode: "PLT-003", PlantName: "Plant Pasuruan", Address: "Kawasan Industri PIER, Pasuruan"},
}

func SeedPlants(db *gorm.DB) {
	for _, p := range defaultPlants {
		var count int64
		db.Model(&masterdomain.Plant{}).Where(&masterdomain.Plant{PlantID: p.PlantID}).Count(&count)
		if count == 0 {
			if err := db.Create(&p).Error; err != nil {
				log.Printf("❌ Failed to seed plant %s: %v", p.PlantName, err)
			} else {
				log.Printf("   ✔ Plant seeded: %s (%s)", p.PlantName, p.PlantID)
			}
		}
	}
}
