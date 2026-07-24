package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func main() {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	f.SetSheetName("Sheet1", sheetName)

	f.SetCellValue(sheetName, "B4", "Tanggal:")
	f.SetCellValue(sheetName, "B5", "Area:")
	f.SetCellValue(sheetName, "B6", "PIC:")

	f.SetCellValue(sheetName, "C4", "[TANGGAL]")
	f.SetCellValue(sheetName, "C5", "[AREA]")
	f.SetCellValue(sheetName, "C6", "[PIC]")

	// Headers for table row 9
	f.SetCellValue(sheetName, "D9", "Uraian ID")
	f.SetCellValue(sheetName, "G9", "Nilai")
	f.SetCellValue(sheetName, "K9", "Keterangan")

	if err := f.SaveAs("../templates/master_gmp.xlsx"); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Template created successfully!")
}
