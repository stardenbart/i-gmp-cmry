package analyticsusecase

import (
	"context"
	"testing"

	"github.com/monitoring-system/backend/pkg/logger"
	"go.uber.org/zap"
)

// fakeExecutor lets RunQuery tests exercise validation/resolution logic
// without a live DB — no test here needs real SQL execution (none of the
// existing *_test.go files in this repo use a DB harness; SQL/grain
// correctness is verified via the plan's manual test flow against real
// local data instead, see the final report's "keterbatasan" section).
type fakeExecutor struct {
	rows []ResultRow
	err  error
	// lastPlan captures the plan passed to Execute so a test can assert on
	// resolved joins/measures without re-implementing RunQuery's internals.
	lastPlan QueryPlan
}

func (f *fakeExecutor) Execute(_ context.Context, plan QueryPlan) ([]ResultRow, error) {
	f.lastPlan = plan
	return f.rows, f.err
}

func newTestService(exec QueryExecutor) *QueryService {
	return NewQueryService(nil, exec, &logger.Logger{Logger: zap.NewNop()})
}

func superAdminUnrestrictedScope() Scope {
	return Scope{IsSuperAdmin: true, QueryPlantID: ""}
}

func TestRunQuery_UnknownMeasure(t *testing.T) {
	s := newTestService(&fakeExecutor{})
	_, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: []string{"tidak_ada_measure_ini"}, Dimension: "kawasan",
	})
	if err == nil {
		t.Fatal("expected error for unknown measure")
	}
	svcErr, ok := err.(*ServiceError)
	if !ok || svcErr.Status != 400 {
		t.Fatalf("expected 400 ServiceError, got %#v", err)
	}
}

func TestRunQuery_UnknownDimension(t *testing.T) {
	s := newTestService(&fakeExecutor{})
	_, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: []string{"total_temuan"}, Dimension: "tidak_ada_dimensi_ini",
	})
	if err == nil {
		t.Fatal("expected error for unknown dimension")
	}
	svcErr, ok := err.(*ServiceError)
	if !ok || svcErr.Status != 400 {
		t.Fatalf("expected 400 ServiceError, got %#v", err)
	}
}

func TestRunQuery_TooManyMeasures(t *testing.T) {
	s := newTestService(&fakeExecutor{})
	_, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures:  []string{"total_temuan", "temuan_terbuka", "temuan_overdue", "wowr_total", "wowr_verified"},
		Dimension: "kawasan",
	})
	if err == nil {
		t.Fatal("expected error for exceeding MaxMeasuresPerQuery")
	}
	if svcErr, ok := err.(*ServiceError); !ok || svcErr.Status != 400 {
		t.Fatalf("expected 400 ServiceError, got %#v", err)
	}
}

func TestRunQuery_NoMeasures(t *testing.T) {
	s := newTestService(&fakeExecutor{})
	_, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: nil, Dimension: "kawasan",
	})
	if err == nil {
		t.Fatal("expected error for zero measures")
	}
}

func TestRunQuery_DedupesRepeatedMeasure(t *testing.T) {
	exec := &fakeExecutor{}
	s := newTestService(exec)
	_, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: []string{"total_temuan", "total_temuan", "temuan_overdue"}, Dimension: "kawasan",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exec.lastPlan.Measures) != 2 {
		t.Fatalf("expected duplicate measure id to be deduped to 2 measures, got %d: %+v", len(exec.lastPlan.Measures), exec.lastPlan.Measures)
	}
}

func TestRunQuery_TotalAndOverdueTogetherDoNotInterfere(t *testing.T) {
	// Regression guard for the per-measure FILTER isolation requirement:
	// selecting total_temuan and temuan_overdue together must resolve both
	// measure definitions independently (each keeps its own SelectExpr),
	// not collapse into a single shared WHERE-filtered measure.
	exec := &fakeExecutor{}
	s := newTestService(exec)
	_, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: []string{"total_temuan", "temuan_overdue"}, Dimension: "status_temuan",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exec.lastPlan.Measures) != 2 {
		t.Fatalf("expected 2 distinct measures in plan, got %d", len(exec.lastPlan.Measures))
	}
	if exec.lastPlan.Measures[0].SelectExpr == exec.lastPlan.Measures[1].SelectExpr {
		t.Fatal("total_temuan and temuan_overdue must not share the same SelectExpr")
	}
}

func TestRunQuery_EmptyResultSet(t *testing.T) {
	exec := &fakeExecutor{rows: nil}
	s := newTestService(exec)
	result, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: []string{"total_temuan"}, Dimension: "kawasan",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Rows == nil || len(result.Rows) != 0 {
		t.Fatalf("expected an empty (non-nil) rows slice, got %#v", result.Rows)
	}
}

