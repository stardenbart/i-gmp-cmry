package exporter

import (
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestIssueReportTemplateRendersHeadersAndItems(t *testing.T) {
	payload := &PlaceholderPayload{
		Headers: map[string]interface{}{
			"generated_at":   "01 September 2026 12:00 WIB",
			"period_filter":  "2026-09-01 s.d. 2026-09-01",
			"status_filter":  "Open",
			"search_filter":  "kebersihan",
			"total":          1,
			"status_summary": "Open: 1",
		},
		Items: []map[string]interface{}{{
			"no": 1, "issue_id": "ISS-001", "created_at": "01-Sep-2026 08:00",
			"plant": "Sentul", "area": "Produksi", "kawasan": "Line 1",
			"detail_kawasan": "Filling", "aspek": "GMP", "detail_aspek": "Kebersihan",
			"pic": "Auditee", "status": "Open", "due_date": "05-Sep-2026",
			"follow_up_delay": "0 hari", "needs_wowr": "Tidak", "wo_id": "-",
			"wr_id": "-", "wowr_status": "None", "hei_detail": "Habit - Kebersihan",
			"finding_description": "Lantai kotor", "keterangan": "Perlu dibersihkan",
			"initial_count": 2, "follow_up_count": 0, "wowr_count": 0,
			"updated_at": "01-Sep-2026 08:00",
		}},
		TotalRow: &TableTotalRow{
			Label:            "TOTAL KESELURUHAN TEMUAN",
			LabelStartColumn: "A",
			LabelEndColumn:   "W",
			ValueColumn:      "X",
			Value:            1,
		},
	}

	buffer, err := GenerateExcelWithPlaceholder("../../templates/issue_report.xlsx", "Laporan Temuan", payload)
	if err != nil {
		t.Fatalf("render issue report: %v", err)
	}
	book, err := excelize.OpenReader(buffer)
	if err != nil {
		t.Fatalf("open generated report: %v", err)
	}
	defer book.Close()

	if got, _ := book.GetCellValue("Laporan Temuan", "B2"); got != "01 September 2026 12:00 WIB" {
		t.Fatalf("unexpected generated_at: %q", got)
	}
	if got, _ := book.GetCellValue("Laporan Temuan", "B9"); got != "ISS-001" {
		t.Fatalf("unexpected issue id: %q", got)
	}
	if got, _ := book.GetCellValue("Laporan Temuan", "A10"); got != "TOTAL KESELURUHAN TEMUAN" {
		t.Fatalf("unexpected total label: %q", got)
	}
	if got, _ := book.GetCellValue("Laporan Temuan", "X10"); got != "1" {
		t.Fatalf("unexpected issue grand total: %q", got)
	}
	rows, err := book.GetRows("Laporan Temuan")
	if err != nil {
		t.Fatalf("read generated rows: %v", err)
	}
	for _, row := range rows {
		for _, cell := range row {
			if strings.Contains(cell, "{item.") {
				t.Fatalf("unresolved item placeholder: %q", cell)
			}
		}
	}
}
