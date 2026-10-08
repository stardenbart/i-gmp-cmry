package inspectionusecase

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"sort"
	"strings"
	"time"
	_ "time/tzdata" // Asia/Jakarta even without OS zoneinfo

	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/domain/pic"
)

// KawasanReport is the "Laporan Hasil Inspeksi GMP" email sent to every PIC
// of a Kawasan once all of its Detail Kawasan are inspected in the current
// period. Layout follows the former Power Automate flow (GMP_Report_Flow):
// inspection info, score per Detail Kawasan + overall average, the issue
// list, and a button back to the website, in the Cimory colours.
type KawasanReport struct {
	AreaName    string
	KawasanName string
	GeneratedAt time.Time
	Details     []DetailScore
	Issues      []ReportIssue
	// AppURL is the website page the main button opens (Temuan Inspeksi).
	AppURL string
}

// DetailScore is Σ Nilai over Σ StandardScore of one Detail Kawasan's latest
// inspection in the period, the same ratio the GMP data page uses.
type DetailScore struct {
	Name  string
	Score int
	Max   int
}

type ReportIssue struct {
	DetailKawasanName string
	Title             string // uraian text
	Keterangan        string
	PICName           string // everyone mapped to the Kawasan, with role
	DueDate           *time.Time
	Status            string // IssueStatus / computed status
	FollowUpBy        string
	FollowUpDate      *time.Time
	URL               string // issue detail page (photos live there)
}

const (
	reportNavy  = "#1b3a6f"
	reportGreen = "#1e8449"
	reportRed   = "#c62833"
	reportAmber = "#9c640c"
	// Same pass mark as the flow.
	reportPassPercent = 80
)

var indonesianMonths = [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

var reportLocation = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		return loc
	}
	return time.FixedZone("WIB", 7*60*60)
}()

// formatReportDate renders "08 Okt 2026" in WIB, whatever the server zone.
func formatReportDate(t time.Time) string {
	t = t.In(reportLocation)
	return fmt.Sprintf("%02d %s %d", t.Day(), indonesianMonths[t.Month()-1], t.Year())
}

// scorePercent rounds down, like the flow's integer div().
func scorePercent(score, max int) int {
	if max <= 0 {
		return 0
	}
	return score * 100 / max
}

func (r KawasanReport) Total() (score, max int) {
	for _, d := range r.Details {
		score += d.Score
		max += d.Max
	}
	return score, max
}

func (r KawasanReport) Subject() string {
	return fmt.Sprintf("[I-GMP] Laporan Inspeksi GMP - %s - %s", r.KawasanName, formatReportDate(r.GeneratedAt))
}

func (i ReportIssue) StatusLabel() string {
	switch i.Status {
	case "OpenOverdue":
		return "Open (Overdue)"
	case "InProgress":
		return "In Progress"
	case "PendingValidation":
		return "Menunggu Validasi"
	case "Verified", "Closed", "ClosedOverdue":
		return "Closed"
	default:
		return "Open"
	}
}

func (i ReportIssue) statusColor() template.CSS {
	switch i.StatusLabel() {
	case "Closed":
		return reportGreen
	case "In Progress", "Menunggu Validasi":
		return reportAmber
	default:
		return reportRed
	}
}

func scoreColor(score, max int) template.CSS {
	if scorePercent(score, max) >= reportPassPercent {
		return reportGreen
	}
	return reportRed
}

