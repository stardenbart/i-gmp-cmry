package inspectionusecase

import (
	"strings"
	"testing"
	"time"
)

func wib() *time.Location { return time.FixedZone("WIB", 7*60*60) }

func sampleReport() KawasanReport {
	due := time.Date(2026, 10, 22, 0, 0, 0, 0, wib())
	fu := time.Date(2026, 10, 9, 0, 0, 0, 0, wib())
	return KawasanReport{
		AreaName:    "Area Produksi",
		KawasanName: "Produksi CMD 1",
		GeneratedAt: time.Date(2026, 10, 8, 14, 30, 0, 0, wib()),
		Details: []DetailScore{
			{Name: "Proses", Score: 4, Max: 4},
			{Name: "Gudang Bahan", Score: 1, Max: 4},
		},
		Issues: []ReportIssue{
			{
				DetailKawasanName: "Gudang Bahan",
				Title:             "Lantai bebas genangan air",
				Keterangan:        "Genangan di dekat palet <script>alert(1)</script>",
				PICName:           "Budi PIC",
				DueDate:           &due,
				Status:            "Open",
				URL:               "http://172.20.240.61:3000/cimory/PLT-SENTUL/dashboard/me/issues/ISS-1",
			},
			{
				DetailKawasanName: "Gudang Bahan",
				Title:             "Label bahan lengkap",
				PICName:           "Sari PIC",
				Status:            "InProgress",
				FollowUpBy:        "Sari PIC",
				FollowUpDate:      &fu,
				URL:               "http://172.20.240.61:3000/cimory/PLT-SENTUL/dashboard/me/issues/ISS-2",
			},
		},
		AppURL: "http://172.20.240.61:3000/cimory/PLT-SENTUL/dashboard/me/issues",
	}
}

func TestKawasanReportSubjectFollowsFlow(t *testing.T) {
	if got, want := sampleReport().Subject(), "[I-GMP] Laporan Inspeksi GMP - Produksi CMD 1 - 08 Okt 2026"; got != want {
		t.Fatalf("subject = %q, want %q", got, want)
	}
}

func TestKawasanReportScores(t *testing.T) {
	r := sampleReport()
	score, max := r.Total()
	if score != 5 || max != 8 {
		t.Fatalf("total = %d/%d, want 5/8", score, max)
	}
	if p := scorePercent(5, 8); p != 62 {
		t.Fatalf("percent = %d, want 62 (rounded down like the flow)", p)
	}
	if p := scorePercent(0, 0); p != 0 {
		t.Fatalf("percent of nothing = %d, want 0", p)
	}
}

func TestRenderKawasanReportMatchesFlowSections(t *testing.T) {
	html, err := RenderKawasanReport(sampleReport())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"I-GMP — Laporan Hasil Inspeksi GMP",
		"Yth. Tim PIC <b>Produksi CMD 1</b>",
		"Informasi Inspeksi", "Area Produksi", "08 Okt 2026",
		"Ringkasan Skor per Detail Kawasan",
		"Proses", "4 / 4", ">100<",
		"Gudang Bahan", "1 / 4", ">25<",
		"Average Keseluruhan", "5 / 8", ">62<",
		"Daftar Issue (2 issue)",
		"Lantai bebas genangan air", "Budi PIC", "22 Okt 2026",
		"Belum ada follow up", "Sari PIC — 09 Okt 2026",
		`href="http://172.20.240.61:3000/cimory/PLT-SENTUL/dashboard/me/issues/ISS-1"`,
		"Lihat Detail &amp; Foto",
		`href="http://172.20.240.61:3000/cimory/PLT-SENTUL/dashboard/me/issues"`,
		"Buka I-GMP",
		"Email ini dikirim otomatis oleh sistem I-GMP",
		"Digital Transformation — Plant Sentul",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("email is missing %q", want)
		}
	}
}

func TestRenderKawasanReportColorsScoresAgainst80Percent(t *testing.T) {
	html, _ := RenderKawasanReport(sampleReport())
	if !strings.Contains(html, `color:#1e8449;">100<`) {
		t.Error("100% should be green")
	}
	if !strings.Contains(html, `color:#c62833;">25<`) {
		t.Error("25% should be red")
	}
}

func TestRenderKawasanReportEscapesUserText(t *testing.T) {
	html, _ := RenderKawasanReport(sampleReport())
	if strings.Contains(html, "<script>") {
		t.Fatal("keterangan must be HTML-escaped")
	}
}

func TestRenderKawasanReportWithoutIssues(t *testing.T) {
	r := sampleReport()
	r.Issues = nil
	html, err := RenderKawasanReport(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "Tidak ada issue ditemukan dalam inspeksi ini. Pertahankan!") {
		t.Error("expected the no-issue message")
	}
	if strings.Contains(html, "Daftar Issue") {
		t.Error("issue list should be hidden when there are no issues")
	}
}

func TestReportIssueStatusLabel(t *testing.T) {
	for status, want := range map[string]string{
		"Open":              "Open",
		"OpenOverdue":       "Open (Overdue)",
		"InProgress":        "In Progress",
		"PendingValidation": "Menunggu Validasi",
		"Verified":          "Closed",
		"Closed":            "Closed",
		"":                  "Open",
	} {
		if got := (ReportIssue{Status: status}).StatusLabel(); got != want {
			t.Errorf("StatusLabel(%q) = %q, want %q", status, got, want)
		}
	}
}

func TestRenderKawasanReportShowsDashForDetailWithoutChecklist(t *testing.T) {
	r := sampleReport()
	r.Details = append(r.Details, DetailScore{Name: "Packing"})
	html, err := RenderKawasanReport(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `Packing</td><td style="padding:8px;border:1px solid #e5e7eb;text-align:center;">-</td>`) {
		t.Error("a detail kawasan without checklist results should show '-' as its score")
	}
}
