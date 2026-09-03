package exporter

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// GMPTableMetadata describes the active filters shown above the exported
// table. It intentionally contains display labels rather than database IDs.
type GMPTableMetadata struct {
	Plant         string
	Area          string
	Kawasan       string
	DetailKawasan string
	StartDate     string
	EndDate       string
	Search        string
	ExportedAt    time.Time
}

// GMPTableRow mirrors the seventeen columns rendered by the Data GMP web table.
// InitialImages and FollowUpImages are embedded into their respective cells.
type GMPTableRow struct {
	InspectionID        string
	Area                string
	KawasanID           string
	Kawasan             string
	DetailKawasanID     string
	DetailKawasan       string
	Aspek               string
	DetailAspek         string
	Uraian              string
	Nilai               int
	StandardScore       int
	TotalNilaiKawasan   int
	CompliancePercent   float64
	IssueUraianCount    int
	InitialImages       []string
	FollowUpImages      []string
	FollowUpDescription string
	KeteranganTemuan    string
	FollowUpDate        string
	DueDate             string
	FollowUpGapDays     *int
}

var gmpTableHeaders = []string{
	"ID Inspeksi",
	"Kawasan",
	"Detail Kawasan",
	"Aspek",
	"Detail Aspek",
	"Uraian",
	"Nilai",
	"Total Nilai (Kawasan)",
	"Persentase Kepatuhan (Detail Kawasan)",
	"Temuan (Uraian)",
	"Visual Temuan Awal",
	"Visual Follow-Up",
	"Keterangan Follow-Up",
	"Keterangan Foto Temuan Awal",
	"Follow-Up Datetime",
	"Due Date",
	"Gap Follow-Up",
}

