package issueusecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/internal/domain/events"
	"github.com/monitoring-system/backend/internal/domain/issue"
	masterdomain "github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/idgen"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/mail"
	"github.com/monitoring-system/backend/pkg/sse"
	"github.com/monitoring-system/backend/pkg/storage"
)

// ── Issue UseCase ─────────────────────────────────────────────────────────

type issueUseCase struct {
	repo         issue.IssueRepository
	userRepo     authdomain.UserRepository // to fetch PIC email
	delegateRepo issue.IssueDelegateRepository
	producer     kafka.EventProducer
	mailer       mail.Mailer
	settingRepo  masterdomain.SettingRepository
	cryptoSvc    *crypto.Service
	sseBroker    *sse.Broker
}

func NewIssueUseCase(repo issue.IssueRepository, producer kafka.EventProducer, mailer mail.Mailer, userRepo authdomain.UserRepository, settingRepo masterdomain.SettingRepository, delegateRepo issue.IssueDelegateRepository, cryptoSvc *crypto.Service, sseBroker *sse.Broker) issue.IssueUseCase {
	return &issueUseCase{repo: repo, producer: producer, mailer: mailer, userRepo: userRepo, settingRepo: settingRepo, delegateRepo: delegateRepo, cryptoSvc: cryptoSvc, sseBroker: sseBroker}
}

func (uc *issueUseCase) GetAll(page, limit int, status, picUserID string, needsWOWR *bool) ([]issue.Issue, int64, error) {
	items, total, err := uc.repo.FindAll(page, limit, status, picUserID, needsWOWR)
	if err == nil {
		for i := range items {
			items[i].Keterangan = uc.cryptoSvc.DecryptWithFallback(items[i].Keterangan)
			for j := range items[i].Photos {
				items[i].Photos[j].ImageUrl = uc.cryptoSvc.DecryptWithFallback(items[i].Photos[j].ImageUrl)
				items[i].Photos[j].FileName = uc.cryptoSvc.DecryptWithFallback(items[i].Photos[j].FileName)
			}
			pic, errPic := uc.userRepo.FindByID(items[i].IssuePICUserID)
			if errPic == nil && pic != nil {
				items[i].PICName = pic.FullName
			}
		}
	}
	return items, total, err
}

func (uc *issueUseCase) GetByID(id string) (*issue.Issue, error) {
	item, err := uc.repo.FindByID(id)
	if err == nil && item != nil {
		item.Keterangan = uc.cryptoSvc.DecryptWithFallback(item.Keterangan)
		for j := range item.Photos {
			item.Photos[j].ImageUrl = uc.cryptoSvc.DecryptWithFallback(item.Photos[j].ImageUrl)
			item.Photos[j].FileName = uc.cryptoSvc.DecryptWithFallback(item.Photos[j].FileName)
		}
		pic, errPic := uc.userRepo.FindByID(item.IssuePICUserID)
		if errPic == nil && pic != nil {
			item.PICName = pic.FullName
		}
	}
	return item, err
}

