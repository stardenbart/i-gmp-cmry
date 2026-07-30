package exporter

import (
	"bytes"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// PlaceholderPayload menyimpan data statis (Header) dan dinamis (Items)
type PlaceholderPayload struct {
	Headers map[string]interface{}
	Items   []map[string]interface{}
}

// GenerateExcelWithPlaceholder adalah engine utama pencarian dan penggantian (search & replace)
func GenerateExcelWithPlaceholder(templatePath, sheetName string, payload *PlaceholderPayload) (*bytes.Buffer, error) {
	f, err := excelize.OpenFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka template: %w", err)
	}
	defer f.Close()

	if sheetName == "" {
		sheetName = f.GetSheetName(f.GetActiveSheetIndex())
	}

	// 1. Cari baris mana yang merupakan "Template Row" tabel (mengandung "{item.")
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca baris: %w", err)
	}

	// Build a set of item key placeholders to detect template row
	itemKeySet := make(map[string]bool)
	if len(payload.Items) > 0 {
		for k := range payload.Items[0] {
			itemKeySet[fmt.Sprintf("{item.%s}", k)] = true
			itemKeySet[fmt.Sprintf("{%s}", k)] = true
		}
	}

	templateRowIdx := -1
	for rIdx, row := range rows {
		for _, cellValue := range row {
			if strings.Contains(cellValue, "{item.") || itemKeySet[cellValue] {
				templateRowIdx = rIdx + 1 // Excel menggunakan indeks 1-based
				break
			}
		}
		if templateRowIdx != -1 {
			break
		}
	}

	// 2. Clone Baris Tabel (jika ditemukan)
	if templateRowIdx != -1 {
		itemsCount := len(payload.Items)
		if itemsCount > 0 {
			// DuplicateRow akan menyisipkan baris berformat persis ke bawahnya, mendorong footer turun.
			for i := 1; i < itemsCount; i++ {
				if err := f.DuplicateRow(sheetName, templateRowIdx); err != nil {
					return nil, fmt.Errorf("gagal kloning baris template: %w", err)
				}
			}
		} else {
			// Jika tidak ada data item sama sekali, hapus baris template agar rapi
			f.RemoveRow(sheetName, templateRowIdx)
		}
	}

	// 3. Ambil ulang seluruh baris setelah proses duplikasi
	rows, _ = f.GetRows(sheetName)

	// 4. Proses Search & Replace Data Header dan Data Tabel
	for rIdx, row := range rows {
		excelRow := rIdx + 1
		// Cek apakah baris ini adalah rentang baris tabel item yang harus di-inject
		isItemRow := templateRowIdx != -1 && len(payload.Items) > 0 &&
			excelRow >= templateRowIdx && excelRow < templateRowIdx+len(payload.Items)

		itemIndex := excelRow - templateRowIdx

		for cIdx, cellValue := range row {
			if cellValue == "" {
				continue // Abaikan sel kosong
			}

			excelCol, _ := excelize.ColumnNumberToName(cIdx + 1)
			cellAxis := excelCol + strconv.Itoa(excelRow)
			newVal := cellValue
			matchedExact := false
			var exactValue interface{}

			// A. Replace Placeholder Item Tabel – support both {item.key} and {key} format
			if isItemRow {
				for k, v := range payload.Items[itemIndex] {
					placeholderWithItem := fmt.Sprintf("{item.%s}", k)
					placeholderDirect := fmt.Sprintf("{%s}", k)

					matched := false
					var matchedPlaceholder string
					if cellValue == placeholderWithItem || strings.Contains(newVal, placeholderWithItem) {
						matchedPlaceholder = placeholderWithItem
						matched = true
					} else if cellValue == placeholderDirect || strings.Contains(newVal, placeholderDirect) {
						matchedPlaceholder = placeholderDirect
						matched = true
					}

					if matched {
						strVal := fmt.Sprintf("%v", v)
						if cellValue == matchedPlaceholder {
							if isImagePath(strVal) {
								insertImage(f, sheetName, cellAxis, strVal, excelRow)
								newVal = ""
							} else {
								matchedExact = true
								exactValue = v
							}
							break
						} else if strings.Contains(newVal, matchedPlaceholder) {
							if isImagePath(strVal) {
								insertImage(f, sheetName, cellAxis, strVal, excelRow)
								newVal = strings.ReplaceAll(newVal, matchedPlaceholder, "")
							} else {
								newVal = strings.ReplaceAll(newVal, matchedPlaceholder, strVal)
							}
						}
					}
				}
			}

			// B. Replace Placeholder Header (selain {item.})
			if !matchedExact {
				for k, v := range payload.Headers {
					placeholder := fmt.Sprintf("{%s}", k)
					if cellValue == placeholder {
						matchedExact = true
						exactValue = v
						break
					} else if strings.Contains(newVal, placeholder) {
						newVal = strings.ReplaceAll(newVal, placeholder, fmt.Sprintf("%v", v))
					}
				}
			}

			// 5. Tulis Hasil Replace ke Excel
			if matchedExact {
				f.SetCellValue(sheetName, cellAxis, exactValue)
			} else if newVal != cellValue {
				f.SetCellValue(sheetName, cellAxis, newVal)
			}
		}
	}

	// Return sebagai Buffer
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("gagal merubah ke buffer: %w", err)
	}

	return buf, nil
}

