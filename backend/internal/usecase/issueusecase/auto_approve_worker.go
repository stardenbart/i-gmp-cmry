package issueusecase

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/internal/domain/events"
	"github.com/monitoring-system/backend/internal/domain/issue"
	masterdomain "github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/mail"
)

// StartAutoApproveWorker starts a background ticker that checks for PendingValidation issues
// (including those awaiting WO/WR sign-off) and auto-approves them if they have
// exceeded the ISSUE_AUTO_APPROVE_DAYS limit.
func StartAutoApproveWorker(ctx context.Context, uc issue.IssueUseCase, interval time.Duration) {
	ticker := time.NewTicker(interval)

	// Safe type assertion because we know uc is an *issueUseCase internally.
	// But to avoid exposing the struct, we can just fetch all pending issues via GetAll
	// since GetAll takes `status`.

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				runAutoApprove(uc)
			}
		}
	}()
}

func runAutoApprove(uc issue.IssueUseCase) {
	// We need access to settingRepo and producer to do it properly.
	// To keep it clean, let's type assert if it's the default implementation.
	usecaseImpl, ok := uc.(*issueUseCase)
	if !ok {
		return
	}

	// 1. Fetch all issues that are PendingValidation
	// In a real app we might paginate if there are thousands, but let's fetch a large batch for now.
	issues, _, err := usecaseImpl.repo.FindAll(1, 1000, "", string(issue.IssueStatusPendingValidation), "", nil)
	if err != nil {
		log.Println("AutoApproveWorker: failed to fetch issues", err)
		return
	}

	for _, i := range issues {
		evaluateIssueForAutoApprove(usecaseImpl, i, "AutoApproveWorker")
	}
}

// TriggerWOWRAutoApproveCheck re-evaluates a single issue for auto-approval
// right away. Call it the moment a PIC uploads WO/WR proof: if the WO/WR
// deadline (ISSUE_WOWR_AUTO_APPROVE_DAYS, counted from when validation was
// requested) had already elapsed while the auditor was waiting on that proof,
// the issue is approved immediately instead of sitting until the next hourly
// tick of the background worker.
func TriggerWOWRAutoApproveCheck(uc issue.IssueUseCase, issueID string) {
	usecaseImpl, ok := uc.(*issueUseCase)
	if !ok {
		return
	}
	i, err := usecaseImpl.repo.FindByID(issueID)
	if err != nil || i == nil || i.IssueStatus != issue.IssueStatusPendingValidation {
		return
	}
	evaluateIssueForAutoApprove(usecaseImpl, *i, "WOWRUploadTrigger")
}

