package inspectionusecase

import (
	"fmt"
	"log"
	"strings"
	"time"
	"unicode"

	"github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/domain/master"
	notificationdomain "github.com/monitoring-system/backend/internal/domain/notification"
	"github.com/monitoring-system/backend/internal/domain/pic"
	"github.com/monitoring-system/backend/pkg/mail"
)

type InspectionEmailNotifier struct {
	mailer         mail.Mailer
	picRepo        pic.PICMappingRepository
	authRepo       auth.UserRepository
	inspectionRepo inspection.InspectionHeaderRepository
	kawasanRepo    master.KawasanRepository
	settingRepo    master.SettingRepository
	notificationUC notificationdomain.NotificationUseCase // in-app bell notifications, alongside email

	// Kawasan report email (score per Detail Kawasan + issue list).
	reportRepo inspection.KawasanReportRepository
	decrypt    func(string) string // opens encrypted issue Keterangan
	appBaseURL string              // website address used for links
}

// WithKawasanReport enables the detailed Kawasan report email. Without it
// the notifier only sends the in-app notification for a completed Kawasan.
func (n *InspectionEmailNotifier) WithKawasanReport(repo inspection.KawasanReportRepository, decrypt func(string) string, appBaseURL string) *InspectionEmailNotifier {
	n.reportRepo = repo
	n.decrypt = decrypt
	n.appBaseURL = appBaseURL
	return n
}

func NewInspectionEmailNotifier(
	mailer mail.Mailer,
	picRepo pic.PICMappingRepository,
	authRepo auth.UserRepository,
	inspectionRepo inspection.InspectionHeaderRepository,
	kawasanRepo master.KawasanRepository,
	settingRepo master.SettingRepository,
	notificationUC notificationdomain.NotificationUseCase,
) *InspectionEmailNotifier {
	return &InspectionEmailNotifier{
		mailer:         mailer,
		picRepo:        picRepo,
		authRepo:       authRepo,
		inspectionRepo: inspectionRepo,
		kawasanRepo:    kawasanRepo,
		settingRepo:    settingRepo,
		notificationUC: notificationUC,
	}
}

// SendInspectionSummary is called when an Area's status becomes Confirmed.
//
// Previously resolved recipients via picRepo.FindByAreaAndKawasan(areaID,
// ""), which filters on PIC_Mapping.AreaID — a column migration 011 already
// deprecated ("PIC mapping now works per Kawasan only"), so that query
// almost certainly returned zero PICs on real data. An Area can span several
// Kawasan, so recipients here are the union of every Kawasan's own
// responsible users (PIC/Manager/whoever is mapped) across the whole Area.
func (n *InspectionEmailNotifier) SendInspectionSummary(areaID string, progress inspection.AreaProgress) error {
	kawasans, err := n.kawasanRepo.FindByAreaID(areaID)
	if err != nil {
		log.Printf("[EmailNotifier] Error retrieving Kawasan for Area %s: %v", areaID, err)
		return err
	}

	var targetEmails []string
	var targetUserIDs []string
	emailSet := make(map[string]bool)
	userSet := make(map[string]bool)
	plantID := ""
	for _, kws := range kawasans {
		users, errUsers := n.picRepo.FindResponsibleUsers(kws.KawasanID, "")
		if errUsers != nil {
			log.Printf("[EmailNotifier] Error retrieving responsible users for Kawasan %s: %v", kws.KawasanID, errUsers)
			continue
		}
		for _, u := range users {
			if plantID == "" && u.PlantID != nil {
				plantID = *u.PlantID
			}
			if u.Email != "" && !emailSet[u.Email] {
				emailSet[u.Email] = true
				targetEmails = append(targetEmails, u.Email)
			}
			if !userSet[u.UserID] {
				userSet[u.UserID] = true
				targetUserIDs = append(targetUserIDs, u.UserID)
			}
		}
	}

	// In-app bell notification, independent of whether the PIC has an email
	// on file (email delivery is best-effort further down).
	if n.notificationUC != nil {
		message := fmt.Sprintf("Inspeksi Area %s telah selesai (%d/%d detail kawasan).", areaID, progress.CompletedDetailKawasan, progress.TotalDetailKawasan)
		for _, userID := range targetUserIDs {
			_ = n.notificationUC.CreateSystemNotification(userID, "success", "Inspeksi Area Selesai", message, "/inspections")
		}
	}

	if len(targetEmails) == 0 {
		log.Printf("[EmailNotifier] No PIC found or missing emails for Area %s", areaID)
		return nil
	}

	// 2. Load the matching plant template (with global fallback).
	setting, err := n.settingRepo.FindByKey(master.SettingKeyEmailTemplateInspectionConfirmed, plantID)
	if err != nil {
		log.Printf("[EmailNotifier] Error retrieving template: %v", err)
		return err
	}
	templateStr := setting.SettingValue

	// 3. Prepare data for the template
	data := map[string]interface{}{
		"AreaID":                 areaID,
		"TotalDetailKawasan":     progress.TotalDetailKawasan,
		"CompletedDetailKawasan": progress.CompletedDetailKawasan,
		"Status":                 string(progress.Status),
	}

	// 4. Send Email via Mailer
	subject := "✅ Inspeksi Area Selesai: " + areaID
	err = n.mailer.SendTemplateForPlant(plantID, targetEmails, subject, templateStr, data)
	if err != nil {
		log.Printf("[EmailNotifier] Failed to send email to %v: %v", targetEmails, err)
		return err
	}

	log.Printf("[EmailNotifier] Successfully sent Area Confirmed email to %v", targetEmails)
	return nil
}

