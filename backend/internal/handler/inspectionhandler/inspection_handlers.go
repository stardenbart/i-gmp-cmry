package inspectionhandler

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/uploadusecase"
	"github.com/monitoring-system/backend/pkg/exporter"
	"github.com/monitoring-system/backend/pkg/pagination"
	"github.com/monitoring-system/backend/pkg/response"
	"github.com/monitoring-system/backend/pkg/validator"
	redis "github.com/redis/go-redis/v9"
)

// ── Inspection Header Handler ─────────────────────────────────────────────

type InspectionHeaderHandler struct {
	uc       inspection.InspectionHeaderUseCase
	resultUC inspection.InspectionResultUseCase
	uploadUC *uploadusecase.UploadUseCase
	rdb      *redis.Client
}

func NewInspectionHeaderHandler(
	uc inspection.InspectionHeaderUseCase,
	resultUC inspection.InspectionResultUseCase,
	uploadUC *uploadusecase.UploadUseCase,
	rdb *redis.Client,
) *InspectionHeaderHandler {
	return &InspectionHeaderHandler{uc: uc, resultUC: resultUC, uploadUC: uploadUC, rdb: rdb}
}

// @Summary Get all inspections
// @Description Get a paginated list of inspections with optional filters
// @Tags Inspections
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Limit per page"
// @Param area_id query string false "Area ID filter"
// @Param status query string false "Status filter"
// @Param inspector_id query string false "Inspector ID filter"
// @Success 200 {object} response.APIResponse "success"
// @Failure 500 {object} response.APIResponse "failed to fetch inspections"
// @Router /inspections [get]
// @Security BearerAuth
func (h *InspectionHeaderHandler) GetAll(c *fiber.Ctx) error {
	p := pagination.FromQuery(c)
	userPlantID, _ := c.Locals("userPlantID").(string)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, userPlantID, c.Query("area_id"), c.Query("status"), c.Query("inspector_id"))
	if err != nil {
		return response.InternalServerError(c, "failed to fetch inspections", err.Error())
	}
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}

// @Summary Get inspection by ID
// @Description Retrieve a specific inspection header by its ID
// @Tags Inspections
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Success 200 {object} response.APIResponse "success"
// @Failure 404 {object} response.APIResponse "inspection not found"
// @Router /inspections/{id} [get]
// @Security BearerAuth
func (h *InspectionHeaderHandler) GetByID(c *fiber.Ctx) error {
	userPlantID, _ := c.Locals("userPlantID").(string)
	item, err := h.uc.GetByIDScoped(c.Params("id"), userPlantID)
	if err != nil {
		if err.Error() == "record not found" || err.Error() == "inspection not found" {
			return response.NotFound(c, "inspection not found")
		}
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "context deadline exceeded") {
			return response.ServiceUnavailable(c, "service busy, please try again", err.Error())
		}
		return response.InternalServerError(c, "failed to fetch inspection", err.Error())
	}
	return response.OK(c, "success", item)
}

// requireInspectionAccess verifies the caller may see the inspection at id
// under their current plant scope (see InspectionHeaderUseCase.GetByIDScoped)
// before an endpoint that doesn't otherwise need the fetched header goes on
// to read/mutate one of its sub-resources. Returns a ready-to-send 404 when
// they may not — nil means "allowed, proceed".
func (h *InspectionHeaderHandler) requireInspectionAccess(c *fiber.Ctx, id string) error {
	userPlantID, _ := c.Locals("userPlantID").(string)
	if _, err := h.uc.GetByIDScoped(id, userPlantID); err != nil {
		return response.NotFound(c, "inspection not found")
	}
	return nil
}

func (h *InspectionHeaderHandler) GetChecklist(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.requireInspectionAccess(c, id); err != nil {
		return err
	}
	item, err := h.uc.GetChecklist(id)
	if err != nil {
		if err.Error() == "record not found" || err.Error() == "inspection not found" {
			return response.NotFound(c, "inspection not found")
		}
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "context deadline exceeded") {
			return response.ServiceUnavailable(c, "service busy, please try again", err.Error())
		}
		return response.InternalServerError(c, "failed to load checklist", err.Error())
	}
	return response.OK(c, "success", item)
}