// Fungsi helper mendeteksi apakah suatu string adalah path gambar
func isImagePath(path string) bool {
	if path == "" {
		return false
	}
	// Strip query parameters for extension check
	cleanPath := path
	if idx := strings.Index(cleanPath, "?"); idx != -1 {
		cleanPath = cleanPath[:idx]
	}
	ext := strings.ToLower(filepath.Ext(cleanPath))
	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" {
		return true
	}
	// For URLs without extension (e.g. MinIO paths like /bucket/issues/ISS-001/17398...)
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return true
	}
	return false
}

// Fungsi helper untuk menyisipkan gambar
func insertImage(f *excelize.File, sheet, cellAxis, imagePath string, excelRow int) {
	// Set row height to give picture room to display
	_ = f.SetRowHeight(sheet, excelRow, 60)

	picFormat := &excelize.GraphicOptions{
		AutoFit:         true,
		OffsetX:         5,
		OffsetY:         5,
		PrintObject:     func() *bool { b := true; return &b }(),
		Locked:          func() *bool { b := false; return &b }(),
		LockAspectRatio: true,
	}

	if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
		urlsToTry := []string{imagePath}
		// Add host fallbacks for Docker vs local environment
		if strings.Contains(imagePath, "localhost:9000") {
			urlsToTry = append(urlsToTry, strings.ReplaceAll(imagePath, "localhost:9000", "minio:9000"))
			urlsToTry = append(urlsToTry, strings.ReplaceAll(imagePath, "localhost:9000", "127.0.0.1:9000"))
		} else if strings.Contains(imagePath, "minio:9000") {
			urlsToTry = append(urlsToTry, strings.ReplaceAll(imagePath, "minio:9000", "localhost:9000"))
			urlsToTry = append(urlsToTry, strings.ReplaceAll(imagePath, "minio:9000", "127.0.0.1:9000"))
		}

		for _, targetURL := range urlsToTry {
			resp, err := http.Get(targetURL)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == 200 {
					body, readErr := io.ReadAll(resp.Body)
					if readErr == nil && len(body) > 0 {
						cleanPath := targetURL
						if idx := strings.Index(cleanPath, "?"); idx != -1 {
							cleanPath = cleanPath[:idx]
						}
						ext := strings.ToLower(filepath.Ext(cleanPath))
						if ext == "" {
							ct := resp.Header.Get("Content-Type")
							switch {
							case strings.Contains(ct, "jpeg") || strings.Contains(ct, "jpg"):
								ext = ".jpg"
							case strings.Contains(ct, "png"):
								ext = ".png"
							case strings.Contains(ct, "gif"):
								ext = ".gif"
							case strings.Contains(ct, "webp"):
								ext = ".webp"
							default:
								ext = ".png"
							}
						}

						errPic := f.AddPictureFromBytes(sheet, cellAxis, &excelize.Picture{
							Extension: ext,
							File:      body,
							Format:    picFormat,
						})
						if errPic == nil {
							return
						}
					}
				}
			}
		}
	}

	// Try relative or local filesystem path
	localPath := imagePath
	if strings.HasPrefix(localPath, "/") {
		localPath = "." + localPath
	}
	if err := f.AddPicture(sheet, cellAxis, localPath, picFormat); err != nil && localPath != imagePath {
		_ = f.AddPicture(sheet, cellAxis, imagePath, picFormat)
	}
}

