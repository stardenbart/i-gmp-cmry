package inspectionusecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/monitoring-system/backend/internal/domain/events"
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/pkg/idgen"
	"github.com/monitoring-system/backend/pkg/kafka"
)

// ── Inspection Header UseCase ─────────────────────────────────────────────

type inspectionHeaderUseCase struct {
	repo     inspection.InspectionHeaderRepository
	producer kafka.EventProducer
}

func NewInspectionHeaderUseCase(repo inspection.InspectionHeaderRepository, producer kafka.EventProducer) inspection.InspectionHeaderUseCase {
	return &inspectionHeaderUseCase{repo: repo, producer: producer}
}

func (uc *inspectionHeaderUseCase) GetAll(page, limit int, areaID, status, inspectorID string) ([]inspection.InspectionHeader, int64, error) {
	return uc.repo.FindAll(page, limit, areaID, status, inspectorID)
}

func (uc *inspectionHeaderUseCase) GetByID(id string) (*inspection.InspectionHeader, error) {
	return uc.repo.FindByID(id)
}

func (uc *inspectionHeaderUseCase) Create(inspectorID string, req *inspection.CreateInspectionRequest) (*inspection.InspectionHeader, error) {
	h := &inspection.InspectionHeader{
		InspectionID:          "INS - " + idgen.Generate(idgen.PrefixInspection),
		AreaID:                 req.AreaID,
		KawasanID:              req.KawasanID,
		DetailKawasanID:        req.DetailKawasanID,
		InspectorID:            inspectorID,
		InspectionHeaderStatus: inspection.InspectionStatusDraft,
	}
	err := uc.repo.Create(h)
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
