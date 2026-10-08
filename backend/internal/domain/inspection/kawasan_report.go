package inspection

import "time"

// KawasanReportRow is one inspection result of the latest Completed/Approved
// inspection of each Detail Kawasan in a Kawasan, for the period's report
// email. Issue fields are set only when the result produced a finding.
type KawasanReportRow struct {
	AreaName          string     `gorm:"column:area_name"`
	KawasanName       string     `gorm:"column:kawasan_name"`
	DetailKawasanID   string     `gorm:"column:detail_kawasan_id"`
	DetailKawasanName string     `gorm:"column:detail_kawasan_name"`
	Nilai             int        `gorm:"column:nilai"`
	StandardScore     int        `gorm:"column:standard_score"`
	UraianText        string     `gorm:"column:uraian_text"`
	IssueID           *string    `gorm:"column:issue_id"`
	Keterangan        string     `gorm:"column:keterangan"` // encrypted at rest
	IssueStatus       string     `gorm:"column:issue_status"`
	DueDate           *time.Time `gorm:"column:due_date"`
	FollowUpBy        *string    `gorm:"column:follow_up_by"`
	FollowUpDate      *time.Time `gorm:"column:follow_up_date"`
}

type KawasanReportRepository interface {
	FindKawasanReportRows(kawasanID string, periodStart, periodEnd time.Time) ([]KawasanReportRow, error)
}
