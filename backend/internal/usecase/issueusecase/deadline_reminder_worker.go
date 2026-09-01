package issueusecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/monitoring-system/backend/internal/domain/issue"
	masterdomain "github.com/monitoring-system/backend/internal/domain/master"
	picdomain "github.com/monitoring-system/backend/internal/domain/pic"
)

// StartDeadlineReminderWorker starts a background ticker that emails/notifies
// PIC & Manager a configurable number of days before an Issue's DueDate.
// Mirrors StartAutoApproveWorker's shape/cadence.
func StartDeadlineReminderWorker(ctx context.Context, uc issue.IssueUseCase, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				runDeadlineReminders(uc)
			}
		}
	}()
}

func runDeadlineReminders(uc issue.IssueUseCase) {
	usecaseImpl, ok := uc.(*issueUseCase)
	if !ok || usecaseImpl.picRepo == nil {
		return
	}

	candidates, err := usecaseImpl.repo.FindReminderCandidates()
	if err != nil {
		log.Println("DeadlineReminderWorker: failed to fetch candidates", err)
		return
	}

	for _, i := range candidates {
		evaluateIssueForDeadlineReminder(usecaseImpl, i)
	}
}

// evaluateIssueForDeadlineReminder resolves the issue PIC's plant-specific
// reminder lead time and sends the reminder once DueDate falls within that
// window — exactly the per-plant resolution pattern used by
// evaluateIssueForAutoApprove/readDaysSetting.
func evaluateIssueForDeadlineReminder(usecaseImpl *issueUseCase, i issue.Issue) {
	if i.DueDate == nil {
		return
	}

	inspector, _ := usecaseImpl.userRepo.FindByID(i.IssuePICUserID)
	plantID := ""
	if inspector != nil && inspector.PlantID != nil {
		plantID = *inspector.PlantID
	}

	daysBefore := readDaysSetting(usecaseImpl, masterdomain.SettingKeyIssueDeadlineReminderDaysBefore, plantID, 3)
	if daysBefore <= 0 {
		return // reminder disabled for this plant
	}

	now := time.Now()
	reminderFrom := i.DueDate.AddDate(0, 0, -daysBefore)
	if now.Before(reminderFrom) {
		return // not within the reminder window yet
	}

	recipients, errRecipients := usecaseImpl.picRepo.FindResponsibleUsersByResultID(i.ResultID, "")
	if errRecipients != nil || len(recipients) == 0 {
		if inspector != nil {
			fallbackPlant := plantID
			recipients = []picdomain.ResponsibleUser{{
				UserID:   inspector.UserID,
				FullName: inspector.FullName,
				Email:    inspector.Email,
				PlantID:  &fallbackPlant,
			}}
		}
	}
	if len(recipients) == 0 {
		return
	}

	daysRemaining := int(time.Until(*i.DueDate).Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}
	dueDateStr := i.DueDate.Format("02 January 2006")

	for _, recipient := range recipients {
		usecaseImpl.notify(recipient.UserID, "warning", "Pengingat Tenggat Temuan",
			fmt.Sprintf("Temuan %s jatuh tempo %s (%d hari lagi).", i.IssueID, dueDateStr, daysRemaining), "/issues/"+i.IssueID)
	}

	go func(issueID, keterangan string, targets []picdomain.ResponsibleUser) {
		for _, target := range targets {
			if target.Email == "" {
				continue
			}
			targetPlant := ""
			if target.PlantID != nil {
				targetPlant = *target.PlantID
			}
			tmpl := ""
			if s, errSet := usecaseImpl.settingRepo.FindByKey(masterdomain.SettingKeyEmailTemplateDeadlineReminder, targetPlant); errSet == nil && s.SettingValue != "" {
				tmpl = s.SettingValue
			}
			if tmpl == "" {
				continue
			}
			_ = usecaseImpl.mailer.SendTemplateForPlant(
				targetPlant,
				[]string{target.Email},
				"[Monitoring Audit] Pengingat Tenggat Temuan",
				tmpl,
				map[string]string{
					"PICName":       target.FullName,
					"IssueID":       issueID,
					"Keterangan":    keterangan,
					"DueDate":       dueDateStr,
					"DaysRemaining": fmt.Sprintf("%d", daysRemaining),
				},
			)
		}
	}(i.IssueID, i.Keterangan, recipients)

	sentAt := now
	i.DeadlineReminderSentAt = &sentAt
	if err := usecaseImpl.repo.Update(&i); err != nil {
		log.Printf("DeadlineReminderWorker: failed to mark Issue %s as reminded: %v", i.IssueID, err)
	}
}
