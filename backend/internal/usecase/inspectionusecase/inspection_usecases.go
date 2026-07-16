package inspectionusecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/monitoring-system/backend/internal/domain/events"
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/idgen"
	"github.com/monitoring-system/backend/pkg/kafka"
)

// ── Inspection Header UseCase ─────────────────────────────────────────────

type inspectionHeaderUseCase struct {
	repo              inspection.InspectionHeaderRepository
	producer          kafka.EventProducer
	detailKawasanRepo master.DetailKawasanRepository
	emailNotifier     *InspectionEmailNotifier
}

func NewInspectionHeaderUseCase(
	repo inspection.InspectionHeaderRepository,
	producer kafka.EventProducer,
	detailKawasanRepo master.DetailKawasanRepository,
	emailNotifier *InspectionEmailNotifier,
) inspection.InspectionHeaderUseCase {
	return &inspectionHeaderUseCase{
		repo:              repo,
		producer:          producer,
		detailKawasanRepo: detailKawasanRepo,
		emailNotifier:     emailNotifier,
	}
}

func (uc *inspectionHeaderUseCase) GetAll(page, limit int, areaID, status, inspectorID string) ([]inspection.InspectionHeader, int64, error) {
	return uc.repo.FindAll(page, limit, areaID, status, inspectorID)
}

func (uc *inspectionHeaderUseCase) GetByID(id string) (*inspection.InspectionHeader, error) {
	return uc.repo.FindByID(id)
}

func (uc *inspectionHeaderUseCase) GetAreaStatus(areaID string) (inspection.AreaProgress, error) {
	return CalculateAreaStatus(areaID, uc.detailKawasanRepo, uc.repo)
}

func (uc *inspectionHeaderUseCase) Create(inspectorID string, req *inspection.CreateInspectionRequest) (*inspection.InspectionHeader, error) {
	now := time.Now()
	
	// 0. Cek apakah kawasan sudah pernah diinspeksi (Selesai/Approved) di bulan yang sama
	completedCount, err := uc.repo.CountCompletedThisMonthByKawasan(req.KawasanID, now.Year(), int(now.Month()))
	if err == nil && completedCount > 0 {
		return nil, errors.New("kawasan ini sudah selesai diinspeksi pada bulan ini, anda baru bisa melakukan inspeksi lagi bulan depan")
	}

	// 1. Cek apakah auditor punya inspeksi aktif di kawasan lain
	activeByMe, err := uc.repo.FindActiveByInspector(inspectorID)
	if err == nil && len(activeByMe) > 0 {
		for _, a := range activeByMe {
			if a.KawasanID != req.KawasanID {
				return nil, errors.New("anda memiliki inspeksi aktif di kawasan lain, harap selesaikan terlebih dahulu")
			}
		}
	}

	// 2. Cek apakah ada auditor lain yang sedang inspeksi kawasan ini
	activeInKawasan, err := uc.repo.FindActiveByKawasan(req.KawasanID)
	var sessionID string
	if err == nil && len(activeInKawasan) > 0 {
		for _, a := range activeInKawasan {
			if a.InspectorID != inspectorID {
				return nil, errors.New("kawasan ini sedang diinspeksi oleh auditor lain")
			}
			if a.SessionID != nil {
				sessionID = *a.SessionID
			}
		}
	}

	if sessionID == "" {
		sessionID = idgen.Generate(idgen.PrefixSession)
	}

	h := &inspection.InspectionHeader{
		InspectionID:          	idgen.Generate(idgen.PrefixInspection),
		AreaID:                 req.AreaID,
		KawasanID:              req.KawasanID,
		DetailKawasanID:        req.DetailKawasanID,
		InspectorID:            inspectorID,
		InspectionHeaderStatus: inspection.InspectionStatusDraft,
		SessionID:              &sessionID,
		LockedAt:               &now,
	}
	err = uc.repo.Create(h)
	if err == nil {
		event := events.InspectionEvent{
			BaseEvent: events.BaseEvent{
				EventID:   uuid.New().String(),
				EventType: events.EventTypeCreated,
				Timestamp: time.Now(),
				ActorID:   inspectorID,
			},
			InspectionID: h.InspectionID,
			PICUserID:    inspectorID, // Using inspector as PIC initially
			Status:       string(h.InspectionHeaderStatus),
			AreaID:       h.AreaID,
		}
		_ = uc.producer.PublishEvent(context.Background(), events.TopicAuditInspections, h.InspectionID, event)
	}
	return h, err
}