func (uc *issueUseCase) Create(actorID string, req *issue.CreateIssueRequest) (*issue.Issue, error) {
	dueDate := req.DueDate
	if dueDate == nil {
		s, err := uc.settingRepo.FindByKey(masterdomain.SettingKeyIssueDeadlineDays)
		days := 14
		if err == nil && s.SettingValue != "" {
			if d, errParse := strconv.Atoi(s.SettingValue); errParse == nil {
				days = d
			}
		}
		dd := time.Now().AddDate(0, 0, days)
		dueDate = &dd
	}

	keteranganEnc := req.Keterangan
	if req.Keterangan != "" {
		if enc, err := uc.cryptoSvc.Encrypt(req.Keterangan); err == nil {
			keteranganEnc = enc
		}
	}

	// WO / WR logic
	var woID, wrID string
	if req.NeedsWOWR {
		woID = strings.ToUpper(strings.TrimSpace(req.WO_ID))
		wrID = strings.ToUpper(strings.TrimSpace(req.WR_ID))
		if woID == "" && wrID == "" {
			return nil, errors.New("jika membutuhkan WO/WR, minimal satu ID (WO atau WR) harus diisi")
		}
	}

	i := &issue.Issue{
		IssueID:        idgen.Generate(idgen.PrefixIssue),
		ResultID:       req.ResultID,
		IssuePICUserID: req.IssuePICUserID,
		DueDate:        dueDate,
		IssueStatus:    issue.IssueStatusOpen,
		Label:          req.Label,
		NeedsWOWR:      req.NeedsWOWR,
		WO_ID:          woID,
		WR_ID:          wrID,
		Keterangan:     keteranganEnc,
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
		if uc.sseBroker != nil {
			uc.sseBroker.Broadcast("ISSUE_UPDATED")
		}

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
	if err != nil {
		return nil, errors.New("issue not found")
	}

	// Validation logic for Follow Up
	if req.IssueStatus == issue.IssueStatusPendingValidation && i.IssueStatus != issue.IssueStatusPendingValidation {
		// Check if actor is PIC, Delegate, Auditor, or mapped PIC
		if actorID != i.IssuePICUserID {
			isDel, _ := uc.delegateRepo.IsDelegate(i.IssueID, actorID)
			if !isDel {
				user, errUser := uc.userRepo.FindByID(actorID)
				if errUser == nil && user != nil {
					isAud := user.RoleID == "ROLE-001" || user.RoleID == "ROLE-002" || user.RoleID == "ADM" || user.RoleID == "ADMIN"
					if !isAud {
						if len(user.PICMappings) == 0 {
							return nil, errors.New("unauthorized: you are not the PIC or delegate for this issue")
						}
					}
				}
			}
		}

		// WOWR Validation: If it needs WOWR, WO/WR ID must be provided before requesting validation
		if i.NeedsWOWR {
			woID := i.WO_ID
			wrID := i.WR_ID
			if req.WO_ID != "" {
				woID = req.WO_ID
			}
			if req.WR_ID != "" {
				wrID = req.WR_ID
			}
			if woID == "" && wrID == "" {
				return nil, errors.New("issue ini membutuhkan WO/WR, harap isi dan simpan Nomor WO/WR terlebih dahulu")
			}
			if i.WOWRStatus != issue.WOWRStatusVerified {
				i.WOWRStatus = issue.WOWRStatusPendingValidation
			}
		}
	}

	if req.IssuePICUserID != "" {
		i.IssuePICUserID = req.IssuePICUserID
	}
	if req.DueDate != nil {
		i.DueDate = req.DueDate
	}
	if req.WOWRStatus != "" {
		i.WOWRStatus = req.WOWRStatus
	}

	if req.IssueStatus != "" {
		// Validation for closing/verifying if NeedsWOWR is true
		if (req.IssueStatus == issue.IssueStatusClosed || req.IssueStatus == issue.IssueStatusVerified) && i.NeedsWOWR {
			hasFollowUp := false
			for _, p := range i.Photos {
				if p.PhotoType == issue.PhotoTypeFollowUp || p.PhotoType == issue.PhotoTypeWOWR {
					hasFollowUp = true
					break
				}
			}
			if !hasFollowUp {
				return nil, errors.New("issue ini membutuhkan WO/WR, wajib mengunggah foto FollowUp sebelum dapat diselesaikan")
			}
		}
		i.IssueStatus = req.IssueStatus
	}

	if req.Label != "" {
		i.Label = req.Label
	}

	// Update WO / WR logic
	if req.NeedsWOWR {
		i.NeedsWOWR = true
		i.WO_ID = strings.ToUpper(strings.TrimSpace(req.WO_ID))
		i.WR_ID = strings.ToUpper(strings.TrimSpace(req.WR_ID))

		if i.WO_ID == "" && i.WR_ID == "" {
			return nil, errors.New("jika membutuhkan WO/WR, minimal satu ID (WO atau WR) harus diisi")
		}
	} else if req.WO_ID != "" || req.WR_ID != "" {
		i.NeedsWOWR = req.NeedsWOWR
		i.WO_ID = strings.ToUpper(strings.TrimSpace(req.WO_ID))
		i.WR_ID = strings.ToUpper(strings.TrimSpace(req.WR_ID))
	}

	if req.Keterangan != "" {
		if enc, err := uc.cryptoSvc.Encrypt(req.Keterangan); err == nil {
			i.Keterangan = enc
		} else {
			i.Keterangan = req.Keterangan
		}
	}
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
		if uc.sseBroker != nil {
			uc.sseBroker.Broadcast("ISSUE_UPDATED")
		}
	}
	return i, err
}