// GenerateGMPTableExcel builds a flat XLSX representation of the Data GMP
// table. The corporate/template report remains handled separately by
// GenerateExcelWithPlaceholder.
func GenerateGMPTableExcel(meta GMPTableMetadata, rows []GMPTableRow) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer f.Close()

	const sheet = "Data GMP"
	defaultSheet := f.GetSheetName(f.GetActiveSheetIndex())
	if err := f.SetSheetName(defaultSheet, sheet); err != nil {
		return nil, fmt.Errorf("gagal membuat sheet Data GMP: %w", err)
	}

	if meta.ExportedAt.IsZero() {
		meta.ExportedAt = time.Now()
	}

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1D4ED8"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuat style judul: %w", err)
	}
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"334155"}, Pattern: 1},
		Border:    tableBorders("CBD5E1"),
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuat style header: %w", err)
	}
	bodyStyle, err := f.NewStyle(&excelize.Style{
		Border:    tableBorders("E2E8F0"),
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuat style isi: %w", err)
	}
	centerStyle, err := f.NewStyle(&excelize.Style{
		Border:    tableBorders("E2E8F0"),
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuat style tengah: %w", err)
	}
	percentStyle, err := f.NewStyle(&excelize.Style{
		Border:    tableBorders("E2E8F0"),
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		NumFmt:    10, // 0.00%
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuat style persen: %w", err)
	}
	lateStyle, _ := statusStyle(f, "FEE2E2", "B91C1C")
	earlyStyle, _ := statusStyle(f, "DCFCE7", "15803D")
	onTimeStyle, _ := statusStyle(f, "DBEAFE", "1D4ED8")
	totalLabelStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1E293B"}, Pattern: 1},
		Border:    tableBorders("64748B"),
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuat style label total: %w", err)
	}
	totalValueStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"0F766E"}, Pattern: 1},
		Border:    tableBorders("64748B"),
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuat style nilai total: %w", err)
	}
	totalPercentStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"0F766E"}, Pattern: 1},
		Border:    tableBorders("64748B"),
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		NumFmt:    10,
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuat style persentase total: %w", err)
	}

	if err := f.MergeCell(sheet, "A1", "Q1"); err != nil {
		return nil, fmt.Errorf("gagal menggabungkan judul: %w", err)
	}
	_ = f.SetCellValue(sheet, "A1", "DATA INSPEKSI (GMP)")
	_ = f.SetCellStyle(sheet, "A1", "Q1", titleStyle)
	_ = f.SetRowHeight(sheet, 1, 28)

	filterText := formatGMPFilterSummary(meta)
	_ = f.MergeCell(sheet, "A2", "Q2")
	_ = f.SetCellValue(sheet, "A2", filterText)
	_ = f.MergeCell(sheet, "A3", "Q3")
	_ = f.SetCellValue(sheet, "A3", fmt.Sprintf("Diekspor: %s | Jumlah baris: %d", meta.ExportedAt.Format("02 Jan 2006 15:04 MST"), len(rows)))

	const headerRow = 5
	for index, header := range gmpTableHeaders {
		cell, _ := excelize.CoordinatesToCellName(index+1, headerRow)
		_ = f.SetCellValue(sheet, cell, header)
	}
	_ = f.SetCellStyle(sheet, "A5", "Q5", headerStyle)
	_ = f.SetRowHeight(sheet, headerRow, 42)

	for index, row := range rows {
		excelRow := headerRow + index + 1
		values := []interface{}{
			joinNonEmpty("\n", row.InspectionID, row.Area),
			row.Kawasan,
			row.DetailKawasan,
			row.Aspek,
			row.DetailAspek,
			row.Uraian,
			row.Nilai,
			row.TotalNilaiKawasan,
			row.CompliancePercent / 100,
			row.IssueUraianCount,
			imagePlaceholder(row.InitialImages),
			imagePlaceholder(row.FollowUpImages),
			fallbackDash(row.FollowUpDescription),
			fallbackDash(row.KeteranganTemuan),
			fallbackDash(row.FollowUpDate),
			fallbackDash(row.DueDate),
			formatGMPGap(row.FollowUpGapDays),
		}
		for column, value := range values {
			cell, _ := excelize.CoordinatesToCellName(column+1, excelRow)
			_ = f.SetCellValue(sheet, cell, value)
		}

		_ = f.SetCellStyle(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("Q%d", excelRow), bodyStyle)
		for _, column := range []string{"G", "H", "J", "K", "L", "O", "P", "Q"} {
			_ = f.SetCellStyle(sheet, fmt.Sprintf("%s%d", column, excelRow), fmt.Sprintf("%s%d", column, excelRow), centerStyle)
		}
		_ = f.SetCellStyle(sheet, fmt.Sprintf("I%d", excelRow), fmt.Sprintf("I%d", excelRow), percentStyle)

		if len(row.InitialImages) > 0 {
			_ = f.SetCellValue(sheet, fmt.Sprintf("K%d", excelRow), "")
			insertImages(f, sheet, fmt.Sprintf("K%d", excelRow), row.InitialImages, excelRow)
		}
		if len(row.FollowUpImages) > 0 {
			_ = f.SetCellValue(sheet, fmt.Sprintf("L%d", excelRow), "")
			insertImages(f, sheet, fmt.Sprintf("L%d", excelRow), row.FollowUpImages, excelRow)
		}

		if row.FollowUpGapDays != nil {
			gapCell := fmt.Sprintf("Q%d", excelRow)
			switch {
			case *row.FollowUpGapDays > 0:
				_ = f.SetCellStyle(sheet, gapCell, gapCell, lateStyle)
			case *row.FollowUpGapDays < 0:
				_ = f.SetCellStyle(sheet, gapCell, gapCell, earlyStyle)
			default:
				_ = f.SetCellStyle(sheet, gapCell, gapCell, onTimeStyle)
			}
		}
	}
	// Columns A-F mirror the hierarchical rowspan merge on the Data GMP web
	// table (ID Inspeksi > Kawasan+Detail Kawasan > Aspek > Detail Aspek >
	// Uraian) — each level's merge is scoped by InspectionID (and every
	// ancestor level above it) so it only ever merges rows belonging to the
	// SAME inspection, never spanning across separate inspection sessions.
	// Rows arrive already sorted this way (gmpDataRelationOrder), so this
	// only merges cells that are already consecutive.
	if err := mergeConsecutiveGMPCells(f, sheet, rows, "A", func(row GMPTableRow) string {
		return normalizedGMPGroupKey(row.InspectionID)
	}); err != nil {
		return nil, fmt.Errorf("gagal menggabungkan ID Inspeksi: %w", err)
	}
	if err := mergeConsecutiveGMPCells(f, sheet, rows, "B", func(row GMPTableRow) string {
		return normalizedGMPGroupKey(row.InspectionID)
	}); err != nil {
		return nil, fmt.Errorf("gagal menggabungkan Kawasan: %w", err)
	}
	if err := mergeConsecutiveGMPCells(f, sheet, rows, "C", func(row GMPTableRow) string {
		return normalizedGMPGroupKey(row.InspectionID)
	}); err != nil {
		return nil, fmt.Errorf("gagal menggabungkan Detail Kawasan: %w", err)
	}
	if err := mergeConsecutiveGMPCells(f, sheet, rows, "D", func(row GMPTableRow) string {
		return normalizedGMPGroupKey(row.InspectionID, row.Aspek)
	}); err != nil {
		return nil, fmt.Errorf("gagal menggabungkan Aspek: %w", err)
	}
	if err := mergeConsecutiveGMPCells(f, sheet, rows, "E", func(row GMPTableRow) string {
		return normalizedGMPGroupKey(row.InspectionID, row.Aspek, row.DetailAspek)
	}); err != nil {
		return nil, fmt.Errorf("gagal menggabungkan Detail Aspek: %w", err)
	}
	if err := mergeConsecutiveGMPCells(f, sheet, rows, "F", func(row GMPTableRow) string {
		return normalizedGMPGroupKey(row.InspectionID, row.Aspek, row.DetailAspek, row.Uraian)
	}); err != nil {
		return nil, fmt.Errorf("gagal menggabungkan Uraian: %w", err)
	}

	// H/I stay scoped by Kawasan/Detail Kawasan alone (not InspectionID) —
	// TotalNilaiKawasan/CompliancePercent are genuine cross-inspection
	// rollups (see dashboard_handler.go: "TotalNilai remains an aggregate
	// per kawasan"), so merging them across separate inspections of the
	// same Kawasan/Detail Kawasan correctly reflects what the number means.
	if err := mergeConsecutiveGMPCells(f, sheet, rows, "H", func(row GMPTableRow) string {
		return normalizedGMPGroupKey(fallbackGMPGroupID(row.KawasanID, row.Kawasan))
	}); err != nil {
		return nil, fmt.Errorf("gagal menggabungkan Total Nilai per Kawasan: %w", err)
	}
	if err := mergeConsecutiveGMPCells(f, sheet, rows, "I", func(row GMPTableRow) string {
		return normalizedGMPGroupKey(
			fallbackGMPGroupID(row.KawasanID, row.Kawasan),
			fallbackGMPGroupID(row.DetailKawasanID, row.DetailKawasan),
		)
	}); err != nil {
		return nil, fmt.Errorf("gagal menggabungkan Persentase Kepatuhan per Detail Kawasan: %w", err)
	}

	totalRow := headerRow + len(rows) + 1
	totalNilai := 0
	totalStandardScore := 0
	totalTemuan := 0
	for _, row := range rows {
		totalNilai += row.Nilai
		totalStandardScore += row.StandardScore
		totalTemuan += row.IssueUraianCount
	}
	if err := f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("F%d", totalRow)); err != nil {
		return nil, fmt.Errorf("gagal menggabungkan label total: %w", err)
	}
	_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("TOTAL KESELURUHAN (%d BARIS)", len(rows)))
	_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", totalRow), totalNilai)
	_ = f.SetCellValue(sheet, fmt.Sprintf("H%d", totalRow), totalNilai)
	if totalStandardScore > 0 {
		_ = f.SetCellValue(sheet, fmt.Sprintf("I%d", totalRow), float64(totalNilai)/float64(totalStandardScore))
	} else {
		_ = f.SetCellValue(sheet, fmt.Sprintf("I%d", totalRow), 0)
	}
	_ = f.SetCellValue(sheet, fmt.Sprintf("J%d", totalRow), totalTemuan)
	_ = f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("F%d", totalRow), totalLabelStyle)
	_ = f.SetCellStyle(sheet, fmt.Sprintf("G%d", totalRow), fmt.Sprintf("H%d", totalRow), totalValueStyle)
	_ = f.SetCellStyle(sheet, fmt.Sprintf("I%d", totalRow), fmt.Sprintf("I%d", totalRow), totalPercentStyle)
	_ = f.SetCellStyle(sheet, fmt.Sprintf("J%d", totalRow), fmt.Sprintf("Q%d", totalRow), totalValueStyle)
	_ = f.SetRowHeight(sheet, totalRow, 26)

	widths := map[string]float64{
		"A": 20, "B": 24, "C": 26, "D": 24, "E": 30, "F": 48,
		"G": 11, "H": 21, "I": 25, "J": 16, "K": 19, "L": 19,
		"M": 38, "N": 38, "O": 22, "P": 17, "Q": 24,
	}
	for column, width := range widths {
		_ = f.SetColWidth(sheet, column, column, width)
	}

	lastRow := headerRow
	if len(rows) > 0 {
		lastRow += len(rows)
	}
	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze: true, Split: true, YSplit: headerRow, TopLeftCell: "A6", ActivePane: "bottomLeft",
	})
	_ = f.AutoFilter(sheet, fmt.Sprintf("A%d:Q%d", headerRow, lastRow), nil)
	if sheetIndex, indexErr := f.GetSheetIndex(sheet); indexErr == nil {
		f.SetActiveSheet(sheetIndex)
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("gagal membuat file tabel GMP: %w", err)
	}
	return buffer, nil
}

