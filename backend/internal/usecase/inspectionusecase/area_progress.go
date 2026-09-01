package inspectionusecase

import (
	"time"

	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/domain/master"
)

// CalculateAreaStatus calculates the overall status of an Area by checking
// how many of its DetailKawasans have been completed within the given
// inspection period [periodStart, periodEnd) — the same admin-configurable
// cutoff-day cycle used by CalculateKawasanStatus (see
// ResolveInspectionPeriod). Previously this checked "completed ever, any
// time" via CountCompletedByAreaAndDetailKawasan, so an Area that was once
// fully inspected stayed permanently "Confirmed" and never re-triggered
// its completion notification again — now it resets every period, exactly
// like Kawasan-level status.
func CalculateAreaStatus(
	areaID string,
	detailKawasanRepo master.DetailKawasanRepository,
	inspectionRepo inspection.InspectionHeaderRepository,
	periodStart, periodEnd time.Time,
) (inspection.AreaProgress, error) {

	// 1. Get all DetailKawasan in this area
	allDK, err := detailKawasanRepo.FindAllByAreaID(areaID)
	if err != nil {
		return inspection.AreaProgress{}, err
	}

	total := len(allDK)
	if total == 0 {
		return inspection.AreaProgress{
			AreaID:                 areaID,
			Status:                 inspection.AreaStatusOpen,
			TotalDetailKawasan:     0,
			CompletedDetailKawasan: 0,
			ProgressPercent:        0,
		}, nil
	}

	// 2. Count how many have at least one completed inspection this period
	completed := 0
	for _, dk := range allDK {
		count, _ := inspectionRepo.CountCompletedInPeriod(dk.DetailKawasanID, periodStart, periodEnd)
		if count > 0 {
			completed++
		}
	}

	// 3. Determine area status
	var status inspection.AreaInspectionStatus
	switch {
	case completed == 0:
		status = inspection.AreaStatusOpen
	case completed < total:
		status = inspection.AreaStatusOnProgress
	default:
		status = inspection.AreaStatusConfirmed
	}

	var progressPercent float64 = 0
	if total > 0 {
		progressPercent = (float64(completed) / float64(total)) * 100
	}

	return inspection.AreaProgress{
		AreaID:                 areaID,
		Status:                 status,
		TotalDetailKawasan:     total,
		CompletedDetailKawasan: completed,
		ProgressPercent:        progressPercent,
	}, nil
}
