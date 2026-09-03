package middleware

import "testing"

func TestRedactSensitivePathHidesPublicKPIToken(t *testing.T) {
	got := redactSensitivePath("/api/v1/public/kpi/kpi_secret-value/bootstrap")
	want := "/api/v1/public/kpi/[redacted]/bootstrap"
	if got != want {
		t.Fatalf("redactSensitivePath() = %q, want %q", got, want)
	}
}
