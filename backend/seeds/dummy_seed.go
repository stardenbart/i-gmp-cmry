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
	log.Println("🌱 Seeding 100+ Dummy Data & Multi-Month Historical Inspections...")

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

			// 7. Current Month Inspection Header
			inspID := fmt.Sprintf("INS-%03d", idx)
			insp := inspectiondomain.InspectionHeader{
				InspectionID:           inspID,
				AreaID:                 areaID,
				KawasanID:              kawasanID,
				DetailKawasanID:        dkID,
				InspectorID:            p.AdminID,
				InspectionHeaderStatus: inspectiondomain.InspectionStatusCompleted,
				InspectionHeaderCreatedAt: now,
			}
			db.Where(&inspectiondomain.InspectionHeader{InspectionID: insp.InspectionID}).FirstOrCreate(&insp)

			// 8. Current Month Inspection Result
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
				issue := issuedomain.Issue{
					IssueID:        issueID,
					ResultID:       resID,
					IssuePICUserID: p.AdminID,
					DueDate:        &dueDate,
					IssueStatus:    issuedomain.IssueStatusOpen,
					Keterangan:     fmt.Sprintf("Segera perbaiki masalah %s ke-%d", p.PlantName, i),
				}
				db.Where(&issuedomain.Issue{IssueID: issue.IssueID}).FirstOrCreate(&issue)

				hei := issuedomain.IssueHEI{
					IssueHEIID:  fmt.Sprintf("HEI-%03d", idx),
					IssueID:     issueID,
					EquipmentID: &eqID,
				}
				db.Where(&issuedomain.IssueHEI{IssueID: hei.IssueID}).FirstOrCreate(&hei)
			}
		}

		// 10. Multi-Month Historical Trend Seeding (Past 5 months with varying compliance rates)
		// Monthly target rates for past months to create natural up/down fluctuations
		monthlyRates := []struct {
			MonthsAgo int
			OKRatio   float64 // e.g. 0.70 = 70% OK
		}{
			{MonthsAgo: 5, OKRatio: 0.72}, // 5 months ago (e.g. March)
			{MonthsAgo: 4, OKRatio: 0.85}, // 4 months ago (e.g. April)
			{MonthsAgo: 3, OKRatio: 0.68}, // 3 months ago (e.g. May)
			{MonthsAgo: 2, OKRatio: 0.91}, // 2 months ago (e.g. June)
			{MonthsAgo: 1, OKRatio: 0.82}, // 1 month ago (e.g. July)
		}

		for _, mr := range monthlyRates {
			histDate := now.AddDate(0, -mr.MonthsAgo, 0)
			for j := 1; j <= 10; j++ {
				idx := p.Offset + j
				areaID := fmt.Sprintf("A%03d", idx)
				kawasanID := fmt.Sprintf("K%03d", idx)
				dkID := fmt.Sprintf("DK%03d", idx)
				uraianID := fmt.Sprintf("UR%03d", idx)

				histInspID := fmt.Sprintf("HIST-%s-M%d-%02d", p.Prefix, mr.MonthsAgo, j)
				histInsp := inspectiondomain.InspectionHeader{
					InspectionID:              histInspID,
					AreaID:                    areaID,
					KawasanID:                 kawasanID,
					DetailKawasanID:           dkID,
					InspectorID:               p.AdminID,
					InspectionHeaderStatus:    inspectiondomain.InspectionStatusCompleted,
					InspectionHeaderCreatedAt: histDate,
				}
				db.Where(&inspectiondomain.InspectionHeader{InspectionID: histInsp.InspectionID}).FirstOrCreate(&histInsp)

				checking := "NG"
				nilai := 40
				if float64(j)/10.0 <= mr.OKRatio {
					checking = "OK"
					nilai = 100
				}

				histResID := fmt.Sprintf("HIST-RES-%s-M%d-%02d", p.Prefix, mr.MonthsAgo, j)
				histRes := inspectiondomain.InspectionResult{
					ResultID:     histResID,
					InspectionID: histInspID,
					UraianID:     uraianID,
					Checking:     checking,
					Nilai:        nilai,
					Keterangan:   fmt.Sprintf("Inspeksi historis %s M-%d ke-%d", p.PlantName, mr.MonthsAgo, j),
				}
				db.Where(&inspectiondomain.InspectionResult{ResultID: histRes.ResultID}).FirstOrCreate(&histRes)
			}
		}
	}

	log.Println("   ✔ Multi-month historical inspection trend data seeded")
}

