package inspectionusecase

import (
	"testing"

	"github.com/monitoring-system/backend/internal/domain/pic"
)

func TestFormatKawasanPICsOrdersByRoleAndLabelsIt(t *testing.T) {
	got := formatKawasanPICs([]pic.ResponsibleUser{
		{UserID: "U3", FullName: "Ariansyah", KategoriPIC: "Staff"},
		{UserID: "U1", FullName: "Hasan Basri", KategoriPIC: "Supervisor"},
		{UserID: "U9", FullName: "Tanpa Kategori", KategoriPIC: ""},
		{UserID: "U2", FullName: "Deden", KategoriPIC: "Manager"},
		{UserID: "U1", FullName: "Hasan Basri", KategoriPIC: "Supervisor"}, // duplicate mapping
	})
	want := "Deden (Manager), Hasan Basri (Supervisor), Ariansyah (Staff), Tanpa Kategori"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFormatKawasanPICsEmpty(t *testing.T) {
	if got := formatKawasanPICs(nil); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestBuildKawasanReportUsesKawasanPICsForEveryIssue(t *testing.T) {
	r := buildKawasanReport(reportRows(), decryptStub, fixedNow(), "http://x", "PLT-SENTUL", "Deden (Manager), Hasan Basri (Supervisor)")
	for _, issue := range r.Issues {
		if issue.PICName != "Deden (Manager), Hasan Basri (Supervisor)" {
			t.Fatalf("issue %q PIC = %q, want the kawasan PICs", issue.Title, issue.PICName)
		}
	}
}
