package dashboardhandler

import (
	"testing"
	"time"
)

func sp(s string) *string { return &s }

func gmpFormSourceFixture() []gmpFormSourceRow {
	loc := dashboardLocation()
	day := time.Date(2026, 10, 9, 8, 27, 0, 0, loc)
	due := time.Date(2026, 10, 23, 0, 0, 0, 0, loc)
	base := gmpFormSourceRow{InspectionID: "INSP-1", Tanggal: day, Area: "Area Produksi", KawasanID: "KWS-010", Kawasan: "Produksi CMD 4", DetailKawasan: "Proses"}
	row := func(uraianID, aspek, detail, text, checking string, nilai int) gmpFormSourceRow {
		r := base
		r.UraianID, r.Aspek, r.Detail, r.Uraian, r.Checking, r.Nilai = uraianID, aspek, detail, text, checking, nilai
		return r
	}
	lainnya := row("CMDXY050", "Lainnya", "Lainnya", "Ketidaksesuaian lain", "NG", 0)
	lainnya.IssueID, lainnya.IssueStatus, lainnya.DueDate, lainnya.Keterangan = sp("ISS-B"), "InProgress", &due, "fallback"
	lantai := row("CMDXY015", "Konstruksi", "Lantai", "Bersih dan kering", "NG", 0)
	lantai.IssueID, lantai.IssueStatus, lantai.DueDate = sp("ISS-A"), "Open", &due

	other := base
	other.InspectionID, other.DetailKawasan, other.Tanggal = "INSP-2", "Filling", day.Add(time.Hour)
	other.UraianID, other.Aspek, other.Detail, other.Uraian, other.Checking, other.Nilai = "CMDXY001", "Lingkungan", "Lokasi", "Bebas banjir", "OK", 2

	// Deliberately unsorted: the form orders uraian by UraianID.
	return []gmpFormSourceRow{
		other,
		lainnya,
		row("CMDXY003", "Lingkungan", "Sarana Jalan", "Saluran air", "NA", 0),
		lantai,
		row("CMDXY001", "Lingkungan", "Lokasi", "Bebas banjir", "OK", 2),
	}
}

func gmpFormEvidenceFixture() (map[string][]gmpInitialEvidence, map[string][]gmpFollowUpEvidence) {
	fu := time.Date(2026, 10, 10, 9, 0, 0, 0, dashboardLocation())
	initial := map[string][]gmpInitialEvidence{
		"ISS-A": {{IssueID: "ISS-A", PhotoID: "P-A1", ImageURL: "/monitoring-audit-bucket/issues/ISS-A/1.jpg", Keterangan: "lantai basah"}},
		"ISS-B": {
			{IssueID: "ISS-B", PhotoID: "P-B1", ImageURL: "/monitoring-audit-bucket/issues/ISS-B/1.jpg", Keterangan: "TESTING #2"},
			{IssueID: "ISS-B", PhotoID: "P-B2", ImageURL: "/monitoring-audit-bucket/issues/ISS-B/2.jpg", Keterangan: ""},
			{IssueID: "ISS-B", PhotoID: "P-B3", ImageURL: "/monitoring-audit-bucket/issues/ISS-B/3.jpg", Keterangan: "TESTING #4"},
		},
	}
	followUps := map[string][]gmpFollowUpEvidence{
		"ISS-B": {{IssueID: "ISS-B", PhotoID: "F-1", RefPhotoID: "P-B3", Keterangan: "sudah dirapikan", FollowUpDate: &fu}},
	}
	return initial, followUps
}

func TestBuildGMPFormSheetsGroupsInspectionsAndOrdersUraian(t *testing.T) {
	initial, followUps := gmpFormEvidenceFixture()
	sheets := buildGMPFormSheets(gmpFormSourceFixture(), initial, followUps,
		map[string]string{"KWS-010": "Abdullah Farauk"}, "PT CISARUA MOUNTAIN DAIRY TBK\nPLANT SENTUL", "http://172.20.240.61:3000/")

	if len(sheets) != 2 {
		t.Fatalf("sheets = %d, want one per inspection (2)", len(sheets))
	}
	s := sheets[0]
	if s.Location != "Proses - Produksi CMD 4" || s.Title != "INSPEKSI GMP (BASIC) - Area Produksi" || s.PIC != "Abdullah Farauk" {
		t.Fatalf("header = %q / %q / %q", s.Location, s.Title, s.PIC)
	}
	if s.Date.Day() != 9 {
		t.Fatalf("date = %v, want 9 Oct (WIB)", s.Date)
	}
	var ids []string
	for _, u := range s.Uraian {
		ids = append(ids, u.UraianID)
	}
	if got := ids; len(got) != 4 || got[0] != "CMDXY001" || got[1] != "CMDXY003" || got[2] != "CMDXY015" || got[3] != "CMDXY050" {
		t.Fatalf("uraian order = %v", got)
	}
	if s.Uraian[1].Nilai != nil {
		t.Fatalf("NA uraian must have no Nilai, got %d", *s.Uraian[1].Nilai)
	}
	if s.Uraian[0].Nilai == nil || *s.Uraian[0].Nilai != 2 {
		t.Fatal("OK uraian keeps its Nilai")
	}
	if sheets[1].Location != "Filling - Produksi CMD 4" {
		t.Fatalf("second sheet = %q", sheets[1].Location)
	}
}

