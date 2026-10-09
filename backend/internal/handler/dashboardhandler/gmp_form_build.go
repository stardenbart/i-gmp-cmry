package dashboardhandler

import (
	"sort"
	"strings"
	"time"

	"github.com/monitoring-system/backend/pkg/exporter"
)

// gmpFormSourceRow is one inspection result as loaded for the GMP form
// export (template format).
type gmpFormSourceRow struct {
	InspectionID  string
	Tanggal       time.Time
	Area          string
	KawasanID     string
	Kawasan       string
	DetailKawasan string
	Aspek         string
	Detail        string
	UraianID      string
	Uraian        string
	Checking      string // OK, NG, NA
	Nilai         int
	IssueID       *string
	IssueStatus   string
	DueDate       *time.Time
	Keterangan    string // issue keterangan (decrypted), fallback for photos without one
}

// buildGMPFormSheets turns export rows into one form sheet per inspection,
// in inspection date order, with uraian in UraianID order and one finding
// per initial photo. Follow-ups attach to the photo they reference
// (RefPhotoID), otherwise to the issue's first finding.
func buildGMPFormSheets(
	rows []gmpFormSourceRow,
	initial map[string][]gmpInitialEvidence,
	followUps map[string][]gmpFollowUpEvidence,
	picsByKawasan map[string]string,
	company, photoBaseURL string,
) []exporter.GMPFormSheet {
	type group struct {
		first gmpFormSourceRow
		rows  []gmpFormSourceRow
	}
	var order []string
	groups := map[string]*group{}
	for _, r := range rows {
		g, ok := groups[r.InspectionID]
		if !ok {
			g = &group{first: r}
			groups[r.InspectionID] = g
			order = append(order, r.InspectionID)
		}
		g.rows = append(g.rows, r)
	}
	sort.SliceStable(order, func(i, j int) bool {
		a, b := groups[order[i]].first, groups[order[j]].first
		if !a.Tanggal.Equal(b.Tanggal) {
			return a.Tanggal.Before(b.Tanggal)
		}
		return a.DetailKawasan < b.DetailKawasan
	})

	base := strings.TrimRight(photoBaseURL, "/")
	sheets := make([]exporter.GMPFormSheet, 0, len(order))
	for _, id := range order {
		g := groups[id]
		sort.SliceStable(g.rows, func(i, j int) bool { return g.rows[i].UraianID < g.rows[j].UraianID })

		sheet := exporter.GMPFormSheet{
			Company:  company,
			Title:    "INSPEKSI GMP (BASIC) - " + g.first.Area,
			Date:     g.first.Tanggal.In(dashboardLocation()),
			Location: g.first.DetailKawasan + " - " + g.first.Kawasan,
			PIC:      picsByKawasan[g.first.KawasanID],
		}
		for _, r := range g.rows {
			u := exporter.GMPFormUraian{UraianID: r.UraianID, Aspek: r.Aspek, Detail: r.Detail, Text: r.Uraian}
			if !strings.EqualFold(r.Checking, "NA") {
				nilai := r.Nilai
				u.Nilai = &nilai
			}
			if r.IssueID != nil && *r.IssueID != "" {
				u.Findings = gmpFormFindings(r, initial[*r.IssueID], followUps[*r.IssueID], base)
			}
			sheet.Uraian = append(sheet.Uraian, u)
		}
		sheets = append(sheets, sheet)
	}
	return sheets
}

func gmpFormFindings(r gmpFormSourceRow, photos []gmpInitialEvidence, followUps []gmpFollowUpEvidence, photoBase string) []exporter.GMPFormFinding {
	status := gmpFormStatusLabel(r.IssueStatus)
	newFinding := func(photoURL, keterangan string) exporter.GMPFormFinding {
		if strings.TrimSpace(keterangan) == "" {
			keterangan = r.Keterangan
		}
		return exporter.GMPFormFinding{PhotoURL: absolutePhotoURL(photoBase, photoURL), Keterangan: strings.TrimSpace(keterangan), Status: status, DueDate: r.DueDate}
	}

	var findings []exporter.GMPFormFinding
	index := map[string]int{}
	for _, p := range photos {
		if p.ImageURL == "" {
			continue
		}
		index[p.PhotoID] = len(findings)
		findings = append(findings, newFinding(p.ImageURL, p.Keterangan))
	}
	if len(findings) == 0 {
		findings = append(findings, newFinding("", ""))
	}

	for _, fu := range followUps {
		target, ok := index[fu.RefPhotoID]
		if !ok {
			target = 0
		}
		line := strings.TrimSpace(fu.Keterangan)
		if fu.FollowUpDate != nil {
			date := fu.FollowUpDate.In(dashboardLocation()).Format("02 Jan 2006")
			if line == "" {
				line = date
			} else {
				line = date + " - " + line
			}
		}
		if line == "" {
			continue
		}
		if findings[target].FollowUp != "" {
			findings[target].FollowUp += "\n"
		}
		findings[target].FollowUp += line
	}
	return findings
}

func absolutePhotoURL(base, path string) string {
	if path == "" || strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") || base == "" {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
}

// gmpFormStatusLabel is the issue status shown before the due date in the
// form ("Open - 03 Oct 2026"); the total row counts the "*Closed*" ones.
func gmpFormStatusLabel(status string) string {
	switch status {
	case "InProgress":
		return "In Progress"
	case "PendingValidation":
		return "Menunggu Validasi"
	case "Verified", "Closed":
		return "Closed"
	case "ClosedOverdue":
		return "Closed (Overdue)"
	case "OpenOverdue":
		return "Open (Overdue)"
	default:
		return "Open"
	}
}
