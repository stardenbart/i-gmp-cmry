package exporter

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestGenerateGMPTemplateKeepsEveryRelationOnItsOwnRow(t *testing.T) {
	items := make([]map[string]interface{}, 0, 6)
	for index := 1; index <= 6; index++ {
		items = append(items, map[string]interface{}{
			"no":                    index,
			"aspekName":             fmt.Sprintf("Aspek %d", index),
			"detailAspekName":       fmt.Sprintf("Detail %d", index),
			"uraianName":            fmt.Sprintf("Uraian %d", index),
			"nilai":                 index,
			"total_nilai_peraspek":  index,
			"total_temuan_peraspek": 0,
			"imageUrl":              "",
			"keterangan":            "",
			"followUp":              "",
			"dueDate":               "",
		})
	}

	buffer, err := GenerateExcelWithPlaceholder("../../templates/master_gmp.xlsx", "Rev 00", &PlaceholderPayload{
		Headers: map[string]interface{}{
			"AreaName":           "Semua Area",
			"PICName":            "Semua PIC",
			"DueDate":            "",
			"datetime.now()":     "03-Sep-26",
			"total_semua_nilai":  21,
			"total_semua_temuan": 0,
		},
		Items:                items,
		FlattenItemRows:      true,
		TrimReservedItemRows: true,
		FixedCells: map[string]interface{}{
			"C1": "PT CISARUA MOUNTAIN DAIRY TBK\nPLANT SENTUL UTAMA",
			"K3": "",
		},
		LogoCell:      "A1",
		LogoPaths:     []string{"../../../frontend/public/Logo_Cimory.png"},
		RemoveColumns: []string{"D"},
	})
	if err != nil {
		t.Fatalf("generate GMP workbook: %v", err)
	}

	workbook, err := excelize.OpenReader(buffer)
	if err != nil {
		t.Fatalf("open generated GMP workbook: %v", err)
	}
	defer workbook.Close()

	for index := 1; index <= 6; index++ {
		row := 9 + index
		assertions := map[string]string{
			fmt.Sprintf("B%d", row): fmt.Sprintf("Aspek %d", index),
			fmt.Sprintf("C%d", row): fmt.Sprintf("Detail %d", index),
			fmt.Sprintf("D%d", row): fmt.Sprintf("Uraian %d", index),
		}
		for cell, expected := range assertions {
			if value, _ := workbook.GetCellValue("Rev 00", cell); value != expected {
				t.Fatalf("expected %s in %s, got %q", expected, cell, value)
			}
		}
	}
	if value, _ := workbook.GetCellValue("Rev 00", "E16"); value != "21" {
		t.Fatalf("expected footer immediately after the final item row, got total %q in E16", value)
	}
	if value, _ := workbook.GetCellValue("Rev 00", "D8"); value != "Uraian" {
		t.Fatalf("expected UraianID column to be removed, got header %q", value)
	}
	if value, _ := workbook.GetCellValue("Rev 00", "E8"); value != "Nilai" {
		t.Fatalf("expected Nilai column to shift left after removing UraianID, got %q", value)
	}
	if value, _ := workbook.GetCellValue("Rev 00", "C1"); value != "PT CISARUA MOUNTAIN DAIRY TBK\nPLANT SENTUL UTAMA" {
		t.Fatalf("unexpected company and plant header: %q", value)
	}
	if value, _ := workbook.GetCellValue("Rev 00", "K3"); value != "" {
		t.Fatalf("expected Berlaku value to be empty, got %q", value)
	}
	pictures, err := workbook.GetPictures("Rev 00", "A1")
	if err != nil || len(pictures) != 1 {
		t.Fatalf("expected Cimory logo in A1, got %d pictures (%v)", len(pictures), err)
	}
}

