package issueusecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/internal/domain/events"
	"github.com/monitoring-system/backend/internal/domain/issue"
	masterdomain "github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/idgen"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/mail"
	"github.com/monitoring-system/backend/pkg/storage"
)

// ── Issue UseCase ─────────────────────────────────────────────────────────

type issueUseCase struct {
	repo        issue.IssueRepository
	producer    kafka.EventProducer
	mailer      mail.Mailer
	userRepo    authdomain.UserRepository // to fetch PIC email
	settingRepo masterdomain.SettingRepository
}

func NewIssueUseCase(repo issue.IssueRepository, producer kafka.EventProducer, mailer mail.Mailer, userRepo authdomain.UserRepository, settingRepo masterdomain.SettingRepository) issue.IssueUseCase {
	return &issueUseCase{repo: repo, producer: producer, mailer: mailer, userRepo: userRepo, settingRepo: settingRepo}
}

func (uc *issueUseCase) GetAll(page, limit int, status, picUserID string) ([]issue.Issue, int64, error) {
	return uc.repo.FindAll(page, limit, status, picUserID)
}

func (uc *issueUseCase) GetByID(id string) (*issue.Issue, error) { return uc.repo.FindByID(id) }

func (uc *issueUseCase) Create(actorID string, req *issue.CreateIssueRequest) (*issue.Issue, error) {
	i := &issue.Issue{
		IssueID:        idgen.Generate(idgen.PrefixIssue),
		ResultID:       req.ResultID,
		IssuePICUserID: req.IssuePICUserID,
		DueDate:        req.DueDate,
		IssueStatus:    issue.IssueStatusOpen,
		Keterangan:     req.Keterangan,
	}
	err := uc.repo.Create(i)
	if err == nil {
		event := events.IssueEvent{
			BaseEvent: events.BaseEvent{
				EventID:   uuid.New().String(),
				EventType: events.EventTypeCreated,
				Timestamp: time.Now(),
				ActorID:   actorID,
			},
			IssueID:        i.IssueID,
			ResultID:       i.ResultID,
			IssuePICUserID: i.IssuePICUserID,
			Status:         string(i.IssueStatus),
			DueDate:        i.DueDate,
		}
		_ = uc.producer.PublishEvent(context.Background(), events.TopicAuditIssues, i.IssueID, event)

		// Send email notification to PIC asynchronously
		go func(issueSnapshot issue.Issue) {
			pic, err := uc.userRepo.FindByID(issueSnapshot.IssuePICUserID)
			if err != nil || pic == nil {
				return
			}
			dueDate := "-"
			if issueSnapshot.DueDate != nil {
				dueDate = issueSnapshot.DueDate.Format("02 January 2006")
			}
			
			tmpl := mail.TmplIssueAssignment
			if s, err := uc.settingRepo.FindByKey(masterdomain.SettingKeyEmailTemplateIssue); err == nil && s.SettingValue != "" {
				tmpl = s.SettingValue
			}

			_ = uc.mailer.SendTemplate(
				[]string{pic.Email},
				"[Monitoring Audit] Issue Baru Ditugaskan kepada Anda",
				tmpl,
				map[string]string{
					"PICName":    pic.FullName,
					"IssueID":    issueSnapshot.IssueID,
					"Keterangan": issueSnapshot.Keterangan,
					"Status":     string(issueSnapshot.IssueStatus),
					"DueDate":    dueDate,
				},
			)
		}(*i)
	}
	return i, err
}


func (uc *issueUseCase) Update(id string, actorID string, req *issue.UpdateIssueRequest) (*issue.Issue, error) {
	i, err := uc.repo.FindByID(id)
	if err != nil { return nil, errors.New("issue not found") }
	if req.IssuePICUserID != "" { i.IssuePICUserID = req.IssuePICUserID }
	if req.DueDate != nil { i.DueDate = req.DueDate }
	if req.IssueStatus != "" { i.IssueStatus = req.IssueStatus }
	if req.Keterangan != "" { i.Keterangan = req.Keterangan }
	err = uc.repo.Update(i)
	if err == nil {
		event := events.IssueEvent{
			BaseEvent: events.BaseEvent{
				EventID:   uuid.New().String(),
				EventType: events.EventTypeUpdated,
				Timestamp: time.Now(),
				ActorID:   actorID,
			},
			IssueID:        i.IssueID,
			ResultID:       i.ResultID,
			IssuePICUserID: i.IssuePICUserID,
			Status:         string(i.IssueStatus),
			DueDate:        i.DueDate,
		}
		_ = uc.producer.PublishEvent(context.Background(), events.TopicAuditIssues, i.IssueID, event)
	}
	return i, err
}

func (uc *issueUseCase) Delete(id string, actorID string) error {
	err := uc.repo.Delete(id)
	if err == nil {
		event := events.IssueEvent{
			BaseEvent: events.BaseEvent{
				EventID:   uuid.New().String(),
				EventType: events.EventTypeDeleted,
				Timestamp: time.Now(),
				ActorID:   actorID,
			},
			IssueID: id,
		}
		_ = uc.producer.PublishEvent(context.Background(), events.TopicAuditIssues, id, event)
	}
	return err
}

// ── Issue Photo UseCase ────────────────────────────────────────────────────

type issuePhotoUseCase struct {
	repo    issue.IssuePhotoRepository
	storage *storage.MinioStorage
}

func NewIssuePhotoUseCase(repo issue.IssuePhotoRepository, s *storage.MinioStorage) issue.IssuePhotoUseCase {
	return &issuePhotoUseCase{repo: repo, storage: s}
}

func (uc *issuePhotoUseCase) GetByIssueID(issueID string) ([]issue.IssuePhoto, error) {
	return uc.repo.FindByIssueID(issueID)
}

func (uc *issuePhotoUseCase) Upload(ctx context.Context, req *issue.UploadPhotoRequest, fileReader io.Reader, fileSize int64, originalFileName, contentType string) (*issue.IssuePhoto, error) {
	if fileSize == 0 { return nil, errors.New("empty file") }

	// Construct a unique object name
	ext := ""
	for i := len(originalFileName) - 1; i >= 0 && originalFileName[i] != '/'; i-- {
		if originalFileName[i] == '.' {
			ext = originalFileName[i:]
			break
		}
	}
	objectName := fmt.Sprintf("issues/%s/%d%s", req.IssueID, time.Now().UnixNano(), ext)

	// Stream to MinIO
	publicURL, err := uc.storage.UploadStream(ctx, objectName, fileReader, fileSize, contentType)
	if err != nil { return nil, err }

	p := &issue.IssuePhoto{
		IssuePhotoID:   idgen.Generate(idgen.PrefixIssuePhoto),
		IssueID:        req.IssueID,
		PICUserID:      req.PICUserID,
		PhotoType:      req.PhotoType,
		ImageUrl:       publicURL,
		FileName:       objectName, // we store objectName here so we can delete it later
		FollowUpDate:   req.FollowUpDate,
		JumlahFollowUp: req.JumlahFollowUp,
	}
	return p, uc.repo.Create(p)
}

func (uc *issuePhotoUseCase) Delete(ctx context.Context, id string) error {
	photo, err := uc.repo.FindByID(id)
	if err != nil { return errors.New("photo not found") }
	_ = uc.storage.Delete(ctx, photo.FileName) // best-effort file deletion from MinIO
	return uc.repo.Delete(id)
}

// Ensure io is imported even if not directly used (for future streaming support)
var _ io.Reader = nil
