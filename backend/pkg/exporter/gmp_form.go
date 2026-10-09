package exporter

import (
	"bytes"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

// GMP inspection form ("INSPEKSI GMP (BASIC)") export, laid out like the
// QA reference workbook GMP_Ref.xlsx: one sheet per inspection ("Hal 00",
// "Hal 01", ...), uraian numbered in UraianID order, and one row per finding
// so every finding keeps its own photo link, keterangan, follow-up and due
// date while the uraian columns (A–I) are merged over its rows. The
// template is built by scripts/gen-gmp-form-template.

const (
	gmpFormBaseSheet = "Hal 00"
	gmpFormDataStart = 10
	gmpFormZeroFill  = "FFD7D7"
	gmpFormOpenFont  = "C00000"
	gmpFormDoneFont  = "1E8449"
)

type GMPFormSheet struct {
	Company  string // C1, e.g. "PT CISARUA MOUNTAIN DAIRY TBK\nPLANT SENTUL"
	Title    string // C4, e.g. "INSPEKSI GMP (BASIC) - Area Produksi"
	Date     time.Time
	Location string // C6, "Detail Kawasan - Kawasan"
	PIC      string // C7
	Uraian   []GMPFormUraian
}

type GMPFormUraian struct {
	UraianID string
	Aspek    string
	Detail   string
	Text     string
	Nilai    *int // nil = NA (left empty)
	Findings []GMPFormFinding
}

type GMPFormFinding struct {
	PhotoURL   string
	Keterangan string
	FollowUp   string
	Status     string // display label, e.g. "Open", "Closed"
	DueDate    *time.Time
}

type gmpFormRow struct {
	uraian  int // index into Uraian
	finding int // index into Findings, -1 when the uraian has none
	first   bool
	span    int // rows of this uraian (set on its first row)
}

func gmpFormLayout(uraian []GMPFormUraian) []gmpFormRow {
	var rows []gmpFormRow
	for i, u := range uraian {
		span := len(u.Findings)
		if span == 0 {
			rows = append(rows, gmpFormRow{uraian: i, finding: -1, first: true, span: 1})
			continue
		}
		for j := range u.Findings {
			rows = append(rows, gmpFormRow{uraian: i, finding: j, first: j == 0, span: span})
		}
	}
	return rows
}

// GenerateGMPFormExcel fills templatePath once per sheet. logoPaths are
// tried in order for the company logo in A1 (nil skips the logo).
func GenerateGMPFormExcel(templatePath string, sheets []GMPFormSheet, logoPaths []string) (*bytes.Buffer, error) {
	if len(sheets) == 0 {
		return nil, fmt.Errorf("tidak ada inspeksi untuk diekspor")
	}
	f, err := excelize.OpenFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka template: %w", err)
	}
	defer f.Close()

	baseIndex, err := f.GetSheetIndex(gmpFormBaseSheet)
	if err != nil || baseIndex < 0 {
		return nil, fmt.Errorf("sheet %q tidak ada di template", gmpFormBaseSheet)
	}
	names := make([]string, len(sheets))
	for i := range sheets {
		names[i] = fmt.Sprintf("Hal %02d", i)
		if i == 0 {
			continue
		}
		idx, err := f.NewSheet(names[i])
		if err != nil {
			return nil, err
		}
		if err := f.CopySheet(baseIndex, idx); err != nil {
			return nil, fmt.Errorf("gagal menyalin sheet: %w", err)
		}
	}
	_ = f.DeleteDefinedName(&excelize.DefinedName{Name: "_xlnm.Print_Area", Scope: gmpFormBaseSheet})

	for i, sheet := range sheets {
		if err := fillGMPFormSheet(f, names[i], sheet, i+1, len(sheets), logoPaths); err != nil {
			return nil, fmt.Errorf("%s: %w", names[i], err)
		}
	}
	f.SetActiveSheet(baseIndex)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}

