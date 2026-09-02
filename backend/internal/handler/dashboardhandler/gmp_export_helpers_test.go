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
