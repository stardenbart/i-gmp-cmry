package seeds

import (
	"fmt"
	"log"
	"time"

	inspectiondomain "github.com/monitoring-system/backend/internal/domain/inspection"
	issuedomain "github.com/monitoring-system/backend/internal/domain/issue"
	masterdomain "github.com/monitoring-system/backend/internal/domain/master"
	"gorm.io/gorm"
)

func SeedDummyData(db *gorm.DB) {
	log.Println("🌱 Seeding 100 Dummy Data (Area, Inspection, Issue)...")

	now := time.Now()

	plants := []struct {
		PlantID   string
		PlantName string
		AdminID   string
		Prefix    string
		Offset    int
	}{
		{PlantID: "PLT-SENTUL", PlantName: "Sentul", AdminID: "USR-ADMIN-SENTUL", Prefix: "SNT", Offset: 0},
		{PlantID: "PLT-CICURUG", PlantName: "Cicurug", AdminID: "USR-ADMIN-CICURUG", Prefix: "CCG", Offset: 20},
		{PlantID: "PLT-PASURUAN", PlantName: "Pasuruan", AdminID: "USR-ADMIN-PASURUAN", Prefix: "PSR", Offset: 40},
	}

	for _, p := range plants {
		for i := 1; i <= 10; i++ {
			idx := p.Offset + i
			// 1. Area
			areaID := fmt.Sprintf("A%03d", idx)
			pID := p.PlantID
			area := masterdomain.Area{AreaID: areaID, PlantID: &pID, AreaName: fmt.Sprintf("Area Pabrik %s %d", p.PlantName, i)}
			db.Where(&masterdomain.Area{AreaID: area.AreaID}).FirstOrCreate(&area)

			// 2. Kawasan
			kawasanID := fmt.Sprintf("K%03d", idx)
			kawasan := masterdomain.Kawasan{KawasanID: kawasanID, AreaID: areaID, KawasanName: fmt.Sprintf("Area Produksi %s %d", p.PlantName, i)}
			db.Where(&masterdomain.Kawasan{KawasanID: kawasan.KawasanID}).FirstOrCreate(&kawasan)

			// 3. DetailKawasan
			dkID := fmt.Sprintf("DK%03d", idx)
			dk := masterdomain.DetailKawasan{DetailKawasanID: dkID, KawasanID: kawasanID, DetailKawasanName: fmt.Sprintf("Line %s %d", p.PlantName, i)}
			db.Where(&masterdomain.DetailKawasan{DetailKawasanID: dk.DetailKawasanID}).FirstOrCreate(&dk)

			// 4. Aspek
			aspekID := fmt.Sprintf("ASP%03d", idx)
			aspek := masterdomain.Aspek{AspekID: aspekID, AreaID: areaID, AspekName: fmt.Sprintf("Kebersihan %s %dR", p.PlantName, i)}
			db.Where(&masterdomain.Aspek{AspekID: aspek.AspekID}).FirstOrCreate(&aspek)

			// 5. Detail (Checklist category)
			detailID := fmt.Sprintf("DET%03d", idx)
			detail := masterdomain.Detail{DetailID: detailID, AspekID: aspekID, DetailName: fmt.Sprintf("Lantai & Saluran Air %s %d", p.PlantName, i)}
			db.Where(&masterdomain.Detail{DetailID: detail.DetailID}).FirstOrCreate(&detail)

			// 6. Uraian (Checklist item)
			uraianID := fmt.Sprintf("UR%03d", idx)
			uraian := masterdomain.Uraian{UraianID: uraianID, DetailID: detailID, UraianText: fmt.Sprintf("Lantai bersih dari genangan air %s %d", p.PlantName, i), StandardScore: 100}
			db.Where(&masterdomain.Uraian{UraianID: uraian.UraianID}).FirstOrCreate(&uraian)

			// 7. Inspection Header
			inspID := fmt.Sprintf("INS-%03d", idx)
			insp := inspectiondomain.InspectionHeader{
				InspectionID:           inspID,
				AreaID:                 areaID,
				KawasanID:              kawasanID,
				DetailKawasanID:        dkID,
				InspectorID:            p.AdminID,
				InspectionHeaderStatus: inspectiondomain.InspectionStatusCompleted,
			}
			db.Where(&inspectiondomain.InspectionHeader{InspectionID: insp.InspectionID}).FirstOrCreate(&insp)

			// 8. Inspection Result
			resID := fmt.Sprintf("RES-%03d", idx)
			nilai := 50
			checking := "NG"
			if i%2 == 0 {
				nilai = 100
				checking = "OK"
			}

			res := inspectiondomain.InspectionResult{
				ResultID:     resID,
				InspectionID: inspID,
				UraianID:     uraianID,
				Checking:     checking,
				Nilai:        nilai,
				Keterangan:   fmt.Sprintf("Keterangan inspeksi %s ke-%d", p.PlantName, i),
			}
			db.Where(&inspectiondomain.InspectionResult{ResultID: res.ResultID}).FirstOrCreate(&res)

			// 9. Issue (Only if NG)
			if checking == "NG" {
				eqID := fmt.Sprintf("EQ%03d", idx)
				eq := issuedomain.Equipment{
					EquipmentID:     eqID,
					KawasanID:       kawasanID,
					EquipmentCode:   fmt.Sprintf("EQ-CODE-%03d", idx),
					EquipmentName:   fmt.Sprintf("Mesin Pompa %s %d", p.PlantName, i),
					EquipmentStatus: "Active",
				}
				db.Where(&issuedomain.Equipment{EquipmentID: eq.EquipmentID}).FirstOrCreate(&eq)

				issueID := fmt.Sprintf("ISSUE-%03d", idx)
				dueDate := now.AddDate(0, 0, 3)
				cat := issuedomain.IssueCategoryEquipment
				issue := issuedomain.Issue{
					IssueID:        issueID,
					ResultID:       resID,
					IssuePICUserID: p.AdminID,
					DueDate:        &dueDate,
					IssueStatus:    issuedomain.IssueStatusOpen,
					IssueCategory:  &cat,
					EquipmentID:    &eqID,
					Keterangan:     fmt.Sprintf("Segera perbaiki masalah %s ke-%d", p.PlantName, i),
				}
				db.Where(&issuedomain.Issue{IssueID: issue.IssueID}).FirstOrCreate(&issue)
			}
		}
	}

	log.Println("   ✔ 100 Dummy data seeded")
}