func fillGMPFormSheet(f *excelize.File, name string, sheet GMPFormSheet, page, pages int, logoPaths []string) error {
	set := func(axis string, value interface{}) error { return f.SetCellValue(name, axis, value) }

	// Header. Excel dates carry no zone, so write the inspection's calendar
	// day as a plain date.
	day := time.Date(sheet.Date.Year(), sheet.Date.Month(), sheet.Date.Day(), 0, 0, 0, 0, time.UTC)
	// Writing a time replaces the cell's number format; keep the template's
	// d-mmm-yy.
	dateStyle, _ := f.GetCellStyle(name, "C5")
	defer func() { _ = f.SetCellStyle(name, "C5", "C5", dateStyle) }()
	for axis, value := range map[string]interface{}{
		"C1": sheet.Company, "C4": sheet.Title, "C5": day, "C6": sheet.Location, "C7": sheet.PIC,
		"M4": fmt.Sprintf(": %d dari %d", page, pages),
	} {
		if err := set(axis, value); err != nil {
			return err
		}
	}

	rows := gmpFormLayout(sheet.Uraian)
	n := len(rows)
	if n == 0 {
		n = 1
	}
	for i := 1; i < n; i++ {
		if err := f.DuplicateRow(name, gmpFormDataStart); err != nil {
			return fmt.Errorf("gagal menambah baris: %w", err)
		}
	}
	last := gmpFormDataStart + n - 1
	if err := unmergeRows(f, name, gmpFormDataStart, last); err != nil {
		return err
	}

	gStyle, _ := f.GetCellStyle(name, fmt.Sprintf("G%d", gmpFormDataStart))
	zeroStyle, err := withFill(f, gStyle, gmpFormZeroFill)
	if err != nil {
		return err
	}
	mStyle, _ := f.GetCellStyle(name, fmt.Sprintf("M%d", gmpFormDataStart))
	openStyle, err := withFontColor(f, mStyle, gmpFormOpenFont)
	if err != nil {
		return err
	}
	doneStyle, err := withFontColor(f, mStyle, gmpFormDoneFont)
	if err != nil {
		return err
	}

	aspekNilai := map[string]int{}
	for _, u := range sheet.Uraian {
		if u.Nilai != nil {
			aspekNilai[u.Aspek] += *u.Nilai
		}
	}

	for i, r := range rows {
		row := gmpFormDataStart + i
		u := sheet.Uraian[r.uraian]
		at := func(col string) string { return fmt.Sprintf("%s%d", col, row) }

		if r.first {
			_ = set(at("A"), r.uraian+1)
			_ = set(at("D"), u.UraianID)
			_ = set(at("E"), u.Text)
			if u.Nilai != nil {
				_ = set(at("G"), *u.Nilai)
				if *u.Nilai == 0 {
					_ = f.SetCellStyle(name, at("G"), at("G"), zeroStyle)
				}
			}
			if len(u.Findings) > 0 {
				_ = set(at("I"), len(u.Findings))
			}
		}
		// Aspek / Detail / Total Nilai are written on every row; the merges
		// below keep only the first value of each group visible.
		_ = set(at("B"), u.Aspek)
		_ = set(at("C"), u.Detail)
		_ = set(at("H"), aspekNilai[u.Aspek])

		if r.finding >= 0 {
			fd := u.Findings[r.finding]
			if fd.PhotoURL != "" {
				_ = set(at("J"), "Lihat Foto")
				_ = f.SetCellHyperLink(name, at("J"), fd.PhotoURL, "External")
			}
			_ = set(at("K"), fd.Keterangan)
			_ = set(at("L"), fd.FollowUp)
			due := fd.Status
			if fd.DueDate != nil {
				due = fmt.Sprintf("%s - %s", fd.Status, fd.DueDate.Format("02 Jan 2006"))
			}
			_ = set(at("M"), due)
			style := openStyle
			if fd.Status == "Closed" {
				style = doneStyle
			}
			_ = f.SetCellStyle(name, at("M"), at("M"), style)
		}
	}

	if err := mergeGMPFormRows(f, name, sheet.Uraian, rows); err != nil {
		return err
	}

	total := last + 1
	for axis, formula := range map[string]string{
		fmt.Sprintf("G%d", total): fmt.Sprintf("SUM(G%d:G%d)", gmpFormDataStart, last),
		fmt.Sprintf("I%d", total): fmt.Sprintf("SUM(I%d:I%d)", gmpFormDataStart, last),
		fmt.Sprintf("M%d", total): fmt.Sprintf(`IFERROR(COUNTIF(M%d:M%d,"*Closed*")/COUNTIF(M%d:M%d,"<>"),0)`, gmpFormDataStart, last, gmpFormDataStart, last),
	} {
		if err := f.SetCellFormula(name, axis, formula); err != nil {
			return err
		}
	}

	// Print area: header through the legend (three rows after a blank row).
	if err := f.SetDefinedName(&excelize.DefinedName{
		Name:     "_xlnm.Print_Area",
		RefersTo: fmt.Sprintf("'%s'!$A$1:$M$%d", name, total+4),
		Scope:    name,
	}); err != nil {
		return err
	}

	// The template carries the logo; add it only where a copied sheet lost it.
	if pics, _ := f.GetPictures(name, "A1"); len(pics) == 0 && len(logoPaths) > 0 {
		insertFirstAvailableImage(f, name, "A1", logoPaths)
	}
	return nil
}

