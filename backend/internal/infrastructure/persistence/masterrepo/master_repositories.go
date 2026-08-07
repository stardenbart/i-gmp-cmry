package masterrepo

import (
	"time"

	"github.com/monitoring-system/backend/internal/domain/master"
	"gorm.io/gorm"
)

type departmentRepository struct{ db *gorm.DB }

func NewDepartmentRepository(db *gorm.DB) master.DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) FindAll(page, limit int, search string) ([]master.Department, int64, error) {
	var items []master.Department
	var total int64
	q := r.db.Model(&master.Department{})
	if search != "" {
		q = q.Where("\"DepartmentName\" LIKE ?", "%"+search+"%")
	}
	q.Count(&total)
	err := q.Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *departmentRepository) FindByID(id string) (*master.Department, error) {
	var item master.Department
	err := r.db.Where("\"DepartmentID\" = ?", id).First(&item).Error
	return &item, err
}

func (r *departmentRepository) Create(d *master.Department) error { return r.db.Create(d).Error }
func (r *departmentRepository) Update(d *master.Department) error { return r.db.Save(d).Error }
func (r *departmentRepository) Delete(id string) error {
	return r.db.Where("\"DepartmentID\" = ?", id).Delete(&master.Department{}).Error
}

// ── Area ──────────────────────────────────────────────────────────────────

type areaRepository struct{ db *gorm.DB }

func NewAreaRepository(db *gorm.DB) master.AreaRepository {
	return &areaRepository{db: db}
}

