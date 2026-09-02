// Package analytics defines the semantic layer catalog for the Custom KPI
// Visualization Builder: a developer-curated set of Measures/Dimensions/
// Joins that non-technical users combine via drag-and-drop (never SQL,
// never free-form formulas). See /kpi's "Tambah Visualisasi" builder.
//
// Design contract (do not weaken without re-reading the approved plan):
//   - Every value here is a developer-authored constant. SelectExpr/KeyExpr/
//     LabelExpr/JoinSQL/CTE are NEVER built from request input — only the
//     IDs (map keys) travel over the wire, and only IDs that exist in the
//     catalog below are ever interpolated into SQL text.
//   - Grain contract: 1 row = 1 Issue. Every JoinDef must therefore
//     guarantee at most one matched row per Issue (pre-aggregate 1:N
//     relations into a CTE before joining) UNLESS FansOut is explicitly
//     true — the one documented exception is `pic_fanout`/`pic`, which
//     intentionally lets a single Issue be counted under more than one
//     PIC row (mirrors the pre-existing GetPICDetail pic_issues CTE in
//     dashboard_handler.go).
package analytics

// JoinID identifies one JoinDef in the catalog.
type JoinID string

const (
	JoinInspectionChain JoinID = "inspection_chain"
	JoinKawasanChain    JoinID = "kawasan_chain"
	JoinAuditor         JoinID = "auditor_join"
	JoinPICFanout       JoinID = "pic_fanout"
	JoinAreaChain       JoinID = "area_chain"
	JoinAspekChain      JoinID = "aspek_chain"
	JoinHEIChain        JoinID = "hei_chain"
)

// ValueType classifies a measure's numeric meaning, driving both display
// formatting and chart-compatibility rules on the frontend.
type ValueType string

const (
	ValueTypeCount   ValueType = "count"
	ValueTypePercent ValueType = "percent" // convention: 0-100, not 0-1 (see compliance_rate elsewhere in the app)
	ValueTypeDays    ValueType = "days"
	ValueTypeFloat   ValueType = "float"
)

// FormatMeta is neutral, engine-agnostic display metadata — never a Go/printf
// format string — so the frontend can render it with Intl.NumberFormat
// without knowing anything about the backend's language or SQL.
type FormatMeta struct {
	Style    string `json:"style"` // "number" | "percent" | "duration" | "currency"
	Decimals int    `json:"decimals"`
	Suffix   string `json:"suffix,omitempty"` // e.g. "hari"
	Unit     string `json:"unit,omitempty"`
}

// DimensionKind distinguishes categorical (bar/pie friendly) from temporal
// (line-chart friendly) dimensions for the frontend's chart-compatibility
// validation.
type DimensionKind string

const (
	KindCategorical DimensionKind = "categorical"
	KindTemporal    DimensionKind = "temporal"
)

// chartsAll/chartsRatio are the two CompatibleCharts sets every measure
// below picks from — see the frontend's VisualizationBuilder for what each
// chart type name means. Kept as shared vars (not re-typed literals per
// measure) so the whole catalog's chart-availability rules move together
// when a chart type is added/removed.
//
// "scatter" is deliberately absent from both lists: its eligibility
// depends on the PAIR of measures selected (exactly 2), not on any single
// measure's own nature, so the frontend adds it back explicitly once 2
// measures are picked (see VisualizationBuilder.tsx's compatibleCharts).
//
// chartsAll: measures whose value is a plain count/absolute number — every
// chart family (including "part of a whole": pie/donut/treemap/funnel)
// makes sense.
var chartsAll = []string{
	"bar", "horizontal_bar", "stacked_bar", "line", "area", "radar",
	"pie", "donut", "treemap", "funnel", "number_card", "gauge",
	"heatmap", "sankey", "table",
}