func (uc *inspectionHeaderUseCase) UpdateStatus(id string, actorID string, req *inspection.UpdateInspectionStatusRequest) (*inspection.InspectionHeader, error) {
	h, err := uc.repo.FindByID(id)
	if err != nil { return nil, errors.New("inspection not found") }
	h.InspectionHeaderStatus = req.Status
	err = uc.repo.Update(h)
	if err == nil {
		event := events.InspectionEvent{
			BaseEvent: events.BaseEvent{
				EventID:   uuid.New().String(),
				EventType: events.EventTypeUpdated,
				Timestamp: time.Now(),
				ActorID:   actorID,
			},
			InspectionID: h.InspectionID,
			Status:       string(h.InspectionHeaderStatus),
			AreaID:       h.AreaID,
		}
		_ = uc.producer.PublishEvent(context.Background(), events.TopicAuditInspections, h.InspectionID, event)

		// Check area progress if status is Completed
		if h.InspectionHeaderStatus == inspection.InspectionStatusCompleted {
			progress, _ := CalculateAreaStatus(h.AreaID, uc.detailKawasanRepo, uc.repo)
			if progress.Status == inspection.AreaStatusConfirmed {
				// Send email summary
				go func(areaID string, p inspection.AreaProgress) {
					_ = uc.emailNotifier.SendInspectionSummary(areaID, p)
				}(h.AreaID, progress)

				// Publish CONFIRMED event
				confirmedEvent := events.BaseEvent{
					EventID:   uuid.New().String(),
					EventType: events.EventTypeConfirmed,
					Timestamp: time.Now(),
					ActorID:   "SYSTEM",
				}
				_ = uc.producer.PublishEvent(context.Background(), events.TopicAuditInspections, h.AreaID, confirmedEvent)
			}
		}
	}
	return h, err
}

func (uc *inspectionHeaderUseCase) Delete(id string) error { return uc.repo.Delete(id) }

// ── Inspection Result UseCase ─────────────────────────────────────────────

type inspectionResultUseCase struct{ repo inspection.InspectionResultRepository }

func NewInspectionResultUseCase(repo inspection.InspectionResultRepository) inspection.InspectionResultUseCase {
	return &inspectionResultUseCase{repo: repo}
}

func (uc *inspectionResultUseCase) GetByInspectionID(id string) ([]inspection.InspectionResult, error) {
	return uc.repo.FindByInspectionID(id)
}

func (uc *inspectionResultUseCase) GetByID(id string) (*inspection.InspectionResult, error) {
	return uc.repo.FindByID(id)
}

func (uc *inspectionResultUseCase) BulkSave(req *inspection.BulkSaveResultRequest) error {
	var results []inspection.InspectionResult
	for _, r := range req.Results {
		results = append(results, inspection.InspectionResult{
			ResultID:     idgen.Generate(idgen.PrefixResult),
			InspectionID: req.InspectionID,
			UraianID:     r.UraianID,
			Checking:     r.Checking,
			Nilai:        r.Nilai,
			Keterangan:   r.Keterangan,
		})
	}
	return uc.repo.BulkCreate(results)
}

func (uc *inspectionResultUseCase) Update(id string, req *inspection.SaveResultRequest) (*inspection.InspectionResult, error) {
	r, err := uc.repo.FindByID(id)
	if err != nil { return nil, errors.New("result not found") }
	r.Checking = req.Checking
	r.Nilai = req.Nilai
	r.Keterangan = req.Keterangan
	return r, uc.repo.Update(r)
}

func (uc *inspectionResultUseCase) Delete(id string) error { return uc.repo.Delete(id) }