var kawasanReportTemplate = template.Must(template.New("kawasan-report").Funcs(template.FuncMap{
	"date":        formatReportDate,
	"percent":     scorePercent,
	"scoreColor":  scoreColor,
	"statusColor": func(i ReportIssue) template.CSS { return i.statusColor() },
}).Parse(`<div style="font-family:Segoe UI,Arial,sans-serif;max-width:660px;color:#1f2937;">
<div style="background:` + reportNavy + `;padding:20px 24px;border-bottom:4px solid #e63946;"><h2 style="color:#ffffff;margin:0;font-size:18pt;">I-GMP — Laporan Hasil Inspeksi GMP</h2></div>
<div style="padding:20px 24px;border:1px solid #e5e7eb;">
<p style="margin-top:0;">Yth. Tim PIC <b>{{.KawasanName}}</b>,</p>
<p>Berikut laporan hasil inspeksi GMP yang telah diselesaikan untuk seluruh detail dalam kawasan ini:</p>

<h3 style="color:` + reportNavy + `;border-bottom:2px solid ` + reportNavy + `;padding-bottom:4px;margin-bottom:8px;">Informasi Inspeksi</h3>
<table style="width:100%;border-collapse:collapse;margin-bottom:16px;">
<tr style="background:#f8fafc;"><td style="padding:8px;border:1px solid #e5e7eb;width:30%;"><b>Area</b></td><td style="padding:8px;border:1px solid #e5e7eb;">{{.AreaName}}</td></tr>
<tr><td style="padding:8px;border:1px solid #e5e7eb;"><b>Kawasan</b></td><td style="padding:8px;border:1px solid #e5e7eb;">{{.KawasanName}}</td></tr>
<tr style="background:#f8fafc;"><td style="padding:8px;border:1px solid #e5e7eb;"><b>Tanggal Generate</b></td><td style="padding:8px;border:1px solid #e5e7eb;">{{date .GeneratedAt}}</td></tr>
</table>

<h3 style="color:` + reportNavy + `;border-bottom:2px solid ` + reportNavy + `;padding-bottom:4px;margin-bottom:8px;">Ringkasan Skor per Detail Kawasan</h3>
<table style="width:100%;border-collapse:collapse;margin-bottom:16px;">
<tr style="background:` + reportNavy + `;"><th style="padding:8px;border:1px solid ` + reportNavy + `;color:#ffffff;text-align:left;">Detail Kawasan</th><th style="padding:8px;border:1px solid ` + reportNavy + `;color:#ffffff;text-align:center;">Skor</th><th style="padding:8px;border:1px solid ` + reportNavy + `;color:#ffffff;text-align:center;">Nilai (%)</th></tr>
{{range .Details}}<tr><td style="padding:8px;border:1px solid #e5e7eb;font-weight:bold;">{{.Name}}</td>{{if .Max}}<td style="padding:8px;border:1px solid #e5e7eb;text-align:center;">{{.Score}} / {{.Max}}</td><td style="padding:8px;border:1px solid #e5e7eb;text-align:center;font-weight:bold;color:{{scoreColor .Score .Max}};">{{percent .Score .Max}}</td>{{else}}<td style="padding:8px;border:1px solid #e5e7eb;text-align:center;">-</td><td style="padding:8px;border:1px solid #e5e7eb;text-align:center;color:#6b7280;">-</td>{{end}}</tr>
{{end}}{{template "total" .}}
</table>

{{if .Issues}}<h3 style="color:` + reportRed + `;border-bottom:2px solid ` + reportRed + `;padding-bottom:4px;margin-bottom:8px;">Daftar Issue ({{len .Issues}} issue)</h3>
<table style="width:100%;border-collapse:collapse;margin-bottom:16px;">
{{range .Issues}}<tr><td colspan="2" style="padding:6px 8px;background:#eaf2f8;color:` + reportNavy + `;font-style:italic;font-size:9pt;border-bottom:1px solid ` + reportNavy + `;">{{.DetailKawasanName}}</td></tr>
<tr><td colspan="2" style="padding:10px 8px;background:` + reportNavy + `;color:#ffffff;font-weight:bold;font-size:10pt;">{{if .Title}}{{.Title}}{{else}}-{{end}}</td></tr>
<tr style="background:#f8fafc;"><td style="padding:6px 8px;border:1px solid #e5e7eb;width:30%;"><b>Keterangan</b></td><td style="padding:6px 8px;border:1px solid #e5e7eb;">{{if .Keterangan}}{{.Keterangan}}{{else}}-{{end}}</td></tr>
<tr><td style="padding:6px 8px;border:1px solid #e5e7eb;"><b>PIC Kawasan</b></td><td style="padding:6px 8px;border:1px solid #e5e7eb;">{{if .PICName}}{{.PICName}}{{else}}-{{end}}</td></tr>
<tr style="background:#f8fafc;"><td style="padding:6px 8px;border:1px solid #e5e7eb;"><b>Due Date</b></td><td style="padding:6px 8px;border:1px solid #e5e7eb;color:` + reportRed + `;font-weight:bold;">{{if .DueDate}}{{date .DueDate}}{{else}}-{{end}}</td></tr>
<tr><td style="padding:6px 8px;border:1px solid #e5e7eb;"><b>Status</b></td><td style="padding:6px 8px;border:1px solid #e5e7eb;color:{{statusColor .}};font-weight:bold;">{{.StatusLabel}}</td></tr>
<tr style="background:#f8fafc;"><td style="padding:6px 8px;border:1px solid #e5e7eb;"><b>Foto Issue</b></td><td style="padding:6px 8px;border:1px solid #e5e7eb;">{{if .URL}}<a href="{{.URL}}" style="color:` + reportNavy + `;font-weight:bold;">Lihat Detail &amp; Foto</a>{{else}}-{{end}}</td></tr>
<tr><td style="padding:6px 8px;border:1px solid #e5e7eb;"><b>Follow Up</b></td><td style="padding:6px 8px;border:1px solid #e5e7eb;">{{if .FollowUpDate}}{{.FollowUpBy}} — {{date .FollowUpDate}}{{else}}Belum ada follow up{{end}}</td></tr>
<tr><td colspan="2" style="padding:4px;border-bottom:3px solid ` + reportNavy + `;">&nbsp;</td></tr>
{{end}}</table>
{{else}}<div style="padding:12px 16px;background:#e9f7ef;border-left:4px solid ` + reportGreen + `;border-radius:4px;margin-bottom:16px;"><p style="margin:0;color:` + reportGreen + `;font-weight:bold;">✓ Tidak ada issue ditemukan dalam inspeksi ini. Pertahankan!</p></div>
{{end}}
<div style="margin-top:16px;padding:14px 16px;background:#eaf2f8;border-left:4px solid ` + reportNavy + `;border-radius:4px;"><p style="margin:0 0 8px 0;font-weight:bold;color:` + reportNavy + `;">Tindak Lanjuti Issue Sekarang:</p><a href="{{.AppURL}}" style="display:inline-block;padding:10px 20px;background:` + reportNavy + `;color:#ffffff;text-decoration:none;border-radius:5px;font-weight:bold;font-size:11pt;">Buka I-GMP</a></div>
<p style="color:#6b7280;font-size:11px;border-top:1px solid #e5e7eb;padding-top:12px;margin-top:16px;">Email ini dikirim otomatis oleh sistem I-GMP. Harap tidak membalas email ini.</p>
<p style="margin-bottom:0;">Regards,<br/><b>Digital Transformation — Plant Sentul</b></p>
</div></div>
{{define "total"}}{{$score := index .TotalPair 0}}{{$max := index .TotalPair 1}}<tr style="background:#eaf2f8;"><td style="padding:8px;border:1px solid #e5e7eb;font-weight:bold;">Average Keseluruhan</td><td style="padding:8px;border:1px solid #e5e7eb;text-align:center;font-weight:bold;">{{$score}} / {{$max}}</td><td style="padding:8px;border:1px solid #e5e7eb;text-align:center;font-size:14pt;font-weight:bold;color:{{scoreColor $score $max}};">{{percent $score $max}}</td></tr>{{end}}`))

