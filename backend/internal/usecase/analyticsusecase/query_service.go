package analyticsusecase

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/monitoring-system/backend/internal/domain/analytics"
	"github.com/monitoring-system/backend/pkg/logger"
	"gorm.io/gorm"
)

// Guardrail constants (see the approved plan §6). Kept here, not scattered
// across handler/repo, so a future adjustment touches exactly one place.
const (
	MaxMeasuresPerQuery = 4
	MaxRowsReturned     = 100
	QueryTimeout        = 10 * time.Second
	maxDateRangeDays    = 365 * 2
)

// ServiceError carries the HTTP status the handler should respond with,
// keeping status-code decisions in the service (which knows WHY a request
// failed) rather than guessed at in the handler.
type ServiceError struct {
	Status int
	Msg    string
}

func (e *ServiceError) Error() string { return e.Msg }

func badRequest(format string, a ...interface{}) error {
	return &ServiceError{Status: 400, Msg: fmt.Sprintf(format, a...)}
}

func forbidden(msg string) error {
	return &ServiceError{Status: 403, Msg: msg}
}

func internalError() error {
	return &ServiceError{Status: 500, Msg: "Terjadi kesalahan saat memproses permintaan"}
}

// Scope is the caller's authorization context, resolved by the HANDLER from
// auth/plant-scope middleware locals and passed in explicitly — the service
// itself never touches *fiber.Ctx, keeping it unit-testable and honoring the
// "handler does HTTP parsing only" layering rule.
type Scope struct {
	IsSuperAdmin bool
	UserPlantID  string
	RoleID       string
	UserID       string
	// QueryPlantID is the `plant_id` query param, meaningful only for
	// SuperAdmin (mirrors dashboard_handler.go's getAllowedAreas).
	QueryPlantID string
}

// isAuditorFamilyRole mirrors middleware.IsAuditorRole's SuperAdmin/Admin/
// Auditor grouping (ROLE-000/001/002 => "unrestricted if no plant
// assigned"). Duplicated intentionally rather than importing the middleware
// package from a usecase — the existing codebase already duplicates this
// exact 3-line check in more than one place (see permission_middleware.go),
// so this follows established convention rather than introducing a new one.
func isAuditorFamilyRole(roleID string) bool {
	r := strings.ToUpper(strings.TrimSpace(roleID))
	return r == "ROLE-000" || r == "ROLE-001" || r == "ROLE-002"
}

// QueryRequest is the semantic (not SQL) request body for POST /analytics/query.
type QueryRequest struct {
	Measures  []string
	Dimension string
	// Dimension2 is optional — only Heatmap/Sankey use a second grouping
	// dimension (a matrix: dim1 x dim2 x 1 measure). "" = classic
	// single-dimension query, unchanged behavior.
	Dimension2 string
	// Filters is an optional set of ad-hoc "dimension = value" equality
	// filters, independent of Dimension/Dimension2 — the primitive that
	// powers the Custom KPI Builder's drill-down (see DynamicKPIWidget):
	// drilling into a category re-issues the query with that category
	// pinned here and the NEXT hierarchy level as the new Dimension. Backend
	// has no concept of "hierarchy" — that ordering lives purely in the
	// frontend/persisted widget config.
	Filters   []DimensionFilter
	AreaID    string
	StartDate string // "YYYY-MM-DD", inclusive; "" = no lower bound
	EndDate   string // "YYYY-MM-DD", exclusive; "" = no upper bound
}

// DimensionFilter pins one catalog dimension to an exact key value (the
// `key` a prior query response returned for some row, never the display
// label) — the generic building block drill-down filters accumulate as the
// user goes deeper.
type DimensionFilter struct {
	DimensionID string
	Value       string
}

// ResultRow is one grouped row returned by a QueryExecutor. Measures maps
// measure ID -> value; a nil pointer means SQL NULL (e.g. AVG over zero
// matching rows, or a NULLIF-guarded division by zero) — never an error.
// Key2/Category2 are nil unless the query used a Dimension2.
type ResultRow struct {
	Key       string
	Category  string
	Key2      *string
	Category2 *string
	Measures  map[string]*float64
}

