package master

import "time"

const (
	ImportTypeAspek   = "aspek"
	ImportTypeDetail  = "detail"
	ImportTypeUraian  = "uraian"
	ImportTypeHEI     = "hei"
	ImportTemplateV1  = "1"
	ImportTokenTTL    = 20 * time.Minute
	ImportMaxFileSize = 5 * 1024 * 1024
	ImportMaxRows     = 5000
)

type ImportActor struct {
	UserID  string
	RoleID  string
	PlantID string
}

type MasterImportBatch struct {
	ImportID         string     `gorm:"column:ImportID;primaryKey" json:"import_id"`
	UserID           string     `gorm:"column:UserID" json:"user_id"`
	PlantID          *string    `gorm:"column:PlantID" json:"plant_id,omitempty"`
	ImportType       string     `gorm:"column:ImportType" json:"import_type"`
	OriginalFileName string     `gorm:"column:OriginalFileName" json:"original_file_name"`
	FileHash         string     `gorm:"column:FileHash" json:"file_hash"`
	TemplateVersion  string     `gorm:"column:TemplateVersion" json:"template_version"`
	TotalRows        int        `gorm:"column:TotalRows" json:"total_rows"`
	ValidRows        int        `gorm:"column:ValidRows" json:"valid_rows"`
	InvalidRows      int        `gorm:"column:InvalidRows" json:"invalid_rows"`
	InsertedRows     int        `gorm:"column:InsertedRows" json:"inserted_rows"`
	Status           string     `gorm:"column:Status" json:"status"`
	ErrorSummary     string     `gorm:"column:ErrorSummary" json:"error_summary,omitempty"`
	CreatedAt        time.Time  `gorm:"column:CreatedAt" json:"created_at"`
	ValidatedAt      *time.Time `gorm:"column:ValidatedAt" json:"validated_at,omitempty"`
	CommittedAt      *time.Time `gorm:"column:CommittedAt" json:"committed_at,omitempty"`
	UpdatedAt        time.Time  `gorm:"column:UpdatedAt" json:"updated_at"`
}

func (MasterImportBatch) TableName() string { return "Master_Import_Batch" }

type MasterImportRow struct {
	ImportID   string `gorm:"column:ImportID;primaryKey"`
	EntityType string `gorm:"column:EntityType;primaryKey"`
	EntityID   string `gorm:"column:EntityID;primaryKey"`
	SourceRow  int    `gorm:"column:SourceRow"`
}

func (MasterImportRow) TableName() string { return "Master_Import_Row" }

type ImportFieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ImportPreviewRow struct {
	Row            int                `json:"row"`
	Status         string             `json:"status"`
	NormalizedData map[string]any     `json:"normalized_data"`
	Errors         []ImportFieldError `json:"errors"`
}

type ImportSummary struct {
	Total               int `json:"total"`
	Valid               int `json:"valid"`
	Invalid             int `json:"invalid"`
	DuplicateInFile     int `json:"duplicate_in_file"`
	DuplicateInDatabase int `json:"duplicate_in_database"`
}

type ImportPreview struct {
	ImportID        string             `json:"import_id"`
	Type            string             `json:"type"`
	TemplateVersion string             `json:"template_version"`
	FileHash        string             `json:"file_hash"`
	ReferenceAt     string             `json:"reference_generated_at,omitempty"`
	Summary         ImportSummary      `json:"summary"`
	Rows            []ImportPreviewRow `json:"rows"`
	CanCommit       bool               `json:"can_commit"`
	ValidationToken string             `json:"validation_token,omitempty"`
}

type ImportCreatedRow struct {
	SourceRow int            `json:"row"`
	EntityID  string         `json:"entity_id"`
	Data      map[string]any `json:"data"`
}

type ImportCommitResult struct {
	ImportID       string             `json:"import_id"`
	Type           string             `json:"type"`
	Inserted       int                `json:"inserted"`
	CreatedRows    []ImportCreatedRow `json:"created_rows"`
	NextImportType string             `json:"next_import_type,omitempty"`
}

type ImportTokenState struct {
	ImportID           string        `json:"import_id"`
	UserID             string        `json:"user_id"`
	PlantID            string        `json:"plant_id"`
	ImportType         string        `json:"import_type"`
	FileHash           string        `json:"file_hash"`
	TemplateVersion    string        `json:"template_version"`
	RowCount           int           `json:"row_count"`
	NormalizedRowsHash string        `json:"normalized_rows_hash"`
	Summary            ImportSummary `json:"summary"`
	Status             string        `json:"status"`
	ExpiresAt          time.Time     `json:"expires_at"`
}
