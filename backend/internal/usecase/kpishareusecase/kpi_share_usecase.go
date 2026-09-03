package kpishareusecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	dashboarddomain "github.com/monitoring-system/backend/internal/domain/dashboard"
	"github.com/monitoring-system/backend/internal/domain/kpishare"
	"github.com/monitoring-system/backend/internal/usecase/analyticsusecase"
)

const maxShareLifetime = 90 * 24 * time.Hour

type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string              { return e.Message }
func fail(status int, message string) error { return &Error{Status: status, Message: message} }

type ResolvedShare struct {
	Entity    *kpishare.PublicShare
	PlantName string
	Layout    []dashboarddomain.WidgetConfig
	Filter    kpishare.FilterSnapshot
}

type Service struct {
	repo       kpishare.Repository
	layoutRepo dashboarddomain.DashboardLayoutRepository
	analytics  *analyticsusecase.QueryService
	now        func() time.Time
}

func New(repo kpishare.Repository, layoutRepo dashboarddomain.DashboardLayoutRepository, analytics *analyticsusecase.QueryService) *Service {
	return &Service{repo: repo, layoutRepo: layoutRepo, analytics: analytics, now: time.Now}
}

func isAdmin(roleID string) bool {
	r := strings.ToUpper(strings.TrimSpace(roleID))
	return r == "ROLE-000" || r == "SUPERADMIN" || r == "ROLE-001"
}

func isSuperAdmin(roleID string) bool {
	r := strings.ToUpper(strings.TrimSpace(roleID))
	return r == "ROLE-000" || r == "SUPERADMIN"
}

func generateToken() (raw, hash, prefix string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", err
	}
	raw = "kpi_" + base64.RawURLEncoding.EncodeToString(b)
	digest := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(digest[:])
	prefix = raw[:12]
	return
}

func hashToken(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}

func validateFilter(filter kpishare.FilterSnapshot, now time.Time) (kpishare.FilterSnapshot, error) {
	filter.Period = strings.ToLower(strings.TrimSpace(filter.Period))
	filter.Granularity = strings.ToLower(strings.TrimSpace(filter.Granularity))
	if filter.Period == "" {
		filter.Period = "range"
	}
	if filter.Period == "quarter" {
		return kpishare.FilterSnapshot{Period: "quarter"}, nil
	}
	if filter.Period != "range" {
		return filter, fail(http.StatusBadRequest, "Periode dashboard tidak valid")
	}
	if filter.Granularity == "" {
		filter.Granularity = "day"
	}
	switch filter.Granularity {
	case "day", "week", "month", "year":
	default:
		return filter, fail(http.StatusBadRequest, "Granularitas dashboard tidak valid")
	}
	if filter.StartDate == "" || filter.EndDate == "" {
		end := now
		start := end.AddDate(0, 0, -29)
		filter.StartDate, filter.EndDate = start.Format("2006-01-02"), end.Format("2006-01-02")
	}
	start, err := time.Parse("2006-01-02", filter.StartDate)
	if err != nil {
		return filter, fail(http.StatusBadRequest, "Tanggal mulai tidak valid")
	}
	end, err := time.Parse("2006-01-02", filter.EndDate)
	if err != nil || end.Before(start) {
		return filter, fail(http.StatusBadRequest, "Tanggal selesai tidak valid")
	}
	return filter, nil
}

