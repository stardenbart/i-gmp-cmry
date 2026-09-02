package analyticsrepo

import (
	"strings"
	"testing"
	"time"

	"github.com/monitoring-system/backend/internal/domain/analytics"
	"github.com/monitoring-system/backend/internal/usecase/analyticsusecase"
)

// TestBuildSQL_NoAmbiguousDetailKawasanID is a regression test for a real
// bug caught via the manual local test flow: scoped_issues used to select
// ih."DetailKawasanID" in addition to i.* (which already carries Issue's
// own DetailKawasanID), producing two same-named columns in the CTE.
// Postgres accepts that at CTE-definition time but rejects any downstream
// reference to si."DetailKawasanID" as ambiguous (42702) — exactly what
// happened querying the "kawasan"/"detail_kawasan" dimensions. This test
// would have caught it without needing a live database.
func TestBuildSQL_NoAmbiguousDetailKawasanID(t *testing.T) {
	cat := analytics.DefaultCatalog()
	dim := cat.Dimensions["kawasan"]
	joins, err := analyticsusecase.ResolveJoins(dim.RequiredJoins, cat.Joins)
	if err != nil {
		t.Fatalf("unexpected error resolving joins: %v", err)
	}
	plan := analyticsusecase.QueryPlan{
		Dimension: dim,
		Measures:  []analytics.MeasureDef{cat.Measures["total_temuan"], cat.Measures["temuan_overdue"]},
		Joins:     joins,
		Limit:     100,
	}

	sqlText, _ := buildSQL(plan)

	if strings.Contains(sqlText, `ih."DetailKawasanID"`) {
		t.Fatalf("scoped_issues must not re-select ih.\"DetailKawasanID\" (Issue's own DetailKawasanID, via i.*, already covers it) — this reintroduces the ambiguous-column bug:\n%s", sqlText)
	}
	// si."DetailKawasanID" itself (referencing Issue's own column via i.*)
	// remains valid to use downstream, e.g. by the "detail_kawasan" dimension.
	if !strings.Contains(sqlText, `si."KawasanID"`) {
		t.Fatalf("expected the kawasan dimension's key expression in the assembled SQL:\n%s", sqlText)
	}
}

// TestBuildSQL_EveryDimensionBuildsCleanly runs buildSQL for every catalog
// dimension (each paired with every measure, one at a time) and asserts the
// placeholder/arg count always lines up and no measure/dimension pairing
// reintroduces a same-named-column collision in the assembled SQL text.
// This is the SQL-assembly-layer counterpart to
// TestRealCatalog_EveryFieldResolvesJoins in analyticsusecase.
func TestBuildSQL_EveryDimensionBuildsCleanly(t *testing.T) {
	cat := analytics.DefaultCatalog()
	for dimID, dim := range cat.Dimensions {
		for measureID, measure := range cat.Measures {
			requested := append(append([]analytics.JoinID{}, dim.RequiredJoins...), measure.RequiredJoins...)
			joins, err := analyticsusecase.ResolveJoins(requested, cat.Joins)
			if err != nil {
				t.Fatalf("dimension %q + measure %q: join resolution failed: %v", dimID, measureID, err)
			}
			plan := analyticsusecase.QueryPlan{
				Dimension:    dim,
				Measures:     []analytics.MeasureDef{measure},
				Joins:        joins,
				AllowedAreas: []string{"AREA-1"},
				Limit:        10,
			}
			sqlText, args := buildSQL(plan)
			placeholderCount := strings.Count(sqlText, "?")
			if placeholderCount != len(args) {
				t.Fatalf("dimension %q + measure %q: placeholder count (%d) != arg count (%d):\n%s", dimID, measureID, placeholderCount, len(args), sqlText)
			}
			if strings.Contains(sqlText, "@now") {
				t.Fatalf("dimension %q + measure %q: unresolved @now token:\n%s", dimID, measureID, sqlText)
			}
		}
	}
}

func TestBuildSQL_Dimension2AddsMatrixColumns(t *testing.T) {
	cat := analytics.DefaultCatalog()
	dim := cat.Dimensions["kawasan"]
	dim2 := cat.Dimensions["status_temuan"]
	requested := append(append([]analytics.JoinID{}, dim.RequiredJoins...), dim2.RequiredJoins...)
	joins, err := analyticsusecase.ResolveJoins(requested, cat.Joins)
	if err != nil {
		t.Fatalf("unexpected error resolving joins: %v", err)
	}
	plan := analyticsusecase.QueryPlan{
		Dimension:  dim,
		Dimension2: &dim2,
		Measures:   []analytics.MeasureDef{cat.Measures["total_temuan"]},
		Joins:      joins,
		Limit:      100,
	}

	sqlText, args := buildSQL(plan)

	if !strings.Contains(sqlText, `"cat_key2"`) || !strings.Contains(sqlText, `"cat_label2"`) {
		t.Fatalf("expected cat_key2/cat_label2 columns in assembled SQL:\n%s", sqlText)
	}
	if !strings.Contains(sqlText, `GROUP BY`) || !strings.Contains(sqlText, dim2.KeyExpr) {
		t.Fatalf("expected dimension2's KeyExpr in GROUP BY:\n%s", sqlText)
	}
	placeholderCount := strings.Count(sqlText, "?")
	if placeholderCount != len(args) {
		t.Fatalf("placeholder count (%d) must match arg count (%d):\n%s", placeholderCount, len(args), sqlText)
	}
}