// chartsRatio: measures whose value is a rate/average/percent — the
// "part of a whole" family is dropped (a percentage or an average isn't a
// slice that sums to a meaningful 100% with other rows). gauge/heatmap/
// sankey stay: gauge especially suits a percent measure (0-100 scale is
// exact, not a heuristic — see buildEChartsOption's computeNiceMax).
var chartsRatio = []string{
	"bar", "horizontal_bar", "stacked_bar", "line", "area", "radar",
	"number_card", "gauge", "heatmap", "sankey", "table",
}

// JoinDef is one fragment of JOIN/CTE SQL, hardcoded by developers. Requested
// joins are resolved (with their DependsOn closure) by
// analyticsusecase.ResolveJoins before any SQL is assembled.
type JoinDef struct {
	ID JoinID
	// CTE is an optional `name AS (...)` fragment prepended to the query's
	// WITH clause (without the leading "WITH" / trailing comma — the
	// assembler joins these together). Empty when the join needs no CTE.
	CTE string
	// JoinSQL is the `JOIN ... ON ...` fragment appended after FROM
	// scoped_issues si. Always joins against si, an already-resolved CTE,
	// or a static master table — never against raw request input.
	JoinSQL string
	// DependsOn lists JoinIDs that must appear (and be positioned) before
	// this one in the final SQL (e.g. kawasan_chain depends on
	// inspection_chain for the `ih` alias it joins against).
	DependsOn []JoinID
	// FansOut marks a join that intentionally may match more than one row
	// per Issue. Must be true ONLY for JoinPICFanout. Any other join
	// violating the 1-row-per-Issue grain is a bug, not a feature.
	FansOut bool
}

// MeasureDef is one pre-built, safe SQL aggregate expression. SelectExpr is
// always a per-measure CASE/FILTER against scoped_issues (or a join reached
// from it) so that combining multiple measures in one query never changes
// any individual measure's result (no shared global WHERE).
type MeasureDef struct {
	ID          string
	Label       string
	Description string // business-language sentence, shown to non-technical users
	// SelectExpr is a full aggregate SQL expression, e.g.
	// `COUNT(*) FILTER (WHERE ...)`. Constant, developer-authored.
	SelectExpr       string
	RequiredJoins    []JoinID
	ValueType        ValueType
	Format           FormatMeta
	CompatibleCharts []string // "bar" | "line" | "pie" | "table"
}

// DimensionDef is one pre-built, safe SQL GROUP BY key + display label.
type DimensionDef struct {
	ID          string
	Label       string
	Description string
	// KeyExpr is the stable grouping key (GROUP BY / ORDER BY tie-break on
	// this), e.g. a UUID or a sortable "YYYY-MM" string.
	KeyExpr string
	// LabelExpr is the human-friendly display value for the same group
	// (may differ from KeyExpr, e.g. periode's sortable key vs. a
	// human month name).
	LabelExpr     string
	RequiredJoins []JoinID
	Kind          DimensionKind
	// FansOut is true only for "pic" — see JoinDef.FansOut docs above.
	FansOut bool
}

// Catalog is the full, immutable set of joins/dimensions/measures the query
// engine and the /analytics/catalog endpoint are allowed to expose. Adding a
// new field here is the ONLY way to make it available to the builder —
// there is deliberately no path from request input to new SQL text.
type Catalog struct {
	Joins      map[JoinID]JoinDef
	Dimensions map[string]DimensionDef
	Measures   map[string]MeasureDef
}

