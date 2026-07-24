package exporter

import (
	"encoding/json"
	"os"
)

// TemplateConfig represents the configuration for exporting an Excel file based on a template.
type TemplateConfig struct {
	TemplateID    string            `json:"template_id"`
	TemplateFile  string            `json:"template_file"` // Path to the .xlsx master file
	SheetName     string            `json:"sheet_name"`
	HeaderMapping map[string]string `json:"header_mappings"`
	TableMapping  TableMapping      `json:"table_mapping"`
}

// TableMapping represents the configuration for the dynamic table section.
type TableMapping struct {
	StartRow    int               `json:"start_row"`
	ColumnMap   map[string]string `json:"column_map"`
	ImageColumn string            `json:"image_column,omitempty"`
}

// LoadConfig reads and parses the JSON configuration file.
func LoadConfig(filePath string) (*TemplateConfig, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config TemplateConfig
	if err := json.Unmarshal(fileBytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
