package inspectionusecase

import (
	"strings"
	"testing"
	"time"

	"github.com/monitoring-system/backend/internal/domain/inspection"
)

func strp(s string) *string { return &s }

func reportRows() []inspection.KawasanReportRow {
	past := time.Date(2026, 10, 1, 0, 0, 0, 0, wib())
	future := time.Date(2026, 10, 22, 0, 0, 0, 0, wib())
	fu := time.Date(2026, 10, 7, 9, 0, 0, 0, wib())
	return []inspection.KawasanReportRow{
		{AreaName: "Area Produksi", KawasanName: "Produksi CMD 1", DetailKawasanID: "DK-1", DetailKawasanName: "Proses", Nilai: 2, StandardScore: 2, UraianText: "Lantai bersih"},
		{AreaName: "Area Produksi", KawasanName: "Produksi CMD 1", DetailKawasanID: "DK-1", DetailKawasanName: "Proses", Nilai: 0, StandardScore: 2, UraianText: "Tidak ada genangan",
			IssueID: strp("ISS-1"), Keterangan: "enc:Genangan dekat palet", IssueStatus: "Open", DueDate: &past},
		{AreaName: "Area Produksi", KawasanName: "Produksi CMD 1", DetailKawasanID: "DK-2", DetailKawasanName: "Gudang", Nilai: 0, StandardScore: 2, UraianText: "Label lengkap",
			IssueID: strp("ISS-2"), Keterangan: "enc:Label hilang", IssueStatus: "InProgress", DueDate: &future,
			FollowUpBy: strp("Sari"), FollowUpDate: &fu},
		{AreaName: "Area Produksi", KawasanName: "Produksi CMD 1", DetailKawasanID: "DK-2", DetailKawasanName: "Gudang", Nilai: 2, StandardScore: 2, UraianText: "Rak rapi"},
	}
}

func decryptStub(s string) string { return strings.TrimPrefix(s, "enc:") }

func TestBuildKawasanReportAggregatesScoresPerDetailKawasan(t *testing.T) {
	now := time.Date(2026, 10, 8, 14, 0, 0, 0, wib())
	r := buildKawasanReport(reportRows(), decryptStub, now, "http://172.20.240.61:3000/", "PLT-SENTUL", "")

	if r.AreaName != "Area Produksi" || r.KawasanName != "Produksi CMD 1" {
		t.Fatalf("names = %q / %q", r.AreaName, r.KawasanName)
	}
	want := []DetailScore{{Name: "Proses", Score: 2, Max: 4}, {Name: "Gudang", Score: 2, Max: 4}}
	if len(r.Details) != 2 || r.Details[0] != want[0] || r.Details[1] != want[1] {
		t.Fatalf("details = %+v, want %+v", r.Details, want)
	}
	if !r.GeneratedAt.Equal(now) {
		t.Fatalf("generated = %v", r.GeneratedAt)
	}
}

func TestBuildKawasanReportListsIssuesWithLinks(t *testing.T) {
	now := time.Date(2026, 10, 8, 14, 0, 0, 0, wib())
	r := buildKawasanReport(reportRows(), decryptStub, now, "http://172.20.240.61:3000/", "PLT-SENTUL", "")

	if len(r.Issues) != 2 {
		t.Fatalf("issues = %d, want 2", len(r.Issues))
	}
	first := r.Issues[0]
	if first.Title != "Tidak ada genangan" || first.Keterangan != "Genangan dekat palet" || first.DetailKawasanName != "Proses" || first.PICName != "" {
		t.Fatalf("first issue = %+v", first)
	}
	// Past due and still open → shown as overdue.
	if first.Status != "OpenOverdue" {
		t.Fatalf("status = %q, want OpenOverdue", first.Status)
	}
	if first.URL != "http://172.20.240.61:3000/cimory/PLT-SENTUL/dashboard/me/issues/ISS-1" {
		t.Fatalf("url = %q", first.URL)
	}
	second := r.Issues[1]
	if second.Status != "InProgress" || second.FollowUpBy != "Sari" || second.FollowUpDate == nil {
		t.Fatalf("second issue = %+v", second)
	}
	if r.AppURL != "http://172.20.240.61:3000/cimory/PLT-SENTUL/dashboard/me/issues" {
		t.Fatalf("app url = %q", r.AppURL)
	}
}

func TestBuildKawasanReportFallsBackToGlobalPlant(t *testing.T) {
	r := buildKawasanReport(reportRows()[:1], decryptStub, time.Now(), "http://x", "", "")
	if r.AppURL != "http://x/cimory/global/dashboard/me/issues" {
		t.Fatalf("app url = %q", r.AppURL)
	}
}

func TestBuildKawasanReportKeepsDetailWithoutResults(t *testing.T) {
	rows := append(reportRows(), inspection.KawasanReportRow{
		AreaName: "Area Produksi", KawasanName: "Produksi CMD 1", DetailKawasanID: "DK-3", DetailKawasanName: "Packing",
	})
	r := buildKawasanReport(rows, decryptStub, time.Now(), "http://x", "PLT-SENTUL", "")
	if len(r.Details) != 3 || r.Details[2] != (DetailScore{Name: "Packing"}) {
		t.Fatalf("details = %+v", r.Details)
	}
}

func fixedNow() time.Time { return time.Date(2026, 10, 8, 14, 0, 0, 0, wib()) }
