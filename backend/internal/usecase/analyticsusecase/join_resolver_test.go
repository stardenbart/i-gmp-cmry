package analyticsusecase

import (
	"testing"

	"github.com/monitoring-system/backend/internal/domain/analytics"
)

func testCatalog() map[analytics.JoinID]analytics.JoinDef {
	return map[analytics.JoinID]analytics.JoinDef{
		"a": {ID: "a"},
		"b": {ID: "b", DependsOn: []analytics.JoinID{"a"}},
		"c": {ID: "c", DependsOn: []analytics.JoinID{"a"}},
		"d": {ID: "d", DependsOn: []analytics.JoinID{"b", "c"}},
	}
}

func TestResolveJoins_DependencyOrder(t *testing.T) {
	got, err := ResolveJoins([]analytics.JoinID{"d"}, testCatalog())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("expected 4 joins (a,b,c,d closure), got %d: %+v", len(got), got)
	}
	pos := make(map[analytics.JoinID]int)
	for i, j := range got {
		pos[j.ID] = i
	}
	if pos["a"] > pos["b"] || pos["a"] > pos["c"] || pos["b"] > pos["d"] || pos["c"] > pos["d"] {
		t.Fatalf("dependency order violated: %+v", pos)
	}
}

func TestResolveJoins_Dedupe(t *testing.T) {
	// Two independent requests both needing "a" through "b" and "c" must
	// not produce "a" twice.
	got, err := ResolveJoins([]analytics.JoinID{"b", "c"}, testCatalog())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := 0
	for _, j := range got {
		if j.ID == "a" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected \"a\" to appear exactly once, appeared %d times: %+v", count, got)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 joins (a,b,c), got %d: %+v", len(got), got)
	}
}

func TestResolveJoins_MissingDependency(t *testing.T) {
	catalog := map[analytics.JoinID]analytics.JoinDef{
		"x": {ID: "x", DependsOn: []analytics.JoinID{"missing"}},
	}
	_, err := ResolveJoins([]analytics.JoinID{"x"}, catalog)
	if err == nil {
		t.Fatal("expected error for missing dependency, got nil")
	}
}

func TestResolveJoins_UnknownRequested(t *testing.T) {
	_, err := ResolveJoins([]analytics.JoinID{"nonexistent"}, testCatalog())
	if err == nil {
		t.Fatal("expected error for unknown requested join, got nil")
	}
}

func TestResolveJoins_CircularDependency(t *testing.T) {
	catalog := map[analytics.JoinID]analytics.JoinDef{
		"a": {ID: "a", DependsOn: []analytics.JoinID{"b"}},
		"b": {ID: "b", DependsOn: []analytics.JoinID{"a"}},
	}
	_, err := ResolveJoins([]analytics.JoinID{"a"}, catalog)
	if err == nil {
		t.Fatal("expected error for circular dependency, got nil")
	}
}

func TestResolveJoins_Empty(t *testing.T) {
	got, err := ResolveJoins(nil, testCatalog())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 joins, got %d", len(got))
	}
}

func TestResolveJoins_RealCatalog(t *testing.T) {
	// Sanity check against the actual production catalog: "pic" dimension's
	// join (pic_fanout) plus every measure (none of which require joins)
	// must resolve without error.
	cat := analytics.DefaultCatalog()
	requested := []analytics.JoinID{analytics.JoinPICFanout}
	got, err := ResolveJoins(requested, cat.Joins)
	if err != nil {
		t.Fatalf("unexpected error resolving real catalog: %v", err)
	}
	if len(got) != 1 || got[0].ID != analytics.JoinPICFanout {
		t.Fatalf("unexpected resolution: %+v", got)
	}

	requestedKawasan := []analytics.JoinID{analytics.JoinKawasanChain}
	gotKawasan, err := ResolveJoins(requestedKawasan, cat.Joins)
	if err != nil {
		t.Fatalf("unexpected error resolving kawasan_chain: %v", err)
	}
	if len(gotKawasan) != 2 {
		t.Fatalf("expected inspection_chain + kawasan_chain, got %+v", gotKawasan)
	}
	if gotKawasan[0].ID != analytics.JoinInspectionChain || gotKawasan[1].ID != analytics.JoinKawasanChain {
		t.Fatalf("unexpected order: %+v", gotKawasan)
	}
}

// TestRealCatalog_EveryFieldResolvesJoins is a self-check that runs
// whenever a new measure/dimension is added to the catalog: every
// RequiredJoins list it declares must resolve without an unknown/circular
// dependency error. Catches a typo'd JoinID at `go test` time instead of
// only surfacing as a 500 the first time that field is actually queried.
func TestRealCatalog_EveryFieldResolvesJoins(t *testing.T) {
	cat := analytics.DefaultCatalog()
	for id, dim := range cat.Dimensions {
		if _, err := ResolveJoins(dim.RequiredJoins, cat.Joins); err != nil {
			t.Errorf("dimension %q: %v", id, err)
		}
	}
	for id, measure := range cat.Measures {
		if _, err := ResolveJoins(measure.RequiredJoins, cat.Joins); err != nil {
			t.Errorf("measure %q: %v", id, err)
		}
	}
}
