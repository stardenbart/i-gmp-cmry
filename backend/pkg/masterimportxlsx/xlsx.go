package masterimportxlsx

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/xuri/excelize/v2"
)

const (
	dataSheet        = "DATA_IMPORT"
	instructionSheet = "PETUNJUK"
	referenceSheet   = "REFERENSI"
	metaSheet        = "_META"
)

var headersByType = map[string][]string{
	master.ImportTypeAspek:  {"area_id", "aspek_name"},
	master.ImportTypeDetail: {"aspek_id", "detail_name"},
	master.ImportTypeUraian: {"detail_id", "uraian_text", "standard_score"},
	master.ImportTypeHEI:    {"category_name", "hei_code", "hei_name", "description", "status"},
}

type ReferenceRow []string

type ParsedRow struct {
	SourceRow int
	Values    map[string]string
}

type ParsedWorkbook struct {
	ImportType      string
	TemplateVersion string
	PlantID         string
	ReferenceAt     string
	Rows            []ParsedRow
}

func Headers(importType string) ([]string, error) {
	headers, ok := headersByType[importType]
	if !ok {
		return nil, fmt.Errorf("jenis import %q tidak didukung", importType)
	}
	return append([]string(nil), headers...), nil
}

func ValidateContainer(fileName string, content []byte) error {
	if strings.ToLower(filepath.Ext(fileName)) != ".xlsx" {
		return errors.New("hanya file .xlsx yang diperbolehkan")
	}
	if len(content) == 0 {
		return errors.New("file Excel kosong")
	}
	if len(content) > master.ImportMaxFileSize {
		return fmt.Errorf("ukuran file melebihi batas %d MB", master.ImportMaxFileSize/(1024*1024))
	}
	if len(content) < 4 || !bytes.Equal(content[:2], []byte("PK")) {
		return errors.New("isi file bukan workbook XLSX yang valid")
	}

	zr, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return errors.New("container XLSX rusak")
	}
	if len(zr.File) > 250 {
		return errors.New("workbook memiliki terlalu banyak komponen")
	}
	var totalUncompressed uint64
	for _, entry := range zr.File {
		name := strings.ToLower(entry.Name)
		totalUncompressed += entry.UncompressedSize64
		if entry.UncompressedSize64 > 20*1024*1024 || totalUncompressed > 50*1024*1024 {
			return errors.New("workbook melebihi batas ukuran hasil ekstraksi")
		}
		if strings.Contains(name, "vbaproject.bin") ||
			strings.Contains(name, "/externallinks/") ||
			strings.Contains(name, "/embeddings/") {
			return errors.New("macro, external link, dan embedded object tidak diperbolehkan")
		}
		if strings.HasSuffix(name, ".rels") && entry.UncompressedSize64 <= 1024*1024 {
			reader, openErr := entry.Open()
			if openErr != nil {
				return errors.New("relasi workbook tidak dapat diperiksa")
			}
			relationContent, readErr := io.ReadAll(io.LimitReader(reader, 1024*1024+1))
			_ = reader.Close()
			if readErr != nil {
				return errors.New("relasi workbook tidak dapat dibaca")
			}
			if strings.Contains(strings.ToLower(string(relationContent)), `targetmode="external"`) {
				return errors.New("external link tidak diperbolehkan")
			}
		}
	}
	return nil
}

