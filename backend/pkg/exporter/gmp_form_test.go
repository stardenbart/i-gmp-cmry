package exporter

import (
	"archive/zip"
	"bytes"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

const gmpFormTemplate = "../../templates/gmp_form.xlsx"

func intp(v int) *int { return &v }

func formFixture() GMPFormSheet {
	due := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	return GMPFormSheet{
		Company:  "PT CISARUA MOUNTAIN DAIRY TBK\nPLANT SENTUL",
		Title:    "INSPEKSI GMP (BASIC) - Area Produksi",
		Date:     time.Date(2026, 9, 19, 9, 0, 0, 0, time.UTC),
		Location: "Proses - Produksi CMD 1",
		PIC:      "Hermaslin Pasaribu, Moh. Rama Gumilar",
		Uraian: []GMPFormUraian{
			{UraianID: "CMDXY001", Aspek: "Lingkungan Sarana Produksi", Detail: "Lokasi Pabrik", Text: "Bebas dari banjir", Nilai: intp(2)},
			{UraianID: "CMDXY002", Aspek: "Lingkungan Sarana Produksi", Detail: "Sarana Jalan", Text: "Jalan bersih", Nilai: intp(2)},
			{UraianID: "CMDXY003", Aspek: "Lingkungan Sarana Produksi", Detail: "Sarana Jalan", Text: "Saluran air baik", Nilai: nil}, // NA
			{UraianID: "CMDXY015", Aspek: "Konstruksi dan Layout Bangunan", Detail: "Lantai", Text: "Bersih dan kering", Nilai: intp(0),
				Findings: []GMPFormFinding{{PhotoURL: "http://host/p1.jpg", Keterangan: "lantai basah", Status: "Open", DueDate: &due}}},
			{UraianID: "CMDXY050", Aspek: "Lainnya", Detail: "Lainnya", Text: "Ketidaksesuaian lain", Nilai: intp(0),
				Findings: []GMPFormFinding{
					{PhotoURL: "http://host/a.jpg", Keterangan: "APD di tempat botol", Status: "Open", DueDate: &due},
					{PhotoURL: "http://host/b.jpg", Keterangan: "Epoxy lantai kitchen", Status: "Open", DueDate: &due, FollowUp: "05 Oct 2026 - sudah dirapikan"},
					{PhotoURL: "http://host/c.jpg", Keterangan: "Gelas berlakban", Status: "Closed", DueDate: &due},
				}},
		},
	}
}

func openForm(t *testing.T, sheets ...GMPFormSheet) *excelize.File {
	t.Helper()
	buf, err := GenerateGMPFormExcel(gmpFormTemplate, sheets, nil)
	if err != nil {
		t.Fatalf("GenerateGMPFormExcel: %v", err)
	}
	f, err := excelize.OpenReader(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func cell(t *testing.T, f *excelize.File, sheet, axis string) string {
	t.Helper()
	v, err := f.GetCellValue(sheet, axis)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func merged(t *testing.T, f *excelize.File, sheet string) map[string]bool {
	t.Helper()
	ms, err := f.GetMergeCells(sheet)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]bool{}
	for _, m := range ms {
		out[m.GetStartAxis()+":"+m.GetEndAxis()] = true
	}
	return out
}

func TestGMPFormHeader(t *testing.T) {
	f := openForm(t, formFixture())
	if got := f.GetSheetList(); len(got) != 1 || got[0] != "Hal 00" {
		t.Fatalf("sheets = %v, want [Hal 00]", got)
	}
	for axis, want := range map[string]string{
		"C1": "PT CISARUA MOUNTAIN DAIRY TBK\nPLANT SENTUL",
		"C4": "INSPEKSI GMP (BASIC) - Area Produksi",
		"C6": "Proses - Produksi CMD 1",
		"C7": "Hermaslin Pasaribu, Moh. Rama Gumilar",
		"M4": ": 1 dari 1",
	} {
		if got := cell(t, f, "Hal 00", axis); got != want {
			t.Errorf("%s = %q, want %q", axis, got, want)
		}
	}
	if got := cell(t, f, "Hal 00", "C5"); got != "19-Sep-26" {
		t.Errorf("C5 date = %q, want 19-Sep-26 (d-mmm-yy)", got)
	}
}

func TestGMPFormNumbersUraianSequentiallyWithUraianID(t *testing.T) {
	f := openForm(t, formFixture())
	// Rows 10..12: three uraian without findings, 13: one finding, 14..16: three findings.
	wantNo := map[int]string{10: "1", 11: "2", 12: "3", 13: "4", 14: "5"}
	for row, want := range wantNo {
		if got := cell(t, f, "Hal 00", "A"+itoa(row)); got != want {
			t.Errorf("A%d = %q, want %q", row, got, want)
		}
	}
	if got := cell(t, f, "Hal 00", "D10"); got != "CMDXY001" {
		t.Errorf("D10 = %q, want the UraianID", got)
	}
	if got := cell(t, f, "Hal 00", "E14"); got != "Ketidaksesuaian lain" {
		t.Errorf("E14 = %q", got)
	}
}

func TestGMPFormSplitsEachFindingIntoItsOwnRow(t *testing.T) {
	f := openForm(t, formFixture())
	for i, want := range []string{"APD di tempat botol", "Epoxy lantai kitchen", "Gelas berlakban"} {
		row := 14 + i
		if got := cell(t, f, "Hal 00", "K"+itoa(row)); got != want {
			t.Errorf("K%d = %q, want %q", row, got, want)
		}
		if got := cell(t, f, "Hal 00", "J"+itoa(row)); got != "Lihat Foto" {
			t.Errorf("J%d = %q, want Lihat Foto", row, got)
		}
		if ok, target, _ := f.GetCellHyperLink("Hal 00", "J"+itoa(row)); !ok || !strings.HasPrefix(target, "http://host/") {
			t.Errorf("J%d hyperlink = %v %q", row, ok, target)
		}
	}
	if got := cell(t, f, "Hal 00", "L15"); got != "05 Oct 2026 - sudah dirapikan" {
		t.Errorf("L15 follow up = %q", got)
	}
	if got := cell(t, f, "Hal 00", "M14"); got != "Open - 03 Oct 2026" {
		t.Errorf("M14 = %q, want 'Open - 03 Oct 2026'", got)
	}
	if got := cell(t, f, "Hal 00", "M16"); got != "Closed - 03 Oct 2026" {
		t.Errorf("M16 = %q", got)
	}

	m := merged(t, f, "Hal 00")
	for _, rng := range []string{"A14:A16", "B14:B16", "C14:C16", "D14:D16", "E14:F16", "G14:G16", "H14:H16", "I14:I16"} {
		if !m[rng] {
			t.Errorf("expected %s merged for the 3-finding uraian", rng)
		}
	}
	for _, col := range []string{"J", "K", "L", "M"} {
		if m[col+"14:"+col+"16"] {
			t.Errorf("%s must stay one cell per finding", col)
		}
	}
}

func TestGMPFormCountsFindingsPerUraianAndScoresPerAspek(t *testing.T) {
	f := openForm(t, formFixture())
	if got := cell(t, f, "Hal 00", "I13"); got != "1" {
		t.Errorf("I13 = %q, want 1 finding", got)
	}
	if got := cell(t, f, "Hal 00", "I14"); got != "3" {
		t.Errorf("I14 = %q, want 3 findings", got)
	}
	if got := cell(t, f, "Hal 00", "I10"); got != "" {
		t.Errorf("I10 = %q, want empty for a uraian without findings", got)
	}
	// Total Nilai per aspek on the aspek's first row, merged over its rows.
	if got := cell(t, f, "Hal 00", "H10"); got != "4" {
		t.Errorf("H10 = %q, want 4 (2+2+NA)", got)
	}
	m := merged(t, f, "Hal 00")
	if !m["B10:B12"] || !m["H10:H12"] || !m["C11:C12"] {
		t.Errorf("aspek/detail groups not merged: %v", m)
	}
	// NA leaves Nilai empty.
	if got := cell(t, f, "Hal 00", "G12"); got != "" {
		t.Errorf("G12 = %q, want empty for NA", got)
	}
}

func TestGMPFormHighlightsZeroScores(t *testing.T) {
	f := openForm(t, formFixture())
	zero, _ := f.GetCellStyle("Hal 00", "G13")
	two, _ := f.GetCellStyle("Hal 00", "G10")
	style, err := f.GetStyle(zero)
	if err != nil {
		t.Fatal(err)
	}
	if zero == two || len(style.Fill.Color) == 0 || !strings.EqualFold(style.Fill.Color[0], "FFD7D7") {
		t.Fatalf("G13 (nilai 0) fill = %+v, want FFD7D7", style.Fill)
	}
}

func TestGMPFormTotalsUseFormulas(t *testing.T) {
	f := openForm(t, formFixture())
	// 7 data rows (10..16) → total row 17.
	if got := cell(t, f, "Hal 00", "A17"); !strings.HasPrefix(got, "Total Score IPGMP") {
		t.Fatalf("A17 = %q, want the total row", got)
	}
	for axis, want := range map[string]string{
		"G17": "SUM(G10:G16)",
		"I17": "SUM(I10:I16)",
		"M17": `IFERROR(COUNTIF(M10:M16,"*Closed*")/COUNTIF(M10:M16,"<>"),0)`,
	} {
		got, err := f.GetCellFormula("Hal 00", axis)
		if err != nil || got != want {
			t.Errorf("%s formula = %q (%v), want %q", axis, got, err, want)
		}
	}
	if got := cell(t, f, "Hal 00", "A19"); !strings.HasPrefix(got, "Keterangan") {
		t.Errorf("legend should follow the total row, A19 = %q", got)
	}
}

func TestGMPFormOneSheetPerInspection(t *testing.T) {
	second := formFixture()
	second.Location = "Filling - Produksi CMD 1"
	second.Uraian = second.Uraian[:1]
	f := openForm(t, formFixture(), second)
	if got := f.GetSheetList(); len(got) != 2 || got[0] != "Hal 00" || got[1] != "Hal 01" {
		t.Fatalf("sheets = %v, want [Hal 00 Hal 01]", got)
	}
	if got := cell(t, f, "Hal 01", "C6"); got != "Filling - Produksi CMD 1" {
		t.Errorf("Hal 01 C6 = %q", got)
	}
	if got := cell(t, f, "Hal 01", "M4"); got != ": 2 dari 2" {
		t.Errorf("Hal 01 M4 = %q", got)
	}
	if got := cell(t, f, "Hal 01", "A11"); !strings.HasPrefix(got, "Total Score IPGMP") {
		t.Errorf("Hal 01 total row should follow its single uraian, A11 = %q", got)
	}
}

func itoa(n int) string { return strconv.Itoa(n) }

// The template already carries the logo; every sheet must end up with
// exactly one, including the copies made for additional inspections.
func TestGMPFormHasOneLogoPerSheet(t *testing.T) {
	second := formFixture()
	second.Uraian = second.Uraian[:1]
	buf, err := GenerateGMPFormExcel(gmpFormTemplate, []GMPFormSheet{formFixture(), second}, []string{"../../../frontend/public/Logo_Cimory.png"})
	if err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenReader(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	for _, sheet := range []string{"Hal 00", "Hal 01"} {
		pics, err := f.GetPictures(sheet, "A1")
		if err != nil || len(pics) != 1 {
			t.Errorf("%s: %d logos at A1 (%v), want 1", sheet, len(pics), err)
		}
	}
}

// The template is cut from a filled reference; none of its inspection text
// may survive in the shared string table of an export.
func TestGMPFormCarriesNoReferenceLeftovers(t *testing.T) {
	buf, err := GenerateGMPFormExcel(gmpFormTemplate, []GMPFormSheet{formFixture()}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, leftover := range []string{"lantai basah di area proses", "H.M.Deden", "Area pengerjaan epoxy lantai kitchen", "sharepoint.com/sites"} {
		if bytes.Contains(unzipAll(t, buf.Bytes()), []byte(leftover)) {
			t.Errorf("export still contains %q from the reference workbook", leftover)
		}
	}
}

func unzipAll(t *testing.T, data []byte) []byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	var all bytes.Buffer
	for _, file := range zr.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(&all, rc)
		rc.Close()
	}
	return all.Bytes()
}

// Total Nilai belongs to each contiguous aspek block, even if an aspek name
// shows up again further down.
func TestGMPFormTotalNilaiPerAspekBlock(t *testing.T) {
	s := formFixture()
	s.Uraian = []GMPFormUraian{
		{UraianID: "U1", Aspek: "A", Detail: "d1", Text: "1", Nilai: intp(2)},
		{UraianID: "U2", Aspek: "A", Detail: "d1", Text: "2", Nilai: intp(2)},
		{UraianID: "U3", Aspek: "B", Detail: "d2", Text: "3", Nilai: intp(0)},
		{UraianID: "U4", Aspek: "A", Detail: "d3", Text: "4", Nilai: intp(2)},
	}
	f := openForm(t, s)
	for axis, want := range map[string]string{"H10": "4", "H12": "0", "H13": "2"} {
		if got := cell(t, f, "Hal 00", axis); got != want {
			t.Errorf("%s = %q, want %q", axis, got, want)
		}
	}
}
