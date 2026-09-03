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
		KawasanID:           "KWS-001",
		Kawasan:             "Line 1",
		DetailKawasanID:     "DKW-001",
		DetailKawasan:       "Filling",
		Aspek:               "Kebersihan",
		DetailAspek:         "Lantai",
		Uraian:              "Lantai harus bersih",
		Nilai:               0,
		StandardScore:       2,
		TotalNilaiKawasan:   80,
		CompliancePercent:   75.5,
		IssueUraianCount:    1,
		FollowUpDescription: "Follow-Up 1: Sudah dibersihkan",
		KeteranganTemuan:    "Lantai kotor\nSaluran kotor",
		FollowUpDate:        "02-Sep-2026 10:00",
		DueDate:             "01-Sep-2026",
		FollowUpGapDays:     &late,
	}, {
		InspectionID:      "INSP-001",
		Area:              "Produksi",
		KawasanID:         "KWS-001",
		Kawasan:           "Line 1",
		DetailKawasanID:   "DKW-001",
		DetailKawasan:     "Filling",
		Aspek:             "Kebersihan",
		DetailAspek:       "Lantai",
		Uraian:            "Saluran harus bersih",
		Nilai:             2,
		StandardScore:     2,
		TotalNilaiKawasan: 80,
		CompliancePercent: 75.5,
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
	assertCell("B5", "Kawasan")
	assertCell("C5", "Detail Kawasan")
	assertCell("F5", "Uraian")
	assertCell("J5", "Temuan (Uraian)")
	assertCell("N5", "Keterangan Foto Temuan Awal")
	assertCell("Q5", "Gap Follow-Up")
	assertCell("A6", "INSP-001\nProduksi")
	assertCell("B6", "Line 1")
	assertCell("C6", "Filling")
	assertCell("F6", "Lantai harus bersih")
	assertCell("J6", "1")
	assertCell("M6", "Follow-Up 1: Sudah dibersihkan")
	assertCell("N6", "Lantai kotor\nSaluran kotor")
	assertCell("Q6", "+2 hari terlambat")
	assertCell("A8", "TOTAL KESELURUHAN (2 BARIS)")
	assertCell("G8", "2")
	assertCell("H8", "2")
	assertCell("I8", "50.00%")
	assertCell("J8", "1")

	merged, mergeErr := workbook.GetMergeCells("Data GMP")
	if mergeErr != nil {
		t.Fatalf("read merged cells: %v", mergeErr)
	}
	mergedRanges := make(map[string]bool, len(merged))
	for _, cell := range merged {
		mergedRanges[cell.GetStartAxis()+":"+cell.GetEndAxis()] = true
	}
	// A-E: both rows share InspectionID/Kawasan/Detail Kawasan/Aspek/Detail
	// Aspek, so every one of those columns merges — mirroring the Data GMP
	// web table's hierarchical rowspan. F (Uraian) must NOT merge: the two
	// rows have different Uraian text, exactly the case that should stay
	// on its own row (same as the web table's uraian-level merge).
	for _, want := range []string{"A6:A7", "B6:B7", "C6:C7", "D6:D7", "E6:E7", "H6:H7", "I6:I7"} {
		if !mergedRanges[want] {
			t.Fatalf("expected merged range %s", want)
		}
	}
	if mergedRanges["F6:F7"] {
		t.Fatal("F6:F7 (Uraian) must not merge — the two rows have different Uraian text")
	}
}

// TestGenerateGMPTableExcelDoesNotMergeAcrossInspections mirrors the web
// table's scoping rule: even if two consecutive rows happen to share the
// same Aspek/Detail Aspek/Uraian text, they must NOT merge together if they
// belong to different inspections (InspectionID differs) — same guarantee
// the frontend's computeGmpRowSpans provides for the on-screen table.
func TestGenerateGMPTableExcelDoesNotMergeAcrossInspections(t *testing.T) {
	row := GMPTableRow{
		KawasanID:       "KWS-001",
		Kawasan:         "Line 1",
		DetailKawasanID: "DKW-001",
		DetailKawasan:   "Filling",
		Aspek:           "Kebersihan",
		DetailAspek:     "Lantai",
		Uraian:          "Lantai harus bersih",
	}
	rowA, rowB := row, row
	rowA.InspectionID, rowB.InspectionID = "INSP-001", "INSP-002"

	buffer, err := GenerateGMPTableExcel(GMPTableMetadata{}, []GMPTableRow{rowA, rowB})
	if err != nil {
		t.Fatalf("generate workbook: %v", err)
	}
	workbook, err := excelize.OpenReader(bytes.NewReader(buffer.Bytes()))
	if err != nil {
		t.Fatalf("open generated workbook: %v", err)
	}
	defer workbook.Close()

	merged, mergeErr := workbook.GetMergeCells("Data GMP")
	if mergeErr != nil {
		t.Fatalf("read merged cells: %v", mergeErr)
	}
	for _, cell := range merged {
		got := cell.GetStartAxis() + ":" + cell.GetEndAxis()
		for _, forbidden := range []string{"A6:A7", "B6:B7", "C6:C7", "D6:D7", "E6:E7", "F6:F7"} {
			if got == forbidden {
				t.Fatalf("rows from different inspections must not merge, but found %s", forbidden)
			}
		}
	}
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
