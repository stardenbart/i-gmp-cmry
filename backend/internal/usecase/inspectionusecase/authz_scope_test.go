package inspectionusecase

import (
	"context"
	"testing"
	"time"

	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/domain/master"
)

// fakeHeaderRepo implements inspection.InspectionHeaderRepository with only
// FindByID doing real work — everything else is an unused stub, same
// pattern as the fakes elsewhere in this codebase's tests.
type fakeHeaderRepo struct {
	byID map[string]*inspection.InspectionHeader
}

func (r *fakeHeaderRepo) FindAll(int, int, string, string, string, string) ([]inspection.InspectionHeader, int64, error) {
	return nil, 0, nil
}
func (r *fakeHeaderRepo) FindByID(id string) (*inspection.InspectionHeader, error) {
	item, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return item, nil
}
func (r *fakeHeaderRepo) FindByIDWithCtx(context.Context, string) (*inspection.InspectionHeader, error) {
	return nil, nil
}
func (r *fakeHeaderRepo) FindActiveByKawasan(string) ([]inspection.InspectionHeader, error) {
	return nil, nil
}
func (r *fakeHeaderRepo) FindActiveByDetailKawasan(string) ([]inspection.InspectionHeader, error) {
	return nil, nil
}
func (r *fakeHeaderRepo) FindActiveByInspector(string) ([]inspection.InspectionHeader, error) {
	return nil, nil
}
func (r *fakeHeaderRepo) FindActiveByInspectorAndDetailKawasan(string, string) ([]inspection.InspectionHeader, error) {
	return nil, nil
}
func (r *fakeHeaderRepo) CountCompletedInPeriod(string, time.Time, time.Time) (int64, error) {
	return 0, nil
}
func (r *fakeHeaderRepo) GetTrendByContext(string, int) ([]inspection.TrendData, error) {
	return nil, nil
}
func (r *fakeHeaderRepo) GetFullChecklist(string, string) (*inspection.FullChecklist, error) {
	return nil, nil
}
func (r *fakeHeaderRepo) GetFullChecklistWithCtx(context.Context, string, string) (*inspection.FullChecklist, error) {
	return nil, nil
}
func (r *fakeHeaderRepo) Create(*inspection.InspectionHeader) error { return nil }
func (r *fakeHeaderRepo) Update(*inspection.InspectionHeader) error { return nil }
func (r *fakeHeaderRepo) Delete(string) error                       { return nil }

// fakeAreaRepo implements master.AreaRepository with only FindByID doing
// real work.
type fakeAreaRepo struct {
	byID map[string]*master.Area
}

func (r *fakeAreaRepo) FindAll(int, int, string, string) ([]master.Area, int64, error) {
	return nil, 0, nil
}
func (r *fakeAreaRepo) FindByID(id string) (*master.Area, error) {
	item, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return item, nil
}
func (r *fakeAreaRepo) Create(*master.Area) error { return nil }
func (r *fakeAreaRepo) Update(*master.Area) error { return nil }
func (r *fakeAreaRepo) Delete(string) error       { return nil }

func plantID(s string) *string { return &s }

func TestGetByIDScopedDeniesCrossPlantCaller(t *testing.T) {
	headerRepo := &fakeHeaderRepo{byID: map[string]*inspection.InspectionHeader{
		"INSP-1": {InspectionID: "INSP-1", AreaID: "AREA-1"},
	}}
	areaRepo := &fakeAreaRepo{byID: map[string]*master.Area{
		"AREA-1": {AreaID: "AREA-1", PlantID: plantID("PLT-A")},
	}}
	uc := NewInspectionHeaderUseCase(headerRepo, nil, nil, nil, areaRepo, nil)

	if _, err := uc.GetByIDScoped("INSP-1", "PLT-B"); err == nil {
		t.Fatal("a caller scoped to a different plant must not see this inspection")
	}
}

func TestGetByIDScopedAllowsSamePlantCaller(t *testing.T) {
	headerRepo := &fakeHeaderRepo{byID: map[string]*inspection.InspectionHeader{
		"INSP-1": {InspectionID: "INSP-1", AreaID: "AREA-1"},
	}}
	areaRepo := &fakeAreaRepo{byID: map[string]*master.Area{
		"AREA-1": {AreaID: "AREA-1", PlantID: plantID("PLT-A")},
	}}
	uc := NewInspectionHeaderUseCase(headerRepo, nil, nil, nil, areaRepo, nil)

	item, err := uc.GetByIDScoped("INSP-1", "PLT-A")
	if err != nil {
		t.Fatalf("GetByIDScoped() error = %v", err)
	}
	if item.InspectionID != "INSP-1" {
		t.Fatal("expected the matching inspection to be returned")
	}
}

func TestGetByIDScopedAllowsUnscopedCaller(t *testing.T) {
	headerRepo := &fakeHeaderRepo{byID: map[string]*inspection.InspectionHeader{
		"INSP-1": {InspectionID: "INSP-1", AreaID: "AREA-1"},
	}}
	areaRepo := &fakeAreaRepo{byID: map[string]*master.Area{
		"AREA-1": {AreaID: "AREA-1", PlantID: plantID("PLT-A")},
	}}
	uc := NewInspectionHeaderUseCase(headerRepo, nil, nil, nil, areaRepo, nil)

	// userPlantID == "" is a Super Admin or unassigned global-scope user —
	// must see every plant's inspections, same as GetAll.
	if _, err := uc.GetByIDScoped("INSP-1", ""); err != nil {
		t.Fatalf("an unscoped (Super Admin / global) caller must be able to see any inspection, got error: %v", err)
	}
}

func TestGetByIDScopedAllowsAreaWithNoPlantOfItsOwn(t *testing.T) {
	headerRepo := &fakeHeaderRepo{byID: map[string]*inspection.InspectionHeader{
		"INSP-1": {InspectionID: "INSP-1", AreaID: "AREA-GLOBAL"},
	}}
	areaRepo := &fakeAreaRepo{byID: map[string]*master.Area{
		"AREA-GLOBAL": {AreaID: "AREA-GLOBAL", PlantID: nil},
	}}
	uc := NewInspectionHeaderUseCase(headerRepo, nil, nil, nil, areaRepo, nil)

	if _, err := uc.GetByIDScoped("INSP-1", "PLT-B"); err != nil {
		t.Fatalf("an area with no PlantID of its own is shared/global data, must be visible to any plant, got error: %v", err)
	}
}

func TestGetByIDScopedNotFoundStaysNotFound(t *testing.T) {
	uc := NewInspectionHeaderUseCase(&fakeHeaderRepo{byID: map[string]*inspection.InspectionHeader{}}, nil, nil, nil, &fakeAreaRepo{}, nil)

	if _, err := uc.GetByIDScoped("INSP-MISSING", "PLT-A"); err == nil {
		t.Fatal("a genuinely missing inspection must still return an error")
	}
}
