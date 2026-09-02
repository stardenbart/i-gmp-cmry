// Package analyticsusecase implements the query-building/validation logic
// for the Custom KPI Visualization Builder's semantic layer. See
// backend/internal/domain/analytics for the catalog these functions operate
// over.
package analyticsusecase

import (
	"fmt"

	"github.com/monitoring-system/backend/internal/domain/analytics"
)

// ResolveJoins collects every JoinDef needed to satisfy `requested` (usually
// the RequiredJoins of one dimension plus every selected measure), resolving
// each entry's DependsOn closure, deduping by JoinID, and preserving
// dependency order (a join's dependencies always appear before it in the
// result — required for SQL that JOINs against an alias only a prior JOIN
// introduces, e.g. kawasan_chain's alias `kc` referencing columns
// inspection_chain guarantees are present on `si`).
//
// Pure function, no DB access — safe to unit test directly.
func ResolveJoins(requested []analytics.JoinID, catalog map[analytics.JoinID]analytics.JoinDef) ([]analytics.JoinDef, error) {
	visited := make(map[analytics.JoinID]bool)
	visiting := make(map[analytics.JoinID]bool)
	var ordered []analytics.JoinDef

	var visit func(id analytics.JoinID) error
	visit = func(id analytics.JoinID) error {
		if visited[id] {
			return nil
		}
		if visiting[id] {
			return fmt.Errorf("circular join dependency: %s", id)
		}
		def, ok := catalog[id]
		if !ok {
			return fmt.Errorf("unknown join: %s", id)
		}
		visiting[id] = true
		for _, dep := range def.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[id] = false
		visited[id] = true
		ordered = append(ordered, def)
		return nil
	}

	for _, id := range requested {
		if err := visit(id); err != nil {
			return nil, err
		}
	}

	return ordered, nil
}