// QueryPlan is the fully-resolved, semantic description of one query: which
// dimension/measures/joins, and what security/date scope applies. It
// carries no SQL text — assembling the actual SQL string is
// analyticsrepo's job (see backend/internal/infrastructure/persistence/analyticsrepo),
// keeping "how SQL gets built" out of the usecase layer.
type QueryPlan struct {
	Dimension analytics.DimensionDef
	// Dimension2 is nil for a classic single-dimension query, or set for a
	// Heatmap/Sankey matrix query (see RunQuery — always paired with
	// exactly 1 measure when set).
	Dimension2 *analytics.DimensionDef
	Measures   []analytics.MeasureDef
	// Filters is the resolved form of QueryRequest.Filters — each entry's
	// DimensionDef is already looked up, so analyticsrepo never touches the
	// catalog map itself (same "resolved, no lookups downstream" rule
	// QueryPlan already follows for Dimension/Dimension2/Measures).
	Filters []PlanFilter
	Joins   []analytics.JoinDef
	// AllowedAreas is nil for unrestricted access, or the exact set of
	// Area_Master.AreaID values the scoped_issues CTE must filter to.
	AllowedAreas []string
	StartDate    string
	EndDate      string
	Limit        int
}

// PlanFilter is one resolved DimensionFilter — see QueryPlan.Filters.
type PlanFilter struct {
	Dimension analytics.DimensionDef
	Value     string
}

// QueryExecutor is the port the service depends on; analyticsrepo.Repository
// implements it. Injected via NewQueryService so the service stays testable
// without a live DB.
type QueryExecutor interface {
	Execute(ctx context.Context, plan QueryPlan) ([]ResultRow, error)
}

// CatalogFieldDTO is the client-facing shape of one catalog entry. It
// NEVER carries SelectExpr/KeyExpr/LabelExpr/JoinSQL/CTE — only what a
// drag-and-drop UI needs to let a non-technical user pick fields safely.
type CatalogFieldDTO struct {
	ID               string                `json:"id"`
	Label            string                `json:"label"`
	Description      string                `json:"description"`
	ValueType        string                `json:"value_type,omitempty"`
	Format           *analytics.FormatMeta `json:"format,omitempty"`
	Kind             string                `json:"kind,omitempty"`
	CompatibleCharts []string              `json:"compatible_charts,omitempty"`
	FansOut          bool                  `json:"fans_out,omitempty"`
}

type CatalogResponse struct {
	Dimensions []CatalogFieldDTO `json:"dimensions"`
	Measures   []CatalogFieldDTO `json:"measures"`
}

type QueryResultDTO struct {
	Dimension  CatalogFieldDTO          `json:"dimension"`
	Dimension2 *CatalogFieldDTO         `json:"dimension2,omitempty"`
	Measures   []CatalogFieldDTO        `json:"measures"`
	Rows       []map[string]interface{} `json:"rows"`
}

// QueryService validates and executes semantic analytics queries. It owns
// the catalog and the (small, metadata-only) area-scope lookups; the actual
// analytics SQL is delegated to the injected QueryExecutor.
type QueryService struct {
	db       *gorm.DB
	executor QueryExecutor
	log      *logger.Logger
	catalog  analytics.Catalog
}

func NewQueryService(db *gorm.DB, executor QueryExecutor, log *logger.Logger) *QueryService {
	return &QueryService{db: db, executor: executor, log: log, catalog: analytics.DefaultCatalog()}
}