func normalizedGMPGroupKey(parts ...string) string {
	normalized := make([]string, len(parts))
	for index, part := range parts {
		normalized[index] = strings.ToLower(strings.TrimSpace(part))
	}
	return strings.Join(normalized, "\x00")
}

func fallbackGMPGroupID(id, name string) string {
	if id = strings.TrimSpace(id); id != "" {
		return id
	}
	return name
}

func mergeConsecutiveGMPCells(
	f *excelize.File,
	sheet string,
	rows []GMPTableRow,
	column string,
	groupKey func(GMPTableRow) string,
) error {
	const firstDataRow = 6
	for start := 0; start < len(rows); {
		end := start + 1
		key := groupKey(rows[start])
		for end < len(rows) && groupKey(rows[end]) == key {
			end++
		}
		if end-start > 1 {
			if err := f.MergeCell(
				sheet,
				fmt.Sprintf("%s%d", column, firstDataRow+start),
				fmt.Sprintf("%s%d", column, firstDataRow+end-1),
			); err != nil {
				return err
			}
		}
		start = end
	}
	return nil
}

func tableBorders(color string) []excelize.Border {
	return []excelize.Border{
		{Type: "left", Color: color, Style: 1},
		{Type: "right", Color: color, Style: 1},
		{Type: "top", Color: color, Style: 1},
		{Type: "bottom", Color: color, Style: 1},
	}
}

