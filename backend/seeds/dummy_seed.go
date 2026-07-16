package seeds

import (
	"log"
	"time"

	inspectiondomain "github.com/monitoring-system/backend/internal/domain/inspection"
	issuedomain "github.com/monitoring-system/backend/internal/domain/issue"
	masterdomain "github.com/monitoring-system/backend/internal/domain/master"
	"gorm.io/gorm"
)

func SeedDummyData(db *gorm.DB) {
	log.Println("🌱 Seeding Dummy Data (Area, Inspection, Issue)...")

	// 1. Area
	area := masterdomain.Area{AreaID: "A01", AreaName: "Pabrik Utama"}
	db.Where("AreaID = ?", area.AreaID).FirstOrCreate(&area)

	// 2. Kawasan
	kawasan := masterdomain.Kawasan{KawasanID: "K01", AreaID: "A01", KawasanName: "Area Produksi A"}
	db.Where("KawasanID = ?", kawasan.KawasanID).FirstOrCreate(&kawasan)

	// 3. DetailKawasan
	dk := masterdomain.DetailKawasan{DetailKawasanID: "DK01", KawasanID: "K01", DetailKawasanName: "Line 1"}
	db.Where("DetailKawasanID = ?", dk.DetailKawasanID).FirstOrCreate(&dk)

	// 4. Aspek
	aspek := masterdomain.Aspek{AspekID: "ASP01", AreaID: "A01", AspekName: "Kebersihan 5R"}
	db.Where("AspekID = ?", aspek.AspekID).FirstOrCreate(&aspek)

	// 5. Detail (Checklist category)
	detail := masterdomain.Detail{DetailID: "DET01", AspekID: "ASP01", DetailName: "Lantai & Saluran Air"}
	db.Where("DetailID = ?", detail.DetailID).FirstOrCreate(&detail)

	// 6. Uraian (Checklist item)
	uraian := masterdomain.Uraian{UraianID: "UR01", DetailID: "DET01", UraianText: "Lantai bersih dari genangan air dan oli", StandardScore: 100}
	db.Where("UraianID = ?", uraian.UraianID).FirstOrCreate(&uraian)

	// 7. Inspection Header
	insp := inspectiondomain.InspectionHeader{
		InspectionID:    "INS-001",
		AreaID:          "A01",
		KawasanID:       "K01",
		DetailKawasanID: "DK01",
		InspectorID:     "USR-001",
		InspectionHeaderStatus: inspectiondomain.InspectionStatusCompleted,
	}
	db.Where("InspectionID = ?", insp.InspectionID).FirstOrCreate(&insp)

	// 8. Inspection Result
	res := inspectiondomain.InspectionResult{
		ResultID:     "RES-001",
		InspectionID: "INS-001",
		UraianID:     "UR01",
		Checking:     "NG",
		Nilai:        50,
		Keterangan:   "Terdapat genangan oli dekat mesin 2",
	}
	db.Where("ResultID = ?", res.ResultID).FirstOrCreate(&res)

	// 9. Issue (Since result was NG)
	now := time.Now()
	dueDate := now.AddDate(0, 0, 3)
	issue := issuedomain.Issue{
		IssueID:        "ISSUE-001",
		ResultID:       "RES-001",
		IssuePICUserID: "USR-001", 
		DueDate:        &dueDate,
		IssueStatus:    issuedomain.IssueStatusOpen,
		Keterangan:     "Segera bersihkan genangan oli dan periksa kebocoran mesin 2",
	}
	db.Where("IssueID = ?", issue.IssueID).FirstOrCreate(&issue)

	dueDate2 := now.AddDate(0, 0, -2)
	issue2 := issuedomain.Issue{
		IssueID:        "ISSUE-002",
		ResultID:       "RES-001",
		IssuePICUserID: "USR-001", 
		DueDate:        &dueDate2,
		IssueStatus:    issuedomain.IssueStatusClosed,
		Keterangan:     "Perbaikan seal mesin 1",
	}
	db.Where("IssueID = ?", issue2.IssueID).FirstOrCreate(&issue2)

	log.Println("   ✔ Dummy data seeded")
}