// Catalog returns the read-only, client-safe catalog for GET /analytics/catalog.
func (s *QueryService) Catalog() CatalogResponse {
	dims := make([]CatalogFieldDTO, 0, len(s.catalog.Dimensions))
	for _, d := range s.catalog.Dimensions {
		dims = append(dims, CatalogFieldDTO{
			ID: d.ID, Label: d.Label, Description: d.Description,
			Kind: string(d.Kind), FansOut: d.FansOut,
		})
	}
	meas := make([]CatalogFieldDTO, 0, len(s.catalog.Measures))
	for _, m := range s.catalog.Measures {
		format := m.Format
		meas = append(meas, CatalogFieldDTO{
			ID: m.ID, Label: m.Label, Description: m.Description,
			ValueType: string(m.ValueType), Format: &format,
			CompatibleCharts: m.CompatibleCharts,
		})
	}
	sort.Slice(dims, func(i, j int) bool { return dims[i].ID < dims[j].ID })
	sort.Slice(meas, func(i, j int) bool { return meas[i].ID < meas[j].ID })
	return CatalogResponse{Dimensions: dims, Measures: meas}
}

// RunQuery validates the request against the catalog + caller's area scope,
// resolves the join graph, and delegates execution to the QueryExecutor.
func (s *QueryService) RunQuery(ctx context.Context, scope Scope, req QueryRequest) (*QueryResultDTO, error) {
	if len(req.Measures) == 0 {
		return nil, badRequest("Minimal 1 measure harus dipilih")
	}
	if len(req.Measures) > MaxMeasuresPerQuery {
		return nil, badRequest("Maksimal %d measure per visualisasi", MaxMeasuresPerQuery)
	}

	dim, ok := s.catalog.Dimensions[req.Dimension]
	if !ok {
		return nil, badRequest("Dimension tidak dikenal: %s", req.Dimension)
	}

	var dim2 *analytics.DimensionDef
	if req.Dimension2 != "" {
		d2, ok := s.catalog.Dimensions[req.Dimension2]
		if !ok {
			return nil, badRequest("Dimension2 tidak dikenal: %s", req.Dimension2)
		}
		dim2 = &d2
	}

	seen := make(map[string]bool, len(req.Measures))
	measureDefs := make([]analytics.MeasureDef, 0, len(req.Measures))
	for _, mID := range req.Measures {
		if seen[mID] {
			continue // silently dedupe a repeated measure id
		}
		seen[mID] = true
		md, ok := s.catalog.Measures[mID]
		if !ok {
			return nil, badRequest("Measure tidak dikenal: %s", mID)
		}
		measureDefs = append(measureDefs, md)
	}
	if len(measureDefs) == 0 {
		return nil, badRequest("Minimal 1 measure harus dipilih")
	}
	// Heatmap/Sankey render a dim1 x dim2 matrix cell per row — a second
	// measure would mean a second matrix, which the UI never exposes, so
	// this is enforced explicitly rather than silently using only the
	// first measure.
	if dim2 != nil && len(measureDefs) != 1 {
		return nil, badRequest("Heatmap/Sankey (dua kategori) hanya mendukung tepat 1 measure")
	}

	planFilters := make([]PlanFilter, 0, len(req.Filters))
	seenFilterDims := map[string]bool{req.Dimension: true}
	for _, f := range req.Filters {
		if seenFilterDims[f.DimensionID] {
			return nil, badRequest("Filter dimension duplikat atau sama dengan dimension yang ditampilkan: %s", f.DimensionID)
		}
		seenFilterDims[f.DimensionID] = true
		fd, ok := s.catalog.Dimensions[f.DimensionID]
		if !ok {
			return nil, badRequest("Filter dimension tidak dikenal: %s", f.DimensionID)
		}
		planFilters = append(planFilters, PlanFilter{Dimension: fd, Value: f.Value})
	}

	startDate, endDate, err := validateDateRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	allowedAreas, err := s.resolveAllowedAreas(scope)
	if err != nil {
		s.log.Error("analytics: failed to resolve allowed areas", logger.Error(err))
		return nil, internalError()
	}

	if req.AreaID != "" {
		if err := s.authorizeAreaID(allowedAreas, req.AreaID); err != nil {
			return nil, err
		}
		allowedAreas = []string{req.AreaID}
	}

	requestedJoins := make([]analytics.JoinID, 0, 4)
	requestedJoins = append(requestedJoins, dim.RequiredJoins...)
	if dim2 != nil {
		requestedJoins = append(requestedJoins, dim2.RequiredJoins...)
	}
	for _, md := range measureDefs {
		requestedJoins = append(requestedJoins, md.RequiredJoins...)
	}
	for _, pf := range planFilters {
		requestedJoins = append(requestedJoins, pf.Dimension.RequiredJoins...)
	}
	joins, err := ResolveJoins(requestedJoins, s.catalog.Joins)
	if err != nil {
		// Only reachable if the catalog itself is malformed (unknown/circular
		// join declared in code) — a developer bug, never caused by request
		// input, so it's logged in detail and given a generic 500.
		s.log.Error("analytics: join resolution failed (catalog bug)", logger.Error(err))
		return nil, internalError()
	}

	plan := QueryPlan{
		Dimension:    dim,
		Dimension2:   dim2,
		Measures:     measureDefs,
		Filters:      planFilters,
		Joins:        joins,
		AllowedAreas: allowedAreas,
		StartDate:    startDate,
		EndDate:      endDate,
		Limit:        MaxRowsReturned,
	}

	measureIDs := make([]string, len(measureDefs))
	for i, m := range measureDefs {
		measureIDs[i] = m.ID
	}

	execCtx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	start := time.Now()
	rows, err := s.executor.Execute(execCtx, plan)
	duration := time.Since(start)
	if err != nil {
		s.log.Error("analytics query failed",
			logger.String("dimension", dim.ID),
			logger.String("measures", strings.Join(measureIDs, ",")),
			logger.Error(err))
		return nil, internalError()
	}
	s.log.Info("analytics query executed",
		logger.String("dimension", dim.ID),
		logger.String("measures", strings.Join(measureIDs, ",")),
		logger.String("duration", duration.String()))

	respRows := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		row := map[string]interface{}{"key": r.Key, "category": r.Category}
		if r.Key2 != nil {
			row["key2"] = *r.Key2
		}
		if r.Category2 != nil {
			row["category2"] = *r.Category2
		}
		for _, mID := range measureIDs {
			if v, ok := r.Measures[mID]; ok && v != nil {
				row[mID] = *v
			} else {
				row[mID] = nil
			}
		}
		respRows = append(respRows, row)
	}

	measureDTOs := make([]CatalogFieldDTO, len(measureDefs))
	for i, m := range measureDefs {
		format := m.Format
		measureDTOs[i] = CatalogFieldDTO{
			ID: m.ID, Label: m.Label, Description: m.Description,
			ValueType: string(m.ValueType), Format: &format,
			CompatibleCharts: m.CompatibleCharts,
		}
	}

	var dim2DTO *CatalogFieldDTO
	if dim2 != nil {
		dim2DTO = &CatalogFieldDTO{ID: dim2.ID, Label: dim2.Label, Description: dim2.Description, Kind: string(dim2.Kind), FansOut: dim2.FansOut}
	}

	return &QueryResultDTO{
		Dimension:  CatalogFieldDTO{ID: dim.ID, Label: dim.Label, Description: dim.Description, Kind: string(dim.Kind), FansOut: dim.FansOut},
		Dimension2: dim2DTO,
		Measures:   measureDTOs,
		Rows:       respRows,
	}, nil
}

