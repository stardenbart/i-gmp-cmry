package kpishareusecase

import (
	"strings"
	"testing"
	"time"

	dashboarddomain "github.com/monitoring-system/backend/internal/domain/dashboard"
	"github.com/monitoring-system/backend/internal/domain/kpishare"
)

type fakeShareRepo struct {
	items map[string]*kpishare.PublicShare
}

func (f *fakeShareRepo) Create(item *kpishare.PublicShare) error {
	if f.items == nil {
		f.items = map[string]*kpishare.PublicShare{}
	}
	copy := *item
	f.items[item.ShareID] = &copy
	return nil
}
func (f *fakeShareRepo) ListByOwner(owner string) ([]kpishare.PublicShare, error) {
	var result []kpishare.PublicShare
	for _, item := range f.items {
		if item.OwnerUserID == owner {
			result = append(result, *item)
		}
	}
	return result, nil
}
func (f *fakeShareRepo) FindByID(id string) (*kpishare.PublicShare, error) { return f.items[id], nil }
func (f *fakeShareRepo) FindByTokenHash(hash string) (*kpishare.PublicShare, error) {
	for _, item := range f.items {
		if item.TokenHash == hash {
			copy := *item
			return &copy, nil
		}
	}
	return nil, nil
}
func (f *fakeShareRepo) Revoke(id, owner string, now time.Time) (bool, error) {
	item := f.items[id]
	if item == nil || item.OwnerUserID != owner || item.RevokedAt != nil {
		return false, nil
	}
	item.RevokedAt = &now
	return true, nil
}
func (f *fakeShareRepo) Rotate(id, owner, hash, prefix string, now time.Time) (*kpishare.PublicShare, error) {
	return nil, nil
}
func (f *fakeShareRepo) TouchAccess(id string, now time.Time) error { return nil }
func (f *fakeShareRepo) PlantName(id string) (string, error) {
	if id == "PLT-1" {
		return "Plant Satu", nil
	}
	return "", errNotFound{}
}

type errNotFound struct{}

func (errNotFound) Error() string { return "not found" }

const visibleCustomLayout = `[{"widget_id":"custom-1","visible":true,"order":0,"custom_query":{"version":1,"title":"Total Issue per Area","measures":["issue_count"],"dimension":"plant","drill_dimensions":["area","status_temuan"]}}]`

func TestHierarchyAtDrillLevelRetainsEveryParentDimension(t *testing.T) {
	cq := &dashboarddomain.CustomQueryConfig{
		Dimension: "plant", DrillDimensions: []string{"area", "status_temuan"},
	}
	tests := []struct {
		level int
		want  string
	}{
		{level: 0, want: "plant"},
		{level: 1, want: "plant,area"},
		{level: 2, want: "plant,area,status_temuan"},
	}
	for _, tt := range tests {
		got, err := hierarchyAtDrillLevel(cq, tt.level)
		if err != nil || strings.Join(got, ",") != tt.want {
			t.Fatalf("hierarchyAtDrillLevel(%d) = %q, %v; want %q", tt.level, got, err, tt.want)
		}
	}
	for _, invalid := range []int{-1, 3} {
		if _, err := hierarchyAtDrillLevel(cq, invalid); err == nil {
			t.Fatalf("hierarchyAtDrillLevel(%d) must reject an out-of-range level", invalid)
		}
	}
}

type fakeLayoutRepo struct {
	layout *dashboarddomain.UserDashboardLayout
}

func (f *fakeLayoutRepo) FindByUserID(_, _ string) (*dashboarddomain.UserDashboardLayout, error) {
	return f.layout, nil
}
func (f *fakeLayoutRepo) Upsert(_, _, _ string) error { return nil }

func validCreate(now time.Time) kpishare.CreateRequest {
	expires := now.Add(7 * 24 * time.Hour)
	return kpishare.CreateRequest{
		ShareName: "Rapat Mingguan", PublicTitle: "KPI Plant Satu", PlantID: "PLT-1",
		Filter:    kpishare.FilterSnapshot{Period: "range", StartDate: "2026-08-01", EndDate: "2026-08-31", Granularity: "day"},
		ExpiresAt: &expires,
	}
}