func (r *areaRepository) FindAll(page, limit int, plantID, search string) ([]master.Area, int64, error) {
	var items []master.Area
	var total int64
	q := r.db.Model(&master.Area{})
	if plantID != "" {
		q = q.Where("\"PlantID\" = ?", plantID)
	}
	if search != "" {
		q = q.Where("\"AreaName\" LIKE ?", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Plant").Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *areaRepository) FindByID(id string) (*master.Area, error) {
	var item master.Area
	err := r.db.Where("\"AreaID\" = ?", id).First(&item).Error
	return &item, err
}

func (r *areaRepository) Create(a *master.Area) error { return r.db.Create(a).Error }
func (r *areaRepository) Update(a *master.Area) error { return r.db.Save(a).Error }
func (r *areaRepository) Delete(id string) error {
	return r.db.Where("\"AreaID\" = ?", id).Delete(&master.Area{}).Error
}

// ── Kawasan ───────────────────────────────────────────────────────────────

type kawasanRepository struct{ db *gorm.DB }

func NewKawasanRepository(db *gorm.DB) master.KawasanRepository {
	return &kawasanRepository{db: db}
}

func (r *kawasanRepository) FindAll(page, limit int, plantID, areaID, search string) ([]master.Kawasan, int64, error) {
	var items []master.Kawasan
	var total int64
	q := r.db.Model(&master.Kawasan{})
	if plantID != "" {
		q = q.Joins(`JOIN "Area_Master" am ON am."AreaID" = "Kawasan_Master"."AreaID"`).Where(`am."PlantID" = ?`, plantID)
	}
	if areaID != "" {
		q = q.Where(`"Kawasan_Master"."AreaID" = ?`, areaID)
	}
	if search != "" {
		q = q.Where(`"Kawasan_Master"."KawasanName" ILIKE ?`, "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Area").Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *kawasanRepository) FindByID(id string) (*master.Kawasan, error) {
	var item master.Kawasan
	err := r.db.Where("\"KawasanID\" = ?", id).Preload("Area").First(&item).Error
	return &item, err
}

func (r *kawasanRepository) FindByAreaID(areaID string) ([]master.Kawasan, error) {
	var items []master.Kawasan
	err := r.db.Where("\"AreaID\" = ?", areaID).Preload("Area").Find(&items).Error
	return items, err
}

func (r *kawasanRepository) Create(k *master.Kawasan) error { return r.db.Create(k).Error }
func (r *kawasanRepository) Update(k *master.Kawasan) error { return r.db.Save(k).Error }
func (r *kawasanRepository) UpdateLastInspection(id string, lastInspection time.Time) error {
	return r.db.Model(&master.Kawasan{}).Where("\"KawasanID\" = ?", id).Update("LastInspection", lastInspection).Error
}
func (r *kawasanRepository) Delete(id string) error {
	return r.db.Where("\"KawasanID\" = ?", id).Delete(&master.Kawasan{}).Error
}

// ── DetailKawasan ─────────────────────────────────────────────────────────

type detailKawasanRepository struct{ db *gorm.DB }

func NewDetailKawasanRepository(db *gorm.DB) master.DetailKawasanRepository {
	return &detailKawasanRepository{db: db}
}

func (r *detailKawasanRepository) FindAll(page, limit int, plantID, kawasanID, search string) ([]master.DetailKawasan, int64, error) {
	var items []master.DetailKawasan
	var total int64
	q := r.db.Model(&master.DetailKawasan{}).
		Select(`"DetailKawasan_Master".*, 
			COALESCE(ih."InspectionHeaderStatus", '') AS "ActiveInspectionStatus",
			(SELECT MAX("InspectionHeaderCreatedAt") FROM "Inspection_Header" WHERE "Inspection_Header"."DetailKawasanID" = "DetailKawasan_Master"."DetailKawasanID" AND "InspectionHeaderStatus" IN ('Completed', 'Approved')) AS "LastInspection"`).
		Joins(`LEFT JOIN (
			SELECT DISTINCT ON ("DetailKawasanID") "DetailKawasanID", "InspectionHeaderStatus"
			FROM "Inspection_Header"
			WHERE "InspectionHeaderStatus" IN ('Ongoing', 'Draft')
			ORDER BY "DetailKawasanID", "InspectionHeaderCreatedAt" DESC
		) ih ON ih."DetailKawasanID" = "DetailKawasan_Master"."DetailKawasanID"`)
	if plantID != "" {
		q = q.Joins(`JOIN "Kawasan_Master" km ON km."KawasanID" = "DetailKawasan_Master"."KawasanID" JOIN "Area_Master" am ON am."AreaID" = km."AreaID"`).
			Where(`am."PlantID" = ?`, plantID)
	}
	if kawasanID != "" {
		q = q.Where(`"DetailKawasan_Master"."KawasanID" = ?`, kawasanID)
	}
	if search != "" {
		q = q.Where(`"DetailKawasan_Master"."DetailKawasanName" ILIKE ?`, "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Kawasan").Preload("Kawasan.Area").Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *detailKawasanRepository) FindByID(id string) (*master.DetailKawasan, error) {
	var item master.DetailKawasan
	err := r.db.Where("\"DetailKawasanID\" = ?", id).First(&item).Error
	return &item, err
}

func (r *detailKawasanRepository) FindByKawasanID(kawasanID string) ([]master.DetailKawasan, error) {
	var items []master.DetailKawasan
	err := r.db.Where("\"KawasanID\" = ?", kawasanID).Find(&items).Error
	return items, err
}

func (r *detailKawasanRepository) FindAllByAreaID(areaID string) ([]master.DetailKawasan, error) {
	var items []master.DetailKawasan
	err := r.db.Joins(`JOIN "Kawasan_Master" ON "Kawasan_Master"."KawasanID" = "DetailKawasan_Master"."KawasanID"`).
		Where("\"Kawasan_Master\".\"AreaID\" = ?", areaID).
		Find(&items).Error
	return items, err
}

func (r *detailKawasanRepository) Create(dk *master.DetailKawasan) error {
	return r.db.Create(dk).Error
}
func (r *detailKawasanRepository) Update(dk *master.DetailKawasan) error { return r.db.Save(dk).Error }
func (r *detailKawasanRepository) UpdateLastInspection(id string, lastInspection time.Time) error {
	return r.db.Model(&master.DetailKawasan{}).Where("\"DetailKawasanID\" = ?", id).Update("LastInspection", lastInspection).Error
}
func (r *detailKawasanRepository) Delete(id string) error {
	return r.db.Where("\"DetailKawasanID\" = ?", id).Delete(&master.DetailKawasan{}).Error
}

// ── Aspek ─────────────────────────────────────────────────────────────────

type aspekRepository struct{ db *gorm.DB }

func NewAspekRepository(db *gorm.DB) master.AspekRepository {
	return &aspekRepository{db: db}
}

func (r *aspekRepository) FindAll(page, limit int, plantID, areaID, search string) ([]master.Aspek, int64, error) {
	var items []master.Aspek
	var total int64
	q := r.db.Model(&master.Aspek{})
	if plantID != "" {
		q = q.Joins(`JOIN "Area_Master" am ON am."AreaID" = "Aspek_Master"."AreaID"`).Where(`am."PlantID" = ?`, plantID)
	}
	if areaID != "" {
		q = q.Where(`"Aspek_Master"."AreaID" = ?`, areaID)
	}
	if search != "" {
		q = q.Where(`"Aspek_Master"."AspekName" ILIKE ?`, "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Area").Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *aspekRepository) FindByID(id string) (*master.Aspek, error) {
	var item master.Aspek
	err := r.db.Where("\"AspekID\" = ?", id).Preload("Area").First(&item).Error
	return &item, err
}

func (r *aspekRepository) FindByAreaID(areaID string) ([]master.Aspek, error) {
	var items []master.Aspek
	err := r.db.Where("\"AreaID\" = ?", areaID).Preload("Area").Find(&items).Error
	return items, err
}

func (r *aspekRepository) Create(a *master.Aspek) error { return r.db.Create(a).Error }
func (r *aspekRepository) Update(a *master.Aspek) error { return r.db.Save(a).Error }
func (r *aspekRepository) Delete(id string) error {
	return r.db.Where("\"AspekID\" = ?", id).Delete(&master.Aspek{}).Error
}

// ── Detail ────────────────────────────────────────────────────────────────

type detailRepository struct{ db *gorm.DB }

func NewDetailRepository(db *gorm.DB) master.DetailRepository {
	return &detailRepository{db: db}
}

func (r *detailRepository) FindAll(page, limit int, plantID, aspekID, search string) ([]master.Detail, int64, error) {
	var items []master.Detail
	var total int64
	q := r.db.Model(&master.Detail{})
	if plantID != "" {
		q = q.Joins(`JOIN "Aspek_Master" asp ON asp."AspekID" = "Detail_Master"."AspekID" JOIN "Area_Master" am ON am."AreaID" = asp."AreaID"`).
			Where(`am."PlantID" = ?`, plantID)
	}
	if aspekID != "" {
		q = q.Where(`"Detail_Master"."AspekID" = ?`, aspekID)
	}
	if search != "" {
		q = q.Where(`"Detail_Master"."DetailName" ILIKE ?`, "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Aspek").Preload("Aspek.Area").Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *detailRepository) FindByID(id string) (*master.Detail, error) {
	var item master.Detail
	err := r.db.Where("\"DetailID\" = ?", id).Preload("Aspek").First(&item).Error
	return &item, err
}

func (r *detailRepository) FindByAspekID(aspekID string) ([]master.Detail, error) {
	var items []master.Detail
	err := r.db.Where("\"AspekID\" = ?", aspekID).Preload("Aspek").Find(&items).Error
	return items, err
}

func (r *detailRepository) Create(d *master.Detail) error { return r.db.Create(d).Error }
func (r *detailRepository) Update(d *master.Detail) error { return r.db.Save(d).Error }
func (r *detailRepository) Delete(id string) error {
	return r.db.Where("\"DetailID\" = ?", id).Delete(&master.Detail{}).Error
}

// ── Uraian ────────────────────────────────────────────────────────────────

type uraianRepository struct{ db *gorm.DB }

func NewUraianRepository(db *gorm.DB) master.UraianRepository {
	return &uraianRepository{db: db}
}

func (r *uraianRepository) FindAll(page, limit int, plantID, detailID, search string) ([]master.Uraian, int64, error) {
	var items []master.Uraian
	var total int64
	q := r.db.Model(&master.Uraian{})
	if plantID != "" {
		q = q.Joins(`JOIN "Detail_Master" dt ON dt."DetailID" = "Uraian_Master"."DetailID" JOIN "Aspek_Master" asp ON asp."AspekID" = dt."AspekID" JOIN "Area_Master" am ON am."AreaID" = asp."AreaID"`).
			Where(`am."PlantID" = ?`, plantID)
	}
	if detailID != "" {
		q = q.Where(`"Uraian_Master"."DetailID" = ?`, detailID)
	}
	if search != "" {
		q = q.Where(`"Uraian_Master"."UraianText" ILIKE ?`, "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Detail").Preload("Detail.Aspek").Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *uraianRepository) FindByID(id string) (*master.Uraian, error) {
	var item master.Uraian
	err := r.db.Where("\"UraianID\" = ?", id).First(&item).Error
	return &item, err
}

func (r *uraianRepository) FindByDetailID(detailID string) ([]master.Uraian, error) {
	var items []master.Uraian
	err := r.db.Where("\"DetailID\" = ?", detailID).Find(&items).Error
	return items, err
}

func (r *uraianRepository) Create(u *master.Uraian) error { return r.db.Create(u).Error }
func (r *uraianRepository) Update(u *master.Uraian) error { return r.db.Save(u).Error }
func (r *uraianRepository) Delete(id string) error {
	return r.db.Where("\"UraianID\" = ?", id).Delete(&master.Uraian{}).Error
}
