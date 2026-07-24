package inspectionrepo

import (
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/domain/master"
)

func (r *inspectionHeaderRepository) GetFullChecklist(areaID, inspectionID string) (*inspection.FullChecklist, error) {
	// 1. Fetch all Aspeks for AreaID, including Details and Uraians
	var aspeks []master.Aspek
	err := r.db.Preload("Details.Urains").Where("\"AreaID\" = ?", areaID).Find(&aspeks).Error
	if err != nil {
		return nil, err
	}

	// 2. Fetch all InspectionResults for the given inspectionID
	var results []inspection.InspectionResult
	if inspectionID != "" {
		r.db.Where("\"InspectionID\" = ?", inspectionID).Find(&results)
	}
	resultMap := make(map[string]*inspection.InspectionResult)
	for i := range results {
		resultMap[results[i].UraianID] = &results[i]
	}

	// 3. Map to DTO
	checklist := &inspection.FullChecklist{
		InspectionID: inspectionID,
		AreaID:       areaID,
		Aspeks:       make([]inspection.ChecklistAspek, 0),
	}

	for _, a := range aspeks {
		ca := inspection.ChecklistAspek{
			AspekID:   a.AspekID,
			AspekName: a.AspekName,
			Details:   make([]inspection.ChecklistDetail, 0),
		}
		for _, d := range a.Details {
			cd := inspection.ChecklistDetail{
				DetailID:   d.DetailID,
				DetailName: d.DetailName,
				Uraians:    make([]inspection.ChecklistUraian, 0),
			}
			for _, u := range d.Urains {
				cu := inspection.ChecklistUraian{
					UraianID:      u.UraianID,
					UraianText:    u.UraianText,
					StandardScore: u.StandardScore,
					Result:        resultMap[u.UraianID],
				}
				cd.Uraians = append(cd.Uraians, cu)
			}
			ca.Details = append(ca.Details, cd)
		}
		checklist.Aspeks = append(checklist.Aspeks, ca)
	}

	return checklist, nil
}
