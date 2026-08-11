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
	kawasanRepo       master.KawasanRepository
	emailNotifier     *InspectionEmailNotifier
}

func NewInspectionHeaderUseCase(
	repo inspection.InspectionHeaderRepository,
	producer kafka.EventProducer,
	detailKawasanRepo master.DetailKawasanRepository,
	kawasanRepo master.KawasanRepository,
	emailNotifier *InspectionEmailNotifier,
) inspection.InspectionHeaderUseCase {
	return &inspectionHeaderUseCase{
		repo:              repo,
		producer:          producer,
		detailKawasanRepo: detailKawasanRepo,
		kawasanRepo:       kawasanRepo,
		emailNotifier:     emailNotifier,
	}
}

func (uc *inspectionHeaderUseCase) GetAll(page, limit int, plantID, areaID, status, inspectorID string) ([]inspection.InspectionHeader, int64, error) {
	return uc.repo.FindAll(page, limit, plantID, areaID, status, inspectorID)
}

func (uc *inspectionHeaderUseCase) GetByID(id string) (*inspection.InspectionHeader, error) {
	return uc.repo.FindByID(id)
}

func (uc *inspectionHeaderUseCase) GetAreaStatus(areaID string) (inspection.AreaProgress, error) {
	return CalculateAreaStatus(areaID, uc.detailKawasanRepo, uc.repo)
}

func (uc *inspectionHeaderUseCase) GetTrend(contextID string, year int) ([]inspection.TrendData, error) {
	return uc.repo.GetTrendByContext(contextID, year)
}

func (uc *inspectionHeaderUseCase) Create(inspectorID string, req *inspection.CreateInspectionRequest) (*inspection.InspectionHeader, error) {
	now := time.Now()

	// 1. Cek apakah ada inspeksi aktif (Draft/Ongoing) di detail kawasan ini (oleh siapapun)
	activeInDK, err := uc.repo.FindActiveByDetailKawasan(req.DetailKawasanID)
	if err == nil && len(activeInDK) > 0 {
		for _, a := range activeInDK {
			if a.InspectorID == inspectorID {
				// Re-use inspeksi aktif milik inspector yang sama
				return &a, nil
			}
			// Auditor lain sedang mengerjakan detail kawasan ini → tolak
			return nil, errors.New("detail kawasan ini sedang diinspeksi oleh auditor lain, harap tunggu sampai inspeksi selesai")
		}
	}

	sessionID := idgen.Generate(idgen.PrefixSession)

	h := &inspection.InspectionHeader{
		InspectionID:           idgen.Generate(idgen.PrefixInspection),
		AreaID:                 req.AreaID,
		KawasanID:              req.KawasanID,
		DetailKawasanID:        req.DetailKawasanID,
		InspectorID:            inspectorID,
		InspectionHeaderStatus: inspection.InspectionStatusOngoing,
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
	if err != nil {
		return nil, errors.New("inspection not found")
	}

	// Reject setting status back to Draft
	if req.Status == inspection.InspectionStatusDraft {
		return nil, errors.New("status inspeksi tidak dapat diubah kembali ke Draft")
	}

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

		// Record LastInspection timestamp on DetailKawasan & Kawasan when Completed
		if h.InspectionHeaderStatus == inspection.InspectionStatusCompleted {
			now := time.Now()
			if uc.detailKawasanRepo != nil {
				_ = uc.detailKawasanRepo.UpdateLastInspection(h.DetailKawasanID, now)
			}
			if uc.kawasanRepo != nil {
				_ = uc.kawasanRepo.UpdateLastInspection(h.KawasanID, now)
			}

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

type inspectionResultUseCase struct {
	repo inspection.InspectionResultRepository
}

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
	// Load existing results so we can re-use their ResultIDs (upsert by UraianID)
	existing, _ := uc.repo.FindByInspectionID(req.InspectionID)
	existingMap := make(map[string]string, len(existing)) // uraianID -> resultID
	for _, e := range existing {
		existingMap[e.UraianID] = e.ResultID
	}

	var results []inspection.InspectionResult
	for _, r := range req.Results {
		resultID := existingMap[r.UraianID]
		if resultID == "" {
			resultID = idgen.Generate(idgen.PrefixResult)
		}
		results = append(results, inspection.InspectionResult{
			ResultID:     resultID,
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
	if err != nil {
		return nil, errors.New("result not found")
	}
	r.Checking = req.Checking
	r.Nilai = req.Nilai
	r.Keterangan = req.Keterangan
	return r, uc.repo.Update(r)
}

func (uc *inspectionResultUseCase) Delete(id string) error { return uc.repo.Delete(id) }
