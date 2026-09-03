package masterimportxlsx

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/xuri/excelize/v2"
)

func TestGenerateAndParseAspekTemplate(t *testing.T) {
	content, err := Generate(master.ImportTypeAspek, "PLT-001", []string{"area_id", "area_name"}, []ReferenceRow{{"AREA-001", "Produksi"}})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer f.Close()
	_ = f.SetCellValue(dataSheet, "A2", "AREA-001")
	_ = f.SetCellValue(dataSheet, "B2", "  Kebersihan   Area  ")
	buffer, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("WriteToBuffer() error = %v", err)
	}

	parsed, err := Parse("aspek.xlsx", buffer.Bytes(), master.ImportTypeAspek)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.PlantID != "PLT-001" || len(parsed.Rows) != 1 {
		t.Fatalf("unexpected parsed workbook: %+v", parsed)
	}
	if parsed.Rows[0].Values["aspek_name"] != "  Kebersihan   Area  " {
		t.Fatalf("parser must preserve raw value for database normalization")
	}
}

func TestParseRejectsFormula(t *testing.T) {
	content, err := Generate(master.ImportTypeAspek, "PLT-001", []string{"area_id"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_ = f.SetCellValue(dataSheet, "A2", "AREA-001")
	_ = f.SetCellFormula(dataSheet, "B2", `="Kebersihan"`)
	buffer, _ := f.WriteToBuffer()

	_, err = Parse("aspek.xlsx", buffer.Bytes(), master.ImportTypeAspek)
	if err == nil || !strings.Contains(err.Error(), "formula") {
		t.Fatalf("expected formula rejection, got %v", err)
	}
}

func TestValidateContainerRejectsMacroPayload(t *testing.T) {
	var buffer bytes.Buffer
	zw := zip.NewWriter(&buffer)
	entry, _ := zw.Create("xl/vbaProject.bin")
	_, _ = entry.Write([]byte("macro"))
	_ = zw.Close()

	err := ValidateContainer("payload.xlsx", buffer.Bytes())
	if err == nil || !strings.Contains(err.Error(), "macro") {
		t.Fatalf("expected macro rejection, got %v", err)
	}
}

func TestParseRejectsWrongTemplateType(t *testing.T) {
	content, err := Generate(master.ImportTypeAspek, "PLT-001", []string{"area_id"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Parse("aspek.xlsx", content, master.ImportTypeDetail)
	if err == nil || !strings.Contains(err.Error(), "jenis template") {
		t.Fatalf("expected template type rejection, got %v", err)
	}
}

func TestGenerateHEITemplateExplainsOpenCategories(t *testing.T) {
	content, err := Generate(
		master.ImportTypeHEI,
		"",
		[]string{"existing_category_name_optional"},
		[]ReferenceRow{{"Habit"}},
	)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows(instructionSheet)
	if err != nil {
		t.Fatalf("GetRows() error = %v", err)
	}
	var instructions []string
	for _, row := range rows {
		instructions = append(instructions, row...)
	}
	joined := strings.Join(instructions, " ")
	if !strings.Contains(joined, "kategori baru") || !strings.Contains(joined, "hanya contoh") {
		t.Fatalf("HEI template must explain open categories, got %q", joined)
	}
}