func statusStyle(f *excelize.File, background, foreground string) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: foreground},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{background}, Pattern: 1},
		Border:    tableBorders("E2E8F0"),
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
}

func formatGMPFilterSummary(meta GMPTableMetadata) string {
	parts := []string{
		"Plant: " + fallbackAll(meta.Plant, "Semua Plant"),
		"Area: " + fallbackAll(meta.Area, "Semua Area"),
		"Kawasan: " + fallbackAll(meta.Kawasan, "Semua Kawasan"),
		"Detail: " + fallbackAll(meta.DetailKawasan, "Semua Detail Kawasan"),
	}
	if meta.StartDate != "" || meta.EndDate != "" {
		parts = append(parts, "Tanggal: "+fallbackAll(meta.StartDate, "awal")+" s.d. "+fallbackAll(meta.EndDate, "sekarang"))
	}
	if strings.TrimSpace(meta.Search) != "" {
		parts = append(parts, "Pencarian: "+strings.TrimSpace(meta.Search))
	}
	return strings.Join(parts, " | ")
}

func formatGMPGap(gap *int) string {
	if gap == nil {
		return "-"
	}
	if *gap > 0 {
		return fmt.Sprintf("+%d hari terlambat", *gap)
	}
	if *gap < 0 {
		return fmt.Sprintf("%d hari lebih cepat", -*gap)
	}
	return "Tepat waktu"
}

func imagePlaceholder(images []string) string {
	if len(images) == 0 {
		return "-"
	}
	return ""
}

func joinNonEmpty(separator string, values ...string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, separator)
}

func fallbackDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func fallbackAll(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