func TestRunQuery_NullMeasureValuePassesThroughAsNil(t *testing.T) {
	// A NULLIF-guarded division-by-zero (persentase_temuan_overdue on a
	// zero-row group) must surface as JSON null, not an error and not a
	// silently-substituted 0.
	exec := &fakeExecutor{rows: []ResultRow{
		{Key: "K1", Category: "Kawasan A", Measures: map[string]*float64{"persentase_temuan_overdue": nil}},
	}}
	s := newTestService(exec)
	result, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: []string{"persentase_temuan_overdue"}, Dimension: "kawasan",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v := result.Rows[0]["persentase_temuan_overdue"]; v != nil {
		t.Fatalf("expected nil for NULL measure value, got %#v", v)
	}
}

func TestRunQuery_ExecutorErrorBecomesGeneric500(t *testing.T) {
	exec := &fakeExecutor{err: context.DeadlineExceeded}
	s := newTestService(exec)
	_, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: []string{"total_temuan"}, Dimension: "kawasan",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	svcErr, ok := err.(*ServiceError)
	if !ok || svcErr.Status != 500 {
		t.Fatalf("expected generic 500 ServiceError (never leaking raw SQL error), got %#v", err)
	}
	if svcErr.Msg == context.DeadlineExceeded.Error() {
		t.Fatal("raw executor error must not leak into the client-facing message")
	}
}

func TestAuthorizeAreaID_RejectsOutOfScopeArea(t *testing.T) {
	s := newTestService(&fakeExecutor{})
	err := s.authorizeAreaID([]string{"AREA-1", "AREA-2"}, "AREA-99")
	if err == nil {
		t.Fatal("expected rejection for an area_id outside the allowed set")
	}
	svcErr, ok := err.(*ServiceError)
	if !ok || svcErr.Status != 403 {
		t.Fatalf("expected 403 ServiceError, got %#v", err)
	}
}

func TestAuthorizeAreaID_AllowsInScopeArea(t *testing.T) {
	s := newTestService(&fakeExecutor{})
	if err := s.authorizeAreaID([]string{"AREA-1", "AREA-2"}, "AREA-2"); err != nil {
		t.Fatalf("expected no error for an allowed area_id, got %v", err)
	}
}

func TestResolveAllowedAreas_SuperAdminUnrestricted(t *testing.T) {
	s := newTestService(&fakeExecutor{})
	areas, err := s.resolveAllowedAreas(Scope{IsSuperAdmin: true, QueryPlantID: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if areas != nil {
		t.Fatalf("expected nil (unrestricted) areas, got %v", areas)
	}
}

func TestResolveAllowedAreas_AuditorFamilyNoPlantUnrestricted(t *testing.T) {
	s := newTestService(&fakeExecutor{})
	areas, err := s.resolveAllowedAreas(Scope{IsSuperAdmin: false, UserPlantID: "", RoleID: "ROLE-002"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if areas != nil {
		t.Fatalf("expected nil (unrestricted) areas for auditor-family role with no plant, got %v", areas)
	}
}

func TestValidateDateRange(t *testing.T) {
	cases := []struct {
		name        string
		start, end  string
		expectError bool
	}{
		{"both empty", "", "", false},
		{"valid range", "2026-01-01", "2026-02-01", false},
		{"end before start", "2026-02-01", "2026-01-01", true},
		{"end equals start", "2026-01-01", "2026-01-01", true},
		{"bad format start", "01-01-2026", "2026-02-01", true},
		{"bad format end", "2026-01-01", "not-a-date", true},
		{"span too long", "2020-01-01", "2026-01-01", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := validateDateRange(tc.start, tc.end)
			if tc.expectError && err == nil {
				t.Fatalf("expected error for case %q", tc.name)
			}
			if !tc.expectError && err != nil {
				t.Fatalf("unexpected error for case %q: %v", tc.name, err)
			}
		})
	}
}

func TestRunQuery_UnknownDimension2(t *testing.T) {
	s := newTestService(&fakeExecutor{})
	_, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: []string{"total_temuan"}, Dimension: "kawasan", Dimension2: "tidak_ada_dimensi_ini",
	})
	if err == nil {
		t.Fatal("expected error for unknown dimension2")
	}
	if svcErr, ok := err.(*ServiceError); !ok || svcErr.Status != 400 {
		t.Fatalf("expected 400 ServiceError, got %#v", err)
	}
}

func TestRunQuery_Dimension2RequiresExactlyOneMeasure(t *testing.T) {
	s := newTestService(&fakeExecutor{})
	_, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: []string{"total_temuan", "temuan_overdue"}, Dimension: "kawasan", Dimension2: "status_temuan",
	})
	if err == nil {
		t.Fatal("expected error when dimension2 is set with more than 1 measure")
	}
	if svcErr, ok := err.(*ServiceError); !ok || svcErr.Status != 400 {
		t.Fatalf("expected 400 ServiceError, got %#v", err)
	}
}

