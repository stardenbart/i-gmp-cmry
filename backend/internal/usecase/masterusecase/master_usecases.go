package masterusecase

import (
	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/idgen"
)

// ── Department ────────────────────────────────────────────────────────────

type departmentUseCase struct{ repo master.DepartmentRepository }

func NewDepartmentUseCase(repo master.DepartmentRepository) master.DepartmentUseCase {
	return &departmentUseCase{repo: repo}
}
func (uc *departmentUseCase) GetAll(page, limit int, search string) ([]master.Department, int64, error) {
	return uc.repo.FindAll(page, limit, search)
}
func (uc *departmentUseCase) GetByID(id string) (*master.Department, error) {
	return uc.repo.FindByID(id)
}
func (uc *departmentUseCase) Create(d *master.Department) error {
	d.DepartmentID = idgen.Generate(idgen.PrefixDepartment)
	return uc.repo.Create(d)
}
func (uc *departmentUseCase) Update(d *master.Department) error { return uc.repo.Update(d) }
func (uc *departmentUseCase) Delete(id string) error            { return uc.repo.Delete(id) }

// ── Area ──────────────────────────────────────────────────────────────────

type areaUseCase struct{ repo master.AreaRepository }

func NewAreaUseCase(repo master.AreaRepository) master.AreaUseCase {
	return &areaUseCase{repo: repo}
}
func (uc *areaUseCase) GetAll(page, limit int, search string) ([]master.Area, int64, error) {
	return uc.repo.FindAll(page, limit, search)
}
func (uc *areaUseCase) GetByID(id string) (*master.Area, error) { return uc.repo.FindByID(id) }
func (uc *areaUseCase) Create(a *master.Area) error {
	a.AreaID = idgen.Generate(idgen.PrefixArea)
	return uc.repo.Create(a)
}
func (uc *areaUseCase) Update(a *master.Area) error { return uc.repo.Update(a) }
func (uc *areaUseCase) Delete(id string) error      { return uc.repo.Delete(id) }

// ── Kawasan ───────────────────────────────────────────────────────────────

type kawasanUseCase struct{ repo master.KawasanRepository }

func NewKawasanUseCase(repo master.KawasanRepository) master.KawasanUseCase {
	return &kawasanUseCase{repo: repo}
}
func (uc *kawasanUseCase) GetAll(page, limit int, areaID, search string) ([]master.Kawasan, int64, error) {
	return uc.repo.FindAll(page, limit, areaID, search)
}
func (uc *kawasanUseCase) GetByID(id string) (*master.Kawasan, error) { return uc.repo.FindByID(id) }
func (uc *kawasanUseCase) GetByAreaID(areaID string) ([]master.Kawasan, error) {
	return uc.repo.FindByAreaID(areaID)
}
func (uc *kawasanUseCase) Create(k *master.Kawasan) error {
	k.KawasanID = idgen.Generate(idgen.PrefixKawasan)
	return uc.repo.Create(k)
}
func (uc *kawasanUseCase) Update(k *master.Kawasan) error { return uc.repo.Update(k) }
func (uc *kawasanUseCase) Delete(id string) error         { return uc.repo.Delete(id) }

// ── DetailKawasan ─────────────────────────────────────────────────────────

type detailKawasanUseCase struct {
	repo master.DetailKawasanRepository
}

func NewDetailKawasanUseCase(repo master.DetailKawasanRepository) master.DetailKawasanUseCase {
	return &detailKawasanUseCase{repo: repo}
}
func (uc *detailKawasanUseCase) GetAll(page, limit int, kawasanID, search string) ([]master.DetailKawasan, int64, error) {
	return uc.repo.FindAll(page, limit, kawasanID, search)
}
func (uc *detailKawasanUseCase) GetByID(id string) (*master.DetailKawasan, error) {
	return uc.repo.FindByID(id)
}
func (uc *detailKawasanUseCase) GetByKawasanID(kawasanID string) ([]master.DetailKawasan, error) {
	return uc.repo.FindByKawasanID(kawasanID)
}
func (uc *detailKawasanUseCase) Create(dk *master.DetailKawasan) error {
	dk.DetailKawasanID = idgen.Generate(idgen.PrefixDetailKawasan)
	return uc.repo.Create(dk)
}
func (uc *detailKawasanUseCase) Update(dk *master.DetailKawasan) error { return uc.repo.Update(dk) }
func (uc *detailKawasanUseCase) Delete(id string) error                { return uc.repo.Delete(id) }

