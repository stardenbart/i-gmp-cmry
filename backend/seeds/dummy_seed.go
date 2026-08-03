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

	for i := 1; i <= 20; i++ {
		// 1. Area (Assigned to PLT-SENTUL)
		areaID := fmt.Sprintf("A%03d", i)
		plantID := "PLT-SENTUL"
		area := masterdomain.Area{AreaID: areaID, PlantID: &plantID, AreaName: fmt.Sprintf("Area Pabrik Sentul %d", i)}
		db.Where(&masterdomain.Area{AreaID: area.AreaID}).FirstOrCreate(&area)

		// 2. Kawasan
		kawasanID := fmt.Sprintf("K%03d", i)
		kawasan := masterdomain.Kawasan{KawasanID: kawasanID, AreaID: areaID, KawasanName: fmt.Sprintf("Area Produksi %d", i)}
		db.Where(&masterdomain.Kawasan{KawasanID: kawasan.KawasanID}).FirstOrCreate(&kawasan)

		// 3. DetailKawasan
		dkID := fmt.Sprintf("DK%03d", i)
		dk := masterdomain.DetailKawasan{DetailKawasanID: dkID, KawasanID: kawasanID, DetailKawasanName: fmt.Sprintf("Line %d", i)}
		db.Where(&masterdomain.DetailKawasan{DetailKawasanID: dk.DetailKawasanID}).FirstOrCreate(&dk)

		// 4. Aspek
		aspekID := fmt.Sprintf("ASP%03d", i)
		aspek := masterdomain.Aspek{AspekID: aspekID, AreaID: areaID, AspekName: fmt.Sprintf("Kebersihan %dR", i)}
		db.Where(&masterdomain.Aspek{AspekID: aspek.AspekID}).FirstOrCreate(&aspek)

		// 5. Detail (Checklist category)
		detailID := fmt.Sprintf("DET%03d", i)
		detail := masterdomain.Detail{DetailID: detailID, AspekID: aspekID, DetailName: fmt.Sprintf("Lantai & Saluran Air %d", i)}
		db.Where(&masterdomain.Detail{DetailID: detail.DetailID}).FirstOrCreate(&detail)

		// 6. Uraian (Checklist item)
		uraianID := fmt.Sprintf("UR%03d", i)
		uraian := masterdomain.Uraian{UraianID: uraianID, DetailID: detailID, UraianText: fmt.Sprintf("Lantai bersih dari genangan air %d", i), StandardScore: 100}
		db.Where(&masterdomain.Uraian{UraianID: uraian.UraianID}).FirstOrCreate(&uraian)

		// 7. Inspection Header
		inspID := fmt.Sprintf("INS-%03d", i)
		insp := inspectiondomain.InspectionHeader{
			InspectionID:           inspID,
			AreaID:                 areaID,
			KawasanID:              kawasanID,
			DetailKawasanID:        dkID,
			InspectorID:            "USR-ADMIN-001",
			InspectionHeaderStatus: inspectiondomain.InspectionStatusCompleted,
		}
		db.Where(&inspectiondomain.InspectionHeader{InspectionID: insp.InspectionID}).FirstOrCreate(&insp)

		// 8. Inspection Result
		resID := fmt.Sprintf("RES-%03d", i)
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
			Keterangan:   fmt.Sprintf("Keterangan inspeksi ke-%d", i),
		}
		db.Where(&inspectiondomain.InspectionResult{ResultID: res.ResultID}).FirstOrCreate(&res)

		// 9. Issue (Only if NG)
		if checking == "NG" {
			issueID := fmt.Sprintf("ISSUE-%03d", i)
			dueDate := now.AddDate(0, 0, 3)
			cat := issuedomain.IssueCategoryEquipment
			issue := issuedomain.Issue{
				IssueID:        issueID,
				ResultID:       resID,
				IssuePICUserID: "USR-ADMIN-001",
				DueDate:        &dueDate,
				IssueStatus:    issuedomain.IssueStatusOpen,
				IssueCategory:  &cat,
				Keterangan:     fmt.Sprintf("Segera perbaiki masalah ke-%d", i),
			}
			db.Where(&issuedomain.Issue{IssueID: issue.IssueID}).FirstOrCreate(&issue)
		}
	}

	log.Println("   ✔ 100 Dummy data seeded")
}