// validateDateRange enforces start < end and a sane maximum span; both
// values are optional (either or both may be "").
func validateDateRange(startDate, endDate string) (string, string, error) {
	if startDate == "" && endDate == "" {
		return "", "", nil
	}
	layout := "2006-01-02"
	var start, end time.Time
	var err error
	if startDate != "" {
		start, err = time.Parse(layout, startDate)
		if err != nil {
			return "", "", badRequest("start_date harus berformat YYYY-MM-DD")
		}
	}
	if endDate != "" {
		end, err = time.Parse(layout, endDate)
		if err != nil {
			return "", "", badRequest("end_date harus berformat YYYY-MM-DD")
		}
	}
	if startDate != "" && endDate != "" {
		if !end.After(start) {
			return "", "", badRequest("end_date harus setelah start_date")
		}
		if end.Sub(start) > time.Duration(maxDateRangeDays)*24*time.Hour {
			return "", "", badRequest("Rentang tanggal maksimal %d hari", maxDateRangeDays)
		}
	}
	return startDate, endDate, nil
}

// resolveAllowedAreas mirrors dashboard_handler.go's getAllowedAreas
// (SuperAdmin unrestricted-or-plant-filtered, plant-scoped users restricted
// to their plant's areas, Auditor-family roles with no plant unrestricted)
// with ONE deliberate correction: the Auditee fallback resolves areas via
// the correct PIC_Mapping.KawasanID -> Kawasan_Master.AreaID join, not the
// deprecated PIC_Mapping.AreaID column getAllowedAreas still reads (see the
// plan's "Temuan Repository" section — that existing bug is left as-is,
// out of scope for this feature, but is not reproduced here).
func (s *QueryService) resolveAllowedAreas(scope Scope) ([]string, error) {
	if scope.IsSuperAdmin {
		if scope.QueryPlantID == "" || scope.QueryPlantID == "all" {
			return nil, nil // unrestricted
		}
		return s.areasForPlant(scope.QueryPlantID)
	}

	if scope.UserPlantID != "" {
		return s.areasForPlant(scope.UserPlantID)
	}

	if isAuditorFamilyRole(scope.RoleID) {
		return nil, nil // unrestricted
	}

	var areaIDs []string
	err := s.db.Table(`"PIC_Mapping" pm`).
		Select(`DISTINCT k."AreaID"`).
		Joins(`JOIN "Kawasan_Master" k ON k."KawasanID" = pm."KawasanID"`).
		Where(`pm."UserID" = ?`, scope.UserID).
		Scan(&areaIDs).Error
	if err != nil {
		return nil, err
	}
	if len(areaIDs) == 0 {
		return []string{"RESTRICTED_NONE"}, nil
	}
	return areaIDs, nil
}