func TestGenerateGMPTemplateMergesDuplicateAspekWithinSameInspection(t *testing.T) {
	items := []map[string]interface{}{
		{"inspectionID": "INSP-001", "aspekName": "Aspek A", "detailAspekName": "Detail 1", "uraianID": "UR-001", "uraianName": "Uraian 1", "total_nilai_peraspek": 4, "total_temuan_peraspek": 2},
		{"inspectionID": "INSP-001", "aspekName": "Aspek A", "detailAspekName": "Detail 1", "uraianID": "UR-002", "uraianName": "Uraian 2", "total_nilai_peraspek": 4, "total_temuan_peraspek": 2},
		{"inspectionID": "INSP-001", "aspekName": "Aspek B", "detailAspekName": "Detail 3", "uraianID": "UR-003", "uraianName": "Uraian 3", "total_nilai_peraspek": 2, "total_temuan_peraspek": 0},
		{"inspectionID": "INSP-002", "aspekName": "Aspek A", "detailAspekName": "Detail 4", "uraianID": "UR-004", "uraianName": "Uraian 4", "total_nilai_peraspek": 2, "total_temuan_peraspek": 1},
	}

	buffer, err := GenerateExcelWithPlaceholder("../../templates/master_gmp.xlsx", "Rev 00", &PlaceholderPayload{
		Headers:                    map[string]interface{}{},
		Items:                      items,
		FlattenItemRows:            true,
		MergeDuplicateItemFields:   []string{"aspekName", "detailAspekName", "total_nilai_peraspek", "total_temuan_peraspek"},
		MergeDuplicateWithinFields: []string{"inspectionID", "aspekName"},
	})
	if err != nil {
		t.Fatalf("generate GMP workbook: %v", err)
	}

	workbook, err := excelize.OpenReader(buffer)
	if err != nil {
		t.Fatalf("open generated GMP workbook: %v", err)
	}
	defer workbook.Close()

	mergedRanges := make(map[string]bool)
	mergeCells, err := workbook.GetMergeCells("Rev 00")
	if err != nil {
		t.Fatalf("read merged cells: %v", err)
	}
	for _, mergedCell := range mergeCells {
		mergedRanges[mergedCell.GetStartAxis()+":"+mergedCell.GetEndAxis()] = true
	}
	if !mergedRanges["B10:B11"] {
		t.Fatalf("expected duplicate Aspek in the same inspection to merge, got %#v", mergedRanges)
	}
	if !mergedRanges["C10:C11"] {
		t.Fatalf("expected duplicate Detail Aspek to merge while Uraian rows remain separate, got %#v", mergedRanges)
	}
	if !mergedRanges["G10:G11"] || !mergedRanges["H10:H11"] {
		t.Fatalf("expected Aspek totals to use the same merged range, got %#v", mergedRanges)
	}
	if value, _ := workbook.GetCellValue("Rev 00", "G10"); value != "4" {
		t.Fatalf("expected merged Total Nilai Aspek to be 4, got %q", value)
	}
	if value, _ := workbook.GetCellValue("Rev 00", "H10"); value != "2" {
		t.Fatalf("expected merged uraian-based Total Temuan Aspek to be 2, got %q", value)
	}
	if mergedRanges["B10:B13"] || mergedRanges["B10:B14"] {
		t.Fatal("Aspek merge must stop when the aspect or inspection changes")
	}
	if value, _ := workbook.GetCellValue("Rev 00", "B12"); value != "Aspek B" {
		t.Fatalf("expected next Aspek to remain visible, got %q", value)
	}
	if value, _ := workbook.GetCellValue("Rev 00", "B13"); value != "Aspek A" {
		t.Fatalf("expected same Aspek in another inspection to remain separate, got %q", value)
	}
}