// ── Aspek ─────────────────────────────────────────────────────────────────

type aspekUseCase struct{ repo master.AspekRepository }

func NewAspekUseCase(repo master.AspekRepository) master.AspekUseCase {
	return &aspekUseCase{repo: repo}
}
func (uc *aspekUseCase) GetAll(page, limit int, areaID, search string) ([]master.Aspek, int64, error) {
	return uc.repo.FindAll(page, limit, areaID, search)
}
func (uc *aspekUseCase) GetByID(id string) (*master.Aspek, error) { return uc.repo.FindByID(id) }
func (uc *aspekUseCase) GetByAreaID(areaID string) ([]master.Aspek, error) {
	return uc.repo.FindByAreaID(areaID)
}
func (uc *aspekUseCase) Create(a *master.Aspek) error {
	a.AspekID = idgen.Generate(idgen.PrefixAspek)
	return uc.repo.Create(a)
}
func (uc *aspekUseCase) Update(a *master.Aspek) error { return uc.repo.Update(a) }
func (uc *aspekUseCase) Delete(id string) error       { return uc.repo.Delete(id) }

// ── Detail ────────────────────────────────────────────────────────────────

type detailUseCase struct{ repo master.DetailRepository }

func NewDetailUseCase(repo master.DetailRepository) master.DetailUseCase {
	return &detailUseCase{repo: repo}
}
func (uc *detailUseCase) GetAll(page, limit int, aspekID, search string) ([]master.Detail, int64, error) {
	return uc.repo.FindAll(page, limit, aspekID, search)
}
func (uc *detailUseCase) GetByID(id string) (*master.Detail, error) { return uc.repo.FindByID(id) }
func (uc *detailUseCase) GetByAspekID(aspekID string) ([]master.Detail, error) {
	return uc.repo.FindByAspekID(aspekID)
}
func (uc *detailUseCase) Create(d *master.Detail) error {
	d.DetailID = idgen.Generate(idgen.PrefixDetail)
	return uc.repo.Create(d)
}
func (uc *detailUseCase) Update(d *master.Detail) error { return uc.repo.Update(d) }
func (uc *detailUseCase) Delete(id string) error        { return uc.repo.Delete(id) }

// ── Uraian ────────────────────────────────────────────────────────────────

type uraianUseCase struct{ repo master.UraianRepository }

func NewUraianUseCase(repo master.UraianRepository) master.UraianUseCase {
	return &uraianUseCase{repo: repo}
}
func (uc *uraianUseCase) GetAll(page, limit int, detailID, search string) ([]master.Uraian, int64, error) {
	return uc.repo.FindAll(page, limit, detailID, search)
}
func (uc *uraianUseCase) GetByID(id string) (*master.Uraian, error) { return uc.repo.FindByID(id) }
func (uc *uraianUseCase) GetByDetailID(detailID string) ([]master.Uraian, error) {
	return uc.repo.FindByDetailID(detailID)
}
func (uc *uraianUseCase) Create(u *master.Uraian) error {
	u.UraianID = idgen.Generate(idgen.PrefixUraian)
	return uc.repo.Create(u)
}
func (uc *uraianUseCase) Update(u *master.Uraian) error { return uc.repo.Update(u) }
func (uc *uraianUseCase) Delete(id string) error        { return uc.repo.Delete(id) }