func Parse(fileName string, content []byte, expectedType string) (*ParsedWorkbook, error) {
	if err := ValidateContainer(fileName, content); err != nil {
		return nil, err
	}
	headers, err := Headers(expectedType)
	if err != nil {
		return nil, err
	}

	f, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		return nil, errors.New("workbook tidak dapat dibaca")
	}
	defer f.Close()

	sheets := make(map[string]bool)
	for _, name := range f.GetSheetList() {
		sheets[name] = true
	}
	if !sheets[dataSheet] || !sheets[metaSheet] {
		return nil, errors.New("template tidak valid: sheet DATA_IMPORT atau metadata tidak ditemukan")
	}
	allowedSheets := map[string]bool{dataSheet: true, instructionSheet: true, referenceSheet: true, metaSheet: true}
	for name := range sheets {
		if !allowedSheets[name] {
			return nil, fmt.Errorf("sheet tambahan %q tidak diperbolehkan", name)
		}
	}
	merged, err := f.GetMergeCells(dataSheet)
	if err != nil {
		return nil, errors.New("struktur cell DATA_IMPORT tidak dapat diperiksa")
	}
	if len(merged) > 0 {
		return nil, errors.New("merged cell tidak diperbolehkan pada DATA_IMPORT")
	}

	actualType, _ := f.GetCellValue(metaSheet, "B1")
	version, _ := f.GetCellValue(metaSheet, "B2")
	plantID, _ := f.GetCellValue(metaSheet, "B3")
	referenceAt, _ := f.GetCellValue(metaSheet, "B4")
	if strings.TrimSpace(actualType) != expectedType {
		return nil, errors.New("jenis template tidak sesuai dengan tab import")
	}
	if strings.TrimSpace(version) != master.ImportTemplateV1 {
		return nil, errors.New("versi template sudah tidak didukung; unduh template terbaru")
	}

	rows, err := f.GetRows(dataSheet)
	if err != nil || len(rows) == 0 {
		return nil, errors.New("sheet DATA_IMPORT tidak dapat dibaca")
	}
	if len(rows[0]) != len(headers) {
		return nil, errors.New("jumlah kolom template tidak sesuai")
	}
	for i, header := range headers {
		if strings.TrimSpace(rows[0][i]) != header {
			return nil, fmt.Errorf("header kolom %d harus %q", i+1, header)
		}
	}

	parsed := make([]ParsedRow, 0, len(rows)-1)
	for rowIndex := 2; rowIndex <= len(rows); rowIndex++ {
		row := rows[rowIndex-1]
		values := make(map[string]string, len(headers))
		nonEmpty := false
		for colIndex, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex)
			formula, formulaErr := f.GetCellFormula(dataSheet, cell)
			if formulaErr != nil {
				return nil, fmt.Errorf("gagal memeriksa formula pada baris %d", rowIndex)
			}
			if formula != "" {
				return nil, fmt.Errorf("formula tidak diperbolehkan pada baris %d kolom %s", rowIndex, header)
			}
			value := ""
			if colIndex < len(row) {
				value = row[colIndex]
			}
			if strings.TrimSpace(value) != "" {
				nonEmpty = true
			}
			values[header] = value
		}
		if nonEmpty {
			parsed = append(parsed, ParsedRow{SourceRow: rowIndex, Values: values})
		}
	}
	if len(parsed) == 0 {
		return nil, errors.New("file tidak berisi data untuk diimport")
	}
	if len(parsed) > master.ImportMaxRows {
		return nil, fmt.Errorf("jumlah data melebihi batas %d baris", master.ImportMaxRows)
	}

	return &ParsedWorkbook{
		ImportType:      actualType,
		TemplateVersion: version,
		PlantID:         strings.TrimSpace(plantID),
		ReferenceAt:     strings.TrimSpace(referenceAt),
		Rows:            parsed,
	}, nil
}

