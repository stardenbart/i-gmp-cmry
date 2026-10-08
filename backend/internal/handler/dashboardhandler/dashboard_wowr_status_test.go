package dashboardhandler

import (
	"strings"
	"testing"
)

// The overview's WO/WR counts must match the follow-up page: an issue whose
// status is PendingValidation but has no WO/WR proof photo yet is still
// waiting for execution proof, not for the auditor.
func TestWOWRStatusConditionsSplitPendingByProof(t *testing.T) {
	c := wowrStatusConditions()

	if !strings.Contains(c.Pending, `"WOWRStatus" = 'PendingValidation'`) || !strings.Contains(c.Pending, "AND EXISTS") {
		t.Fatalf("pending must require PendingValidation with proof, got %q", c.Pending)
	}
	if !strings.Contains(c.Awaiting, "NOT EXISTS") || !strings.Contains(c.Awaiting, "'PendingValidation'") {
		t.Fatalf("awaiting must include PendingValidation without proof, got %q", c.Awaiting)
	}
	if !strings.Contains(c.Awaiting, "'None'") || !strings.Contains(c.Awaiting, "IS NULL") {
		t.Fatalf("awaiting must still include None/NULL, got %q", c.Awaiting)
	}
	for _, sql := range []string{c.Pending, c.Awaiting} {
		if !strings.Contains(sql, `"PhotoType" = 'WOWR'`) || !strings.Contains(sql, `"ImageUrl"`) {
			t.Fatalf("proof check must look for a WOWR photo with an image, got %q", sql)
		}
	}
}
