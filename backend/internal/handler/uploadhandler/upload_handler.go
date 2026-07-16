package uploadhandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/upload"
	"github.com/monitoring-system/backend/internal/usecase/uploadusecase"
	"github.com/monitoring-system/backend/pkg/response"
)

// UploadHandler handles HTTP requests for file uploads
type UploadHandler struct {
	uc *uploadusecase.UploadUseCase
}

// NewUploadHandler creates a new upload handler instance
func NewUploadHandler(uc *uploadusecase.UploadUseCase) *UploadHandler {
	return &UploadHandler{uc: uc}
}

// Upload handles file upload requests
// POST /api/v1/uploads/file
// @Summary Upload a file
// @Description Upload a file associated with an inspection
// @Tags uploads
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param file_type formData string true "File Type"
// @Param inspection_id formData string true "Inspection ID"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 413 {object} map[string]response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/uploads/file [post]
// @Security BearerAuth
func (h *UploadHandler) Upload(c *fiber.Ctx) error {
	// Get the uploaded file
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "no file provided", nil)
	}

	// Get file type from form
	fileType := c.FormValue("file_type")
	if fileType == "" {
		return response.BadRequest(c, "file_type is required", nil)
	}

	// Get inspection ID from form
	inspectionID := c.FormValue("inspection_id")
	if inspectionID == "" {
		return response.BadRequest(c, "inspection_id is required", nil)
	}

	// Build request
	req := &upload.UploadRequest{
		FileType:     fileType,
		InspectionID: inspectionID,
	}

	// Call use case
	result, err := h.uc.UploadFile(c.Context(), fileHeader, req)
	if err != nil {
		// Map error to appropriate response
		switch err {
		case upload.ErrInvalidExtension:
			return response.BadRequest(c, "file extension not allowed", nil)
		case upload.ErrInvalidMIMEType:
			return response.BadRequest(c, "invalid file type", nil)
		case upload.ErrFileTooLarge:
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"status":  "error",
				"message": "file size exceeds maximum",
			})
		case upload.ErrCorruptedDOCX:
			return response.BadRequest(c, "corrupted or invalid docx file", nil)
		case upload.ErrInvalidMagicBytes:
			return response.BadRequest(c, "invalid magic bytes", nil)
		default:
			return response.InternalServerError(c, "failed to upload file", err.Error())
		}
	}

	return response.Created(c, "file uploaded successfully", result)
}

// GetStatus retrieves the status of an upload
// GET /api/v1/uploads/:id/status
// @Summary Get upload status
// @Description Get the status of a specific upload
// @Tags uploads
// @Accept json
// @Produce json
// @Param id path string true "Upload ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/uploads/{id}/status [get]
// @Security BearerAuth
func (h *UploadHandler) GetStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "upload id is required", nil)
	}

	upload, err := h.uc.GetByID(id)
	if err != nil {
		return response.NotFound(c, "upload not found")
	}

	return response.OK(c, "success", fiber.Map{
		"id":            upload.ID,
		"status":        upload.Status,
		"processed_url": upload.ProcessedURL,
		"file_url":      upload.FilePath,
	})
}

// GetByInspection retrieves all uploads for an inspection
// GET /api/v1/uploads/inspection/:inspectionId
// @Summary Get uploads by inspection
// @Description Get all uploads associated with a specific inspection
// @Tags uploads
// @Accept json
// @Produce json
// @Param inspectionId path string true "Inspection ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/uploads/inspection/{inspectionId} [get]
// @Security BearerAuth
func (h *UploadHandler) GetByInspection(c *fiber.Ctx) error {
	inspectionID := c.Params("inspectionId")
	if inspectionID == "" {
		return response.BadRequest(c, "inspection_id is required", nil)
	}

	uploads, err := h.uc.GetByInspectionID(inspectionID)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch uploads", err.Error())
	}

	return response.OK(c, "success", uploads)
}