// SendKawasanInspectionSummary is called when a Kawasan's status becomes
// Confirmed (every DetailKawasan under it completed in the period). Everyone
// mapped to that Kawasan (PIC, Manager, ...) gets one "Laporan Hasil
// Inspeksi GMP" email: score per Detail Kawasan, the Kawasan average, the
// issue list and links to the website, with the plant's CC list copied. This
// replaces the former Power Automate flow (GMP_Report_Flow).
func (n *InspectionEmailNotifier) SendKawasanInspectionSummary(kawasanID string, progress inspection.KawasanProgress, periodStart, periodEnd time.Time) error {
	responsible, err := n.picRepo.FindResponsibleUsers(kawasanID, "")
	if err != nil {
		log.Printf("[EmailNotifier] Error retrieving responsible users for Kawasan %s: %v", kawasanID, err)
		return err
	}

	var targetEmails []string
	var targetUserIDs []string
	emailSet := make(map[string]bool)
	plantID := ""
	for _, m := range responsible {
		if plantID == "" && m.PlantID != nil {
			plantID = *m.PlantID
		}
		if m.Email != "" && !emailSet[strings.ToLower(m.Email)] {
			emailSet[strings.ToLower(m.Email)] = true
			targetEmails = append(targetEmails, m.Email)
		}
		targetUserIDs = append(targetUserIDs, m.UserID)
	}

	var report KawasanReport
	if n.reportRepo != nil {
		rows, errRows := n.reportRepo.FindKawasanReportRows(kawasanID, periodStart, periodEnd)
		if errRows != nil {
			log.Printf("[EmailNotifier] Error loading report for Kawasan %s: %v", kawasanID, errRows)
			return errRows
		}
		report = buildKawasanReport(rows, n.decrypt, time.Now(), n.appBaseURL, plantID, formatKawasanPICs(responsible))
	}
	kawasanName := report.KawasanName
	if kawasanName == "" {
		kawasanName = kawasanID
	}

	if n.notificationUC != nil {
		message := fmt.Sprintf("Kawasan %s telah selesai diinspeksi (%d/%d detail kawasan). Laporan skor dan daftar issue dikirim ke email.", kawasanName, progress.CompletedDetailKawasan, progress.TotalDetailKawasan)
		for _, userID := range targetUserIDs {
			_ = n.notificationUC.CreateSystemNotification(userID, "success", "Kawasan Selesai Diinspeksi", message, "/issues")
		}
	}

	if len(targetEmails) == 0 {
		log.Printf("[EmailNotifier] No PIC/Manager found or missing emails for Kawasan %s", kawasanID)
		return nil
	}
	if n.reportRepo == nil {
		log.Printf("[EmailNotifier] Kawasan report not configured; skipping email for Kawasan %s", kawasanID)
		return nil
	}

	html, err := RenderKawasanReport(report)
	if err != nil {
		return err
	}

	var cc []string
	if setting, errCC := n.settingRepo.FindByKey(master.SettingKeyEmailCCKawasanReport, plantID); errCC == nil && setting != nil {
		for _, addr := range parseEmailList(setting.SettingValue) {
			if !emailSet[strings.ToLower(addr)] {
				cc = append(cc, addr)
			}
		}
	}

	msg := mail.HTMLMessage{To: targetEmails, Cc: cc, Subject: report.Subject(), HTML: html, HighImportance: true}
	if err := n.mailer.SendHTMLForPlant(plantID, msg); err != nil {
		log.Printf("[EmailNotifier] Failed to send Kawasan report to %v (cc %v): %v", targetEmails, cc, err)
		return err
	}

	log.Printf("[EmailNotifier] Sent Kawasan report for %s to %v (cc %v)", kawasanID, targetEmails, cc)
	return nil
}

// parseEmailList splits a setting value on ";", "," and whitespace, keeping
// valid-looking addresses once (case-insensitive) in their original order.
func parseEmailList(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == ';' || r == ',' || unicode.IsSpace(r)
	})
	seen := make(map[string]bool)
	var out []string
	for _, f := range fields {
		f = strings.TrimSpace(f)
		at := strings.Index(f, "@")
		if at <= 0 || at == len(f)-1 || seen[strings.ToLower(f)] {
			continue
		}
		seen[strings.ToLower(f)] = true
		out = append(out, f)
	}
	return out
}