func (s *QueryService) areasForPlant(plantID string) ([]string, error) {
	var areaIDs []string
	q := s.db.Table(`"Area_Master"`).Select(`"AreaID"`)
	if plantID == "GLOBAL" || plantID == "NULL" {
		q = q.Where(`"PlantID" IS NULL OR "PlantID" = ''`)
	} else {
		q = q.Where(`"PlantID" = ?`, plantID)
	}
	if err := q.Scan(&areaIDs).Error; err != nil {
		return nil, err
	}
	if len(areaIDs) == 0 {
		return []string{"RESTRICTED_NONE"}, nil
	}
	return areaIDs, nil
}

// authorizeAreaID rejects an explicitly-requested area_id the caller isn't
// allowed to see. Unlike getAllowedAreas' silent RESTRICTED_NONE (fine for
// "no area at all"), an explicit out-of-scope area_id is a real
// unauthorized-access attempt and must be rejected loudly (403), never
// answered with a quiet empty result.
func (s *QueryService) authorizeAreaID(allowedAreas []string, areaID string) error {
	if allowedAreas == nil {
		// Unrestricted caller — still confirm the area actually exists so a
		// typo'd/garbage id gets a clear 400 instead of a silently-empty result.
		var count int64
		if err := s.db.Table(`"Area_Master"`).Where(`"AreaID" = ?`, areaID).Count(&count).Error; err != nil {
			s.log.Error("analytics: failed to verify area_id", logger.Error(err))
			return internalError()
		}
		if count == 0 {
			return badRequest("area_id tidak ditemukan")
		}
		return nil
	}
	for _, a := range allowedAreas {
		if a == areaID {
			return nil
		}
	}
	return forbidden("Anda tidak memiliki akses ke area ini")
}
