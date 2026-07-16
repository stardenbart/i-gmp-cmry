package inspection

// AreaInspectionStatus defines the aggregated progress status for an entire Area.
type AreaInspectionStatus string

const (
	AreaStatusOpen       AreaInspectionStatus = "Open"
	AreaStatusOnProgress AreaInspectionStatus = "OnProgress"
	AreaStatusConfirmed  AreaInspectionStatus = "Confirmed"
)

// AreaProgress represents the calculation of how many DetailKawasans in an Area
// have been inspected vs the total number of DetailKawasans in that Area.
type AreaProgress struct {
	AreaID                 string               `json:"area_id"`
	Status                 AreaInspectionStatus `json:"status"`
	TotalDetailKawasan     int                  `json:"total_detail_kawasan"`
	CompletedDetailKawasan int                  `json:"completed_detail_kawasan"`
	ProgressPercent        float64              `json:"progress_percent"`
}
