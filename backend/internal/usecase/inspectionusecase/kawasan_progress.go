package inspectionusecase

import (
	"time"

	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/domain/master"
)

// CalculateKawasanStatus calculates the overall status of a Kawasan by
// checking how many of its DetailKawasans have been completed within the
// given inspection period [periodStart, periodEnd) — the admin-configurable
// cutoff-day cycle resolved by the caller (see ResolveInspectionPeriod),
// the same "current cycle" convention already enforced when creating a new
// inspection (see CountCompletedInPeriod in inspectionHeaderUseCase.Create).
// Mirrors CalculateAreaStatus, but scoped one level narrower: a single
// Kawasan instead of every Kawasan under an Area.
func CalculateKawasanStatus(
	kawasanID string,
	detailKawasanRepo master.DetailKawasanRepository,
	inspectionRepo inspection.InspectionHeaderRepository,
	periodStart, periodEnd time.Time,
) (inspection.KawasanProgress, error) {
	allDK, err := detailKawasanRepo.FindByKawasanID(kawasanID)
	if err != nil {
		return inspection.KawasanProgress{}, err
	}

	total := len(allDK)
	if total == 0 {
		return inspection.KawasanProgress{
			KawasanID: kawasanID,
			Status:    inspection.KawasanStatusOpen,
		}, nil
	}

	completed := 0
	for _, dk := range allDK {
		count, _ := inspectionRepo.CountCompletedInPeriod(dk.DetailKawasanID, periodStart, periodEnd)
		if count > 0 {
			completed++
		}
	}

	var status inspection.KawasanInspectionStatus
	switch {
	case completed == 0:
		status = inspection.KawasanStatusOpen
	case completed < total:
		status = inspection.KawasanStatusOnProgress
	default:
		status = inspection.KawasanStatusConfirmed
	}

	return inspection.KawasanProgress{
		KawasanID:              kawasanID,
		Status:                 status,
		TotalDetailKawasan:     total,
		CompletedDetailKawasan: completed,
		ProgressPercent:        (float64(completed) / float64(total)) * 100,
	}, nil
}