func TestCreateFreezesOnlyVisibleLayoutAndResolvesOpaqueToken(t *testing.T) {
	now := time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)
	repo := &fakeShareRepo{}
	layout := `[{"widget_id":"kpi-summary-cards","visible":true,"order":0},{"widget_id":"custom-1","visible":true,"order":1,"custom_query":{"version":1,"title":"Total Issue per Area","measures":["issue_count"],"dimension":"area"}},{"widget_id":"custom-hidden","visible":false,"order":2,"custom_query":{"version":1,"measures":["issue_count"],"dimension":"area"}}]`
	service := New(repo, &fakeLayoutRepo{layout: &dashboarddomain.UserDashboardLayout{LayoutJSON: layout}}, nil)
	service.now = func() time.Time { return now }

	created, err := service.Create(kpishare.Actor{UserID: "USR-1", RoleID: "ROLE-001", PlantID: "PLT-1"}, validCreate(now))
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !strings.HasPrefix(created.RawToken, "kpi_") || len(created.RawToken) < 40 {
		t.Fatalf("raw token is not a strong public token: %q", created.RawToken)
	}
	if strings.Contains(repo.items[created.ShareID].TokenHash, created.RawToken) {
		t.Fatal("raw token must not be stored")
	}
	resolved, err := service.Resolve(created.RawToken, false)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(resolved.Layout) != 1 || resolved.Layout[0].WidgetID != "custom-1" {
		t.Fatalf("unexpected frozen layout: %#v", resolved.Layout)
	}
}

func TestCreateRejectsUnauthorizedScopeAndAllPlant(t *testing.T) {
	now := time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)
	service := New(&fakeShareRepo{}, &fakeLayoutRepo{layout: &dashboarddomain.UserDashboardLayout{LayoutJSON: visibleCustomLayout}}, nil)
	service.now = func() time.Time { return now }
	tests := []struct {
		name   string
		actor  kpishare.Actor
		mutate func(*kpishare.CreateRequest)
		status int
	}{
		{name: "auditee", actor: kpishare.Actor{UserID: "USR-3", RoleID: "ROLE-003", PlantID: "PLT-1"}, mutate: func(*kpishare.CreateRequest) {}, status: 403},
		{name: "admin other plant", actor: kpishare.Actor{UserID: "USR-1", RoleID: "ROLE-001", PlantID: "PLT-2"}, mutate: func(*kpishare.CreateRequest) {}, status: 403},
		{name: "all plant", actor: kpishare.Actor{UserID: "USR-0", RoleID: "ROLE-000"}, mutate: func(req *kpishare.CreateRequest) { req.PlantID = "all" }, status: 400},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := validCreate(now)
			tc.mutate(&req)
			_, err := service.Create(tc.actor, req)
			typed, ok := err.(*Error)
			if !ok || typed.Status != tc.status {
				t.Fatalf("error = %#v, want status %d", err, tc.status)
			}
		})
	}
}

func TestResolveUsesSameNotFoundForRevokedAndExpired(t *testing.T) {
	now := time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)
	service := New(&fakeShareRepo{}, &fakeLayoutRepo{layout: &dashboarddomain.UserDashboardLayout{LayoutJSON: visibleCustomLayout}}, nil)
	service.now = func() time.Time { return now }
	created, err := service.Create(kpishare.Actor{UserID: "USR-1", RoleID: "ROLE-001", PlantID: "PLT-1"}, validCreate(now))
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Revoke(kpishare.Actor{UserID: "USR-1", RoleID: "ROLE-001"}, created.ShareID); err != nil {
		t.Fatal(err)
	}
	_, err = service.Resolve(created.RawToken, false)
	typed, ok := err.(*Error)
	if !ok || typed.Status != 404 || typed.Message != "Dashboard publik tidak ditemukan" {
		t.Fatalf("unexpected public error: %#v", err)
	}
}