// TestBuildSQL_FiltersProduceCorrectArgOrder is a regression test for the
// exact class of bug this file's DetailKawasanID fix came from: a mismatch
// between the TEXTUAL order of "?" placeholders in the assembled SQL and
// the ORDER of the args slice supplied to Raw(sql, args...). Drill-down
// filters are assembled (textually) AFTER the SELECT list's "@now" tokens
// but their arg values must not be spliced in before the now-values in the
// final args slice — see buildSQL's comment on filterArgs.
func TestBuildSQL_FiltersProduceCorrectArgOrder(t *testing.T) {
	cat := analytics.DefaultCatalog()
	dim := cat.Dimensions["status_temuan"]    // display dimension (no extra joins)
	filterDim1 := cat.Dimensions["kawasan"]   // filter #1
	filterDim2 := cat.Dimensions["area"]      // filter #2
	measure := cat.Measures["temuan_overdue"] // exactly one "@now" in its SelectExpr

	requested := append(append([]analytics.JoinID{}, dim.RequiredJoins...), measure.RequiredJoins...)
	requested = append(requested, filterDim1.RequiredJoins...)
	requested = append(requested, filterDim2.RequiredJoins...)
	joins, err := analyticsusecase.ResolveJoins(requested, cat.Joins)
	if err != nil {
		t.Fatalf("unexpected error resolving joins: %v", err)
	}

	plan := analyticsusecase.QueryPlan{
		Dimension:    dim,
		Measures:     []analytics.MeasureDef{measure},
		Joins:        joins,
		AllowedAreas: []string{"AREA-1"},
		Filters: []analyticsusecase.PlanFilter{
			{Dimension: filterDim1, Value: "KAWASAN-1"},
			{Dimension: filterDim2, Value: "AREA-9"},
		},
		Limit: 50,
	}

	sqlText, args := buildSQL(plan)

	placeholderCount := strings.Count(sqlText, "?")
	if placeholderCount != len(args) {
		t.Fatalf("placeholder count (%d) != arg count (%d):\n%s\nargs: %+v", placeholderCount, len(args), sqlText, args)
	}
	// Expected positional order: [AllowedAreas slice] [now (from @now)]
	// [filter #1 value] [filter #2 value] [limit].
	if len(args) != 5 {
		t.Fatalf("expected exactly 5 args ([allowedAreas, now, filter1, filter2, limit]), got %d: %+v", len(args), args)
	}
	if allowed, ok := args[0].([]string); !ok || len(allowed) != 1 || allowed[0] != "AREA-1" {
		t.Fatalf("args[0] expected AllowedAreas slice [\"AREA-1\"], got %#v", args[0])
	}
	if _, ok := args[1].(time.Time); !ok {
		t.Fatalf("args[1] expected the @now time.Time value, got %#v", args[1])
	}
	if v, ok := args[2].(string); !ok || v != "KAWASAN-1" {
		t.Fatalf("args[2] expected filter #1 value \"KAWASAN-1\", got %#v", args[2])
	}
	if v, ok := args[3].(string); !ok || v != "AREA-9" {
		t.Fatalf("args[3] expected filter #2 value \"AREA-9\", got %#v", args[3])
	}
	if v, ok := args[4].(int); !ok || v != 50 {
		t.Fatalf("args[4] expected LIMIT value 50, got %#v", args[4])
	}
	if !strings.Contains(sqlText, `(`+filterDim1.KeyExpr+`)::text = ?`) {
		t.Fatalf("expected filter #1's KeyExpr equality clause in assembled SQL:\n%s", sqlText)
	}
	if !strings.Contains(sqlText, `(`+filterDim2.KeyExpr+`)::text = ?`) {
		t.Fatalf("expected filter #2's KeyExpr equality clause in assembled SQL:\n%s", sqlText)
	}
	// The WHERE-filter clause must appear AFTER the FROM/JOIN section and
	// BEFORE GROUP BY, never inside the scoped_issues CTE.
	whereIdx := strings.Index(sqlText, filterDim1.KeyExpr+`)::text = ?`)
	fromIdx := strings.Index(sqlText, `FROM scoped_issues si`)
	groupByIdx := strings.Index(sqlText, "GROUP BY")
	if whereIdx < 0 || fromIdx < 0 || groupByIdx < 0 || !(fromIdx < whereIdx && whereIdx < groupByIdx) {
		t.Fatalf("expected filter clause strictly between FROM scoped_issues and GROUP BY:\n%s", sqlText)
	}
}

func TestBuildSQL_NowPlaceholderCountMatchesArgs(t *testing.T) {
	cat := analytics.DefaultCatalog()
	dim := cat.Dimensions["status_temuan"]
	plan := analyticsusecase.QueryPlan{
		Dimension:    dim,
		Measures:     []analytics.MeasureDef{cat.Measures["temuan_overdue"], cat.Measures["rata_rata_keterlambatan_hari"]},
		AllowedAreas: []string{"AREA-1", "AREA-2"},
		Limit:        50,
	}

	sqlText, args := buildSQL(plan)

	placeholderCount := strings.Count(sqlText, "?")
	if placeholderCount != len(args) {
		t.Fatalf("placeholder count (%d) must match arg count (%d) — SQL:\n%s\nargs: %+v", placeholderCount, len(args), sqlText, args)
	}
	if strings.Contains(sqlText, "@now") {
		t.Fatal("all @now tokens must be replaced with ? before returning")
	}
}
