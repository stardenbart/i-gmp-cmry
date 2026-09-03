package inspectionusecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/monitoring-system/backend/internal/authz"
	"github.com/monitoring-system/backend/internal/domain/events"
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/domain/issue"
	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/inspectionrepo"
	"github.com/monitoring-system/backend/pkg/idgen"
	"github.com/monitoring-system/backend/pkg/kafka"
)

// ── Inspection Header UseCase ─────────────────────────────────────────────

type inspectionHeaderUseCase struct {
	repo              inspection.InspectionHeaderRepository
	producer          kafka.EventProducer
	detailKawasanRepo master.DetailKawasanRepository
	kawasanRepo       master.KawasanRepository
	areaRepo          master.AreaRepository
	emailNotifier     *InspectionEmailNotifier
}

func NewInspectionHeaderUseCase(
	repo inspection.InspectionHeaderRepository,
	producer kafka.EventProducer,
	detailKawasanRepo master.DetailKawasanRepository,
	kawasanRepo master.KawasanRepository,
	areaRepo master.AreaRepository,
	emailNotifier *InspectionEmailNotifier,
) inspection.InspectionHeaderUseCase {
	return &inspectionHeaderUseCase{
		repo:              repo,
		producer:          producer,
		detailKawasanRepo: detailKawasanRepo,
		kawasanRepo:       kawasanRepo,
		areaRepo:          areaRepo,
		emailNotifier:     emailNotifier,
	}
}

// resolvePeriodForArea resolves the [start, end) boundaries of the
// currently-running inspection period for the plant that owns `areaID`,
// using that plant's INSPECTION_PERIOD_CUTOFF_DAY setting (per-plant, with
// the usual fallback to the global default when unset). Falls back to an
// exact calendar month (cutoffDay=1) if the Area or setting can't be
// resolved for any reason — never blocks the caller.
func (uc *inspectionHeaderUseCase) resolvePeriodForArea(areaID string) inspection.InspectionPeriodInfo {
	plantID := ""
	if uc.areaRepo != nil {
		if area, err := uc.areaRepo.FindByID(areaID); err == nil && area != nil && area.PlantID != nil {
			plantID = *area.PlantID
		}
	}

	cutoffDay := DefaultInspectionPeriodCutoffDay
	if uc.emailNotifier != nil && uc.emailNotifier.settingRepo != nil {
		if s, err := uc.emailNotifier.settingRepo.FindByKey(master.SettingKeyInspectionPeriodCutoffDay, plantID); err == nil && s != nil {
			if d, errConv := strconv.Atoi(s.SettingValue); errConv == nil {
				cutoffDay = d
			}
		}
	}

	start, end := ResolveInspectionPeriod(time.Now(), cutoffDay)
	return inspection.InspectionPeriodInfo{CutoffDay: cutoffDay, PeriodStart: start, PeriodEnd: end}
}

func (uc *inspectionHeaderUseCase) GetAll(page, limit int, plantID, areaID, status, inspectorID string) ([]inspection.InspectionHeader, int64, error) {
	return uc.repo.FindAll(page, limit, plantID, areaID, status, inspectorID)
}

func (uc *inspectionHeaderUseCase) GetByID(id string) (*inspection.InspectionHeader, error) {
	// Delegasikan langsung ke repo — repo layer sudah punya singleflight + TTL cache 5 menit sendiri.
	// Menambah singleflight kedua di sini justru menciptakan serial blocking dengan shared 3s deadline.
	return uc.repo.FindByID(id)
}

// GetByIDScoped is GetByID plus a per-record plant check — see the
// interface doc comment for why this exists as a separate method rather
// than changing GetByID's behavior.
func (uc *inspectionHeaderUseCase) GetByIDScoped(id, userPlantID string) (*inspection.InspectionHeader, error) {
	item, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("inspection not found")
	}
	if !uc.callerMayAccessArea(userPlantID, item.AreaID) {
		return nil, errors.New("inspection not found")
	}
	return item, nil
}

// callerMayAccessArea resolves areaID's owning plant and checks it against
// userPlantID. Fails CLOSED: if the area can't be resolved at all (bad
// data, repo error), a plant-scoped caller is denied rather than let
// through — an unscoped caller (userPlantID == "") is unaffected either
// way, matching authz.PlantMatches' own contract.
func (uc *inspectionHeaderUseCase) callerMayAccessArea(userPlantID, areaID string) bool {
	if userPlantID == "" {
		return true
	}
	if uc.areaRepo == nil {
		return false
	}
	area, err := uc.areaRepo.FindByID(areaID)
	if err != nil || area == nil {
		return false
	}
	return authz.PlantMatches(userPlantID, area.PlantID)
}