// CleanupDummyData removes all dummy inspection, issue, area, and checklist seed records from database.
func CleanupDummyData(db *gorm.DB) {
	log.Println("🧹 Cleaning up dummy seed data from database...")

	// 1. Delete Issue Photos
	db.Exec(`DELETE FROM "Issue_Photo" WHERE "IssueID" LIKE 'ISSUE-%' OR "IssueID" LIKE 'HIST-%' OR "IssueID" IN (SELECT "IssueID" FROM "Issue" WHERE "IssueID" LIKE 'ISSUE-%' OR "IssueID" LIKE 'HIST-%' OR "ResultID" LIKE 'RES-%' OR "ResultID" LIKE 'HIST-%')`)

	// 2. Delete Issue HEI
	db.Exec(`DELETE FROM "Issue_HEI" WHERE "IssueHEIID" LIKE 'HEI-%' OR "IssueID" LIKE 'ISSUE-%' OR "IssueID" LIKE 'HIST-%' OR "IssueID" IN (SELECT "IssueID" FROM "Issue" WHERE "IssueID" LIKE 'ISSUE-%' OR "IssueID" LIKE 'HIST-%' OR "ResultID" LIKE 'RES-%' OR "ResultID" LIKE 'HIST-%')`)

	// 3. Delete Issue Delegates
	db.Exec(`DELETE FROM "Issue_Delegate" WHERE "IssueID" LIKE 'ISSUE-%' OR "IssueID" LIKE 'HIST-%' OR "IssueID" IN (SELECT "IssueID" FROM "Issue" WHERE "IssueID" LIKE 'ISSUE-%' OR "IssueID" LIKE 'HIST-%' OR "ResultID" LIKE 'RES-%' OR "ResultID" LIKE 'HIST-%')`)

	// 4. Delete Issues
	db.Exec(`DELETE FROM "Issue" WHERE "IssueID" LIKE 'ISSUE-%' OR "IssueID" LIKE 'HIST-%' OR "ResultID" LIKE 'RES-%' OR "ResultID" LIKE 'HIST-%'`)

	// 5. Delete Inspection Results
	db.Exec(`DELETE FROM "Inspection_Result" WHERE "ResultID" LIKE 'RES-%' OR "ResultID" LIKE 'HIST-%' OR "InspectionID" LIKE 'INS-%' OR "InspectionID" LIKE 'HIST-%'`)

	// 6. Delete Inspection Headers
	db.Exec(`DELETE FROM "Inspection_Header" WHERE "InspectionID" LIKE 'INS-%' OR "InspectionID" LIKE 'HIST-%'`)

	// 7. Delete Equipment Master dummy records
	db.Exec(`DELETE FROM "Equipment_Master" WHERE "EquipmentID" LIKE 'EQ%' AND ("EquipmentCode" LIKE 'EQ-CODE-%' OR "KawasanID" LIKE 'K%')`)

	// 8. Delete Checklist Uraian
	db.Exec(`DELETE FROM "Uraian_Master" WHERE "UraianID" LIKE 'UR%' OR "UraianText" LIKE 'Lantai bersih dari genangan air%'`)

	// 9. Delete Checklist Detail
	db.Exec(`DELETE FROM "Detail_Master" WHERE "DetailID" LIKE 'DET%' OR "DetailName" LIKE 'Lantai & Saluran Air%'`)

	// 10. Delete Checklist Aspek
	db.Exec(`DELETE FROM "Aspek_Master" WHERE "AspekID" LIKE 'ASP%' OR "AspekName" LIKE 'Kebersihan%'`)

	// 11. Delete DetailKawasan
	db.Exec(`DELETE FROM "DetailKawasan_Master" WHERE "DetailKawasanID" LIKE 'DK%' OR "DetailKawasanName" LIKE 'Line%'`)

	// 12. Delete Kawasan
	db.Exec(`DELETE FROM "Kawasan_Master" WHERE "KawasanID" LIKE 'K%' OR "KawasanName" LIKE 'Area Produksi%'`)

	// 13. Delete Area
	db.Exec(`DELETE FROM "Area_Master" WHERE "AreaID" LIKE 'A%' AND "AreaName" LIKE 'Area Pabrik%'`)

	// 14. Recalculate and sync LastInspection for remaining DetailKawasan and Kawasan
	db.Exec(`
		UPDATE "DetailKawasan_Master" 
		SET "LastInspection" = (
			SELECT MAX("InspectionHeaderCreatedAt") 
			FROM "Inspection_Header" 
			WHERE "Inspection_Header"."DetailKawasanID" = "DetailKawasan_Master"."DetailKawasanID" 
			  AND "InspectionHeaderStatus" IN ('Completed', 'Approved')
		)
	`)
	db.Exec(`
		UPDATE "Kawasan_Master" 
		SET "LastInspection" = (
			SELECT MAX("InspectionHeaderCreatedAt") 
			FROM "Inspection_Header" 
			WHERE "Inspection_Header"."KawasanID" = "Kawasan_Master"."KawasanID" 
			  AND "InspectionHeaderStatus" IN ('Completed', 'Approved')
		)
	`)

	log.Println("✅ All dummy seed records successfully cleaned up!")
}
