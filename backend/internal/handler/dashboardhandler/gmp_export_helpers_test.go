package dashboardhandler

import "testing"

func TestNormalizeGMPExportFormat(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{"", gmpExportFormatTemplate, true},
		{" TEMPLATE ", gmpExportFormatTemplate, true},
		{"table", gmpExportFormatTable, true},
		{"csv", "", false},
	}
	for _, test := range tests {
		got, ok := normalizeGMPExportFormat(test.input)
		if got != test.want || ok != test.ok {
			t.Fatalf("normalize %q = (%q, %v), want (%q, %v)", test.input, got, ok, test.want, test.ok)
		}
	}
}

func TestGMPAspectGroupKeyNormalizesAspectButKeepsInspectionBoundary(t *testing.T) {
	if gmpAspectGroupKey("INSP-001", "  Kebersihan ") != gmpAspectGroupKey("INSP-001", "KEBERSIHAN") {
		t.Fatal("expected equivalent aspect labels in one inspection to share a group")
	}
	if gmpAspectGroupKey("INSP-001", "Kebersihan") == gmpAspectGroupKey("INSP-002", "Kebersihan") {
		t.Fatal("expected identical aspect labels from different inspections to remain separate")
	}
}

func TestResolveGMPExportPlantIDPrefersAuthenticatedUserPlant(t *testing.T) {
	if plantID := resolveGMPExportPlantID("PLT-SENTUL", "PLT-PASURUAN"); plantID != "PLT-SENTUL" {
		t.Fatalf("expected authenticated user's plant, got %q", plantID)
	}
	if plantID := resolveGMPExportPlantID("", "PLT-CICURUG"); plantID != "PLT-CICURUG" {
		t.Fatalf("expected explicit plant filter, got %q", plantID)
	}
	if plantID := resolveGMPExportPlantID("", "all"); plantID != "" {
		t.Fatalf("expected all-plant selection to remain global, got %q", plantID)
	}
}

func TestValidateGMPDateRange(t *testing.T) {
	if err := validateGMPDateRange("2026-09-01", "2026-09-02"); err != nil {
		t.Fatalf("expected valid range: %v", err)
	}
	if err := validateGMPDateRange("2026/09/01", ""); err == nil {
		t.Fatal("expected invalid date format")
	}
	if err := validateGMPDateRange("2026-09-03", "2026-09-02"); err == nil {
		t.Fatal("expected reversed range to fail")
	}
}

func TestMatchesGMPSearchIncludesFollowUpDescription(t *testing.T) {
	evidence := []gmpFollowUpEvidence{{Keterangan: "Mesin sudah dibersihkan"}}
	if !matchesGMPSearch("DIBERSIHKAN", []string{"INSP-001"}, evidence) {
		t.Fatal("expected case-insensitive follow-up description match")
	}
	if matchesGMPSearch("tidak ada", []string{"INSP-001"}, evidence) {
		t.Fatal("unexpected match")
	}
}