// @Summary Get area status
// @Description Calculate and get the inspection status progress for a specific area
// @Tags Inspections
// @Accept json
// @Produce json
// @Param areaId path string true "Area ID"
// @Success 200 {object} response.APIResponse "success"
// @Failure 500 {object} response.APIResponse "failed to calculate area status"
// @Router /inspections/area/{areaId}/status [get]
// @Security BearerAuth
func (h *InspectionHeaderHandler) GetAreaStatus(c *fiber.Ctx) error {
	progress, err := h.uc.GetAreaStatus(c.Params("areaId"))
	if err != nil {
		return response.InternalServerError(c, "failed to calculate area status", err.Error())
	}
	return response.OK(c, "success", progress)
}

// @Summary Get current inspection period info
// @Description Resolve the currently-running inspection period (cutoff-day cycle) for the plant that owns this Area, so clients never have to re-derive the cutoff-day math themselves
// @Tags Inspections
// @Accept json
// @Produce json
// @Param area_id query string true "Area ID"
// @Success 200 {object} response.APIResponse "success"
// @Failure 400 {object} response.APIResponse "area_id wajib diisi"
// @Failure 500 {object} response.APIResponse "failed to resolve current period"
// @Router /inspections/period-info [get]
// @Security BearerAuth
func (h *InspectionHeaderHandler) GetCurrentPeriodInfo(c *fiber.Ctx) error {
	areaID := c.Query("area_id")
	if areaID == "" {
		return response.BadRequest(c, "area_id wajib diisi", nil)
	}
	info, err := h.uc.GetCurrentPeriodInfo(areaID)
	if err != nil {
		return response.InternalServerError(c, "failed to resolve current period", err.Error())
	}
	return response.OK(c, "success", info)
}

// @Summary Get inspection trends
// @Description Get inspection trends by context and year
// @Tags analytics
// @Accept json
// @Produce json
// @Param context_id query string true "Context ID (e.g. Inspector ID)"
// @Param year query int true "Year"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/analytics/inspections-trend [get]
// @Security BearerAuth
func (h *InspectionHeaderHandler) GetTrend(c *fiber.Ctx) error {
	contextID := c.Query("context_id")
	year := c.QueryInt("year", time.Now().Year())

	if contextID == "" {
		return response.BadRequest(c, "context_id is required", nil)
	}

	trend, err := h.uc.GetTrend(contextID, year)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch inspection trend", err.Error())
	}
	return response.OK(c, "Inspection trend fetched successfully", trend)
}

// @Summary Create inspection
// @Description Create a new inspection header
// @Tags Inspections
// @Accept json
// @Produce json
// @Param body body inspection.CreateInspectionRequest true "Inspection creation request"
// @Success 201 {object} response.APIResponse "inspection created"
// @Failure 400 {object} response.APIResponse "bad request"
// @Router /inspections [post]
// @Security BearerAuth
func (h *InspectionHeaderHandler) Create(c *fiber.Ctx) error {
	var req inspection.CreateInspectionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid body", err.Error())
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.BadRequest(c, "validation failed", errs)
	}
	inspectorID := middleware.GetUserID(c)
	item, err := h.uc.Create(inspectorID, &req)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.Created(c, "inspection created", item)
}