func TestBuildGMPFormSheetsSplitsFindingsPerPhoto(t *testing.T) {
	initial, followUps := gmpFormEvidenceFixture()
	sheets := buildGMPFormSheets(gmpFormSourceFixture(), initial, followUps, nil, "PT", "http://172.20.240.61:3000")
	lainnya := sheets[0].Uraian[3]
	if len(lainnya.Findings) != 3 {
		t.Fatalf("findings = %d, want 3 (one per initial photo)", len(lainnya.Findings))
	}
	first := lainnya.Findings[0]
	if first.PhotoURL != "http://172.20.240.61:3000/monitoring-audit-bucket/issues/ISS-B/1.jpg" {
		t.Fatalf("photo url = %q", first.PhotoURL)
	}
	if first.Keterangan != "TESTING #2" || first.Status != "In Progress" || first.DueDate == nil {
		t.Fatalf("first finding = %+v", first)
	}
	if lainnya.Findings[1].Keterangan != "fallback" {
		t.Fatalf("a photo without its own description falls back to the issue keterangan, got %q", lainnya.Findings[1].Keterangan)
	}
	if lainnya.Findings[2].FollowUp != "10 Oct 2026 - sudah dirapikan" || lainnya.Findings[0].FollowUp != "" {
		t.Fatalf("follow-up must sit on the photo it refers to: %q / %q", lainnya.Findings[0].FollowUp, lainnya.Findings[2].FollowUp)
	}
	if sheets[0].Uraian[2].Findings[0].Status != "Open" {
		t.Fatalf("status = %q", sheets[0].Uraian[2].Findings[0].Status)
	}
}

func TestGMPFormStatusLabel(t *testing.T) {
	for status, want := range map[string]string{
		"Open": "Open", "": "Open", "InProgress": "In Progress", "PendingValidation": "Menunggu Validasi",
		"Verified": "Closed", "Closed": "Closed", "ClosedOverdue": "Closed (Overdue)", "OpenOverdue": "Open (Overdue)",
	} {
		if got := gmpFormStatusLabel(status); got != want {
			t.Errorf("gmpFormStatusLabel(%q) = %q, want %q", status, got, want)
		}
	}
}

// Production UraianIDs are random (URN-261005-<hash>), so the form must
// follow the checklist order instead: Aspek, then Detail (master IDs), then
// the order the uraian were created in.
func TestBuildGMPFormSheetsFollowsChecklistOrderNotRandomUraianIDs(t *testing.T) {
	loc := dashboardLocation()
	day := time.Date(2026, 10, 9, 8, 0, 0, 0, loc)
	created := time.Date(2026, 10, 5, 3, 7, 53, 0, loc)
	mk := func(uraianID, aspekID, aspek, detailID, detail string, createdOffset int) gmpFormSourceRow {
		return gmpFormSourceRow{
			InspectionID: "INSP-1", Tanggal: day, Area: "Area Produksi", KawasanID: "KWS-1", Kawasan: "CMD 1", DetailKawasan: "Proses",
			AspekID: aspekID, Aspek: aspek, DetailID: detailID, Detail: detail,
			UraianID: uraianID, Uraian: uraianID, UraianCreatedAt: created.Add(time.Duration(createdOffset) * time.Millisecond),
			Checking: "OK", Nilai: 2,
		}
	}
	rows := []gmpFormSourceRow{
		mk("URN-261005-fc1825bd", "ASP-001", "Lingkungan Sarana Produksi", "DET-002", "Sarana Jalan", 3),
		mk("URN-261005-2282384b", "ASP-002", "Konstruksi dan Layout Bangunan", "DET-005", "Dinding", 6),
		mk("URN-261005-18585069", "ASP-001", "Lingkungan Sarana Produksi", "DET-002", "Sarana Jalan", 1),
		mk("URN-261005-0d706332", "ASP-001", "Lingkungan Sarana Produksi", "DET-003", "Lingkungan", 4),
		mk("URN-261005-4169c6cc", "ASP-002", "Konstruksi dan Layout Bangunan", "DET-005", "Dinding", 5),
		mk("URN-261005-3e16eda1", "ASP-001", "Lingkungan Sarana Produksi", "DET-002", "Sarana Jalan", 2),
	}
	sheets := buildGMPFormSheets(rows, nil, nil, nil, "PT", "")
	var got []string
	for _, u := range sheets[0].Uraian {
		got = append(got, u.UraianID)
	}
	want := []string{
		"URN-261005-18585069", "URN-261005-3e16eda1", "URN-261005-fc1825bd", // DET-002 by creation
		"URN-261005-0d706332",                        // DET-003
		"URN-261005-4169c6cc", "URN-261005-2282384b", // ASP-002 / DET-005 by creation
	}
	if len(got) != len(want) {
		t.Fatalf("order = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}