// mergeGMPFormRows merges each uraian's columns over its finding rows, and
// Aspek (B), Detail (C) and Total Nilai (H) over consecutive equal groups.
func mergeGMPFormRows(f *excelize.File, name string, uraian []GMPFormUraian, rows []gmpFormRow) error {
	merge := func(col1, col2 string, from, to int) error {
		if from == to && col1 == col2 {
			return nil
		}
		return f.MergeCell(name, fmt.Sprintf("%s%d", col1, from), fmt.Sprintf("%s%d", col2, to))
	}

	for i, r := range rows {
		if !r.first {
			continue
		}
		from, to := gmpFormDataStart+i, gmpFormDataStart+i+r.span-1
		if err := merge("E", "F", from, to); err != nil {
			return err
		}
		if r.span > 1 {
			for _, col := range []string{"A", "D", "G", "I"} {
				if err := merge(col, col, from, to); err != nil {
					return err
				}
			}
		}
	}

	groups := []struct {
		cols []string
		key  func(GMPFormUraian) string
	}{
		{[]string{"B", "H"}, func(u GMPFormUraian) string { return u.Aspek }},
		{[]string{"C"}, func(u GMPFormUraian) string { return u.Aspek + "\x00" + u.Detail }},
	}
	for _, g := range groups {
		start := 0
		for i := 1; i <= len(rows); i++ {
			if i < len(rows) && g.key(uraian[rows[i].uraian]) == g.key(uraian[rows[start].uraian]) {
				continue
			}
			if i-1 > start {
				for _, col := range g.cols {
					if err := merge(col, col, gmpFormDataStart+start, gmpFormDataStart+i-1); err != nil {
						return err
					}
				}
			}
			start = i
		}
	}
	return nil
}

// unmergeRows removes merges that lie within [from, to] (copied from the
// prototype row) so the per-inspection merges can be rebuilt cleanly.
func unmergeRows(f *excelize.File, name string, from, to int) error {
	merges, err := f.GetMergeCells(name)
	if err != nil {
		return err
	}
	for _, m := range merges {
		_, r1, _ := excelize.CellNameToCoordinates(m.GetStartAxis())
		_, r2, _ := excelize.CellNameToCoordinates(m.GetEndAxis())
		if r1 >= from && r2 <= to {
			if err := f.UnmergeCell(name, m.GetStartAxis(), m.GetEndAxis()); err != nil {
				return err
			}
		}
	}
	return nil
}

func withFill(f *excelize.File, styleID int, color string) (int, error) {
	style, err := f.GetStyle(styleID)
	if err != nil {
		return 0, err
	}
	style.Fill = excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{color}}
	return f.NewStyle(style)
}

func withFontColor(f *excelize.File, styleID int, color string) (int, error) {
	style, err := f.GetStyle(styleID)
	if err != nil {
		return 0, err
	}
	if style.Font == nil {
		style.Font = &excelize.Font{}
	}
	style.Font.Color = color
	return f.NewStyle(style)
}