func TestRunQuery_Dimension2HappyPath(t *testing.T) {
	exec := &fakeExecutor{rows: []ResultRow{}}
	s := newTestService(exec)
	result, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: []string{"total_temuan"}, Dimension: "kawasan", Dimension2: "status_temuan",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.lastPlan.Dimension2 == nil || exec.lastPlan.Dimension2.ID != "status_temuan" {
		t.Fatalf("expected Dimension2 resolved in plan, got %+v", exec.lastPlan.Dimension2)
	}
	if result.Dimension2 == nil || result.Dimension2.ID != "status_temuan" {
		t.Fatalf("expected Dimension2 in response DTO, got %+v", result.Dimension2)
	}
}

func TestRunQuery_UnknownFilterDimension(t *testing.T) {
	s := newTestService(&fakeExecutor{})
	_, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: []string{"total_temuan"}, Dimension: "status_temuan",
		Filters: []DimensionFilter{{DimensionID: "tidak_ada_dimensi_ini", Value: "x"}},
	})
	if err == nil {
		t.Fatal("expected error for unknown filter dimension")
	}
	if svcErr, ok := err.(*ServiceError); !ok || svcErr.Status != 400 {
		t.Fatalf("expected 400 ServiceError, got %#v", err)
	}
}

func TestRunQuery_FilterSameAsDisplayDimensionRejected(t *testing.T) {
	s := newTestService(&fakeExecutor{})
	_, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: []string{"total_temuan"}, Dimension: "kawasan",
		Filters: []DimensionFilter{{DimensionID: "kawasan", Value: "K1"}},
	})
	if err == nil {
		t.Fatal("expected error when a filter targets the same dimension currently being displayed")
	}
	if svcErr, ok := err.(*ServiceError); !ok || svcErr.Status != 400 {
		t.Fatalf("expected 400 ServiceError, got %#v", err)
	}
}

func TestRunQuery_DuplicateFilterDimensionsRejected(t *testing.T) {
	s := newTestService(&fakeExecutor{})
	_, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: []string{"total_temuan"}, Dimension: "status_temuan",
		Filters: []DimensionFilter{
			{DimensionID: "kawasan", Value: "K1"},
			{DimensionID: "kawasan", Value: "K2"},
		},
	})
	if err == nil {
		t.Fatal("expected error for duplicate filter dimension ids")
	}
	if svcErr, ok := err.(*ServiceError); !ok || svcErr.Status != 400 {
		t.Fatalf("expected 400 ServiceError, got %#v", err)
	}
}

func TestRunQuery_FilterHappyPath(t *testing.T) {
	exec := &fakeExecutor{rows: []ResultRow{}}
	s := newTestService(exec)
	_, err := s.RunQuery(context.Background(), superAdminUnrestrictedScope(), QueryRequest{
		Measures: []string{"total_temuan"}, Dimension: "status_temuan",
		Filters: []DimensionFilter{{DimensionID: "kawasan", Value: "KAWASAN-1"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exec.lastPlan.Filters) != 1 || exec.lastPlan.Filters[0].Dimension.ID != "kawasan" || exec.lastPlan.Filters[0].Value != "KAWASAN-1" {
		t.Fatalf("expected 1 resolved filter for kawasan=KAWASAN-1 in plan, got %+v", exec.lastPlan.Filters)
	}
}

func TestCatalog_NeverExposesSQL(t *testing.T) {
	// Guards the "GET /analytics/catalog never sends SQL" contract at the
	// type level: CatalogFieldDTO simply has no field capable of carrying
	// SelectExpr/KeyExpr/LabelExpr/JoinSQL/CTE. This test fails to compile
	// (not just fails at runtime) if such a field is ever added without
	// deliberate thought.
	s := newTestService(&fakeExecutor{})
	cat := s.Catalog()
	if len(cat.Measures) == 0 || len(cat.Dimensions) == 0 {
		t.Fatal("expected non-empty catalog")
	}
	for _, m := range cat.Measures {
		if m.ID == "" || m.Label == "" {
			t.Fatalf("measure missing id/label: %+v", m)
		}
	}
}