func (uc *inspectionHeaderUseCase) GetAreaStatus(areaID string) (inspection.AreaProgress, error) {
	period := uc.resolvePeriodForArea(areaID)
	return CalculateAreaStatus(areaID, uc.detailKawasanRepo, uc.repo, period.PeriodStart, period.PeriodEnd)
}

// GetCurrentPeriodInfo exposes the resolved inspection period boundaries so
// the frontend never has to re-derive the cutoff-day math itself (that
// duplication is exactly what caused the "Selesai Bulan Ini" dropdown check
// on the Buat Inspeksi page to drift from the backend's actual gate).
func (uc *inspectionHeaderUseCase) GetCurrentPeriodInfo(areaID string) (inspection.InspectionPeriodInfo, error) {
	return uc.resolvePeriodForArea(areaID), nil
}

func (uc *inspectionHeaderUseCase) GetTrend(contextID string, year int) ([]inspection.TrendData, error) {
	return uc.repo.GetTrendByContext(contextID, year)
}

func (uc *inspectionHeaderUseCase) Create(inspectorID string, req *inspection.CreateInspectionRequest) (*inspection.InspectionHeader, error) {
	now := time.Now()

	// 0. Cek apakah detail kawasan sudah pernah diinspeksi (Selesai/Approved)
	// di periode inspeksi yang sedang berjalan (cutoff day admin-configurable
	// per plant — lihat resolvePeriodForArea).
	period := uc.resolvePeriodForArea(req.AreaID)
	completedCount, err := uc.repo.CountCompletedInPeriod(req.DetailKawasanID, period.PeriodStart, period.PeriodEnd)
	if err == nil && completedCount > 0 {
		return nil, fmt.Errorf("detail kawasan ini sudah selesai diinspeksi pada periode berjalan, anda baru bisa melakukan inspeksi lagi mulai %s", period.PeriodEnd.Format("2 January 2006"))
	}

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

		// Invalidasi header cache — status berubah, cache lama tidak valid
		inspectionrepo.InvalidateHeaderCache(h.InspectionID)

		// Record LastInspection timestamp on DetailKawasan & Kawasan when Completed
		if h.InspectionHeaderStatus == inspection.InspectionStatusCompleted {
			now := time.Now()
			if uc.detailKawasanRepo != nil {
				_ = uc.detailKawasanRepo.UpdateLastInspection(h.DetailKawasanID, now)
			}
			if uc.kawasanRepo != nil {
				_ = uc.kawasanRepo.UpdateLastInspection(h.KawasanID, now)
			}

			// Resolved once from this inspection's Area — one InspectionHeader
			// always belongs to exactly one Area (and thus one plant), so the
			// same period boundaries apply to both the Area- and
			// Kawasan-level checks below.
			period := uc.resolvePeriodForArea(h.AreaID)

			progress, _ := CalculateAreaStatus(h.AreaID, uc.detailKawasanRepo, uc.repo, period.PeriodStart, period.PeriodEnd)
			if progress.Status == inspection.AreaStatusConfirmed {
				// Admin-configurable via Settings: skip the "Area selesai"
				// email/notification entirely when explicitly turned off.
				// Missing setting (fresh install without the migration yet,
				// or the key not seeded) defaults to enabled.
				if uc.isCompletionNotifyEnabled(master.SettingKeyNotifyOnAreaComplete) {
					// Send email summary
					go func(areaID string, p inspection.AreaProgress) {
						_ = uc.emailNotifier.SendInspectionSummary(areaID, p)
					}(h.AreaID, progress)
				}

				// Publish CONFIRMED event
				confirmedEvent := events.BaseEvent{
					EventID:   uuid.New().String(),
					EventType: events.EventTypeConfirmed,
					Timestamp: time.Now(),
					ActorID:   "SYSTEM",
				}
				_ = uc.producer.PublishEvent(context.Background(), events.TopicAuditInspections, h.AreaID, confirmedEvent)
			}

			// Narrower than the Area-wide check above: notify the Kawasan's
			// Manager (KategoriPIC "Manager") specifically once every
			// DetailKawasan under THIS Kawasan is done this month — an Area
			// can contain several Kawasan, so "Area confirmed" and "Kawasan
			// confirmed" are different events with different audiences.
			kawasanProgress, _ := CalculateKawasanStatus(h.KawasanID, uc.detailKawasanRepo, uc.repo, period.PeriodStart, period.PeriodEnd)
			if kawasanProgress.Status == inspection.KawasanStatusConfirmed {
				if uc.isCompletionNotifyEnabled(master.SettingKeyNotifyOnKawasanComplete) {
					go func(kawasanID string, p inspection.KawasanProgress) {
						_ = uc.emailNotifier.SendKawasanInspectionSummary(kawasanID, p)
					}(h.KawasanID, kawasanProgress)
				}
			}
		}
	}
	return h, err
}

