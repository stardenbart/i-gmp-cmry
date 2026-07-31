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
	q := r.db.Model(&master.Area{}).Preload("Plant")
	if plantID != "" {
		q = q.Where("\"PlantID\" = ?", plantID)
	}
	if search != "" {
		q = q.Where("\"AreaName\" LIKE ?", "%"+search+"%")
	}
	q.Count(&total)
	err := q.Offset((page - 1) * limit).Limit(limit).Find(&items).Error
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

func (r *kawasanRepository) FindAll(page, limit int, areaID, search string) ([]master.Kawasan, int64, error) {
	var items []master.Kawasan
	var total int64
	q := r.db.Model(&master.Kawasan{}).Preload("Area")
	if areaID != "" {
		q = q.Where("\"AreaID\" = ?", areaID)
	}
	if search != "" {
		q = q.Where("\"KawasanName\" LIKE ?", "%"+search+"%")
	}
	q.Count(&total)
	err := q.Offset((page - 1) * limit).Limit(limit).Find(&items).Error
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

func (r *detailKawasanRepository) FindAll(page, limit int, kawasanID, search string) ([]master.DetailKawasan, int64, error) {
	var items []master.DetailKawasan
	var total int64
	q := r.db.Model(&master.DetailKawasan{}).
		Preload("Kawasan").Preload("Kawasan.Area").
		Select(`"DetailKawasan_Master".*, COALESCE(ih."InspectionHeaderStatus", '') AS "ActiveInspectionStatus"`).
		Joins(`LEFT JOIN (
			SELECT DISTINCT ON ("DetailKawasanID") "DetailKawasanID", "InspectionHeaderStatus"
			FROM "Inspection_Header"
			WHERE "InspectionHeaderStatus" IN ('Ongoing', 'Draft')
			ORDER BY "DetailKawasanID", "InspectionHeaderCreatedAt" DESC
		) ih ON ih."DetailKawasanID" = "DetailKawasan_Master"."DetailKawasanID"`)
	if kawasanID != "" {
		q = q.Where(`"DetailKawasan_Master"."KawasanID" = ?`, kawasanID)
	}
	if search != "" {
		q = q.Where(`"DetailKawasan_Master"."DetailKawasanName" ILIKE ?`, "%"+search+"%")
	}
	q.Count(&total)
	err := q.Offset((page - 1) * limit).Limit(limit).Find(&items).Error
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

func (r *aspekRepository) FindAll(page, limit int, areaID, search string) ([]master.Aspek, int64, error) {
	var items []master.Aspek
	var total int64
	q := r.db.Model(&master.Aspek{}).Preload("Area")
	if areaID != "" {
		q = q.Where("\"AreaID\" = ?", areaID)
	}
	if search != "" {
		q = q.Where("\"AspekName\" LIKE ?", "%"+search+"%")
	}
	q.Count(&total)
	err := q.Offset((page - 1) * limit).Limit(limit).Find(&items).Error
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

func (r *detailRepository) FindAll(page, limit int, aspekID, search string) ([]master.Detail, int64, error) {
	var items []master.Detail
	var total int64
	q := r.db.Model(&master.Detail{}).Preload("Aspek").Preload("Aspek.Area")
	if aspekID != "" {
		q = q.Where("\"AspekID\" = ?", aspekID)
	}
	if search != "" {
		q = q.Where("\"DetailName\" LIKE ?", "%"+search+"%")
	}
	q.Count(&total)
	err := q.Offset((page - 1) * limit).Limit(limit).Find(&items).Error
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

func (r *uraianRepository) FindAll(page, limit int, detailID, search string) ([]master.Uraian, int64, error) {
	var items []master.Uraian
	var total int64
	q := r.db.Model(&master.Uraian{}).Preload("Detail").Preload("Detail.Aspek")
	if detailID != "" {
		q = q.Where("\"DetailID\" = ?", detailID)
	}
	if search != "" {
		q = q.Where("\"UraianText\" LIKE ?", "%"+search+"%")
	}
	q.Count(&total)
	err := q.Offset((page - 1) * limit).Limit(limit).Find(&items).Error
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
