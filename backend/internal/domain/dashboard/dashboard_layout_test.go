package dashboard

import (
	"fmt"
	"strings"
	"testing"
)

func validCustomQuery() *CustomQueryConfig {
	return &CustomQueryConfig{
		Version:   1,
		Title:     "  Kepatuhan per Area  ",
		Measures:  []string{"total_issues"},
		Dimension: "area",
	}
}

func TestValidateCustomQueryNormalizesTitle(t *testing.T) {
	cq := validCustomQuery()
	if err := ValidateCustomQuery(cq, nil); err != nil {
		t.Fatalf("ValidateCustomQuery() error = %v", err)
	}
	if cq.Title != "Kepatuhan per Area" {
		t.Fatalf("title = %q, want trimmed title", cq.Title)
	}
}

func TestValidateCustomQueryRejectsInvalidTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
	}{
		{name: "too long", title: strings.Repeat("a", maxCustomQueryTitleLength+1)},
		{name: "control character", title: "Temuan\nPer Area"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cq := validCustomQuery()
			cq.Title = tt.title
			if err := ValidateCustomQuery(cq, nil); err == nil {
				t.Fatal("expected title validation error")
			}
		})
	}
}

func TestValidateCustomQueryAcceptsLegacyEmptyTitle(t *testing.T) {
	cq := validCustomQuery()
	cq.Title = ""
	if err := ValidateCustomQuery(cq, nil); err != nil {
		t.Fatalf("legacy empty title must remain valid: %v", err)
	}
}

func TestValidateCustomQueryRejectsOverLimitMeasuresForNonTable(t *testing.T) {
	cq := validCustomQuery()
	cq.Measures = make([]string, maxCustomQueryMeasures+1)
	for i := range cq.Measures {
		cq.Measures[i] = fmt.Sprintf("measure_%d", i)
	}
	vizType := "bar"
	if err := ValidateCustomQuery(cq, &vizType); err == nil {
		t.Fatal("expected error for exceeding maxCustomQueryMeasures on a non-table chart")
	}
	// Legacy rows with no VizType at all must be treated the same as a
	// non-table chart, not optimistically allowed through.
	if err := ValidateCustomQuery(cq, nil); err == nil {
		t.Fatal("expected error for exceeding maxCustomQueryMeasures when vizType is nil")
	}
}

func TestValidateCustomQueryAllowsOverLimitMeasuresForTable(t *testing.T) {
	cq := validCustomQuery()
	cq.Measures = make([]string, maxCustomQueryMeasures+1)
	for i := range cq.Measures {
		cq.Measures[i] = fmt.Sprintf("measure_%d", i)
	}
	vizType := "table"
	if err := ValidateCustomQuery(cq, &vizType); err != nil {
		t.Fatalf("table should accept more than %d measures: %v", maxCustomQueryMeasures, err)
	}
}