func (s *Service) Create(actor kpishare.Actor, req kpishare.CreateRequest) (*kpishare.CreateResponse, error) {
	if !isAdmin(actor.RoleID) {
		return nil, fail(http.StatusForbidden, "Hanya Admin atau Super Admin yang dapat membagikan dashboard KPI")
	}
	req.ShareName, req.PublicTitle, req.PlantID = strings.TrimSpace(req.ShareName), strings.TrimSpace(req.PublicTitle), strings.TrimSpace(req.PlantID)
	if req.ShareName == "" || utf8.RuneCountInString(req.ShareName) > 100 {
		return nil, fail(http.StatusBadRequest, "Nama link wajib diisi dan maksimal 100 karakter")
	}
	if req.PublicTitle == "" || utf8.RuneCountInString(req.PublicTitle) > 150 {
		return nil, fail(http.StatusBadRequest, "Judul publik wajib diisi dan maksimal 150 karakter")
	}
	if strings.IndexFunc(req.ShareName, unicode.IsControl) >= 0 || strings.IndexFunc(req.PublicTitle, unicode.IsControl) >= 0 {
		return nil, fail(http.StatusBadRequest, "Nama atau judul publik mengandung karakter yang tidak diizinkan")
	}
	if req.PlantID == "" || strings.EqualFold(req.PlantID, "all") {
		return nil, fail(http.StatusBadRequest, "Pilih satu plant untuk link publik")
	}
	if !isSuperAdmin(actor.RoleID) && (actor.PlantID == "" || actor.PlantID != req.PlantID) {
		return nil, fail(http.StatusForbidden, "Anda tidak memiliki akses untuk membagikan plant ini")
	}

	plantName, err := s.repo.PlantName(req.PlantID)
	if err != nil {
		return nil, fail(http.StatusBadRequest, "Plant tidak ditemukan")
	}
	now := s.now()
	if req.ExpiresAt == nil || !req.ExpiresAt.After(now) || req.ExpiresAt.Sub(now) > maxShareLifetime {
		return nil, fail(http.StatusBadRequest, "Masa berlaku link harus di masa depan dan maksimal 90 hari")
	}
	filter, err := validateFilter(req.Filter, now)
	if err != nil {
		return nil, err
	}

	layout := dashboarddomain.DefaultLayoutKPI()
	saved, err := s.layoutRepo.FindByUserID(actor.UserID, dashboarddomain.DashboardKeyKPI)
	if err != nil {
		return nil, fail(http.StatusInternalServerError, "Gagal membaca tata letak dashboard")
	}
	if saved != nil {
		if err := json.Unmarshal([]byte(saved.LayoutJSON), &layout); err != nil {
			return nil, fail(http.StatusInternalServerError, "Tata letak dashboard tersimpan tidak valid")
		}
	}
	visible := make([]dashboarddomain.WidgetConfig, 0, len(layout))
	for _, widget := range layout {
		// The KPI page no longer duplicates fixed widgets from the main
		// dashboard. Public snapshots therefore contain custom visualizations
		// only, including when the owner's saved layout predates this change.
		if !widget.Visible || widget.CustomQuery == nil {
			continue
		}
		if err := dashboarddomain.ValidateCustomQuery(widget.CustomQuery, widget.VizType); err != nil {
			return nil, fail(http.StatusBadRequest, "Konfigurasi widget "+widget.WidgetID+" tidak valid")
		}
		visible = append(visible, widget)
	}
	if len(visible) == 0 {
		return nil, fail(http.StatusBadRequest, "Dashboard tidak memiliki widget yang dapat dibagikan")
	}
	layoutJSON, _ := json.Marshal(visible)
	filterJSON, _ := json.Marshal(filter)
	raw, tokenHash, prefix, err := generateToken()
	if err != nil {
		return nil, fail(http.StatusInternalServerError, "Gagal membuat token publik")
	}
	entity := &kpishare.PublicShare{
		ShareID: "KSHARE-" + strings.ToUpper(uuid.NewString()[:12]), OwnerUserID: actor.UserID,
		ShareName: req.ShareName, PublicTitle: req.PublicTitle, TokenHash: tokenHash, TokenPrefix: prefix,
		PlantID: req.PlantID, LayoutJSON: string(layoutJSON), FilterJSON: string(filterJSON),
		AllowPeriodChange: req.AllowPeriodChange, SnapshotVersion: 1, ExpiresAt: req.ExpiresAt, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.repo.Create(entity); err != nil {
		return nil, fail(http.StatusInternalServerError, "Gagal menyimpan link publik")
	}
	return &kpishare.CreateResponse{ShareResponse: mapShare(entity, plantName), RawToken: raw}, nil
}

func mapShare(item *kpishare.PublicShare, plantName string) kpishare.ShareResponse {
	return kpishare.ShareResponse{
		ShareID: item.ShareID, ShareName: item.ShareName, PublicTitle: item.PublicTitle,
		TokenPrefix: item.TokenPrefix, PlantID: item.PlantID, PlantName: plantName,
		AllowPeriodChange: item.AllowPeriodChange, ExpiresAt: item.ExpiresAt, RevokedAt: item.RevokedAt,
		LastAccessedAt: item.LastAccessedAt, AccessCount: item.AccessCount, CreatedAt: item.CreatedAt,
	}
}

func (s *Service) List(actor kpishare.Actor) ([]kpishare.ShareResponse, error) {
	if !isAdmin(actor.RoleID) {
		return nil, fail(http.StatusForbidden, "Hanya Admin atau Super Admin yang dapat melihat link publik")
	}
	items, err := s.repo.ListByOwner(actor.UserID)
	if err != nil {
		return nil, fail(http.StatusInternalServerError, "Gagal membaca link publik")
	}
	result := make([]kpishare.ShareResponse, 0, len(items))
	for i := range items {
		plantName, _ := s.repo.PlantName(items[i].PlantID)
		result = append(result, mapShare(&items[i], plantName))
	}
	return result, nil
}

func (s *Service) Revoke(actor kpishare.Actor, shareID string) error {
	if !isAdmin(actor.RoleID) {
		return fail(http.StatusForbidden, "Hanya Admin atau Super Admin yang dapat mencabut link publik")
	}
	ok, err := s.repo.Revoke(strings.TrimSpace(shareID), actor.UserID, s.now())
	if err != nil {
		return fail(http.StatusInternalServerError, "Gagal mencabut link publik")
	}
	if !ok {
		return fail(http.StatusNotFound, "Link publik tidak ditemukan")
	}
	return nil
}

func (s *Service) Rotate(actor kpishare.Actor, shareID string) (*kpishare.CreateResponse, error) {
	if !isAdmin(actor.RoleID) {
		return nil, fail(http.StatusForbidden, "Hanya Admin atau Super Admin yang dapat merotasi link publik")
	}
	shareID = strings.TrimSpace(shareID)
	existing, err := s.repo.FindByID(shareID)
	if err != nil {
		return nil, fail(http.StatusInternalServerError, "Gagal membaca link publik")
	}
	if existing == nil || existing.OwnerUserID != actor.UserID {
		return nil, fail(http.StatusNotFound, "Link publik tidak ditemukan")
	}
	if existing.ExpiresAt != nil && !existing.ExpiresAt.After(s.now()) {
		return nil, fail(http.StatusBadRequest, "Link sudah kedaluwarsa. Buat link baru dengan masa berlaku baru.")
	}
	raw, tokenHash, prefix, err := generateToken()
	if err != nil {
		return nil, fail(http.StatusInternalServerError, "Gagal membuat token publik")
	}
	item, err := s.repo.Rotate(shareID, actor.UserID, tokenHash, prefix, s.now())
	if err != nil {
		return nil, fail(http.StatusInternalServerError, "Gagal merotasi link publik")
	}
	if item == nil {
		return nil, fail(http.StatusNotFound, "Link publik tidak ditemukan")
	}
	plantName, _ := s.repo.PlantName(item.PlantID)
	return &kpishare.CreateResponse{ShareResponse: mapShare(item, plantName), RawToken: raw}, nil
}

func (s *Service) Resolve(rawToken string, touch bool) (*ResolvedShare, error) {
	if !strings.HasPrefix(rawToken, "kpi_") || len(rawToken) < 40 {
		return nil, fail(http.StatusNotFound, "Dashboard publik tidak ditemukan")
	}
	item, err := s.repo.FindByTokenHash(hashToken(rawToken))
	if err != nil || item == nil || item.RevokedAt != nil || (item.ExpiresAt != nil && !item.ExpiresAt.After(s.now())) {
		return nil, fail(http.StatusNotFound, "Dashboard publik tidak ditemukan")
	}
	var layout []dashboarddomain.WidgetConfig
	var filter kpishare.FilterSnapshot
	if json.Unmarshal([]byte(item.LayoutJSON), &layout) != nil || json.Unmarshal([]byte(item.FilterJSON), &filter) != nil {
		return nil, fail(http.StatusNotFound, "Dashboard publik tidak ditemukan")
	}
	customLayout := make([]dashboarddomain.WidgetConfig, 0, len(layout))
	for _, widget := range layout {
		if widget.Visible && widget.CustomQuery != nil {
			customLayout = append(customLayout, widget)
		}
	}
	layout = customLayout
	plantName, err := s.repo.PlantName(item.PlantID)
	if err != nil {
		return nil, fail(http.StatusNotFound, "Dashboard publik tidak ditemukan")
	}
	if touch {
		_ = s.repo.TouchAccess(item.ShareID, s.now())
	}
	return &ResolvedShare{Entity: item, PlantName: plantName, Layout: layout, Filter: filter}, nil
}

func hierarchyAtDrillLevel(cq *dashboarddomain.CustomQueryConfig, drillLevel int) ([]string, error) {
	if drillLevel < 0 || drillLevel > len(cq.DrillDimensions) {
		return nil, fail(http.StatusBadRequest, "Level drill-down tidak valid")
	}
	hierarchy := make([]string, 0, drillLevel+1)
	hierarchy = append(hierarchy, cq.Dimension)
	if drillLevel > 0 {
		hierarchy = append(hierarchy, cq.DrillDimensions[:drillLevel]...)
	}
	return hierarchy, nil
}

func (s *Service) RunCustomQuery(ctx context.Context, rawToken, widgetID string, drillLevel int) (*analyticsusecase.QueryResultDTO, error) {
	resolved, err := s.Resolve(rawToken, false)
	if err != nil {
		return nil, err
	}
	var widget *dashboarddomain.WidgetConfig
	for i := range resolved.Layout {
		if resolved.Layout[i].Visible && resolved.Layout[i].WidgetID == widgetID && resolved.Layout[i].CustomQuery != nil {
			widget = &resolved.Layout[i]
			break
		}
	}
	if widget == nil {
		return nil, fail(http.StatusNotFound, "Widget publik tidak ditemukan")
	}
	cq := widget.CustomQuery
	hierarchyDimensions, err := hierarchyAtDrillLevel(cq, drillLevel)
	if err != nil {
		return nil, err
	}
	currentDimension := hierarchyDimensions[len(hierarchyDimensions)-1]
	req := analyticsusecase.QueryRequest{
		Measures: cq.Measures, Dimension: currentDimension, Dimension2: cq.Dimension2,
		HierarchyDimensions: hierarchyDimensions,
	}
	return s.analytics.RunQuery(ctx, analyticsusecase.Scope{UserPlantID: resolved.Entity.PlantID, RoleID: "ROLE-001"}, req)
}

func (s *Service) Bootstrap(rawToken string) (*kpishare.BootstrapResponse, error) {
	resolved, err := s.Resolve(rawToken, true)
	if err != nil {
		return nil, err
	}
	return &kpishare.BootstrapResponse{
		ShareID: resolved.Entity.ShareID, PublicTitle: resolved.Entity.PublicTitle,
		PlantID: resolved.Entity.PlantID, PlantName: resolved.PlantName, Layout: resolved.Layout,
		Filter: resolved.Filter, AllowPeriodChange: resolved.Entity.AllowPeriodChange,
		ExpiresAt: resolved.Entity.ExpiresAt, UpdatedAt: resolved.Entity.UpdatedAt, Catalog: s.analytics.Catalog(),
	}, nil
}

func ErrorStatus(err error) (int, string) {
	if typed, ok := err.(*Error); ok {
		return typed.Status, typed.Message
	}
	return http.StatusInternalServerError, fmt.Sprintf("%s", "Terjadi kesalahan saat memproses permintaan")
}
