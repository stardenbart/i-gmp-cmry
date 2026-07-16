package inspectionrepo

import (
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"gorm.io/gorm"
)

// ── Inspection Header ─────────────────────────────────────────────────────

type inspectionHeaderRepository struct{ db *gorm.DB }

func NewInspectionHeaderRepository(db *gorm.DB) inspection.InspectionHeaderRepository {
	return &inspectionHeaderRepository{db: db}
}

func (r *inspectionHeaderRepository) FindAll(page, limit int, areaID, status, inspectorID string) ([]inspection.InspectionHeader, int64, error) {
	var items []inspection.InspectionHeader
	var total int64
	q := r.db.Model(&inspection.InspectionHeader{})
	if areaID != "" { q = q.Where("AreaID = ?", areaID) }
	if status != "" { q = q.Where("InspectionHeaderStatus = ?", status) }
	if inspectorID != "" { q = q.Where("InspectorID = ?", inspectorID) }
	q.Count(&total)
	err := q.Order("InspectionHeaderCreatedAt DESC").Offset((page-1)*limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *inspectionHeaderRepository) FindByID(id string) (*inspection.InspectionHeader, error) {
	var item inspection.InspectionHeader
	err := r.db.Preload("Results").Where("InspectionID = ?", id).First(&item).Error
	return &item, err
}

func (r *inspectionHeaderRepository) FindActiveByKawasan(kawasanID string) ([]inspection.InspectionHeader, error) {
	var items []inspection.InspectionHeader
	err := r.db.Where("KawasanID = ? AND InspectionHeaderStatus IN (?, ?)", 
		kawasanID, inspection.InspectionStatusDraft, inspection.InspectionStatusOngoing).
		Find(&items).Error
	return items, err
}

func (r *inspectionHeaderRepository) FindActiveByInspector(inspectorID string) ([]inspection.InspectionHeader, error) {
	var items []inspection.InspectionHeader
	err := r.db.Where("InspectorID = ? AND InspectionHeaderStatus IN (?, ?)", 
		inspectorID, inspection.InspectionStatusDraft, inspection.InspectionStatusOngoing).
		Find(&items).Error
	return items, err
}

func (r *inspectionHeaderRepository) CountCompletedThisMonthByKawasan(kawasanID string, year int, month int) (int64, error) {
	var count int64
	err := r.db.Model(&inspection.InspectionHeader{}).
		Where("KawasanID = ? AND InspectionHeaderStatus IN (?, ?)", kawasanID, inspection.InspectionStatusCompleted, inspection.InspectionStatusApproved).
		Where("EXTRACT(YEAR FROM InspectionHeaderCreatedAt) = ? AND EXTRACT(MONTH FROM InspectionHeaderCreatedAt) = ?", year, month).
		Count(&count).Error
	return count, err
}

func (r *inspectionHeaderRepository) CountCompletedByAreaAndDetailKawasan(areaID, detailKawasanID string) (int64, error) {
	var count int64
	err := r.db.Model(&inspection.InspectionHeader{}).
		Where("AreaID = ? AND DetailKawasanID = ? AND InspectionHeaderStatus = ?", 
		areaID, detailKawasanID, inspection.InspectionStatusCompleted).
		Count(&count).Error
	return count, err
}

func (r *inspectionHeaderRepository) Create(h *inspection.InspectionHeader) error { return r.db.Create(h).Error }
func (r *inspectionHeaderRepository) Update(h *inspection.InspectionHeader) error { return r.db.Save(h).Error }
func (r *inspectionHeaderRepository) Delete(id string) error {
	return r.db.Where("InspectionID = ?", id).Delete(&inspection.InspectionHeader{}).Error
}

// ── Inspection Result ─────────────────────────────────────────────────────

type inspectionResultRepository struct{ db *gorm.DB }

func NewInspectionResultRepository(db *gorm.DB) inspection.InspectionResultRepository {
	return &inspectionResultRepository{db: db}
}

func (r *inspectionResultRepository) FindByInspectionID(inspectionID string) ([]inspection.InspectionResult, error) {
	var items []inspection.InspectionResult
	err := r.db.Where("InspectionID = ?", inspectionID).Find(&items).Error
	return items, err
}

func (r *inspectionResultRepository) FindByID(id string) (*inspection.InspectionResult, error) {
	var item inspection.InspectionResult
	err := r.db.Where("ResultID = ?", id).First(&item).Error
	return &item, err
}

func (r *inspectionResultRepository) Create(res *inspection.InspectionResult) error { return r.db.Create(res).Error }
func (r *inspectionResultRepository) BulkCreate(rs []inspection.InspectionResult) error {
	return r.db.CreateInBatches(&rs, 50).Error
}
func (r *inspectionResultRepository) Update(res *inspection.InspectionResult) error { return r.db.Save(res).Error }
func (r *inspectionResultRepository) Delete(id string) error {
	return r.db.Where("ResultID = ?", id).Delete(&inspection.InspectionResult{}).Error
}