func TestNormalizeExcelImageConvertsWebPToPNG(t *testing.T) {
	webPBase64 := "UklGRrIBAABXRUJQVlA4TKUBAAAvSsAYAA8w//M///MfeJAkbXvaSG7m8Q3GfYSBJekwQztm/IcZlgwnmWImn2BK7aFmBtnVir6q//8VOkFE/xm4baTIu8c48ArEo6+B3zFKYln3pqClSCKX0begFTAXFOLXHSyF8cCNcZEG4OywuA4KVVfJCiArU7GAgJI8+lJP/OKMT/fBAjevg1cYB7YVkFuWga2lyPi5I0HFy5YTpWIHg0RZpkniRVW9odHAKOwosWuOGdxIyn2OvaCDvhg/we6TwadPBPbqBV58MsLmMJ8yZnOWk8SRz4N+QoyPL+MnamzMvcE1rHNEr91F9GKZPVUcS9w7PhhH36suB9qPeYb/oLk6cuTiJ0wOK3m5h1cKjW6EVZCYMK7dxcKCBdgP9HkKr9gkAO2P8GKZGWVdIAatQa+1IDpt6qyorVwdy01xdW8Jkfk6xjEXmVQQ+HQdFr6OKhIN34dXWq0+0qr6EJSCeeVLH9+gvGTLyqM65PQ44ihzlTXxQKjKbAvshXgir7Lil9w4L2bvMycmjQcqXaMCO6BlY28i+FOLzbfI1vEqxAhotocAAA=="
	webPData, err := base64.StdEncoding.DecodeString(webPBase64)
	if err != nil {
		t.Fatalf("decode WebP fixture: %v", err)
	}

	converted, extension, ok := normalizeExcelImage(webPData, ".jpg")
	if !ok {
		t.Fatal("expected WebP image to be normalized")
	}
	if extension != ".png" {
		t.Fatalf("expected PNG extension, got %q", extension)
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(converted))
	if err != nil {
		t.Fatalf("decode converted image: %v", err)
	}
	if format != "png" {
		t.Fatalf("expected PNG bytes, got %q", format)
	}
}

func TestImageURLsToTryResolvesRelativeMinIOPath(t *testing.T) {
	t.Setenv("MINIO_BUCKET", "test-bucket")
	t.Setenv("MINIO_ENDPOINT", "minio:9000")

	urls := imageURLsToTry("/test-bucket/issues/ISS-001/evidence.png")
	if len(urls) != 2 {
		t.Fatalf("expected internal and local MinIO URLs, got %#v", urls)
	}
	if urls[0] != "http://minio:9000/test-bucket/issues/ISS-001/evidence.png" {
		t.Fatalf("unexpected internal MinIO URL: %q", urls[0])
	}
	if urls[1] != "http://localhost:9000/test-bucket/issues/ISS-001/evidence.png" {
		t.Fatalf("unexpected local MinIO fallback URL: %q", urls[1])
	}
}

func TestGenerateExcelWithPlaceholderInsertsAllItemImages(t *testing.T) {
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "template.xlsx")

	template := excelize.NewFile()
	sheetName := "Rev 00"
	template.SetSheetName("Sheet1", sheetName)
	if err := template.SetCellValue(sheetName, "A1", "{imageUrl}"); err != nil {
		t.Fatalf("set template placeholder: %v", err)
	}
	if err := template.SaveAs(templatePath); err != nil {
		t.Fatalf("save template: %v", err)
	}
	_ = template.Close()

	imagePaths := make([]string, 0, 2)
	for index, fill := range []color.RGBA{{R: 255, A: 255}, {G: 255, A: 255}} {
		imagePath := filepath.Join(tempDir, "evidence-"+string(rune('1'+index))+".png")
		file, err := os.Create(imagePath)
		if err != nil {
			t.Fatalf("create image: %v", err)
		}
		canvas := image.NewRGBA(image.Rect(0, 0, 120, 80))
		for y := 0; y < 80; y++ {
			for x := 0; x < 120; x++ {
				canvas.Set(x, y, fill)
			}
		}
		if err := png.Encode(file, canvas); err != nil {
			_ = file.Close()
			t.Fatalf("encode image: %v", err)
		}
		_ = file.Close()
		imagePaths = append(imagePaths, imagePath)
	}

	buffer, err := GenerateExcelWithPlaceholder(templatePath, sheetName, &PlaceholderPayload{
		Headers: map[string]interface{}{},
		Items: []map[string]interface{}{
			{"imageUrl": imagePaths},
		},
	})
	if err != nil {
		t.Fatalf("generate workbook: %v", err)
	}

	workbook, err := excelize.OpenReader(buffer)
	if err != nil {
		t.Fatalf("open generated workbook: %v", err)
	}
	defer workbook.Close()

	pictures, err := workbook.GetPictures(sheetName, "A1")
	if err != nil {
		t.Fatalf("read generated pictures: %v", err)
	}
	if len(pictures) != len(imagePaths) {
		t.Fatalf("expected %d pictures, got %d", len(imagePaths), len(pictures))
	}
}

