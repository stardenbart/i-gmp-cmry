package dashboardhandler

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/response"
	"gorm.io/gorm"
)

type trendGranularity string

const (
	trendDay     trendGranularity = "day"
	trendWeek    trendGranularity = "week"
	trendMonth   trendGranularity = "month"
	trendQuarter trendGranularity = "quarter"
)

type trendPeriodSpec struct {
	Period      string
	Label       string
	Granularity trendGranularity
	Start       time.Time
	End         time.Time
	Buckets     []trendBucket
}

type trendBucket struct {
	Start time.Time
	End   time.Time
	Key   string
	Label string
}

type trendPoint struct {
	BucketStart      string  `json:"bucket_start"`
	BucketEnd        string  `json:"bucket_end"`
	Label            string  `json:"label"`
	TotalInspections int64   `json:"total_inspections"`
	TotalIssues      int64   `json:"total_issues"`
	ComplianceRate   float64 `json:"compliance_rate"`
}

type inspectionTrendAggregate struct {
	BucketKey        string `gorm:"column:bucket_key"`
	TotalInspections int64  `gorm:"column:total_inspections"`
	TotalChecks      int64  `gorm:"column:total_checks"`
	TotalOK          int64  `gorm:"column:total_ok"`
}

type issueTrendAggregate struct {
	BucketKey   string `gorm:"column:bucket_key"`
	TotalIssues int64  `gorm:"column:total_issues"`
}

var trendPeriodAliases = map[string]string{
	"1m": "weekly", "3m": "monthly", "6m": "monthly", "1y": "monthly", "12m": "monthly",
}

func dashboardLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*60*60)
	}
	return location
}

func startOfDay(value time.Time, location *time.Location) time.Time {
	localized := value.In(location)
	return time.Date(localized.Year(), localized.Month(), localized.Day(), 0, 0, 0, 0, location)
}

func startOfWeek(value time.Time, location *time.Location) time.Time {
	day := startOfDay(value, location)
	daysSinceMonday := (int(day.Weekday()) + 6) % 7
	return day.AddDate(0, 0, -daysSinceMonday)
}

func startOfQuarter(value time.Time, location *time.Location) time.Time {
	localized := value.In(location)
	month := time.Month(((int(localized.Month()) - 1) / 3 * 3) + 1)
	return time.Date(localized.Year(), month, 1, 0, 0, 0, 0, location)
}

func buildTrendPeriod(period string, now time.Time, customStart, customEnd string) (trendPeriodSpec, error) {
	location := dashboardLocation()
	normalized := strings.ToLower(strings.TrimSpace(period))
	if normalized == "" {
		normalized = "monthly"
	}
	if alias, ok := trendPeriodAliases[normalized]; ok {
		normalized = alias
	}

	today := startOfDay(now, location)
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, location)
	var spec trendPeriodSpec
	spec.Period = normalized

	switch normalized {
	case "daily":
		spec.Label, spec.Granularity = "14 Hari Terakhir", trendDay
		spec.Start, spec.End = today.AddDate(0, 0, -13), today.AddDate(0, 0, 1)
	case "weekly":
		spec.Label, spec.Granularity = "12 Minggu Terakhir", trendWeek
		currentWeek := startOfWeek(today, location)
		spec.Start, spec.End = currentWeek.AddDate(0, 0, -11*7), currentWeek.AddDate(0, 0, 7)
	case "monthly":
		spec.Label, spec.Granularity = "12 Bulan Terakhir", trendMonth
		spec.Start, spec.End = monthStart.AddDate(0, -11, 0), monthStart.AddDate(0, 1, 0)
	case "quarter":
		spec.Label, spec.Granularity = "8 Kuartal Terakhir", trendQuarter
		currentQuarter := startOfQuarter(today, location)
		spec.Start, spec.End = currentQuarter.AddDate(0, -21, 0), currentQuarter.AddDate(0, 3, 0)
	case "custom":
		// User-picked date range (replaces the old "previous_week/month/year"
		// single-past-period presets) — granularity auto-scales so the chart
		// stays readable whether they pick a week or a couple of years.
		start, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(customStart), location)
		if err != nil {
			return trendPeriodSpec{}, errors.New("start_date wajib diisi (format YYYY-MM-DD) untuk periode custom")
		}
		endDay, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(customEnd), location)
		if err != nil {
			return trendPeriodSpec{}, errors.New("end_date wajib diisi (format YYYY-MM-DD) untuk periode custom")
		}
		start = startOfDay(start, location)
		endExclusive := startOfDay(endDay, location).AddDate(0, 0, 1)
		if !endExclusive.After(start) {
			return trendPeriodSpec{}, errors.New("end_date harus setelah start_date")
		}

		totalDays := int(endExclusive.Sub(start).Hours() / 24)
		switch {
		case totalDays <= 31:
			spec.Granularity = trendDay
		case totalDays <= 180:
			spec.Granularity = trendWeek
		default:
			spec.Granularity = trendMonth
		}
		spec.Label = fmt.Sprintf("%s – %s", start.Format("2 Jan 2006"), endDay.Format("2 Jan 2006"))
		spec.Start, spec.End = start, endExclusive
	default:
		return trendPeriodSpec{}, errors.New("periode tren tidak valid")
	}

	for cursor := spec.Start; cursor.Before(spec.End); {
		next := nextTrendBucket(cursor, spec.Granularity)
		if next.After(spec.End) {
			next = spec.End
		}
		spec.Buckets = append(spec.Buckets, trendBucket{
			Start: cursor,
			End:   next,
			Key:   trendBucketKey(cursor, spec.Granularity),
			Label: trendBucketLabel(cursor, next, spec.Granularity),
		})
		cursor = next
	}
	return spec, nil
}

