package inspection

// KawasanInspectionStatus defines the aggregated progress status for a
// single Kawasan (narrower than AreaInspectionStatus, which rolls up every
// Kawasan under an Area — one Area can contain several Kawasan).
type KawasanInspectionStatus string

const (
	KawasanStatusOpen       KawasanInspectionStatus = "Open"
	KawasanStatusOnProgress KawasanInspectionStatus = "OnProgress"
	KawasanStatusConfirmed  KawasanInspectionStatus = "Confirmed"
)

// KawasanProgress represents how many DetailKawasans in a Kawasan have been
// inspected (this month) vs the total number of DetailKawasans in it.
type KawasanProgress struct {
	KawasanID              string                  `json:"kawasan_id"`
	Status                 KawasanInspectionStatus `json:"status"`
	TotalDetailKawasan     int                     `json:"total_detail_kawasan"`
	CompletedDetailKawasan int                     `json:"completed_detail_kawasan"`
	ProgressPercent        float64                 `json:"progress_percent"`
}
