package exporter

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

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
	for _, cell := range []string{"P9", "R9"} {
		pictures, pictureErr := workbook.GetPictures("Laporan WO-WR", cell)
		if pictureErr != nil || len(pictures) != 1 {
			t.Fatalf("expected an embedded picture in %s, got %d (%v)", cell, len(pictures), pictureErr)
		}
	}
}
