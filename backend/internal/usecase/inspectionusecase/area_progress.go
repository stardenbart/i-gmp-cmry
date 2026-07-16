package inspectionusecase

import (
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/domain/master"
)

// CalculateAreaStatus calculates the overall status of an Area by checking
// how many of its DetailKawasans have at least one completed inspection.
func CalculateAreaStatus(
	areaID string,
	detailKawasanRepo master.DetailKawasanRepository,
	inspectionRepo inspection.InspectionHeaderRepository,
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

	// 2. Count how many have at least one completed inspection
	completed := 0
	for _, dk := range allDK {
		count, _ := inspectionRepo.CountCompletedByAreaAndDetailKawasan(areaID, dk.DetailKawasanID)
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
