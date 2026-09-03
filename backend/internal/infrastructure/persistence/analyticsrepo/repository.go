// Package analyticsrepo assembles and executes the raw SQL for one resolved
// analyticsusecase.QueryPlan. This is the ONLY place semantic-layer SQL
// text gets built — every fragment it concatenates (CTEs, JoinSQL,
// SelectExpr/KeyExpr/LabelExpr) comes from the developer-authored catalog
// (backend/internal/domain/analytics), never from request input. The only
// request-derived values that reach the database are bound as `?`
// parameters (allowed area IDs, date range, current time, row limit) —
// exactly the same GORM Raw(query, args...) pattern already used by
// dashboard_handler.go's GetPICDetail/GetAuditorDetail.
package analyticsrepo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/monitoring-system/backend/internal/usecase/analyticsusecase"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Execute builds the final SQL for plan and runs it with db.WithContext(ctx)
// so a caller-cancelled context (e.g. React Query aborting a superseded
// request) actually cancels the in-flight DB statement, not just the HTTP
// response.
func (r *Repository) Execute(ctx context.Context, plan analyticsusecase.QueryPlan) ([]analyticsusecase.ResultRow, error) {
	sqlText, args := buildSQL(plan)

	rows, err := r.db.WithContext(ctx).Raw(sqlText, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	measureIDs := make([]string, len(plan.Measures))
	for i, m := range plan.Measures {
		measureIDs[i] = m.ID
	}

	hasDim2 := plan.Dimension2 != nil

	var out []analyticsusecase.ResultRow
	for rows.Next() {
		var key, category, key2, category2 *string
		var totalRows int
		values := make([]*float64, len(measureIDs))
		hierarchyKeys := make([]*string, len(plan.Hierarchy))
		hierarchyCategories := make([]*string, len(plan.Hierarchy))
		scanArgs := make([]interface{}, 0, 5+len(values)+len(plan.Hierarchy)*2)
		scanArgs = append(scanArgs, &key, &category)
		if hasDim2 {
			scanArgs = append(scanArgs, &key2, &category2)
		}
		for i := range plan.Hierarchy {
			scanArgs = append(scanArgs, &hierarchyKeys[i], &hierarchyCategories[i])
		}
		for i := range values {
			scanArgs = append(scanArgs, &values[i])
		}
		scanArgs = append(scanArgs, &totalRows)
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		row := analyticsusecase.ResultRow{Measures: make(map[string]*float64, len(measureIDs))}
		if key != nil {
			row.Key = *key
		}
		if category != nil {
			row.Category = *category
		}
		row.Key2 = key2
		row.Category2 = category2
		row.TotalRows = totalRows
		row.Hierarchy = make([]analyticsusecase.HierarchyValue, 0, len(plan.Hierarchy))
		for i, hierarchyDimension := range plan.Hierarchy {
			item := analyticsusecase.HierarchyValue{DimensionID: hierarchyDimension.ID}
			if hierarchyKeys[i] != nil {
				item.Key = *hierarchyKeys[i]
			}
			if hierarchyCategories[i] != nil {
				item.Category = *hierarchyCategories[i]
			}
			row.Hierarchy = append(row.Hierarchy, item)
		}
		for i, mID := range measureIDs {
			row.Measures[mID] = values[i]
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// buildSQL assembles the full WITH/SELECT/GROUP BY/ORDER BY/LIMIT statement
// for plan, returning the SQL text (with `?` placeholders, in the exact
// left-to-right order the returned args must be supplied) and the args
// themselves.
//
// scoped_issues is the mandatory security backbone: EVERY dimension/measure
// in the catalog reads from it (directly or via a join reached from it), so
// no measure/dimension combination can ever bypass the area-scope filter.
//
// Measure SelectExprs may contain the literal token "@now" (never a SQL
// `?`, since a measure can be selected 0..1 times but its own SelectExpr
// text may reference "@now" more than once — see temuan_overdue /
// rata_rata_keterlambatan_hari in the catalog). "@now" always appears
// strictly after the scoped_issues CTE's own `?` placeholders in the
// assembled text (it only ever appears inside the SELECT list, which is
// textually built after the WITH clause), and the LIMIT `?` is always the
// very last placeholder — so counting "@now" occurrences in the fully
// assembled string and replacing them with `?` left-to-right, then
// supplying [allowedAreas?, startDate?, endDate?, now×count, limit] as
// args, always lines up correctly. This is covered by
// query_service_test.go / repository tests exercising every measure that
// uses "@now".
func buildSQL(plan analyticsusecase.QueryPlan) (string, []interface{}) {
	var b strings.Builder
	var args []interface{}

	// NOTE: ih."DetailKawasanID" is deliberately NOT selected here — Issue
	// already has its own DetailKawasanID (picked up by i.*), and selecting
	// both produced two same-named columns in the scoped_issues CTE, which
	// Postgres allows at CTE-definition time but then rejects as "column
	// reference is ambiguous" the moment anything downstream reads
	// si."DetailKawasanID" (caught via the local manual test flow, not by
	// a Go unit test — no SQL-executing test harness exists in this repo).
	b.WriteString(`WITH scoped_issues AS (
		SELECT i.*, ih."KawasanID", ih."AreaID", ih."InspectorID", ih."InspectionHeaderCreatedAt", ir."UraianID"
		FROM "Issue" i
		JOIN "Inspection_Result" ir ON ir."ResultID" = i."ResultID"
		JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"
		WHERE 1=1`)
	if plan.AllowedAreas != nil {
		b.WriteString(` AND ih."AreaID" IN ?`)
		args = append(args, plan.AllowedAreas)
	}
	if plan.StartDate != "" {
		b.WriteString(` AND ih."InspectionHeaderCreatedAt" >= ?`)
		args = append(args, plan.StartDate)
	}
	if plan.EndDate != "" {
		b.WriteString(` AND ih."InspectionHeaderCreatedAt" < ?`)
		args = append(args, plan.EndDate)
	}
	b.WriteString(`
	)`)

	for _, j := range plan.Joins {
		if strings.TrimSpace(j.CTE) != "" {
			b.WriteString(", ")
			b.WriteString(j.CTE)
		}
	}

	b.WriteString(`
	SELECT `)
	fmt.Fprintf(&b, "%s AS \"cat_key\", %s AS \"cat_label\"", plan.Dimension.KeyExpr, plan.Dimension.LabelExpr)
	if plan.Dimension2 != nil {
		fmt.Fprintf(&b, ", %s AS \"cat_key2\", %s AS \"cat_label2\"", plan.Dimension2.KeyExpr, plan.Dimension2.LabelExpr)
	}
	for i, hierarchyDimension := range plan.Hierarchy {
		fmt.Fprintf(&b, ", %s AS \"hierarchy_key_%d\", %s AS \"hierarchy_label_%d\"", hierarchyDimension.KeyExpr, i, hierarchyDimension.LabelExpr, i)
	}
	for i, m := range plan.Measures {
		fmt.Fprintf(&b, ", (%s)::double precision AS \"m%d\"", m.SelectExpr, i)
	}
	b.WriteString(`, COUNT(*) OVER() AS "total_rows"`)

	b.WriteString(`
	FROM scoped_issues si`)
	for _, j := range plan.Joins {
		if strings.TrimSpace(j.JoinSQL) != "" {
			b.WriteString("\n\t")
			b.WriteString(j.JoinSQL)
		}
	}

	// Ad-hoc equality filters ("dimension = value"), applied at the outer
	// query level (not inside scoped_issues) because a filter dimension's
	// KeyExpr may reference a column only available after the LEFT JOINs
	// above (e.g. arc."AreaName", aspc."AspekID") — never inside the CTE.
	//
	// filterArgs is collected SEPARATELY from args and only merged in below,
	// AFTER the now×nowCount block. This matters: "@now" tokens (written
	// into the SELECT list above, converted to literal "?" only at the very
	// end of this function) are textually positioned BEFORE this WHERE
	// clause, so the args slice must place the now-values before the filter
	// values even though this WHERE clause is assembled after the SELECT
	// text. Appending filterArgs to `args` directly at write-time here would
	// put them in the wrong position relative to the not-yet-materialized
	// "@now" placeholders — the same class of "textual position vs. args
	// order" mistake this file's DetailKawasanID bug came from.
	var filterArgs []interface{}
	if len(plan.Filters) > 0 {
		b.WriteString("\n\tWHERE ")
		for i, f := range plan.Filters {
			if i > 0 {
				b.WriteString(" AND ")
			}
			fmt.Fprintf(&b, "(%s)::text = ?", f.Dimension.KeyExpr)
			filterArgs = append(filterArgs, f.Value)
		}
	}

	fmt.Fprintf(&b, "\n\tGROUP BY %s, %s", plan.Dimension.KeyExpr, plan.Dimension.LabelExpr)
	if plan.Dimension2 != nil {
		fmt.Fprintf(&b, ", %s, %s", plan.Dimension2.KeyExpr, plan.Dimension2.LabelExpr)
	}
	for _, hierarchyDimension := range plan.Hierarchy {
		fmt.Fprintf(&b, ", %s, %s", hierarchyDimension.KeyExpr, hierarchyDimension.LabelExpr)
	}
	b.WriteString("\n\tORDER BY ")
	for i := range plan.Hierarchy {
		fmt.Fprintf(&b, "\"hierarchy_label_%d\" ASC NULLS LAST, ", i)
	}
	b.WriteString(`"m0" DESC NULLS LAST, "cat_label" ASC
	LIMIT ?`)

	sqlText := b.String()
	nowCount := strings.Count(sqlText, "@now")
	sqlText = strings.ReplaceAll(sqlText, "@now", "?")

	// Reassemble args in the exact order their placeholders now appear:
	// [scoped_issues WHERE args already appended above] -> now×nowCount
	// (from the SELECT list) -> filterArgs (from the WHERE clause above,
	// textually AFTER the SELECT list) -> limit.
	now := time.Now()
	for i := 0; i < nowCount; i++ {
		args = append(args, now)
	}
	args = append(args, filterArgs...)
	args = append(args, plan.Limit)

	return sqlText, args
}
