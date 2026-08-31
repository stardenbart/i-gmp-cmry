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
	notificationdomain "github.com/monitoring-system/backend/internal/domain/notification"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/eventstore"
	"github.com/monitoring-system/backend/pkg/idgen"
	"github.com/monitoring-system/backend/pkg/kafka"
	"github.com/monitoring-system/backend/pkg/mail"
	"github.com/monitoring-system/backend/pkg/storage"
	redis "github.com/redis/go-redis/v9"
)

// ── Issue UseCase ─────────────────────────────────────────────────────────

type issueUseCase struct {
	repo           issue.IssueRepository
	photoRepo      issue.IssuePhotoRepository
	heiRepo        issue.IssueHEIRepository
	storage        *storage.MinioStorage
	userRepo       authdomain.UserRepository // to fetch PIC email
	delegateRepo   issue.IssueDelegateRepository
	producer       kafka.EventProducer
	mailer         mail.Mailer
	settingRepo    masterdomain.SettingRepository
	cryptoSvc      *crypto.Service
	rdb            *redis.Client
	notificationUC notificationdomain.NotificationUseCase // in-app bell notifications, alongside email
}

func NewIssueUseCase(repo issue.IssueRepository, photoRepo issue.IssuePhotoRepository, heiRepo issue.IssueHEIRepository, storage *storage.MinioStorage, producer kafka.EventProducer, mailer mail.Mailer, userRepo authdomain.UserRepository, settingRepo masterdomain.SettingRepository, delegateRepo issue.IssueDelegateRepository, cryptoSvc *crypto.Service, rdb *redis.Client, notificationUC notificationdomain.NotificationUseCase) issue.IssueUseCase {
	return &issueUseCase{repo: repo, photoRepo: photoRepo, heiRepo: heiRepo, storage: storage, producer: producer, mailer: mailer, userRepo: userRepo, settingRepo: settingRepo, delegateRepo: delegateRepo, cryptoSvc: cryptoSvc, rdb: rdb, notificationUC: notificationUC}
}

// notify creates an in-app bell notification for a user, best-effort — a
// failure here must never break the issue action it's attached to (email
// already carries the same information).
func (uc *issueUseCase) notify(userID, nType, title, message, link string) {
	if uc.notificationUC == nil || userID == "" {
		return
	}
	_ = uc.notificationUC.CreateSystemNotification(userID, nType, title, message, link)
}

func (uc *issueUseCase) ConsolidateDuplicateActiveIssues() error {
	return uc.repo.ConsolidateDuplicateActiveIssues()
}