func TestWOWRReportTemplateRendersDetailsAndEvidence(t *testing.T) {
	templatePath := filepath.Join("..", "..", "templates", "wowr_report.xlsx")
	imagePath := filepath.Join(t.TempDir(), "evidence.png")
	file, err := os.Create(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, image.NewRGBA(image.Rect(0, 0, 120, 80))); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	_ = file.Close()

	buffer, err := GenerateExcelWithPlaceholder(templatePath, "Laporan WO-WR", &PlaceholderPayload{
		Headers: map[string]interface{}{
			"generated_at": "27 August 2026 10:00 WIB", "area_filter": "Produksi",
			"period_filter": "Semua periode", "total": 1, "verified": 1,
			"pending": 0, "rejected": 0, "awaiting": 0,
		},
		Items: []map[string]interface{}{{
			"no": 1, "issue_id": "ISS-001", "photo_id": "PHOTO-001",
			"wo_number": "WO-001", "wr_number": "-", "wowr_status": "Verified",
			"issue_status": "Closed", "location": "Produksi\nLine 1", "pic_name": "Teknisi",
			"finding_detail": "Kebersihan\nLantai", "hei_detail": "Infrastructure: Lantai",
			"description": "Lantai retak", "due_date": "30-Aug-2026", "created_at": "27-Aug-2026",
			"initial_details": "1. Bukti awal", "completion_details": "1. Bukti selesai",
			"initial_images": []string{imagePath}, "completion_images": []string{imagePath},
			"initial_count": 1, "completion_count": 1,
		}},
		TotalRow: &TableTotalRow{
			Label:            "TOTAL KESELURUHAN WO/WR",
			LabelStartColumn: "A",
			LabelEndColumn:   "R",
			ValueColumn:      "S",
			Value:            1,
		},
	})
	if err != nil {
		t.Fatalf("generate WO/WR workbook: %v", err)
	}

	workbook, err := excelize.OpenReader(buffer)
	if err != nil {
		t.Fatal(err)
	}
	defer workbook.Close()
	if value, _ := workbook.GetCellValue("Laporan WO-WR", "B9"); value != "ISS-001" {
		t.Fatalf("expected issue detail in B9, got %q", value)
	}
	if value, _ := workbook.GetCellValue("Laporan WO-WR", "S9"); value != "Selesai 1 / Awal 1" {
		t.Fatalf("expected evidence summary in S9, got %q", value)
	}
	if value, _ := workbook.GetCellValue("Laporan WO-WR", "A10"); value != "TOTAL KESELURUHAN WO/WR" {
		t.Fatalf("expected total label in A10, got %q", value)
	}
	if value, _ := workbook.GetCellValue("Laporan WO-WR", "S10"); value != "1" {
		t.Fatalf("expected WO/WR grand total in S10, got %q", value)
	}
	for _, cell := range []string{"P9", "R9"} {
		pictures, pictureErr := workbook.GetPictures("Laporan WO-WR", cell)
		if pictureErr != nil || len(pictures) != 1 {
			t.Fatalf("expected an embedded picture in %s, got %d (%v)", cell, len(pictures), pictureErr)
		}
	}
}