func (uc *issueUseCase) ExtendDueDate(id string, actorID string, newDueDate time.Time) (*issue.Issue, error) {
	i, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("issue not found")
	}

	i.DueDate = &newDueDate
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
		if uc.sseBroker != nil {
			uc.sseBroker.Broadcast("ISSUE_UPDATED")
		}
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
		if uc.sseBroker != nil {
			uc.sseBroker.Broadcast("ISSUE_UPDATED")
		}
	}
	return err
}

// ── Issue Photo UseCase ────────────────────────────────────────────────────

type issuePhotoUseCase struct {
	repo      issue.IssuePhotoRepository
	storage   *storage.MinioStorage
	cryptoSvc *crypto.Service
	sseBroker *sse.Broker
}

func NewIssuePhotoUseCase(repo issue.IssuePhotoRepository, s *storage.MinioStorage, cryptoSvc *crypto.Service, sseBroker *sse.Broker) issue.IssuePhotoUseCase {
	return &issuePhotoUseCase{repo: repo, storage: s, cryptoSvc: cryptoSvc, sseBroker: sseBroker}
}

func (uc *issuePhotoUseCase) GetByIssueID(issueID string) ([]issue.IssuePhoto, error) {
	photos, err := uc.repo.FindByIssueID(issueID)
	if err == nil {
		for i := range photos {
			photos[i].ImageUrl = uc.cryptoSvc.DecryptWithFallback(photos[i].ImageUrl)
			photos[i].FileName = uc.cryptoSvc.DecryptWithFallback(photos[i].FileName)
		}
	}
	return photos, err
}

func (uc *issuePhotoUseCase) Upload(ctx context.Context, req *issue.UploadPhotoRequest, fileReader io.Reader, fileSize int64, originalFileName, contentType string) (*issue.IssuePhoto, error) {
	if fileSize == 0 {
		return nil, errors.New("empty file")
	}

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
	if err != nil {
		return nil, err
	}

	// Encrypt sensitive info
	encPublicURL := publicURL
	if enc, err := uc.cryptoSvc.Encrypt(publicURL); err == nil {
		encPublicURL = enc
	}

	encObjectName := objectName
	if enc, err := uc.cryptoSvc.Encrypt(objectName); err == nil {
		encObjectName = enc
	}

	p := &issue.IssuePhoto{
		IssuePhotoID:   idgen.GenerateRandom(idgen.PrefixIssuePhoto),
		IssueID:        req.IssueID,
		PICUserID:      req.PICUserID,
		PhotoType:      req.PhotoType,
		ImageUrl:       encPublicURL,
		FileName:       encObjectName, // we store objectName here so we can delete it later
		FollowUpDate:   req.FollowUpDate,
		JumlahFollowUp: req.JumlahFollowUp,
	}
	err = uc.repo.Create(p)
	if err == nil && uc.sseBroker != nil {
		uc.sseBroker.Broadcast("ISSUE_UPDATED")
	}
	return p, err
}

func (uc *issuePhotoUseCase) Delete(ctx context.Context, id string) error {
	photo, err := uc.repo.FindByID(id)
	if err != nil {
		return errors.New("photo not found")
	}

	// Decrypt the filename before deleting from MinIO
	plainFileName := uc.cryptoSvc.DecryptWithFallback(photo.FileName)
	_ = uc.storage.Delete(ctx, plainFileName) // best-effort file deletion from MinIO

	err = uc.repo.Delete(id)
	if err == nil && uc.sseBroker != nil {
		uc.sseBroker.Broadcast("ISSUE_UPDATED")
	}
	return err
}

// Ensure io is imported even if not directly used (for future streaming support)
var _ io.Reader = nil