// @Summary Update inspection status
// @Description Update the status of an existing inspection
// @Tags Inspections
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Param body body inspection.UpdateInspectionStatusRequest true "Status update request"
// @Success 200 {object} response.APIResponse "status updated"
// @Failure 400 {object} response.APIResponse "bad request"
// @Router /inspections/{id}/status [patch]
// @Security BearerAuth
func (h *InspectionHeaderHandler) UpdateStatus(c *fiber.Ctx) error {
	var req inspection.UpdateInspectionStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid body", err.Error())
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.BadRequest(c, "validation failed", errs)
	}
	actorID := middleware.GetUserID(c)
	inspectionID := c.Params("id")

	if err := h.requireInspectionAccess(c, inspectionID); err != nil {
		return err
	}

	// Jika status berubah ke Completed: finalisasi foto base64 WebP dari Redis ke MinIO
	if req.Status == "Completed" && h.uploadUC != nil {
		go func() {
			// Jalankan di background goroutine agar tidak memblok response
			// Frontend sudah punya isFinalizing guard untuk mencegah race condition
			_ = c.Request() // capture context before handler returns
			_, _ = h.uploadUC.FinalizeInspectionPhotos(
				c.Context(),
				h.rdb,
				inspectionID,
			)
		}()
	}

	item, err := h.uc.UpdateStatus(inspectionID, actorID, &req)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "status updated", item)
}

// @Summary Delete inspection
// @Description Delete an inspection by ID
// @Tags Inspections
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Success 200 {object} response.APIResponse "inspection deleted"
// @Failure 400 {object} response.APIResponse "bad request"
// @Router /inspections/{id} [delete]
// @Security BearerAuth
func (h *InspectionHeaderHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.requireInspectionAccess(c, id); err != nil {
		return err
	}
	if err := h.uc.Delete(id); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "inspection deleted", nil)
}

// @Summary Export inspection to Excel
// @Description Export inspection details and results to an Excel template
// @Tags Inspections
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param id path string true "Inspection ID"
// @Router /inspections/{id}/export [get]
// @Security BearerAuth
func (h *InspectionHeaderHandler) ExportExcel(c *fiber.Ctx) error {
	id := c.Params("id")
	userPlantID, _ := c.Locals("userPlantID").(string)

	// 1. Fetch Header
	header, err := h.uc.GetByIDScoped(id, userPlantID)
	if err != nil {
		return response.NotFound(c, "inspection not found")
	}

	// 2. Fetch Results
	results, err := h.resultUC.GetByInspectionID(id)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch results", err.Error())
	}

	// 3. Load Template Config
	// In a real app, you'd choose the config based on the Area/Form type.
	// We'll use the example config for now.
	config, err := exporter.LoadConfig("./templates/template_config.json")
	if err != nil {
		return response.InternalServerError(c, "failed to load template config", err.Error())
	}

	// 4. Map Data to Payload
	areaName := header.AreaName
	if areaName == "" {
		areaName = header.AreaID
	}
	picName := header.InspectorName
	if picName == "" {
		picName = header.InspectorID
	}
	kawasanName := header.KawasanName
	if kawasanName == "" {
		kawasanName = header.KawasanID
	}

	payload := &exporter.ExportPayload{
		TemplateID: config.TemplateID,
		Headers: map[string]interface{}{
			"tanggal": header.InspectionHeaderCreatedAt.Format("2006-01-02"),
			"area":    areaName,
			"pic":     picName,
			"kawasan": kawasanName,
		},
		TableData: make([]exporter.TableRow, 0, len(results)),
	}

	for i, res := range results {
		payload.TableData = append(payload.TableData, exporter.TableRow{
			Data: map[string]interface{}{
				"no":         i + 1,
				"uraian_id":  res.UraianID,
				"uraian":     res.Keterangan,
				"nilai":      res.Nilai,
				"keterangan": res.Keterangan,
			},
			ImagePath: "", // Photo logic can be added here if needed
		})
	}

	// 5. Generate Excel
	buf, err := exporter.GenerateExcel(config, payload)
	if err != nil {
		return response.InternalServerError(c, "failed to generate excel", err.Error())
	}

	// 6. Return as downloadable file
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=\"report_inspeksi_"+id+".xlsx\"")
	return c.SendStream(buf)
}

// ── Inspection Result Handler ─────────────────────────────────────────────

type InspectionResultHandler struct {
	uc       inspection.InspectionResultUseCase
	headerUC inspection.InspectionHeaderUseCase
}