func (uc *inspectionHeaderUseCase) Delete(id string) error { return uc.repo.Delete(id) }

// isCompletionNotifyEnabled reads a global (plant-agnostic) on/off toggle
// for the "fully inspected" email/notification triggers, e.g.
// NOTIFY_ON_KAWASAN_COMPLETE / NOTIFY_ON_AREA_COMPLETE. Missing setting
// (fresh install predating migration 041, or a plant that never had the
// row seeded) is treated as enabled — only an explicit "false" disables it.
func (uc *inspectionHeaderUseCase) isCompletionNotifyEnabled(settingKey string) bool {
	if uc.emailNotifier == nil || uc.emailNotifier.settingRepo == nil {
		return true
	}
	setting, err := uc.emailNotifier.settingRepo.FindByKey(settingKey, "")
	if err != nil || setting == nil {
		return true
	}
	return setting.SettingValue != "false"
}

// ── Inspection Result UseCase ─────────────────────────────────────────────

type inspectionResultUseCase struct {
	repo     inspection.InspectionResultRepository
	issueUC  issue.IssueUseCase
	headerUC inspection.InspectionHeaderUseCase
}

func NewInspectionResultUseCase(
	repo inspection.InspectionResultRepository,
	issueUC issue.IssueUseCase,
	headerUC inspection.InspectionHeaderUseCase,
) inspection.InspectionResultUseCase {
	return &inspectionResultUseCase{repo: repo, issueUC: issueUC, headerUC: headerUC}
}

func (uc *inspectionResultUseCase) GetByInspectionID(id string) ([]inspection.InspectionResult, error) {
	return uc.repo.FindByInspectionID(id)
}

func (uc *inspectionResultUseCase) GetByID(id string) (*inspection.InspectionResult, error) {
	return uc.repo.FindByID(id)
}

func (uc *inspectionResultUseCase) BulkSave(req *inspection.BulkSaveResultRequest) ([]inspection.InspectionResult, error) {
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
	if err := uc.repo.BulkCreate(results); err != nil {
		return nil, err
	}

	// Invalidasi cache agar GetChecklist + GetByID berikutnya fetch data fresh.
	inspectionrepo.InvalidateChecklistCache(req.InspectionID)
	inspectionrepo.InvalidateHeaderCache(req.InspectionID)

	// Issue adalah bagian wajib dari finalisasi hasil NG. Jangan menelan error di
	// sini: frontend tidak boleh melanjutkan status ke Completed bila sinkronisasi
	// issue gagal. Operasi ini idempotent sehingga bulk-save aman untuk diulang.
	if uc.issueUC == nil || uc.headerUC == nil {
		return nil, errors.New("issue synchronization is not configured")
	}

	header, err := uc.headerUC.GetByID(req.InspectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load inspection owner for issue synchronization: %w", err)
	}
	if header == nil {
		return nil, errors.New("inspection not found during issue synchronization")
	}
	if header.InspectorID == "" {
		return nil, errors.New("inspection has no inspector for issue assignment")
	}

	for _, r := range results {
		switch r.Checking {
		case "NG":
			if _, err := uc.issueUC.Create(header.InspectorID, &issue.CreateIssueRequest{
				ResultID:       r.ResultID,
				IssuePICUserID: header.InspectorID,
				Keterangan:     r.Keterangan,
			}); err != nil {
				return nil, fmt.Errorf("failed to synchronize issue for result %s: %w", r.ResultID, err)
			}
		case "OK":
			if err := uc.issueUC.CloseByResultID(r.ResultID, header.InspectorID); err != nil {
				return nil, fmt.Errorf("failed to close issue for result %s: %w", r.ResultID, err)
			}
		}
	}

	return results, nil
}

func (uc *inspectionResultUseCase) Update(id string, req *inspection.SaveResultRequest) (*inspection.InspectionResult, error) {
	r, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("result not found")
	}
	r.Checking = req.Checking
	r.Nilai = req.Nilai
	r.Keterangan = req.Keterangan
	updated, err := r, uc.repo.Update(r)
	if err == nil {
		// Invalidasi cache checklist dan header saat satu item hasil diupdate
		inspectionrepo.InvalidateChecklistCache(r.InspectionID)
		inspectionrepo.InvalidateHeaderCache(r.InspectionID) // skor bisa berubah
	}
	return updated, err
}

func (uc *inspectionResultUseCase) Delete(id string) error { return uc.repo.Delete(id) }
