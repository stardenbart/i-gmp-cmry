package dashboardhandler

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const (
	gmpExportFormatTemplate = "template"
	gmpExportFormatTable    = "table"
)

func normalizeGMPExportFormat(value string) (string, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return gmpExportFormatTemplate, true
	}
	switch value {
	case gmpExportFormatTemplate, gmpExportFormatTable:
		return value, true
	default:
		return "", false
	}
}

func validateGMPDateRange(startDate, endDate string) error {
	var start, end time.Time
	var err error
	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return fmt.Errorf("tanggal mulai harus menggunakan format YYYY-MM-DD")
		}
	}
	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return fmt.Errorf("tanggal selesai harus menggunakan format YYYY-MM-DD")
		}
	}
	if !start.IsZero() && !end.IsZero() && start.After(end) {
		return fmt.Errorf("tanggal mulai tidak boleh melewati tanggal selesai")
	}
	return nil
}

func (h *DashboardHandler) applyGMPDataFilters(
	c *fiber.Ctx,
	query *gorm.DB,
	areaID, kawasanID, detailKawasanID, startDate, endDate string,
) *gorm.DB {
	if areaID != "" {
		query = query.Where(`ih."AreaID" = ?`, areaID)
	}
	if kawasanID != "" {
		query = query.Where(`ih."KawasanID" = ?`, kawasanID)
	}
	if detailKawasanID != "" {
		query = query.Where(`ih."DetailKawasanID" = ?`, detailKawasanID)
	}
	if startDate != "" {
		query = query.Where(`ih."InspectionHeaderCreatedAt" >= ?`, startDate+" 00:00:00")
	}
	if endDate != "" {
		query = query.Where(`ih."InspectionHeaderCreatedAt" <= ?`, endDate+" 23:59:59")
	}
	if allowedAreas := h.getAllowedAreas(c); allowedAreas != nil {
		query = query.Where(`ih."AreaID" IN ?`, allowedAreas)
	}
	return query
}

// matchesGMPSearch is shared by preview and export so a query entered in the
// Data GMP search box selects the same rows in both places.
func matchesGMPSearch(query string, values []string, evidence []gmpFollowUpEvidence) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}
	for _, item := range evidence {
		if strings.Contains(strings.ToLower(item.Keterangan), query) {
			return true
		}
	}
	return false
}
