package exporter

// ExportPayload represents the data to be injected into the Excel template.
type ExportPayload struct {
	TemplateID string                 `json:"template_id"`
	Headers    map[string]interface{} `json:"headers"`
	TableData  []TableRow             `json:"table_data"`
}

// TableRow represents a single row in the dynamic table section.
type TableRow struct {
	Data      map[string]interface{} `json:"data"`
	ImagePath string                 `json:"image_path,omitempty"` // Local file path to the image to insert
}