func NewInspectionResultHandler(uc inspection.InspectionResultUseCase, headerUC inspection.InspectionHeaderUseCase) *InspectionResultHandler {
	return &InspectionResultHandler{uc: uc, headerUC: headerUC}
}

// requireInspectionAccess mirrors InspectionHeaderHandler's helper of the
// same name — results/photos/etc. carry their own resource IDs, but access
// to them is still governed by the plant that owns their parent inspection.
func (h *InspectionResultHandler) requireInspectionAccess(c *fiber.Ctx, inspectionID string) error {
	userPlantID, _ := c.Locals("userPlantID").(string)
	if _, err := h.headerUC.GetByIDScoped(inspectionID, userPlantID); err != nil {
		return response.NotFound(c, "inspection not found")
	}
	return nil
}

// @Summary Get inspection results by inspection ID
// @Description Retrieve all results associated with a specific inspection
// @Tags Inspection Results
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Success 200 {object} response.APIResponse "success"
// @Failure 500 {object} response.APIResponse "failed to fetch results"
// @Router /inspections/{id}/results [get]
// @Security BearerAuth
func (h *InspectionResultHandler) GetByInspectionID(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.requireInspectionAccess(c, id); err != nil {
		return err
	}
	items, err := h.uc.GetByInspectionID(id)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch results", err.Error())
	}
	return response.OK(c, "success", items)
}

// @Summary Bulk save inspection results
// @Description Save multiple results for a specific inspection in bulk
// @Tags Inspection Results
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Param body body inspection.BulkSaveResultRequest true "Bulk save request"
// @Success 200 {object} response.APIResponse "results saved"
// @Failure 400 {object} response.APIResponse "bad request"
// @Failure 500 {object} response.APIResponse "internal server error"
// @Router /inspections/{id}/results/bulk [post]
// @Security BearerAuth
func (h *InspectionResultHandler) BulkSave(c *fiber.Ctx) error {
	var req inspection.BulkSaveResultRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid body", err.Error())
	}
	req.InspectionID = c.Params("id")
	if err := h.requireInspectionAccess(c, req.InspectionID); err != nil {
		return err
	}
	items, err := h.uc.BulkSave(&req)
	if err != nil {
		return response.InternalServerError(c, err.Error(), nil)
	}
	return response.OK(c, "results and issues synchronized", items)
}

// @Summary Update inspection result
// @Description Update a specific inspection result
// @Tags Inspection Results
// @Accept json
// @Produce json
// @Param result_id path string true "Result ID"
// @Param body body inspection.SaveResultRequest true "Update result request"
// @Success 200 {object} response.APIResponse "result updated"
// @Failure 400 {object} response.APIResponse "bad request"
// @Router /inspection-results/{result_id} [put]
// @Security BearerAuth
func (h *InspectionResultHandler) Update(c *fiber.Ctx) error {
	var req inspection.SaveResultRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid body", err.Error())
	}
	resultID := c.Params("result_id")
	existing, err := h.uc.GetByID(resultID)
	if err != nil {
		return response.NotFound(c, "result not found")
	}
	if err := h.requireInspectionAccess(c, existing.InspectionID); err != nil {
		return err
	}
	item, err := h.uc.Update(resultID, &req)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "result updated", item)
}

// @Summary Delete inspection result
// @Description Delete a specific inspection result
// @Tags Inspection Results
// @Accept json
// @Produce json
// @Param result_id path string true "Result ID"
// @Success 200 {object} response.APIResponse "result deleted"
// @Failure 400 {object} response.APIResponse "bad request"
// @Router /inspection-results/{result_id} [delete]
// @Security BearerAuth
func (h *InspectionResultHandler) Delete(c *fiber.Ctx) error {
	resultID := c.Params("result_id")
	existing, err := h.uc.GetByID(resultID)
	if err != nil {
		return response.NotFound(c, "result not found")
	}
	if err := h.requireInspectionAccess(c, existing.InspectionID); err != nil {
		return err
	}
	if err := h.uc.Delete(resultID); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "result deleted", nil)
}
