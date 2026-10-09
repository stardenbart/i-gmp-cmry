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
	// Checklist order for every area: master Aspek, then Detail, then the
	// order the uraian were created in (UraianIDs are random hashes).
	gmpDataRelationOrder    = `ih."InspectionHeaderCreatedAt" DESC, ih."InspectionID" DESC, am."AspekID" ASC, dm."DetailID" ASC, um."UraianCreatedAt" ASC, um."UraianID" ASC, ir."ResultID" ASC`
	gmpCompanyName          = "PT CISARUA MOUNTAIN DAIRY TBK"
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

func gmpAspectGroupKey(inspectionID, aspectName string) string {
	return strings.TrimSpace(inspectionID) + "\x00" + strings.ToLower(strings.TrimSpace(aspectName))
}

func resolveGMPExportPlantID(userPlantID, queryPlantID string) string {
	if plantID := strings.TrimSpace(userPlantID); plantID != "" {
		return plantID
	}
	plantID := strings.TrimSpace(queryPlantID)
	if strings.EqualFold(plantID, "all") || strings.EqualFold(plantID, "global") || strings.EqualFold(plantID, "null") {
		return ""
	}
	return plantID
}

func (h *DashboardHandler) resolveGMPExportPlantName(c *fiber.Ctx, areaID string) string {
	userPlantID, _ := c.Locals("userPlantID").(string)
	plantID := resolveGMPExportPlantID(userPlantID, c.Query("plant_id"))
	if plantID == "" && areaID != "" {
		_ = h.db.Table(`"Area_Master"`).Select(`"PlantID"`).Where(`"AreaID" = ?`, areaID).Scan(&plantID).Error
	}
	if plantID == "" {
		return "Semua Plant"
	}

	plantName := plantID
	_ = h.db.Table(`"Plant_Master"`).Select(`"PlantName"`).Where(`"PlantID" = ?`, plantID).Scan(&plantName).Error
	return plantName
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
