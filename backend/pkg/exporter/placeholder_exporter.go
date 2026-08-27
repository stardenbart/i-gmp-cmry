package exporter

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	_ "golang.org/x/image/webp"
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
						imagePaths := getImagePaths(v)
						strVal := fmt.Sprintf("%v", v)
						if cellValue == matchedPlaceholder {
							if len(imagePaths) > 0 {
								insertImages(f, sheetName, cellAxis, imagePaths, excelRow)
								newVal = ""
							} else {
								matchedExact = true
								exactValue = v
							}
							break
						} else if strings.Contains(newVal, matchedPlaceholder) {
							if len(imagePaths) > 0 {
								insertImages(f, sheetName, cellAxis, imagePaths, excelRow)
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

// getImagePaths normalizes a placeholder value so one template cell can
// contain all evidence photos belonging to an uraian.
func getImagePaths(value interface{}) []string {
	var candidates []string
	switch typed := value.(type) {
	case string:
		candidates = []string{typed}
	case []string:
		candidates = typed
	case []interface{}:
		for _, item := range typed {
			if path, ok := item.(string); ok {
				candidates = append(candidates, path)
			}
		}
	}

	paths := make([]string, 0, len(candidates))
	seen := make(map[string]bool)
	for _, candidate := range candidates {
		if isImagePath(candidate) && !seen[candidate] {
			paths = append(paths, candidate)
			seen[candidate] = true
		}
	}
	return paths
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

// insertImages stacks all evidence photos vertically in the template's image
// cell. The row grows with the number of photos, keeping other report columns
// and the user-provided Excel layout intact.
func insertImages(f *excelize.File, sheet, cellAxis string, imagePaths []string, excelRow int) {
	if len(imagePaths) == 0 {
		return
	}
	requiredHeight := float64(len(imagePaths)) * 60
	if currentHeight, err := f.GetRowHeight(sheet, excelRow); err != nil || currentHeight < requiredHeight {
		_ = f.SetRowHeight(sheet, excelRow, requiredHeight)
	}
	for index, imagePath := range imagePaths {
		insertImageAt(f, sheet, cellAxis, imagePath, index*80)
	}
}

func insertImageAt(f *excelize.File, sheet, cellAxis, imagePath string, offsetY int) {
	data, extension, ok := loadImageData(imagePath)
	if !ok {
		return
	}
	data, extension, ok = normalizeExcelImage(data, extension)
	if !ok {
		return
	}

	scale := 0.15
	if config, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil && config.Width > 0 && config.Height > 0 {
		widthScale := 105.0 / float64(config.Width)
		heightScale := 70.0 / float64(config.Height)
		scale = widthScale
		if heightScale < scale {
			scale = heightScale
		}
	}

	picFormat := &excelize.GraphicOptions{
		OffsetX:         5,
		OffsetY:         5 + offsetY,
		ScaleX:          scale,
		ScaleY:          scale,
		PrintObject:     func() *bool { b := true; return &b }(),
		Locked:          func() *bool { b := false; return &b }(),
		LockAspectRatio: true,
		Positioning:     "oneCell",
	}

	_ = f.AddPictureFromBytes(sheet, cellAxis, &excelize.Picture{
		Extension: extension,
		File:      data,
		Format:    picFormat,
	})
}

// normalizeExcelImage detects the binary format instead of trusting the file
// name. Existing MinIO objects can have a .jpg name while containing WebP
// bytes; Excel does not support WebP pictures, so they are converted to PNG.
func normalizeExcelImage(data []byte, extension string) ([]byte, string, bool) {
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return data, extension, extension != ""
	}

	switch strings.ToLower(format) {
	case "webp":
		decoded, _, decodeErr := image.Decode(bytes.NewReader(data))
		if decodeErr != nil {
			return nil, "", false
		}
		var converted bytes.Buffer
		if encodeErr := png.Encode(&converted, decoded); encodeErr != nil {
			return nil, "", false
		}
		return converted.Bytes(), ".png", true
	case "jpeg":
		return data, ".jpg", true
	case "png":
		return data, ".png", true
	case "gif":
		return data, ".gif", true
	default:
		return data, extension, extension != ""
	}
}

func loadImageData(imagePath string) ([]byte, string, bool) {
	urlsToTry := imageURLsToTry(imagePath)
	if len(urlsToTry) > 0 {
		// Add host fallbacks for Docker vs local environment
		if strings.Contains(imagePath, "localhost:9000") {
			urlsToTry = append(urlsToTry, strings.ReplaceAll(imagePath, "localhost:9000", "minio:9000"))
			urlsToTry = append(urlsToTry, strings.ReplaceAll(imagePath, "localhost:9000", "127.0.0.1:9000"))
		} else if strings.Contains(imagePath, "minio:9000") {
			urlsToTry = append(urlsToTry, strings.ReplaceAll(imagePath, "minio:9000", "localhost:9000"))
			urlsToTry = append(urlsToTry, strings.ReplaceAll(imagePath, "minio:9000", "127.0.0.1:9000"))
		}

		client := &http.Client{Timeout: 15 * time.Second}
		for _, targetURL := range urlsToTry {
			resp, err := client.Get(targetURL)
			if err == nil {
				body, readErr := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				if resp.StatusCode == 200 {
					if readErr == nil && len(body) > 0 {
						ext := imageExtension(targetURL, resp.Header.Get("Content-Type"))
						return body, ext, true
					}
				}
			}
		}
		return nil, "", false
	}

	localPath := imagePath
	if strings.HasPrefix(localPath, "/") {
		localPath = "." + localPath
	}
	data, err := os.ReadFile(localPath)
	if err != nil && localPath != imagePath {
		data, err = os.ReadFile(imagePath)
	}
	if err != nil || len(data) == 0 {
		return nil, "", false
	}
	return data, imageExtension(imagePath, ""), true
}

// imageURLsToTry resolves both absolute image URLs and the relative MinIO URLs
// stored in Issue_Photo (for example /monitoring-audit-bucket/issues/...).
// The browser resolves the latter through a Next.js proxy, while an Excel
// export must fetch the object directly from MinIO inside the backend service.
func imageURLsToTry(imagePath string) []string {
	if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
		return []string{imagePath}
	}

	bucket := strings.TrimSpace(os.Getenv("MINIO_BUCKET"))
	if bucket == "" {
		bucket = "monitoring-audit-bucket"
	}
	bucketPath := "/" + strings.Trim(bucket, "/") + "/"
	pathIndex := strings.Index(imagePath, bucketPath)
	if pathIndex == -1 {
		return nil
	}
	objectPath := imagePath[pathIndex:]

	endpoint := strings.TrimSpace(os.Getenv("MINIO_ENDPOINT"))
	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	baseURL := strings.TrimRight(endpoint, "/")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		scheme := "http://"
		if strings.EqualFold(strings.TrimSpace(os.Getenv("MINIO_USE_SSL")), "true") {
			scheme = "https://"
		}
		baseURL = scheme + baseURL
	}

	urls := []string{baseURL + objectPath}
	if strings.Contains(baseURL, "minio:9000") {
		urls = append(urls, strings.Replace(baseURL, "minio:9000", "localhost:9000", 1)+objectPath)
	} else if strings.Contains(baseURL, "localhost:9000") {
		urls = append(urls, strings.Replace(baseURL, "localhost:9000", "minio:9000", 1)+objectPath)
	}
	return urls
}

func imageExtension(path, contentType string) string {
	cleanPath := path
	if idx := strings.Index(cleanPath, "?"); idx != -1 {
		cleanPath = cleanPath[:idx]
	}
	ext := strings.ToLower(filepath.Ext(cleanPath))
	if ext != "" {
		return ext
	}
	switch {
	case strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg"):
		return ".jpg"
	case strings.Contains(contentType, "gif"):
		return ".gif"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	default:
		return ".png"
	}
}