func Generate(importType, plantID string, referenceHeaders []string, references []ReferenceRow) ([]byte, error) {
	headers, err := Headers(importType)
	if err != nil {
		return nil, err
	}
	f := excelize.NewFile()
	defer f.Close()

	dataIndex, err := f.NewSheet(dataSheet)
	if err != nil {
		return nil, err
	}
	if _, err = f.NewSheet(instructionSheet); err != nil {
		return nil, err
	}
	if _, err = f.NewSheet(referenceSheet); err != nil {
		return nil, err
	}
	if _, err = f.NewSheet(metaSheet); err != nil {
		return nil, err
	}
	f.DeleteSheet("Sheet1")
	f.SetActiveSheet(dataIndex)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"2563EB"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(dataSheet, cell, header)
		_ = f.SetCellStyle(dataSheet, cell, cell, headerStyle)
		_ = f.SetColWidth(dataSheet, string(rune('A'+i)), string(rune('A'+i)), columnWidth(header))
	}
	_ = f.SetPanes(dataSheet, &excelize.Panes{Freeze: true, Split: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	_ = f.AutoFilter(dataSheet, fmt.Sprintf("A1:%s1", string(rune('A'+len(headers)-1))), nil)

	instructions := []string{"Gunakan hanya template resmi dan jangan mengubah nama sheet atau header."}
	if importType == master.ImportTypeHEI {
		instructions = append(instructions,
			"category_name boleh memakai kategori yang sudah ada atau nama kategori baru.",
			"Daftar kategori pada sheet REFERENSI hanya contoh dan tidak membatasi nilai category_name.",
		)
	} else {
		instructions = append(instructions,
			"ID parent harus dipilih dari sheet REFERENSI.",
			"Template referensi adalah snapshot. Unduh ulang template setelah import parent berhasil.",
		)
	}
	instructions = append(instructions,
		"Seluruh baris harus valid; satu error akan membatalkan seluruh commit.",
		"Formula, macro, external link, dan embedded object tidak diperbolehkan.",
		"Perbedaan kapitalisasi atau spasi tidak membuat data menjadi unik.",
	)
	_ = f.SetCellValue(instructionSheet, "A1", "PETUNJUK IMPORT MASTER DATA")
	_ = f.SetCellStyle(instructionSheet, "A1", "A1", headerStyle)
	for i, instruction := range instructions {
		_ = f.SetCellValue(instructionSheet, fmt.Sprintf("A%d", i+3), fmt.Sprintf("%d. %s", i+1, instruction))
	}
	_ = f.SetColWidth(instructionSheet, "A", "A", 110)

	for i, header := range referenceHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(referenceSheet, cell, header)
		_ = f.SetCellStyle(referenceSheet, cell, cell, headerStyle)
		_ = f.SetColWidth(referenceSheet, string(rune('A'+i)), string(rune('A'+i)), 30)
	}
	for rowIndex, ref := range references {
		for colIndex, value := range ref {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+2)
			_ = f.SetCellValue(referenceSheet, cell, value)
		}
	}
	_ = f.SetPanes(referenceSheet, &excelize.Panes{Freeze: true, Split: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})

	generatedAt := time.Now().Format(time.RFC3339)
	metadata := [][2]string{
		{"import_type", importType},
		{"template_version", master.ImportTemplateV1},
		{"plant_id", plantID},
		{"reference_generated_at", generatedAt},
	}
	for i, item := range metadata {
		_ = f.SetCellValue(metaSheet, fmt.Sprintf("A%d", i+1), item[0])
		_ = f.SetCellValue(metaSheet, fmt.Sprintf("B%d", i+1), item[1])
	}
	_ = f.SetSheetVisible(metaSheet, false)

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func SortedSheetNames(content []byte) ([]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names := f.GetSheetList()
	sort.Strings(names)
	return names, nil
}

func ReadLimited(reader io.Reader) ([]byte, error) {
	content, err := io.ReadAll(io.LimitReader(reader, master.ImportMaxFileSize+1))
	if err != nil {
		return nil, err
	}
	if len(content) > master.ImportMaxFileSize {
		return nil, fmt.Errorf("ukuran file melebihi batas %d MB", master.ImportMaxFileSize/(1024*1024))
	}
	return content, nil
}

func columnWidth(header string) float64 {
	switch header {
	case "uraian_text", "description":
		return 60
	case "aspek_name", "detail_name", "hei_name":
		return 35
	default:
		return 24
	}
}
