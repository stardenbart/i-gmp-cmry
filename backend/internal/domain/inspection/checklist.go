package inspection

// ChecklistUraian represents a single inspection point and its current result (if any).
type ChecklistUraian struct {
	UraianID      string            `json:"uraian_id"`
	UraianText    string            `json:"uraian_text"`
	StandardScore int               `json:"standard_score"`
	Result        *InspectionResult `json:"result,omitempty"` // populated if already checked
}

// ChecklistDetail groups uraians by Detail_Master
type ChecklistDetail struct {
	DetailID   string            `json:"detail_id"`
	DetailName string            `json:"detail_name"`
	Uraians    []ChecklistUraian `json:"uraians"`
}

// ChecklistAspek groups details by Aspek_Master
type ChecklistAspek struct {
	AspekID   string            `json:"aspek_id"`
	AspekName string            `json:"aspek_name"`
	Details   []ChecklistDetail `json:"details"`
}

// FullChecklist is the entire tree returned to the frontend.
type FullChecklist struct {
	InspectionID string           `json:"inspection_id"`
	AreaID       string           `json:"area_id"`
	Aspeks       []ChecklistAspek `json:"aspeks"`
}
