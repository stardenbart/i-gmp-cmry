package dashboard

import (
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
	if err := ValidateCustomQuery(cq); err != nil {
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
			if err := ValidateCustomQuery(cq); err == nil {
				t.Fatal("expected title validation error")
			}
		})
	}
}

func TestValidateCustomQueryAcceptsLegacyEmptyTitle(t *testing.T) {
	cq := validCustomQuery()
	cq.Title = ""
	if err := ValidateCustomQuery(cq); err != nil {
		t.Fatalf("legacy empty title must remain valid: %v", err)
	}
}