func (uc *issueUseCase) GetAll(page, limit int, plantID, status, picUserID string, needsWOWR *bool) ([]issue.Issue, int64, error) {
	items, total, err := uc.repo.FindAll(page, limit, plantID, status, picUserID, needsWOWR)
	if err == nil {
		now := time.Now()
		for i := range items {
			items[i].ComputedIssueStatus = items[i].ComputedStatus(now)
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

	// Deduplicate items by IssueID to prevent SQL join duplication
	seen := make(map[string]bool)
	unique := make([]issue.Issue, 0, len(items))
	for _, item := range items {
		if item.IssueID != "" && !seen[item.IssueID] {
			seen[item.IssueID] = true
			unique = append(unique, item)
		}
	}
	items = unique

	return items, total, err
}

func (uc *issueUseCase) GetByID(id string) (*issue.Issue, error) {
	item, err := uc.repo.FindByID(id)
	if err == nil && item != nil {
		item.ComputedIssueStatus = item.ComputedStatus(time.Now())
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

func (uc *issueUseCase) GetByResultID(resultID string) (*issue.Issue, error) {
	item, err := uc.repo.FindByResultID(resultID)
	if err == nil && item != nil {
		item.ComputedIssueStatus = item.ComputedStatus(time.Now())
		item.Keterangan = uc.cryptoSvc.DecryptWithFallback(item.Keterangan)
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
		// Resolve the deadline from the PIC's plant setting first, falling back to
		// the global default (FindByKey already does this) — each plant can define
		// its own ISSUE_DEADLINE_DAYS instead of sharing one hardcoded value.
		plantID := ""
		if pic, errPic := uc.userRepo.FindByID(req.IssuePICUserID); errPic == nil && pic != nil && pic.PlantID != nil {
			plantID = *pic.PlantID
		}
		s, err := uc.settingRepo.FindByKey(masterdomain.SettingKeyIssueDeadlineDays, plantID)
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

	// A finding is uniquely identified by its inspection ResultID. Different NG
	// uraian in the same location must remain separate issues so their photos and
	// follow-up histories cannot overwrite each other.
	var existing *issue.Issue
	if item, errExist := uc.repo.FindByResultID(req.ResultID); errExist == nil && item != nil {
		existing = item
	}

	if existing != nil {
		if req.IssuePICUserID != "" {
			existing.IssuePICUserID = req.IssuePICUserID
		}
		if req.Keterangan != "" {
			if existing.Keterangan != "" {
				decExisting := uc.cryptoSvc.DecryptWithFallback(existing.Keterangan)
				if decExisting != "" && decExisting != req.Keterangan && !strings.Contains(decExisting, req.Keterangan) {
					combined := decExisting + "; " + req.Keterangan
					if enc, err := uc.cryptoSvc.Encrypt(combined); err == nil {
						keteranganEnc = enc
					}
				}
			}
			existing.Keterangan = keteranganEnc
		}
		if req.Label != "" {
			existing.Label = req.Label
		}
		// Re-open issue when inspection NG finding is edited or re-submitted
		existing.IssueStatus = issue.IssueStatusOpen
		existing.FollowUpDelay = nil

		if errUpdate := uc.repo.Update(existing); errUpdate == nil {
			event := events.IssueEvent{
				BaseEvent: events.BaseEvent{
					EventID:   uuid.New().String(),
					EventType: events.EventTypeUpdated,
					Timestamp: time.Now(),
					ActorID:   actorID,
				},
				IssueID:        existing.IssueID,
				ResultID:       existing.ResultID,
				IssuePICUserID: existing.IssuePICUserID,
				Status:         string(existing.IssueStatus),
				DueDate:        existing.DueDate,
			}
			_ = uc.producer.PublishEvent(context.Background(), events.TopicAuditIssues, existing.IssueID, event)
			eventstore.GetEventStore(uc.rdb).PushGlobal("ISSUE_UPDATED")
			return existing, nil
		} else {
			return nil, fmt.Errorf("failed to update existing issue: %w", errUpdate)
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
		eventstore.GetEventStore(uc.rdb).PushGlobal("ISSUE_UPDATED")

		// Use req.Keterangan (plaintext) rather than i.Keterangan, which holds the
		// encrypted value stored on the record.
		uc.notify(i.IssuePICUserID, "info", "Issue Baru Ditugaskan",
			fmt.Sprintf("Anda mendapat temuan baru: %s", req.Keterangan), "/issues/"+i.IssueID)

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

			plantID := ""
			if pic.PlantID != nil {
				plantID = *pic.PlantID
			}
			tmpl := mail.TmplIssueAssignment
			if s, err := uc.settingRepo.FindByKey(masterdomain.SettingKeyEmailTemplateIssue, plantID); err == nil && s.SettingValue != "" {
				tmpl = s.SettingValue
			}

			_ = uc.mailer.SendTemplateForPlant(
				plantID,
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

func (uc *issueUseCase) CloseByResultID(resultID string, actorID string) error {
	existing, errExist := uc.repo.FindByResultID(resultID)
	if errExist != nil || existing == nil {
		return nil
	}
	if existing.IssueStatus == issue.IssueStatusClosed || existing.IssueStatus == issue.IssueStatusVerified {
		return nil
	}
	existing.IssueStatus = issue.IssueStatusClosed
	if errUpdate := uc.repo.Update(existing); errUpdate == nil {
		event := events.IssueEvent{
			BaseEvent: events.BaseEvent{
				EventID:   uuid.New().String(),
				EventType: events.EventTypeUpdated,
				Timestamp: time.Now(),
				ActorID:   actorID,
			},
			IssueID:        existing.IssueID,
			ResultID:       existing.ResultID,
			IssuePICUserID: existing.IssuePICUserID,
			Status:         string(existing.IssueStatus),
			DueDate:        existing.DueDate,
		}
		_ = uc.producer.PublishEvent(context.Background(), events.TopicAuditIssues, existing.IssueID, event)
		eventstore.GetEventStore(uc.rdb).PushGlobal("ISSUE_UPDATED")
	}
	return nil
}

func (uc *issueUseCase) Update(id string, actorID string, req *issue.UpdateIssueRequest) (*issue.Issue, error) {
	i, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("issue not found")
	}

	actorIsAuditor := false
	if actorID != "" {
		if actor, actorErr := uc.userRepo.FindByID(actorID); actorErr == nil && actor != nil {
			roleID := strings.ToUpper(strings.TrimSpace(actor.RoleID))
			actorIsAuditor = roleID == "ROLE-000" || roleID == "ROLE-001" || roleID == "ROLE-002" ||
				roleID == "SUPERADMIN" || roleID == "ADMIN" || roleID == "AUDITOR"
		}
	}

	if (req.IssueStatus == issue.IssueStatusClosed || req.IssueStatus == issue.IssueStatusVerified) && !actorIsAuditor {
		return nil, errors.New("hanya Admin atau Auditor yang dapat menyetujui dan menutup temuan")
	}
	if (req.WOWRStatus == issue.WOWRStatusVerified || req.WOWRStatus == issue.WOWRStatusRejected) && !actorIsAuditor {
		return nil, errors.New("hanya Admin atau Auditor yang dapat memvalidasi WO/WR")
	}
	if req.WOWRStatus == issue.WOWRStatusVerified {
		if uc.photoRepo == nil {
			return nil, errors.New("WO/WR tidak dapat diverifikasi: layanan foto bukti tidak tersedia")
		}
		photos, photoErr := uc.photoRepo.FindByIssueID(i.IssueID)
		if photoErr != nil {
			return nil, errors.New("gagal memeriksa foto bukti WO/WR")
		}
		// IssuePICUserID is always the inspector who ran the audit, never
		// the Auditee/PIC who actually uploads WO/WR evidence — comparing
		// the photo's uploader against it here always failed, permanently
		// blocking verification even when a valid photo existed.
		hasWOWRProof := false
		for _, photo := range photos {
			if photo.PhotoType == issue.PhotoTypeWOWR &&
				(photo.ImageUrl != "" || photo.FileName != "") {
				hasWOWRProof = true
				break
			}
		}
		if !hasWOWRProof {
			return nil, errors.New("WO/WR tidak dapat diverifikasi: Auditee belum mengunggah foto bukti penyelesaian WO/WR")
		}
	}

	// Validation logic for Follow Up
	if req.IssueStatus == issue.IssueStatusPendingValidation && i.IssueStatus != issue.IssueStatusPendingValidation {
		// Check if actor is PIC, Delegate, Auditor, or mapped PIC
		if actorID != i.IssuePICUserID {
			isDel, _ := uc.delegateRepo.IsDelegate(i.IssueID, actorID)
			if !isDel {
				user, errUser := uc.userRepo.FindByID(actorID)
				if errUser == nil && user != nil {
					r := strings.ToUpper(user.RoleID)
					// Strict check against valid system Role IDs:
					// ROLE-000 (Super Admin), ROLE-001 (Admin), ROLE-002 (Auditor)
					isAud := r == "ROLE-000" || r == "ROLE-001" || r == "ROLE-002"
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
				submittedAt := time.Now()
				i.WOWRSubmittedAt = &submittedAt
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
		if req.WOWRStatus != issue.WOWRStatusPendingValidation {
			// Resolved (Verified/Rejected/None) — stop counting toward the WO/WR
			// auto-approve timeout until it's resubmitted for validation again.
			i.WOWRSubmittedAt = nil
		}
		// If Auditor rejects WOWR proof, automatically delete existing WOWR proof photos from DB & MinIO
		if req.WOWRStatus == issue.WOWRStatusRejected && uc.photoRepo != nil {
			photos, _ := uc.photoRepo.FindByIssueID(i.IssueID)
			for _, p := range photos {
				if p.PhotoType == issue.PhotoTypeWOWR || p.PhotoType == issue.PhotoTypeFollowUp {
					_ = uc.photoRepo.Delete(p.IssuePhotoID)
					if uc.storage != nil && p.FileName != "" {
						_ = uc.storage.Delete(context.Background(), p.FileName)
					}
				}
			}
			uc.notify(i.IssuePICUserID, "warning", "Bukti WO/WR Ditolak",
				"Bukti WO/WR yang Anda unggah ditolak Auditor dan foto bukti telah dihapus. Silakan unggah ulang.", "/issues/"+i.IssueID)
		}
	}

	if req.IssueStatus != "" {
		// Validasi penyelesaian temuan (Closed, Verified, PendingValidation)
		if req.IssueStatus == issue.IssueStatusClosed || req.IssueStatus == issue.IssueStatusVerified || req.IssueStatus == issue.IssueStatusPendingValidation {
			// 1. Validasi Bukti Foto Follow-Up perbaikan (Wajib 1-banding-1 untuk seluruh foto temuan awal)
			photosToCheck := i.Photos
			if len(photosToCheck) == 0 && uc.photoRepo != nil {
				if dbPhotos, errP := uc.photoRepo.FindByIssueID(i.IssueID); errP == nil {
					photosToCheck = dbPhotos
				}
			}

			initialCount := 0
			followUpCount := 0
			initialPhotosMap := make(map[string]bool)

			isFinalApproval := req.IssueStatus == issue.IssueStatusClosed || req.IssueStatus == issue.IssueStatusVerified
			for _, p := range photosToCheck {
				if p.PhotoType == issue.PhotoTypeInitial {
					initialCount++
					initialPhotosMap[p.IssuePhotoID] = false
					if p.NeedsWOWR || p.WO_ID != "" || p.WR_ID != "" {
						if p.WO_ID == "" && p.WR_ID == "" {
							return nil, errors.New("gagal mengajukan temuan: terdapat foto yang menggunakan WO/WR tetapi Nomor WO/WR belum diisi")
						}
						if isFinalApproval && p.WOWRStatus != issue.WOWRStatusVerified {
							return nil, errors.New("gagal menutup temuan: terdapat WO/WR foto yang belum diverifikasi Admin atau Auditor")
						}
					}
				} else if p.PhotoType == issue.PhotoTypeFollowUp || p.PhotoType == issue.PhotoTypeWOWR {
					followUpCount++
					if p.RefPhotoID != nil && *p.RefPhotoID != "" {
						initialPhotosMap[*p.RefPhotoID] = true
					}
				}
			}

			if initialCount > 0 {
				missingFollowUpCount := 0
				for _, isFollowedUp := range initialPhotosMap {
					if !isFollowedUp {
						missingFollowUpCount++
					}
				}

				if followUpCount < initialCount || missingFollowUpCount > 0 {
					return nil, fmt.Errorf("gagal menyelesaikan temuan: terdapat %d foto bukti temuan awal, namun baru %d foto perbaikan (follow-up) yang diunggah. Seluruh %d foto temuan awal wajib memiliki bukti foto follow-up perbaikan", initialCount, followUpCount, initialCount)
				}
			} else if followUpCount == 0 {
				return nil, errors.New("gagal menyelesaikan temuan: wajib mengunggah setidaknya 1 bukti foto Follow-Up perbaikan terlebih dahulu")
			}

			// 2. Validasi WO/WR jika temuan memerlukan WO/WR
			if i.NeedsWOWR || i.WO_ID != "" || i.WR_ID != "" {
				if i.WO_ID == "" && i.WR_ID == "" {
					return nil, errors.New("gagal menyelesaikan temuan: Nomor WO/WR wajib diisi terlebih dahulu")
				}
				if isFinalApproval && i.WOWRStatus != issue.WOWRStatusVerified {
					return nil, errors.New("gagal menyelesaikan temuan: temuan ini menggunakan WO/WR. Harap tunggu konfirmasi (persetujuan) WO/WR oleh Auditor terlebih dahulu")
				}
			}
		}
		i.IssueStatus = req.IssueStatus

		// Calculate FollowUpDelay when closing/verifying issue
		if req.IssueStatus == issue.IssueStatusClosed || req.IssueStatus == issue.IssueStatusVerified {
			if i.DueDate != nil {
				delayDays := int(time.Since(*i.DueDate).Hours() / 24)
				if delayDays < 0 {
					delayDays = 0
				}
				i.FollowUpDelay = &delayDays
			}
		}
	}

	if req.Label != "" {
		i.Label = req.Label
	}

	// Update WO / WR logic
	if req.NeedsWOWR != nil {
		if *req.NeedsWOWR {
			i.NeedsWOWR = true
			if req.WO_ID != "" || req.WR_ID != "" {
				i.WO_ID = strings.ToUpper(strings.TrimSpace(req.WO_ID))
				i.WR_ID = strings.ToUpper(strings.TrimSpace(req.WR_ID))
			}
			if i.WO_ID == "" && i.WR_ID == "" {
				return nil, errors.New("jika membutuhkan WO/WR, minimal satu ID (WO atau WR) harus diisi")
			}
		} else {
			i.NeedsWOWR = false
			i.WO_ID = ""
			i.WR_ID = ""
			i.WOWRStatus = issue.WOWRStatusNone
		}
	} else if req.WO_ID != "" || req.WR_ID != "" {
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
		i.ComputedIssueStatus = i.ComputedStatus(time.Now())
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
		eventstore.GetEventStore(uc.rdb).PushGlobal("ISSUE_UPDATED")
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
		eventstore.GetEventStore(uc.rdb).PushGlobal("ISSUE_UPDATED")
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
		eventstore.GetEventStore(uc.rdb).PushGlobal("ISSUE_UPDATED")
	}
	return err
}

// ── Issue Photo UseCase ────────────────────────────────────────────────────

type issuePhotoUseCase struct {
	repo      issue.IssuePhotoRepository
	issueRepo issue.IssueRepository
	storage   *storage.MinioStorage
	cryptoSvc *crypto.Service
	rdb       *redis.Client
	producer  kafka.EventProducer
	// issueUC lets a freshly-uploaded WO/WR proof immediately re-check the
	// issue's auto-approve deadline (see TriggerWOWRAutoApproveCheck) instead
	// of waiting for the next hourly worker tick.
	issueUC issue.IssueUseCase
}

func NewIssuePhotoUseCase(repo issue.IssuePhotoRepository, issueRepo issue.IssueRepository, s *storage.MinioStorage, cryptoSvc *crypto.Service, rdb *redis.Client, producer kafka.EventProducer, issueUC issue.IssueUseCase) issue.IssuePhotoUseCase {
	return &issuePhotoUseCase{repo: repo, issueRepo: issueRepo, storage: s, cryptoSvc: cryptoSvc, rdb: rdb, producer: producer, issueUC: issueUC}
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

	// Check if issue is closed/locked
	if uc.issueRepo != nil {
		iss, errIss := uc.issueRepo.FindByID(req.IssueID)
		if errIss == nil && iss != nil {
			if iss.IssueStatus == issue.IssueStatusClosed || iss.IssueStatus == issue.IssueStatusVerified {
				return nil, errors.New("issue ini sudah ditutup (Closed) dan foto tidak dapat diunggah")
			}
			// Auto-transition Open/Overdue issue to InProgress upon first follow-up/WOWR photo upload
			if req.PhotoType == issue.PhotoTypeFollowUp || req.PhotoType == issue.PhotoTypeWOWR || req.RefPhotoID != nil {
				if iss.IssueStatus == issue.IssueStatusOpen || iss.IssueStatus == issue.IssueStatusOpenOverdue {
					iss.IssueStatus = issue.IssueStatusInProgress
					_ = uc.issueRepo.Update(iss)
				}
			}
		}
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

	// 1. STEP KAFKA: Publish event to Kafka event topic
	if uc.producer != nil {
		event := events.IssueEvent{
			BaseEvent: events.BaseEvent{
				EventID:   uuid.New().String(),
				EventType: events.EventTypeUpdated,
				Timestamp: time.Now(),
				ActorID:   req.PICUserID,
			},
			IssueID: req.IssueID,
		}
		_ = uc.producer.PublishEvent(ctx, events.TopicAuditIssues, req.IssueID, event)
	}

	// 2. STEP REDIS: Push real-time event notification to Redis SSE eventstore
	if uc.rdb != nil {
		eventstore.GetEventStore(uc.rdb).PushGlobal("ISSUE_UPDATED")
	}

	// 3. STEP MINIO: Stream & store binary image file to MinIO object storage
	publicURL := fmt.Sprintf("/uploads/%s", objectName)
	if uc.storage != nil {
		uploadedURL, err := uc.storage.UploadStream(ctx, objectName, fileReader, fileSize, contentType)
		if err != nil {
			return nil, fmt.Errorf("failed to upload image to MinIO: %w", err)
		}
		publicURL = uploadedURL
	}

	// 4. STEP POSTGRESQL: Save record into PostgreSQL database (store clean public URL and object name)
	p := &issue.IssuePhoto{
		IssuePhotoID: idgen.GenerateRandom(idgen.PrefixIssuePhoto),
		IssueID:      req.IssueID,
		RefPhotoID:   req.RefPhotoID,
		PICUserID:    req.PICUserID,
		PhotoType:    req.PhotoType,
		ImageUrl:     publicURL,
		FileName:     objectName,
		Keterangan:   req.Keterangan,
		HEIID:        req.HEIID,
		HEICategory: func() string {
			if req.HEICategory != nil {
				return *req.HEICategory
			}
			return ""
		}(),
		HabitID:          req.HabitID,
		EquipmentID:      req.EquipmentID,
		InfrastructureID: req.InfrastructureID,
		FollowUpDate:     req.FollowUpDate,
		JumlahFollowUp:   req.JumlahFollowUp,
	}
	err := uc.repo.Create(p)
	if err == nil {
		p.ImageUrl = publicURL
		p.FileName = objectName

		// WO/WR proof just landed — re-check this issue's auto-approve deadline
		// right away instead of waiting for the next hourly worker tick. If the
		// deadline had already elapsed while the auditor was waiting on this
		// proof, the issue is approved immediately.
		if req.PhotoType == issue.PhotoTypeWOWR && uc.issueUC != nil {
			go TriggerWOWRAutoApproveCheck(uc.issueUC, req.IssueID)
		}
	}
	return p, err
}

func (uc *issuePhotoUseCase) Update(ctx context.Context, photoID string, keterangan string) (*issue.IssuePhoto, error) {
	photo, err := uc.repo.FindByID(photoID)
	if err != nil {
		return nil, errors.New("photo not found")
	}

	// Check if issue is closed/locked (allow Initial photo updates when re-submitting inspection)
	if uc.issueRepo != nil && photo.PhotoType != "Initial" {
		iss, errIss := uc.issueRepo.FindByID(photo.IssueID)
		if errIss == nil && iss != nil {
			if iss.IssueStatus == issue.IssueStatusClosed || iss.IssueStatus == issue.IssueStatusVerified {
				return nil, errors.New("issue ini sudah ditutup (Closed) dan foto tidak dapat diubah")
			}
		}
	}

	photo.Keterangan = keterangan
	err = uc.repo.Update(photo)
	if err == nil {
		eventstore.GetEventStore(uc.rdb).PushGlobal("ISSUE_UPDATED")
	}
	return photo, err
}

func (uc *issuePhotoUseCase) UpdateHEI(ctx context.Context, photoID string, req *issue.UpdatePhotoHEIRequest) (*issue.IssuePhoto, error) {
	photo, err := uc.repo.FindByID(photoID)
	if err != nil {
		return nil, errors.New("photo not found")
	}

	// Check if issue is closed/locked (allow Initial photo updates when re-submitting inspection)
	if uc.issueRepo != nil && photo.PhotoType != "Initial" {
		iss, errIss := uc.issueRepo.FindByID(photo.IssueID)
		if errIss == nil && iss != nil {
			if iss.IssueStatus == issue.IssueStatusClosed || iss.IssueStatus == issue.IssueStatusVerified {
				return nil, errors.New("issue ini sudah ditutup (Closed) dan foto tidak dapat diubah")
			}
		}
	}

	if req.HEIID != nil {
		if *req.HEIID == "" {
			photo.HEIID = nil
		} else {
			photo.HEIID = req.HEIID
		}
	}
	if req.HEICategory != nil {
		photo.HEICategory = *req.HEICategory
	}
	if req.HabitID != nil {
		if *req.HabitID == "" {
			photo.HabitID = nil
		} else {
			photo.HabitID = req.HabitID
		}
	}
	if req.EquipmentID != nil {
		if *req.EquipmentID == "" {
			photo.EquipmentID = nil
		} else {
			photo.EquipmentID = req.EquipmentID
		}
	}
	if req.InfrastructureID != nil {
		if *req.InfrastructureID == "" {
			photo.InfrastructureID = nil
		} else {
			photo.InfrastructureID = req.InfrastructureID
		}
	}

	err = uc.repo.Update(photo)
	if err == nil {
		if updated, errFind := uc.repo.FindByID(photoID); errFind == nil {
			photo = updated
		}
		eventstore.GetEventStore(uc.rdb).PushGlobal("ISSUE_UPDATED")
	}
	return photo, err
}

func (uc *issuePhotoUseCase) UpdateWOWR(ctx context.Context, photoID string, req *issue.UpdatePhotoWOWRRequest) (*issue.IssuePhoto, error) {
	photo, err := uc.repo.FindByID(photoID)
	if err != nil {
		return nil, errors.New("photo not found")
	}

	if req.WOWRStatus == issue.WOWRStatusVerified {
		if uc.issueRepo == nil {
			return nil, errors.New("WO/WR tidak dapat diverifikasi: data PIC/Auditee tidak tersedia")
		}
		parentIssue, issueErr := uc.issueRepo.FindByID(photo.IssueID)
		if issueErr != nil || parentIssue == nil {
			return nil, errors.New("WO/WR tidak dapat diverifikasi: Issue tidak ditemukan")
		}
		photos, proofErr := uc.repo.FindByIssueID(photo.IssueID)
		if proofErr != nil {
			return nil, errors.New("gagal memeriksa foto bukti WO/WR")
		}
		// IssuePICUserID is always the inspector who ran the audit, never
		// the Auditee/PIC who actually uploads WO/WR evidence — comparing
		// the photo's uploader against it here always failed, permanently
		// blocking verification even when a valid photo existed.
		hasLinkedWOWRProof := false
		for _, proof := range photos {
			if proof.PhotoType == issue.PhotoTypeWOWR &&
				proof.RefPhotoID != nil && *proof.RefPhotoID == photoID &&
				(proof.ImageUrl != "" || proof.FileName != "") {
				hasLinkedWOWRProof = true
				break
			}
		}
		if !hasLinkedWOWRProof {
			return nil, errors.New("WO/WR tidak dapat diverifikasi: Auditee belum mengunggah foto bukti penyelesaian untuk temuan ini")
		}
	}

	if req.NeedsWOWR != nil {
		photo.NeedsWOWR = *req.NeedsWOWR
	}
	photo.WO_ID = req.WO_ID
	photo.WR_ID = req.WR_ID
	if req.WOWRStatus != "" {
		photo.WOWRStatus = req.WOWRStatus
	} else if photo.NeedsWOWR {
		photo.WOWRStatus = issue.WOWRStatusPendingValidation
	} else {
		photo.WOWRStatus = issue.WOWRStatusNone
	}

	err = uc.repo.Update(photo)
	if err == nil {
		if updated, errFind := uc.repo.FindByID(photoID); errFind == nil {
			photo = updated
		}
		eventstore.GetEventStore(uc.rdb).PushGlobal("ISSUE_UPDATED")
	}
	return photo, err
}

func (uc *issuePhotoUseCase) Delete(ctx context.Context, id string) error {
	photo, err := uc.repo.FindByID(id)
	if err != nil {
		return errors.New("photo not found")
	}

	// Verified WO/WR evidence is locked permanently, independent of the
	// parent issue's status — it's the audit trail proving the repair was
	// checked and approved, and must not be removable afterward. Rejected
	// evidence has no such lock: the PIC needs to delete it and re-upload.
	if photo.PhotoType == issue.PhotoTypeWOWR && photo.WOWRStatus == issue.WOWRStatusVerified {
		return errors.New("bukti WO/WR yang sudah diverifikasi tidak dapat dihapus")
	}

	// Check if issue is closed/locked
	if uc.issueRepo != nil {
		iss, errIss := uc.issueRepo.FindByID(photo.IssueID)
		if errIss == nil && iss != nil {
			if iss.IssueStatus == issue.IssueStatusClosed || iss.IssueStatus == issue.IssueStatusVerified {
				return errors.New("issue ini sudah ditutup (Closed) dan foto tidak dapat dihapus")
			}
		}
	}

	// Decrypt the filename before deleting from MinIO
	plainFileName := uc.cryptoSvc.DecryptWithFallback(photo.FileName)
	_ = uc.storage.Delete(ctx, plainFileName) // best-effort file deletion from MinIO

	err = uc.repo.Delete(id)
	if err == nil {
		eventstore.GetEventStore(uc.rdb).PushGlobal("ISSUE_UPDATED")
	}
	return err
}

// Ensure io is imported even if not directly used (for future streaming support)
var _ io.Reader = nil