// evaluateIssueForAutoApprove checks one PendingValidation issue against the
// follow-up and WO/WR auto-approve deadlines and, when every gate that
// applies has timed out, approves it exactly as an auditor would.
func evaluateIssueForAutoApprove(usecaseImpl *issueUseCase, i issue.Issue, source string) {
	now := time.Now()

	// Resolve the PIC's plant so each plant's own auto-approve settings apply
	// (FindByKey falls back to the global default when the plant has no override).
	pic, errPic := usecaseImpl.userRepo.FindByID(i.IssuePICUserID)
	plantID := ""
	if errPic == nil && pic != nil && pic.PlantID != nil {
		plantID = *pic.PlantID
	}

	days := readDaysSetting(usecaseImpl, masterdomain.SettingKeyIssueAutoApproveDays, plantID, 3)

	// The follow-up deadline (regular temuan) and the WO/WR deadline are two
	// independent clocks, each with its own setting — a plant may want
	// auditors to review WO/WR proof faster/slower than a plain follow-up.
	followUpDeadline := i.IssueUpdatedAt.AddDate(0, 0, days)
	followUpDue := now.After(followUpDeadline)

	wowrPending := i.NeedsWOWR && i.WOWRStatus == issue.WOWRStatusPendingValidation
	wowrDue := false
	if wowrPending {
		wowrDays := readDaysSetting(usecaseImpl, masterdomain.SettingKeyIssueWOWRAutoApproveDays, plantID, 3)
		wowrBase := i.IssueUpdatedAt
		if i.WOWRSubmittedAt != nil {
			wowrBase = *i.WOWRSubmittedAt
		}
		wowrDue = now.After(wowrBase.AddDate(0, 0, wowrDays))
	}

	// Only auto-approve once every gate that applies to this issue has timed
	// out — an issue can't close while WO/WR is still within its own grace
	// period, even if the follow-up timeout already elapsed, and vice versa.
	if !followUpDue || (wowrPending && !wowrDue) {
		return
	}

	// If this issue requires WO/WR sign-off, the auditor has to verify that
	// proof too — the same mechanism now covers it instead of silently
	// skipping it (which previously let the issue close without the WO/WR
	// ever actually being marked verified).
	if wowrPending {
		if !usecaseImpl.hasWOWRProof(i.IssueID, i.IssuePICUserID) {
			// PIC never uploaded WO/WR proof — nothing to auto-approve yet.
			// (Once they do, the upload path calls TriggerWOWRAutoApproveCheck
			// so this deadline gets re-checked immediately, not just hourly.)
			log.Printf("%s: Issue %s skipped — WO/WR proof not yet uploaded", source, i.IssueID)
			return
		}
		i.WOWRStatus = issue.WOWRStatusVerified
		i.WOWRSubmittedAt = nil
	}

	// Auto approve!
	i.IssueStatus = issue.IssueStatusVerified
	i.Keterangan = "Auto-Approved by System (Timeout)"

	errUpdate := usecaseImpl.repo.Update(&i)
	if errUpdate != nil {
		return
	}
	log.Printf("%s: Issue %s auto-approved", source, i.IssueID)

	usecaseImpl.notify(i.IssuePICUserID, "success", "Issue Disetujui Otomatis",
		"Follow-up perbaikan temuan Anda telah disetujui otomatis oleh sistem karena tidak diulas auditor dalam batas waktu.", "/issues/"+i.IssueID)

	// Publish Event
	event := events.IssueEvent{
		BaseEvent: events.BaseEvent{
			EventID:   uuid.New().String(),
			EventType: events.EventTypeUpdated,
			Timestamp: time.Now(),
			ActorID:   "SYSTEM",
		},
		IssueID:        i.IssueID,
		ResultID:       i.ResultID,
		IssuePICUserID: i.IssuePICUserID,
		Status:         string(i.IssueStatus),
		DueDate:        i.DueDate,
	}
	_ = usecaseImpl.producer.PublishEvent(context.Background(), events.TopicAuditIssues, i.IssueID, event)

	// Send Email Notification (pic/plantID already resolved above)
	go func(issueSnapshot issue.Issue, pic *authdomain.User, plantID string) {
		if pic == nil {
			return
		}

		dueDate := "-"
		if issueSnapshot.DueDate != nil {
			dueDate = issueSnapshot.DueDate.Format("02 January 2006")
		}

		// Re-use assignment template or fallback
		tmpl := mail.TmplIssueAssignment
		if s, errSet := usecaseImpl.settingRepo.FindByKey(masterdomain.SettingKeyEmailTemplateIssue, plantID); errSet == nil && s.SettingValue != "" {
			tmpl = s.SettingValue
		}

		_ = usecaseImpl.mailer.SendTemplateForPlant(
			plantID,
			[]string{pic.Email},
			"[Monitoring Audit] Issue Telah Disetujui (Auto-Approve)",
			tmpl,
			map[string]string{
				"PICName":    pic.FullName,
				"IssueID":    issueSnapshot.IssueID,
				"Keterangan": issueSnapshot.Keterangan,
				"Status":     string(issueSnapshot.IssueStatus),
				"DueDate":    dueDate,
			},
		)
	}(i, pic, plantID)
}

// readDaysSetting resolves a "days" setting for the given plant, falling back
// to defaultDays when unset, unparsable, or missing entirely.
func readDaysSetting(uc *issueUseCase, key, plantID string, defaultDays int) int {
	s, err := uc.settingRepo.FindByKey(key, plantID)
	if err != nil || s.SettingValue == "" {
		return defaultDays
	}
	d, err := strconv.Atoi(s.SettingValue)
	if err != nil {
		return defaultDays
	}
	return d
}

// hasWOWRProof reports whether a WO/WR completion photo has been uploaded
// for the issue. Mirrors the check in issueUseCase.Update so the
// auto-approve timeout never verifies WO/WR work that was never submitted.
//
// Does not compare the photo's uploader against picUserID: IssuePICUserID
// is always the inspector who ran the audit, never the Auditee/PIC who
// actually uploads WO/WR evidence, so that comparison always failed and
// permanently blocked auto-approval even when a valid photo existed.
func (uc *issueUseCase) hasWOWRProof(issueID, _ string) bool {
	if uc.photoRepo == nil {
		return false
	}
	photos, err := uc.photoRepo.FindByIssueID(issueID)
	if err != nil {
		return false
	}
	for _, photo := range photos {
		if photo.PhotoType == issue.PhotoTypeWOWR &&
			(photo.ImageUrl != "" || photo.FileName != "") {
			return true
		}
	}
	return false
}