// DefaultCatalog returns the V1 semantic layer catalog.
//
// scoped_issues (built by the query service, not here — see
// analyticsusecase.query_service.go) is always:
//
//	WITH scoped_issues AS (
//	  SELECT i.*, ih."KawasanID", ih."AreaID",
//	         ih."InspectorID", ih."InspectionHeaderCreatedAt"
//	  FROM "Issue" i
//	  JOIN "Inspection_Result" ir ON ir."ResultID" = i."ResultID"
//	  JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"
//	  WHERE ih."AreaID" = ANY(@allowed_area_ids) [AND date range]
//	)
//
// every join below is written to join against the `si` alias for that CTE.
func DefaultCatalog() Catalog {
	joins := map[JoinID]JoinDef{
		JoinInspectionChain: {
			ID: JoinInspectionChain,
			// scoped_issues already carries KawasanID/AreaID/InspectorID/
			// InspectionHeaderCreatedAt (from Inspection_Header) plus
			// DetailKawasanID (from Issue itself, via i.* — deliberately
			// NOT re-selected from Inspection_Header too, since Postgres
			// only rejects the resulting same-named-column ambiguity when
			// something downstream reads si."DetailKawasanID", not at CTE
			// definition time — see analyticsrepo/repository.go's buildSQL
			// comment). This join needs no SQL of its own — it exists as
			// an explicit dependency anchor other joins declare DependsOn
			// against, documenting that they rely on those columns being
			// present on si.
			JoinSQL: "",
		},
		JoinKawasanChain: {
			ID: JoinKawasanChain,
			CTE: `kawasan_chain AS (
				SELECT k."KawasanID", k."KawasanName", dk."DetailKawasanID", dk."DetailKawasanName"
				FROM "Kawasan_Master" k
				LEFT JOIN "DetailKawasan_Master" dk ON dk."KawasanID" = k."KawasanID"
			)`,
			JoinSQL:   `LEFT JOIN kawasan_chain kc ON kc."KawasanID" = si."KawasanID" AND kc."DetailKawasanID" = si."DetailKawasanID"`,
			DependsOn: []JoinID{JoinInspectionChain},
		},
		JoinAuditor: {
			ID:        JoinAuditor,
			JoinSQL:   `LEFT JOIN "Users" au ON au."UserID" = si."InspectorID"`,
			DependsOn: []JoinID{JoinInspectionChain},
		},
		JoinPICFanout: {
			ID: JoinPICFanout,
			// Deliberately scoped to scoped_issues (si), not raw "Issue" —
			// mirrors GetPICDetail's pic_issues CTE (dashboard_handler.go)
			// but restricted to only the issues this caller is allowed to
			// see, and the ONLY join in this catalog allowed to fan out.
			CTE: `pic_fanout AS (
				SELECT si."IssueID", si."IssuePICUserID" AS "UserID" FROM scoped_issues si
				UNION
				SELECT d."IssueID", d."DelegateUserID" AS "UserID"
				FROM "Issue_Delegate" d
				JOIN scoped_issues si2 ON si2."IssueID" = d."IssueID"
				UNION
				SELECT si3."IssueID", pm."UserID"
				FROM scoped_issues si3
				JOIN "PIC_Mapping" pm ON pm."KawasanID" = si3."KawasanID"
			)`,
			JoinSQL: `JOIN pic_fanout pf ON pf."IssueID" = si."IssueID"
				JOIN "Users" pu ON pu."UserID" = pf."UserID"`,
			FansOut: true,
		},
		JoinAreaChain: {
			ID: JoinAreaChain,
			// Area_Master.AreaID is a PK and Plant_Master.PlantID is a PK,
			// both LEFT JOINs — always exactly 0 or 1 match per si row,
			// never fans out.
			CTE: `area_chain AS (
				SELECT ar."AreaID", ar."AreaName", ar."PlantID", pl."PlantName"
				FROM "Area_Master" ar
				LEFT JOIN "Plant_Master" pl ON pl."PlantID" = ar."PlantID"
			)`,
			JoinSQL:   `LEFT JOIN area_chain arc ON arc."AreaID" = si."AreaID"`,
			DependsOn: []JoinID{JoinInspectionChain},
		},
		JoinAspekChain: {
			ID: JoinAspekChain,
			// Issue -> Inspection_Result.UraianID -> Uraian_Master ->
			// Detail_Master -> Aspek_Master is 1:1 all the way down (every
			// FK here is a PK on the far side), so this never fans out.
			CTE: `aspek_chain AS (
				SELECT u."UraianID", u."UraianText", a."AspekID", a."AspekName"
				FROM "Uraian_Master" u
				JOIN "Detail_Master" d ON d."DetailID" = u."DetailID"
				JOIN "Aspek_Master" a ON a."AspekID" = d."AspekID"
			)`,
			JoinSQL:   `LEFT JOIN aspek_chain aspc ON aspc."UraianID" = si."UraianID"`,
			DependsOn: []JoinID{JoinInspectionChain},
		},
		JoinHEIChain: {
			ID: JoinHEIChain,
			// Issue_HEI.IssueID is UNIQUE (issue.go's IssueHEI struct) and
			// HEI_Master.HEIID is a PK — 1:1, never fans out. Joined
			// straight off si (not off inspection_chain) since Issue_HEI
			// references Issue directly.
			CTE: `hei_chain AS (
				SELECT ihei."IssueID", h."HEIID", h."CategoryName", h."HEIName"
				FROM "Issue_HEI" ihei
				JOIN "HEI_Master" h ON h."HEIID" = ihei."HEIID"
			)`,
			JoinSQL: `LEFT JOIN hei_chain hc ON hc."IssueID" = si."IssueID"`,
		},
	}

	dimensions := map[string]DimensionDef{
		"kawasan": {
			ID:            "kawasan",
			Label:         "Kawasan",
			Description:   "Kelompokkan berdasarkan Kawasan.",
			KeyExpr:       `si."KawasanID"`,
			LabelExpr:     `kc."KawasanName"`,
			RequiredJoins: []JoinID{JoinKawasanChain},
			Kind:          KindCategorical,
		},
		"detail_kawasan": {
			ID:            "detail_kawasan",
			Label:         "Detail Kawasan",
			Description:   "Kelompokkan berdasarkan Detail Kawasan.",
			KeyExpr:       `si."DetailKawasanID"`,
			LabelExpr:     `kc."DetailKawasanName"`,
			RequiredJoins: []JoinID{JoinKawasanChain},
			Kind:          KindCategorical,
		},
		"area": {
			ID:            "area",
			Label:         "Area",
			Description:   "Kelompokkan berdasarkan Area.",
			KeyExpr:       `si."AreaID"`,
			LabelExpr:     `COALESCE(arc."AreaName", si."AreaID")`,
			RequiredJoins: []JoinID{JoinAreaChain},
			Kind:          KindCategorical,
		},
		"plant": {
			ID:            "plant",
			Label:         "Plant",
			Description:   "Kelompokkan berdasarkan Plant (lintas Area).",
			KeyExpr:       `COALESCE(arc."PlantID", 'GLOBAL')`,
			LabelExpr:     `COALESCE(arc."PlantName", 'Global')`,
			RequiredJoins: []JoinID{JoinAreaChain},
			Kind:          KindCategorical,
		},
		"auditor": {
			ID:            "auditor",
			Label:         "Auditor",
			Description:   "Kelompokkan berdasarkan auditor yang menjalankan inspeksi.",
			KeyExpr:       `au."UserID"`,
			LabelExpr:     `au."FullName"`,
			RequiredJoins: []JoinID{JoinAuditor},
			Kind:          KindCategorical,
		},
		"status_temuan": {
			ID:            "status_temuan",
			Label:         "Status Temuan",
			Description:   "Kelompokkan berdasarkan status temuan (Open, InProgress, Closed, Verified, PendingValidation).",
			KeyExpr:       `si."IssueStatus"`,
			LabelExpr:     `si."IssueStatus"`,
			RequiredJoins: []JoinID{JoinInspectionChain},
			Kind:          KindCategorical,
		},
		"periode": {
			ID:            "periode",
			Label:         "Periode",
			Description:   "Kelompokkan berdasarkan bulan inspeksi dibuat.",
			KeyExpr:       `TO_CHAR(si."InspectionHeaderCreatedAt", 'YYYY-MM')`,
			LabelExpr:     `TO_CHAR(si."InspectionHeaderCreatedAt", 'TMMonth YYYY')`,
			RequiredJoins: []JoinID{JoinInspectionChain},
			Kind:          KindTemporal,
		},
		"pic": {
			ID:    "pic",
			Label: "PIC",
			Description: "Kelompokkan berdasarkan PIC (penanggung jawab tindak lanjut). " +
				"Satu temuan bisa dihitung di lebih dari satu PIC kalau kawasannya punya " +
				"lebih dari satu penanggung jawab — total di sini bisa lebih besar dari Total Temuan.",
			KeyExpr:       `pu."UserID"`,
			LabelExpr:     `pu."FullName"`,
			RequiredJoins: []JoinID{JoinPICFanout},
			Kind:          KindCategorical,
			FansOut:       true,
		},
		"aspek": {
			ID:            "aspek",
			Label:         "Aspek",
			Description:   "Kelompokkan berdasarkan Aspek checklist yang diperiksa.",
			KeyExpr:       `aspc."AspekID"`,
			LabelExpr:     `aspc."AspekName"`,
			RequiredJoins: []JoinID{JoinAspekChain},
			Kind:          KindCategorical,
		},
		"uraian": {
			ID:            "uraian",
			Label:         "Uraian Checklist",
			Description:   "Kelompokkan berdasarkan butir checklist (Uraian) spesifik yang menghasilkan temuan. Cocok untuk tabel — kategorinya bisa sangat banyak.",
			KeyExpr:       `aspc."UraianID"`,
			LabelExpr:     `aspc."UraianText"`,
			RequiredJoins: []JoinID{JoinAspekChain},
			Kind:          KindCategorical,
		},
		"hei_kategori": {
			ID:            "hei_kategori",
			Label:         "Kategori HEI",
			Description:   "Kelompokkan berdasarkan kategori HEI (Human Error Investigation) yang terkait, jika ada.",
			KeyExpr:       `COALESCE(hc."CategoryName", 'Tidak Ada HEI')`,
			LabelExpr:     `COALESCE(hc."CategoryName", 'Tidak Ada HEI')`,
			RequiredJoins: []JoinID{JoinHEIChain},
			Kind:          KindCategorical,
		},
		"hei": {
			ID:            "hei",
			Label:         "HEI",
			Description:   "Kelompokkan berdasarkan HEI spesifik yang terkait, jika ada.",
			KeyExpr:       `COALESCE(hc."HEIID", 'NONE')`,
			LabelExpr:     `COALESCE(hc."HEIName", 'Tidak Ada HEI')`,
			RequiredJoins: []JoinID{JoinHEIChain},
			Kind:          KindCategorical,
		},
	}

	measures := map[string]MeasureDef{
		"total_temuan": {
			ID:               "total_temuan",
			Label:            "Total Temuan",
			Description:      "Jumlah seluruh temuan.",
			SelectExpr:       `COUNT(*)`,
			ValueType:        ValueTypeCount,
			Format:           FormatMeta{Style: "number", Decimals: 0},
			CompatibleCharts: chartsAll,
		},
		"temuan_terbuka": {
			ID:               "temuan_terbuka",
			Label:            "Temuan Terbuka",
			Description:      "Jumlah temuan yang belum Closed/Verified.",
			SelectExpr:       `COUNT(*) FILTER (WHERE si."IssueStatus" NOT IN ('Closed','Verified'))`,
			ValueType:        ValueTypeCount,
			Format:           FormatMeta{Style: "number", Decimals: 0},
			CompatibleCharts: chartsAll,
		},
		"temuan_overdue": {
			ID:               "temuan_overdue",
			Label:            "Temuan Overdue",
			Description:      "Jumlah temuan yang sudah melewati batas waktu (due date) dan belum selesai.",
			SelectExpr:       `COUNT(*) FILTER (WHERE si."DueDate" IS NOT NULL AND si."DueDate" < @now AND si."IssueStatus" NOT IN ('Closed','Verified'))`,
			ValueType:        ValueTypeCount,
			Format:           FormatMeta{Style: "number", Decimals: 0},
			CompatibleCharts: chartsAll,
		},
		"rata_rata_keterlambatan_hari": {
			ID:          "rata_rata_keterlambatan_hari",
			Label:       "Rata-rata Keterlambatan (hari)",
			Description: "Rata-rata jumlah hari keterlambatan dari temuan yang overdue.",
			SelectExpr: `AVG(CASE WHEN si."DueDate" IS NOT NULL AND si."DueDate" < @now AND si."IssueStatus" NOT IN ('Closed','Verified')
				THEN EXTRACT(EPOCH FROM (@now::timestamptz - si."DueDate")) / 86400.0 END)`,
			ValueType:        ValueTypeDays,
			Format:           FormatMeta{Style: "duration", Decimals: 1, Suffix: "hari"},
			CompatibleCharts: chartsRatio,
		},
		"persentase_temuan_overdue": {
			ID:          "persentase_temuan_overdue",
			Label:       "Persentase Temuan Overdue",
			Description: "Persentase temuan overdue dibandingkan total temuan (siap pakai, tanpa perlu menyusun rumus manual).",
			SelectExpr: `(COUNT(*) FILTER (WHERE si."DueDate" IS NOT NULL AND si."DueDate" < @now AND si."IssueStatus" NOT IN ('Closed','Verified')) * 100.0)
				/ NULLIF(COUNT(*), 0)`,
			ValueType:        ValueTypePercent,
			Format:           FormatMeta{Style: "percent", Decimals: 1, Suffix: "%"},
			CompatibleCharts: chartsRatio,
		},
		"wowr_total": {
			ID:               "wowr_total",
			Label:            "Total WO/WR",
			Description:      "Jumlah temuan yang membutuhkan WO/WR.",
			SelectExpr:       `COUNT(*) FILTER (WHERE si."NeedsWOWR" = true)`,
			ValueType:        ValueTypeCount,
			Format:           FormatMeta{Style: "number", Decimals: 0},
			CompatibleCharts: chartsAll,
		},
		"wowr_verified": {
			ID:               "wowr_verified",
			Label:            "WO/WR Verified",
			Description:      "Jumlah WO/WR yang sudah diverifikasi.",
			SelectExpr:       `COUNT(*) FILTER (WHERE si."NeedsWOWR" = true AND si."WOWRStatus" = 'Verified')`,
			ValueType:        ValueTypeCount,
			Format:           FormatMeta{Style: "number", Decimals: 0},
			CompatibleCharts: chartsAll,
		},
		"wowr_pending": {
			ID:               "wowr_pending",
			Label:            "WO/WR Pending",
			Description:      "Jumlah WO/WR yang menunggu validasi.",
			SelectExpr:       `COUNT(*) FILTER (WHERE si."NeedsWOWR" = true AND si."WOWRStatus" = 'PendingValidation')`,
			ValueType:        ValueTypeCount,
			Format:           FormatMeta{Style: "number", Decimals: 0},
			CompatibleCharts: chartsAll,
		},
		"wowr_rejected": {
			ID:               "wowr_rejected",
			Label:            "WO/WR Rejected",
			Description:      "Jumlah WO/WR yang ditolak.",
			SelectExpr:       `COUNT(*) FILTER (WHERE si."NeedsWOWR" = true AND si."WOWRStatus" = 'Rejected')`,
			ValueType:        ValueTypeCount,
			Format:           FormatMeta{Style: "number", Decimals: 0},
			CompatibleCharts: chartsAll,
		},
		"wowr_awaiting": {
			ID:               "wowr_awaiting",
			Label:            "WO/WR Belum Bukti",
			Description:      "Jumlah temuan yang butuh WO/WR tapi buktinya belum diunggah sama sekali.",
			SelectExpr:       `COUNT(*) FILTER (WHERE si."NeedsWOWR" = true AND si."WOWRStatus" = 'None')`,
			ValueType:        ValueTypeCount,
			Format:           FormatMeta{Style: "number", Decimals: 0},
			CompatibleCharts: chartsAll,
		},
		"persentase_wowr_verified": {
			ID:          "persentase_wowr_verified",
			Label:       "Persentase WO/WR Verified",
			Description: "Persentase WO/WR yang sudah diverifikasi, dari seluruh temuan yang butuh WO/WR.",
			SelectExpr: `(COUNT(*) FILTER (WHERE si."NeedsWOWR" = true AND si."WOWRStatus" = 'Verified') * 100.0)
				/ NULLIF(COUNT(*) FILTER (WHERE si."NeedsWOWR" = true), 0)`,
			ValueType:        ValueTypePercent,
			Format:           FormatMeta{Style: "percent", Decimals: 1, Suffix: "%"},
			CompatibleCharts: chartsRatio,
		},
		"temuan_closed": {
			ID:               "temuan_closed",
			Label:            "Temuan Closed",
			Description:      "Jumlah temuan berstatus Closed.",
			SelectExpr:       `COUNT(*) FILTER (WHERE si."IssueStatus" = 'Closed')`,
			ValueType:        ValueTypeCount,
			Format:           FormatMeta{Style: "number", Decimals: 0},
			CompatibleCharts: chartsAll,
		},
		"temuan_verified": {
			ID:               "temuan_verified",
			Label:            "Temuan Verified",
			Description:      "Jumlah temuan berstatus Verified.",
			SelectExpr:       `COUNT(*) FILTER (WHERE si."IssueStatus" = 'Verified')`,
			ValueType:        ValueTypeCount,
			Format:           FormatMeta{Style: "number", Decimals: 0},
			CompatibleCharts: chartsAll,
		},
		"temuan_in_progress": {
			ID:               "temuan_in_progress",
			Label:            "Temuan In Progress",
			Description:      "Jumlah temuan berstatus InProgress.",
			SelectExpr:       `COUNT(*) FILTER (WHERE si."IssueStatus" = 'InProgress')`,
			ValueType:        ValueTypeCount,
			Format:           FormatMeta{Style: "number", Decimals: 0},
			CompatibleCharts: chartsAll,
		},
		"temuan_pending_validation": {
			ID:               "temuan_pending_validation",
			Label:            "Temuan Pending Validation",
			Description:      "Jumlah temuan berstatus PendingValidation.",
			SelectExpr:       `COUNT(*) FILTER (WHERE si."IssueStatus" = 'PendingValidation')`,
			ValueType:        ValueTypeCount,
			Format:           FormatMeta{Style: "number", Decimals: 0},
			CompatibleCharts: chartsAll,
		},
		"persentase_temuan_selesai": {
			ID:          "persentase_temuan_selesai",
			Label:       "Persentase Temuan Selesai",
			Description: "Persentase temuan yang sudah Closed atau Verified, dari total temuan.",
			SelectExpr: `(COUNT(*) FILTER (WHERE si."IssueStatus" IN ('Closed','Verified')) * 100.0)
				/ NULLIF(COUNT(*), 0)`,
			ValueType:        ValueTypePercent,
			Format:           FormatMeta{Style: "percent", Decimals: 1, Suffix: "%"},
			CompatibleCharts: chartsRatio,
		},
		"rata_rata_keterlambatan_penyelesaian_hari": {
			ID:    "rata_rata_keterlambatan_penyelesaian_hari",
			Label: "Rata-rata Keterlambatan Penyelesaian (hari)",
			Description: "Rata-rata jumlah hari keterlambatan tindak lanjut dihitung dari FollowUpDelay saat " +
				"temuan ditutup/diverifikasi (0 berarti tidak terlambat).",
			SelectExpr:       `AVG(si."FollowUpDelay") FILTER (WHERE si."FollowUpDelay" IS NOT NULL)`,
			ValueType:        ValueTypeDays,
			Format:           FormatMeta{Style: "duration", Decimals: 1, Suffix: "hari"},
			CompatibleCharts: chartsRatio,
		},
	}

	return Catalog{Joins: joins, Dimensions: dimensions, Measures: measures}
}
