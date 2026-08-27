package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

func main() {
	output := flag.String("output", "templates/wowr_report.xlsx", "output XLSX path")
	flag.Parse()
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		panic(err)
	}

	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Laporan WO-WR"
	f.SetSheetName("Sheet1", sheet)

	border := []excelize.Border{
		{Type: "left", Color: "CBD5E1", Style: 1},
		{Type: "right", Color: "CBD5E1", Style: 1},
		{Type: "top", Color: "CBD5E1", Style: 1},
		{Type: "bottom", Color: "CBD5E1", Style: 1},
	}
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Family: "Arial", Size: 20, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"17365D"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	metaStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "Arial", Size: 10, Color: "334155"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"EAF2F8"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border:    border,
	})
	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Family: "Arial", Size: 10, Color: "17365D"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D9EAF7"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    border,
	})
	sectionStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Family: "Arial", Size: 11, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"2F75B5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Family: "Arial", Size: 9, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1F4E78"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    border,
	})
	itemStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "Arial", Size: 9, Color: "1E293B"},
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
		Border:    border,
	})
	centerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "Arial", Size: 9, Color: "1E293B"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "top", WrapText: true},
		Border:    border,
	})

	f.MergeCell(sheet, "A1", "S2")
	f.SetCellValue(sheet, "A1", "LAPORAN DETAIL WORK ORDER / WORK REQUEST")
	f.SetCellStyle(sheet, "A1", "S2", titleStyle)
	f.SetRowHeight(sheet, 1, 28)
	f.SetRowHeight(sheet, 2, 16)

	f.MergeCell(sheet, "A3", "F3")
	f.MergeCell(sheet, "G3", "L3")
	f.MergeCell(sheet, "M3", "S3")
	f.SetCellValue(sheet, "A3", "Dibuat: {generated_at}")
	f.SetCellValue(sheet, "G3", "Area: {area_filter}")
	f.SetCellValue(sheet, "M3", "Periode: {period_filter}")
	f.SetCellStyle(sheet, "A3", "S3", metaStyle)

	summaries := []struct{ cell, value string }{
		{"A5", "TOTAL\n{total}"}, {"D5", "TERVERIFIKASI\n{verified}"},
		{"G5", "MENUNGGU VALIDASI\n{pending}"}, {"J5", "DITOLAK\n{rejected}"},
		{"M5", "MENUNGGU BUKTI\n{awaiting}"},
	}
	for _, summary := range summaries {
		endCol := map[string]string{"A5": "C5", "D5": "F5", "G5": "I5", "J5": "L5", "M5": "S5"}[summary.cell]
		f.MergeCell(sheet, summary.cell, endCol)
		f.SetCellValue(sheet, summary.cell, summary.value)
	}
	f.SetCellStyle(sheet, "A5", "S5", summaryStyle)
	f.SetRowHeight(sheet, 5, 34)

	f.MergeCell(sheet, "A7", "S7")
	f.SetCellValue(sheet, "A7", "DETAIL TEMUAN DAN BUKTI PENYELESAIAN")
	f.SetCellStyle(sheet, "A7", "S7", sectionStyle)
	f.SetRowHeight(sheet, 7, 22)

	headers := []string{
		"No", "Issue ID", "Referensi Foto", "Nomor WO", "Nomor WR", "Status WO/WR",
		"Status Issue", "Lokasi", "PIC", "Detail Temuan", "Detail HEI", "Keterangan",
		"Tenggat", "Dibuat", "Rincian Bukti Awal", "Foto Temuan Awal",
		"Rincian Bukti Penyelesaian", "Foto Penyelesaian WO/WR", "Kelengkapan Bukti",
	}
	for index, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(index+1, 8)
		f.SetCellValue(sheet, cell, header)
	}
	f.SetCellStyle(sheet, "A8", "S8", headerStyle)
	f.SetRowHeight(sheet, 8, 38)

	placeholders := []string{
		"{item.no}", "{item.issue_id}", "{item.photo_id}", "{item.wo_number}", "{item.wr_number}",
		"{item.wowr_status}", "{item.issue_status}", "{item.location}", "{item.pic_name}",
		"{item.finding_detail}", "{item.hei_detail}", "{item.description}", "{item.due_date}",
		"{item.created_at}", "{item.initial_details}", "{item.initial_images}",
		"{item.completion_details}", "{item.completion_images}",
		"Selesai {item.completion_count} / Awal {item.initial_count}",
	}
	for index, placeholder := range placeholders {
		cell, _ := excelize.CoordinatesToCellName(index+1, 9)
		f.SetCellValue(sheet, cell, placeholder)
	}
	f.SetCellStyle(sheet, "A9", "S9", itemStyle)
	for _, col := range []string{"A", "C", "D", "E", "F", "G", "M", "N", "S"} {
		f.SetCellStyle(sheet, col+"9", col+"9", centerStyle)
	}
	f.SetRowHeight(sheet, 9, 72)

	widths := map[string]float64{
		"A": 5, "B": 18, "C": 18, "D": 16, "E": 16, "F": 18, "G": 16,
		"H": 25, "I": 20, "J": 36, "K": 26, "L": 35, "M": 14, "N": 18,
		"O": 34, "P": 24, "Q": 34, "R": 24, "S": 17,
	}
	for col, width := range widths {
		f.SetColWidth(sheet, col, col, width)
	}
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 8, TopLeftCell: "A9", ActivePane: "bottomLeft"})
	f.AutoFilter(sheet, "A8:S9", nil)
	f.SetSheetView(sheet, 0, &excelize.ViewOptions{ShowGridLines: boolPointer(false)})
	f.SetPageLayout(sheet, &excelize.PageLayoutOptions{
		Size:        intPointer(8), // A3
		Orientation: stringPointer("landscape"),
		FitToWidth:  intPointer(1),
		FitToHeight: intPointer(0),
	})
	f.SetPageMargins(sheet, &excelize.PageLayoutMarginsOptions{
		Left: floatPointer(0.25), Right: floatPointer(0.25),
		Top: floatPointer(0.5), Bottom: floatPointer(0.5),
		Header: floatPointer(0.2), Footer: floatPointer(0.2),
		Horizontally: boolPointer(true),
	})

	if err := f.SaveAs(*output); err != nil {
		panic(err)
	}
	fmt.Println("Template WO/WR generated:", *output)
}

func boolPointer(value bool) *bool        { return &value }
func intPointer(value int) *int           { return &value }
func stringPointer(value string) *string  { return &value }
func floatPointer(value float64) *float64 { return &value }
