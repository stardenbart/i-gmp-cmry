package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func main() {
	f := excelize.NewFile()
	sheet := "Sheet1"

	// Corporate Header Styling
	styleHeader, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#0f172a"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Family: "Arial", Size: 16, Color: "#ffffff"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "F2", styleHeader)
	f.MergeCell(sheet, "A1", "F2")
	f.SetCellValue(sheet, "A1", "CORPORATE AUDIT REPORT - MONITORING SYSTEM")

	// Field Labels styling
	styleLabel, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#334155"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#f1f5f9"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "cccccc", Style: 1}, {Type: "right", Color: "cccccc", Style: 1},
			{Type: "top", Color: "cccccc", Style: 1}, {Type: "bottom", Color: "cccccc", Style: 1},
		},
	})

	// Placeholder styling (where data goes)
	styleData, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "#000000"},
		Border: []excelize.Border{
			{Type: "left", Color: "cccccc", Style: 1}, {Type: "right", Color: "cccccc", Style: 1},
			{Type: "top", Color: "cccccc", Style: 1}, {Type: "bottom", Color: "cccccc", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	// Row 4: Global Stats
	f.SetCellValue(sheet, "B4", "Total Inspeksi")
	f.SetCellStyle(sheet, "B4", "B4", styleLabel)
	f.SetCellStyle(sheet, "C4", "C4", styleData)
	f.SetCellValue(sheet, "C4", "[DATA]") // C4

	f.SetCellValue(sheet, "D4", "Inspeksi Berjalan")
	f.SetCellStyle(sheet, "D4", "D4", styleLabel)
	f.SetCellStyle(sheet, "E4", "E4", styleData)
	f.SetCellValue(sheet, "E4", "[DATA]") // E4

	// Row 6: Issue Stats
	f.SetCellValue(sheet, "B6", "Total Open Issues")
	f.SetCellStyle(sheet, "B6", "B6", styleLabel)
	f.SetCellStyle(sheet, "C6", "C6", styleData)
	f.SetCellValue(sheet, "C6", "[DATA]") // C6

	f.SetCellValue(sheet, "D6", "Issue Overdue")
	f.SetCellStyle(sheet, "D6", "D6", styleLabel)
	f.SetCellStyle(sheet, "E6", "E6", styleData)
	f.SetCellValue(sheet, "E6", "[DATA]") // E6

	// Set column widths
	f.SetColWidth(sheet, "B", "E", 20)
	f.SetColWidth(sheet, "A", "A", 5)
	f.SetColWidth(sheet, "F", "F", 5)

	if err := f.SaveAs("templates/report_template.xlsx"); err != nil {
		fmt.Println("Error saving template:", err)
		return
	}
	fmt.Println("Template generated at templates/report_template.xlsx")
}
