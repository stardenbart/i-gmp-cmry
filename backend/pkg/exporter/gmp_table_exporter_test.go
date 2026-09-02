package exporter

import (
	"bytes"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestGenerateGMPTableExcelMatchesWebColumns(t *testing.T) {
	late := 2
	buffer, err := GenerateGMPTableExcel(GMPTableMetadata{
		Plant:      "Sentul",
		Area:       "Produksi",
		StartDate:  "2026-09-01",
		EndDate:    "2026-09-02",
		Search:     "lantai",
		ExportedAt: time.Date(2026, 9, 2, 10, 30, 0, 0, time.FixedZone("WIB", 7*60*60)),
	}, []GMPTableRow{{
		InspectionID:        "INSP-001",
		Area:                "Produksi",
		Kawasan:             "Line 1",
		DetailKawasan:       "Filling",
		Aspek:               "Kebersihan",
		DetailAspek:         "Lantai",
		UraianID:            "UR-001",
		Nilai:               0,
		TotalNilaiKawasan:   80,
		CompliancePercent:   75.5,
		Temuan:              1,
		FollowUpDescription: "Follow-Up 1: Sudah dibersihkan",
		KeteranganTemuan:    "Lantai kotor",
		FollowUpDate:        "02-Sep-2026 10:00",
		DueDate:             "01-Sep-2026",
		FollowUpGapDays:     &late,
	}})
	if err != nil {
		t.Fatalf("generate workbook: %v", err)
	}

	workbook, err := excelize.OpenReader(bytes.NewReader(buffer.Bytes()))
	if err != nil {
		t.Fatalf("open generated workbook: %v", err)
	}
	defer workbook.Close()

	assertCell := func(cell, want string) {
		t.Helper()
		got, cellErr := workbook.GetCellValue("Data GMP", cell)
		if cellErr != nil {
			t.Fatalf("read %s: %v", cell, cellErr)
		}
		if got != want {
			t.Fatalf("%s: got %q, want %q", cell, got, want)
		}
	}

	assertCell("A1", "DATA INSPEKSI (GMP)")
	assertCell("A5", "ID Inspeksi")
	assertCell("P5", "Gap Follow-Up")
	assertCell("A6", "INSP-001\nProduksi")
	assertCell("B6", "Line 1 / Filling")
	assertCell("L6", "Follow-Up 1: Sudah dibersihkan")
	assertCell("P6", "+2 hari terlambat")
}

func TestGenerateGMPTableExcelSupportsEmptyRows(t *testing.T) {
	buffer, err := GenerateGMPTableExcel(GMPTableMetadata{}, nil)
	if err != nil {
		t.Fatalf("generate empty workbook: %v", err)
	}
	if buffer.Len() == 0 {
		t.Fatal("expected a non-empty workbook")
	}
}
