// Package authz holds small, dependency-free authorization helpers shared
// across usecases that need to check per-record access, as opposed to the
// module-level permission checks in internal/middleware (which only answer
// "can this role use this feature at all", never "does this specific
// record belong to a plant/person this caller may see").
package authz

import "strings"

// PlantMatches reports whether a resource whose owning area has
// resourcePlantID may be seen by a caller scoped to userPlantID.
//
// userPlantID == "" means the caller is either a Super Admin or an
// unassigned "global scope" user (see middleware.PlantScopeMiddleware) —
// both are unrestricted, exactly like every existing plant-filtered list
// query in this codebase (e.g. inspectionrepo.FindAll only adds its WHERE
// clause `if plantID != ""`). A resourcePlantID that is nil/empty means the
// owning area itself isn't tied to any one plant — shared/global master
// data, visible regardless of the caller's plant — matching that same
// query's `"PlantID" = ? OR "PlantID" IS NULL OR "PlantID" = <empty>`
// clause.
func PlantMatches(userPlantID string, resourcePlantID *string) bool {
	if userPlantID == "" {
		return true
	}
	if resourcePlantID == nil || strings.TrimSpace(*resourcePlantID) == "" {
		return true
	}
	return *resourcePlantID == userPlantID
}
