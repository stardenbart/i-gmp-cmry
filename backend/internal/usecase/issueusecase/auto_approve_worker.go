package issueusecase

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/monitoring-system/backend/internal/domain/events"
	"github.com/monitoring-system/backend/internal/domain/issue"
	masterdomain "github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/mail"
)

// StartAutoApproveWorker starts a background ticker that checks for PendingValidation issues
// and auto-approves them if they have exceeded the ISSUE_AUTO_APPROVE_DAYS limit.
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

	// 1. Get setting ISSUE_AUTO_APPROVE_DAYS
	days := 3
	s, err := usecaseImpl.settingRepo.FindByKey(masterdomain.SettingKeyIssueAutoApproveDays)
	if err == nil && s.SettingValue != "" {
		if d, errParse := strconv.Atoi(s.SettingValue); errParse == nil {
			days = d
		}
	}

	// 2. Fetch all issues that are PendingValidation
	// In a real app we might paginate if there are thousands, but let's fetch a large batch for now.
	issues, _, err := usecaseImpl.repo.FindAll(1, 1000, string(issue.IssueStatusPendingValidation), "")
	if err != nil {
		log.Println("AutoApproveWorker: failed to fetch issues", err)
		return
	}

	now := time.Now()
	for _, i := range issues {
		// Calculate deadline from IssueUpdatedAt + days
		deadline := i.IssueUpdatedAt.AddDate(0, 0, days)
		if now.After(deadline) {
			// Auto approve!
			i.IssueStatus = issue.IssueStatusVerified
			i.Keterangan = "Auto-Approved by System (Timeout)"
			
			errUpdate := usecaseImpl.repo.Update(&i)
			if errUpdate == nil {
				log.Printf("AutoApproveWorker: Issue %s auto-approved", i.IssueID)
				
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

				// Send Email Notification
				go func(issueSnapshot issue.Issue) {
					pic, errPic := usecaseImpl.userRepo.FindByID(issueSnapshot.IssuePICUserID)
					if errPic != nil || pic == nil {
						return
					}
					
					dueDate := "-"
					if issueSnapshot.DueDate != nil {
						dueDate = issueSnapshot.DueDate.Format("02 January 2006")
					}
					
					// Re-use assignment template or fallback
					tmpl := mail.TmplIssueAssignment
					if s, errSet := usecaseImpl.settingRepo.FindByKey(masterdomain.SettingKeyEmailTemplateIssue); errSet == nil && s.SettingValue != "" {
						tmpl = s.SettingValue
					}

					_ = usecaseImpl.mailer.SendTemplate(
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
				}(i)
			}
		}
	}
}
