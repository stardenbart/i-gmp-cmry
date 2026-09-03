package authz

import "testing"

func TestPlantMatches(t *testing.T) {
	plt := func(s string) *string { return &s }

	cases := []struct {
		name            string
		userPlantID     string
		resourcePlantID *string
		want            bool
	}{
		{"unscoped caller sees everything", "", plt("PLT-A"), true},
		{"unscoped caller sees area with no plant", "", nil, true},
		{"same plant matches", "PLT-A", plt("PLT-A"), true},
		{"different plant denied", "PLT-A", plt("PLT-B"), false},
		{"area with no plant is shared/global", "PLT-A", nil, true},
		{"area with empty-string plant is shared/global", "PLT-A", plt(""), true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := PlantMatches(c.userPlantID, c.resourcePlantID); got != c.want {
				t.Fatalf("PlantMatches(%q, %v) = %v, want %v", c.userPlantID, c.resourcePlantID, got, c.want)
			}
		})
	}
}