// TotalPair exposes Total() to the template as [score, max].
func (r KawasanReport) TotalPair() []int {
	score, max := r.Total()
	return []int{score, max}
}

// RenderKawasanReport returns the email HTML.
func RenderKawasanReport(r KawasanReport) (string, error) {
	var buf bytes.Buffer
	if err := kawasanReportTemplate.Execute(&buf, r); err != nil {
		return "", fmt.Errorf("render kawasan report: %w", err)
	}
	return buf.String(), nil
}

// buildKawasanReport turns the report rows into the email model. decrypt
// opens the at-rest-encrypted issue Keterangan; appBaseURL is the website
// address (APP_BASE_URL). Links use "me" as the user segment: the website's
// ScopeGuard swaps it for whoever is logged in, so one email serves every PIC.
func buildKawasanReport(rows []inspection.KawasanReportRow, decrypt func(string) string, now time.Time, appBaseURL, plantID, kawasanPICs string) KawasanReport {
	if plantID == "" {
		plantID = "global"
	}
	issuesURL := strings.TrimRight(appBaseURL, "/") + "/cimory/" + url.PathEscape(plantID) + "/dashboard/me/issues"

	r := KawasanReport{GeneratedAt: now, AppURL: issuesURL}
	detailIndex := make(map[string]int)
	today := time.Date(now.In(reportLocation).Year(), now.In(reportLocation).Month(), now.In(reportLocation).Day(), 0, 0, 0, 0, reportLocation)

	for _, row := range rows {
		if r.AreaName == "" {
			r.AreaName = row.AreaName
		}
		if r.KawasanName == "" {
			r.KawasanName = row.KawasanName
		}

		idx, ok := detailIndex[row.DetailKawasanID]
		if !ok {
			idx = len(r.Details)
			detailIndex[row.DetailKawasanID] = idx
			r.Details = append(r.Details, DetailScore{Name: row.DetailKawasanName})
		}
		r.Details[idx].Score += row.Nilai
		r.Details[idx].Max += row.StandardScore

		if row.IssueID == nil || *row.IssueID == "" {
			continue
		}
		status := row.IssueStatus
		if (status == "" || status == "Open" || status == "InProgress") && row.DueDate != nil && row.DueDate.Before(today) {
			status = "OpenOverdue"
		}
		keterangan := row.Keterangan
		if decrypt != nil && keterangan != "" {
			keterangan = decrypt(keterangan)
		}
		issue := ReportIssue{
			DetailKawasanName: row.DetailKawasanName,
			Title:             row.UraianText,
			Keterangan:        keterangan,
			PICName:           kawasanPICs,
			DueDate:           row.DueDate,
			Status:            status,
			FollowUpDate:      row.FollowUpDate,
			URL:               issuesURL + "/" + url.PathEscape(*row.IssueID),
		}
		if row.FollowUpBy != nil {
			issue.FollowUpBy = *row.FollowUpBy
		}
		r.Issues = append(r.Issues, issue)
	}
	return r
}

var picRoleOrder = map[string]int{"Manager": 0, "Supervisor": 1, "Primary PIC": 2, "PIC": 2, "Staff": 3}

// formatKawasanPICs lists everyone mapped to the Kawasan for the report's
// "PIC Kawasan" field: "Deden (Manager), Hasan (Supervisor), Ari (Staff)",
// most senior role first, each person once.
func formatKawasanPICs(users []pic.ResponsibleUser) string {
	type entry struct {
		label string
		rank  int
	}
	seen := make(map[string]bool)
	var entries []entry
	for _, u := range users {
		if u.FullName == "" || seen[u.UserID] {
			continue
		}
		seen[u.UserID] = true
		rank, ok := picRoleOrder[u.KategoriPIC]
		if !ok {
			rank = len(picRoleOrder) + 1
		}
		label := u.FullName
		if u.KategoriPIC != "" {
			label += " (" + u.KategoriPIC + ")"
		}
		entries = append(entries, entry{label, rank})
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].rank < entries[j].rank })
	labels := make([]string, len(entries))
	for i, e := range entries {
		labels[i] = e.label
	}
	return strings.Join(labels, ", ")
}