func nextTrendBucket(value time.Time, granularity trendGranularity) time.Time {
	switch granularity {
	case trendDay:
		return value.AddDate(0, 0, 1)
	case trendWeek:
		return value.AddDate(0, 0, 7)
	case trendMonth:
		return value.AddDate(0, 1, 0)
	default:
		return value.AddDate(0, 3, 0)
	}
}

func trendBucketKey(value time.Time, granularity trendGranularity) string {
	switch granularity {
	case trendDay:
		return value.Format("2006-01-02")
	case trendWeek:
		year, week := value.ISOWeek()
		return fmt.Sprintf("%04d-%02d", year, week)
	case trendMonth:
		return value.Format("2006-01")
	default:
		return fmt.Sprintf("%04d-Q%d", value.Year(), (int(value.Month())-1)/3+1)
	}
}

var indonesianMonths = [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
var indonesianDays = [...]string{"Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"}

func trendBucketLabel(start, end time.Time, granularity trendGranularity) string {
	switch granularity {
	case trendDay:
		return fmt.Sprintf("%s, %d %s", indonesianDays[start.Weekday()], start.Day(), indonesianMonths[start.Month()-1])
	case trendWeek:
		lastDay := end.AddDate(0, 0, -1)
		return fmt.Sprintf("%d %s–%d %s", start.Day(), indonesianMonths[start.Month()-1], lastDay.Day(), indonesianMonths[lastDay.Month()-1])
	case trendMonth:
		return fmt.Sprintf("%s %d", indonesianMonths[start.Month()-1], start.Year())
	default:
		return fmt.Sprintf("Q%d %d", (int(start.Month())-1)/3+1, start.Year())
	}
}

func trendBucketSQL(column string, granularity trendGranularity) string {
	switch granularity {
	case trendDay:
		return fmt.Sprintf(`TO_CHAR(%s, 'YYYY-MM-DD')`, column)
	case trendWeek:
		return fmt.Sprintf(`TO_CHAR(%s, 'IYYY-IW')`, column)
	case trendMonth:
		return fmt.Sprintf(`TO_CHAR(%s, 'YYYY-MM')`, column)
	default:
		return fmt.Sprintf(`TO_CHAR(%s, 'YYYY') || '-Q' || EXTRACT(QUARTER FROM %s)::int`, column, column)
	}
}

func applyTrendScope(query *gorm.DB, allowedAreas []string, areaID, inspectorID string) *gorm.DB {
	if allowedAreas != nil {
		query = query.Where(`ih."AreaID" IN ?`, allowedAreas)
	}
	if areaID != "" {
		query = query.Where(`ih."AreaID" = ?`, areaID)
	}
	if inspectorID != "" {
		query = query.Where(`ih."InspectorID" = ?`, inspectorID)
	}
	return query
}

func resolveTrendInspector(c *fiber.Ctx) (string, error) {
	requested := strings.TrimSpace(c.Query("inspector_id"))
	callerID := middleware.GetUserID(c)
	roleID := middleware.GetRoleID(c)
	// IsAuditorRole also includes Admin/Super Admin because those roles may
	// perform inspections. Only ROLE-002 must be restricted to its own data.
	if strings.EqualFold(strings.TrimSpace(roleID), "ROLE-002") {
		if requested != "" && requested != callerID {
			return "", errors.New("tidak diizinkan melihat tren auditor lain")
		}
		return callerID, nil
	}
	if requested != "" && !isDashboardConfigAdmin(roleID) {
		return "", errors.New("filter auditor hanya tersedia untuk admin")
	}
	return requested, nil
}

// GetTrend returns a shared chart dataset for admin and auditor dashboards.
func (h *DashboardHandler) GetTrend(c *fiber.Ctx) error {
	spec, err := buildTrendPeriod(c.Query("period", "monthly"), time.Now(), c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	inspectorID, err := resolveTrendInspector(c)
	if err != nil {
		return response.Forbidden(c, err.Error())
	}
	allowedAreas := h.getAllowedAreas(c)
	areaID := strings.TrimSpace(c.Query("area_id"))

	inspectionBucketExpr := trendBucketSQL(`ih."InspectionHeaderCreatedAt"`, spec.Granularity)
	inspectionQuery := h.db.Table(`"Inspection_Header" ih`).
		Select(inspectionBucketExpr+` AS bucket_key,
			COUNT(DISTINCT ih."InspectionID") AS total_inspections,
			COUNT(ir."ResultID") AS total_checks,
			COUNT(ir."ResultID") FILTER (WHERE ir."Checking" = 'OK') AS total_ok`).
		Joins(`LEFT JOIN "Inspection_Result" ir ON ir."InspectionID" = ih."InspectionID"`).
		Where(`ih."InspectionHeaderCreatedAt" >= ? AND ih."InspectionHeaderCreatedAt" < ?`, spec.Start, spec.End).
		Group(inspectionBucketExpr)
	inspectionQuery = applyTrendScope(inspectionQuery, allowedAreas, areaID, inspectorID)

	var inspectionRows []inspectionTrendAggregate
	if err := inspectionQuery.Scan(&inspectionRows).Error; err != nil {
		h.log.Error("failed to aggregate dashboard inspection trend", logger.Error(err))
		return response.InternalServerError(c, "Gagal memuat tren inspeksi", err.Error())
	}

	issueBucketExpr := trendBucketSQL(`ip."PhotoCreatedAt"`, spec.Granularity)
	issueQuery := h.db.Table(`"Issue_Photo" ip`).
		Select(issueBucketExpr+` AS bucket_key, COUNT(ip."IssuePhotoID") AS total_issues`).
		Joins(`JOIN "Issue" i ON i."IssueID" = ip."IssueID"`).
		Joins(`JOIN "Inspection_Result" ir ON ir."ResultID" = i."ResultID"`).
		Joins(`JOIN "Inspection_Header" ih ON ih."InspectionID" = ir."InspectionID"`).
		Where(`ip."PhotoType" = 'Initial' AND ip."PhotoCreatedAt" >= ? AND ip."PhotoCreatedAt" < ?`, spec.Start, spec.End).
		Group(issueBucketExpr)
	issueQuery = applyTrendScope(issueQuery, allowedAreas, areaID, inspectorID)

	var issueRows []issueTrendAggregate
	if err := issueQuery.Scan(&issueRows).Error; err != nil {
		h.log.Error("failed to aggregate dashboard issue trend", logger.Error(err))
		return response.InternalServerError(c, "Gagal memuat tren issue", err.Error())
	}

	inspectionByBucket := make(map[string]inspectionTrendAggregate, len(inspectionRows))
	for _, row := range inspectionRows {
		inspectionByBucket[row.BucketKey] = row
	}
	issuesByBucket := make(map[string]int64, len(issueRows))
	for _, row := range issueRows {
		issuesByBucket[row.BucketKey] = row.TotalIssues
	}

	points := make([]trendPoint, 0, len(spec.Buckets))
	var totalInspections, totalIssues, totalChecks, totalOK int64
	for _, bucket := range spec.Buckets {
		inspectionData := inspectionByBucket[bucket.Key]
		complianceRate := 0.0
		if inspectionData.TotalChecks > 0 {
			complianceRate = float64(inspectionData.TotalOK) / float64(inspectionData.TotalChecks) * 100
		}
		issueCount := issuesByBucket[bucket.Key]
		points = append(points, trendPoint{
			BucketStart:      bucket.Start.Format("2006-01-02"),
			BucketEnd:        bucket.End.Format("2006-01-02"),
			Label:            bucket.Label,
			TotalInspections: inspectionData.TotalInspections,
			TotalIssues:      issueCount,
			ComplianceRate:   float64(int(complianceRate*10)) / 10,
		})
		totalInspections += inspectionData.TotalInspections
		totalIssues += issueCount
		totalChecks += inspectionData.TotalChecks
		totalOK += inspectionData.TotalOK
	}

	averageCompliance := 0.0
	if totalChecks > 0 {
		averageCompliance = float64(totalOK) / float64(totalChecks) * 100
	}
	return response.OK(c, "success", fiber.Map{
		"period":       spec.Period,
		"period_label": spec.Label,
		"granularity":  spec.Granularity,
		"timezone":     dashboardLocation().String(),
		"range_start":  spec.Start.Format("2006-01-02"),
		"range_end":    spec.End.Format("2006-01-02"),
		"inspector_id": inspectorID,
		"points":       points,
		"summary": fiber.Map{
			"total_inspections":  totalInspections,
			"total_issues":       totalIssues,
			"average_compliance": float64(int(averageCompliance*10)) / 10,
		},
	})
}
