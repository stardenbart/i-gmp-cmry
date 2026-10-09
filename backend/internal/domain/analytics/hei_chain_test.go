package analytics

import (
	"strings"
	"testing"
)

// HEI is picked per photo when a finding is entered (Issue_Photo.HEIID);
// Issue_HEI stays empty for those issues. The KPI builder's HEI and
// Kategori HEI dimensions must read the photo's HEI, while keeping the
// catalog's one-row-per-issue grain.
func TestHEIChainReadsPhotoHEIOnePerIssue(t *testing.T) {
	join := DefaultCatalog().Joins[JoinHEIChain]
	cte := join.CTE
	for _, want := range []string{`"Issue_Photo"`, `"PhotoType" = 'Initial'`, `"HEIID"`, `"Issue_HEI"`, `DISTINCT ON`} {
		if !strings.Contains(cte, want) {
			t.Errorf("hei_chain CTE is missing %s:\n%s", want, cte)
		}
	}
	if join.FansOut {
		t.Error("hei_chain must not fan out")
	}
}
