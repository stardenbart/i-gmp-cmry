package issuehandler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/issue"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/pagination"
	"github.com/monitoring-system/backend/pkg/response"
	"github.com/monitoring-system/backend/pkg/validator"
)

// Helper to check if a role is Auditor/Admin (strict RoleID match)
func isAuditor(roleID string) bool {
	return middleware.IsAuditorRole(roleID)
}

// ── Issue Handler ─────────────────────────────────────────────────────────

type IssueHandler struct{ uc issue.IssueUseCase }

func NewIssueHandler(uc issue.IssueUseCase) *IssueHandler { return &IssueHandler{uc: uc} }

// @Summary Get all issues
// @Description Get a paginated list of issues, optionally filtered by status or pic_user_id
// @Tags issues
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param status query string false "Filter by status"
// @Param pic_user_id query string false "Filter by PIC User ID"
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/issues [get]
// @Security BearerAuth
func (h *IssueHandler) GetAll(c *fiber.Ctx) error {
	p := pagination.FromQuery(c)
	var needsWOWR *bool
	if needsStr := c.Query("needs_wo_wr"); needsStr != "" {
		b, err := strconv.ParseBool(needsStr)
		if err == nil {
			needsWOWR = &b
		}
	}

	picUserID := c.Query("pic_user_id")
	actorRole := middleware.GetRoleID(c)
	actorID := middleware.GetUserID(c)

	// SECURITY: If the user is not an auditor (i.e. they are an auditee),
	// strictly enforce that they can only view issues assigned to them.
	if !isAuditor(actorRole) {
		picUserID = actorID
	}

	// Plant scope enforcement for non-SuperAdmin users
	userPlantID, _ := c.Locals("userPlantID").(string)

	items, total, err := h.uc.GetAll(p.Page, p.Limit, userPlantID, c.Query("status"), picUserID, needsWOWR)
	if err != nil {
		return response.InternalServerError(c, "failed to fetch issues", err.Error())
	}
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}

// @Summary Get issue by ID
// @Description Get a specific issue by its ID
// @Tags issues
// @Accept json
// @Produce json
// @Param id path string true "Issue ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/issues/{id} [get]
// @Security BearerAuth
func (h *IssueHandler) GetByID(c *fiber.Ctx) error {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil {
		return response.NotFound(c, "issue not found")
	}
	return response.OK(c, "success", item)
}

// @Summary Create a new issue
// @Description Create a new issue with the provided details
// @Tags issues
// @Accept json
// @Produce json
// @Param request body issue.CreateIssueRequest true "Create Issue Request"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/issues [post]
// @Security BearerAuth
func (h *IssueHandler) Create(c *fiber.Ctx) error {
	var req issue.CreateIssueRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid body", err.Error())
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.BadRequest(c, "validation failed", errs)
	}
	actorID := middleware.GetUserID(c)
	item, err := h.uc.Create(actorID, &req)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.Created(c, "issue created", item)
}

// @Summary Update an issue
// @Description Update an existing issue by its ID
// @Tags issues
// @Accept json
// @Produce json
// @Param id path string true "Issue ID"
// @Param request body issue.UpdateIssueRequest true "Update Issue Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/issues/{id} [put]
// @Security BearerAuth
func (h *IssueHandler) Update(c *fiber.Ctx) error {
	var req issue.UpdateIssueRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid body", err.Error())
	}
	actorID := middleware.GetUserID(c)
	item, err := h.uc.Update(c.Params("id"), actorID, &req)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "issue updated", item)
}

type ExtendDueDateRequest struct {
	DueDate string `json:"due_date" validate:"required"`
}

// @Summary Extend issue due date
// @Description Extend the due date of a specific issue
// @Tags issues
// @Accept json
// @Produce json
// @Param id path string true "Issue ID"
// @Param request body ExtendDueDateRequest true "Extend Due Date Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/issues/{id}/extend-due-date [patch]
// @Security BearerAuth
func (h *IssueHandler) ExtendDueDate(c *fiber.Ctx) error {
	var req ExtendDueDateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid body", err.Error())
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.BadRequest(c, "validation failed", errs)
	}

	actorID := middleware.GetUserID(c)

	// Parse Time manually since it's an extension
	importTime, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		return response.BadRequest(c, "invalid date format, must be RFC3339", err.Error())
	}

	item, err := h.uc.ExtendDueDate(c.Params("id"), actorID, importTime)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "due date extended", item)
}

// @Summary Delete an issue
// @Description Delete an issue by its ID
// @Tags issues
// @Accept json
// @Produce json
// @Param id path string true "Issue ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/issues/{id} [delete]
// @Security BearerAuth
func (h *IssueHandler) Delete(c *fiber.Ctx) error {
	actorID := middleware.GetUserID(c)
	if err := h.uc.Delete(c.Params("id"), actorID); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "issue deleted", nil)
}

// ── Issue Photo Handler ───────────────────────────────────────────────────

type IssuePhotoHandler struct{ uc issue.IssuePhotoUseCase }

func NewIssuePhotoHandler(uc issue.IssuePhotoUseCase) *IssuePhotoHandler {
	return &IssuePhotoHandler{uc: uc}
}

// @Summary Get photos by issue ID
// @Description Get all photos associated with a specific issue
// @Tags issue-photos
// @Accept json
// @Produce json
// @Param id path string true "Issue ID"
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/issues/{id}/photos [get]
// @Security BearerAuth
func (h *IssuePhotoHandler) GetByIssueID(c *fiber.Ctx) error {
	photos, err := h.uc.GetByIssueID(c.Params("id"))
	if err != nil {
		return response.InternalServerError(c, "failed to fetch photos", err.Error())
	}
	return response.OK(c, "success", photos)
}

// Upload handles multipart/form-data file upload for issue photos.
// @Summary Upload an issue photo
// @Description Upload a photo for a specific issue
// @Tags issue-photos
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Issue ID"
// @Param photo formData file true "Photo file"
// @Param photo_type formData string true "Photo Type"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/issues/{id}/photos [post]
// @Security BearerAuth
func (h *IssuePhotoHandler) Upload(c *fiber.Ctx) error {
	// Support both 'photo' and 'file' field names
	header, err := c.FormFile("photo")
	if err != nil {
		header, err = c.FormFile("file")
		if err != nil {
			return response.BadRequest(c, "file is required", err.Error())
		}
	}

	file, err := header.Open()
	if err != nil {
		return response.InternalServerError(c, "failed to open file", err.Error())
	}
	defer file.Close()

	photoTypeStr := c.FormValue("photo_type")
	photoType := issue.PhotoType(photoTypeStr)
	if photoType != issue.PhotoTypeInitial && photoType != issue.PhotoTypeFollowUp && photoType != issue.PhotoTypeWOWR {
		photoType = issue.PhotoTypeFollowUp
	}

	picUserID := middleware.GetUserID(c)
	issueID := c.Params("id")
	if picUserID == "" {
		picUserID = "SYSTEM"
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Parse follow_up_date from form data
	var followUpDate *time.Time
	if fuDateStr := c.FormValue("follow_up_date"); fuDateStr != "" {
		parsed, parseErr := time.Parse(time.RFC3339, fuDateStr)
		if parseErr != nil {
			// Try other common formats
			parsed, parseErr = time.Parse("2006-01-02", fuDateStr)
		}
		if parseErr == nil {
			followUpDate = &parsed
		}
	}

	// Parse jumlah_follow_up from form data
	var jumlahFollowUp *int
	if jfuStr := c.FormValue("jumlah_follow_up"); jfuStr != "" {
		if jfu, parseErr := strconv.Atoi(jfuStr); parseErr == nil {
			jumlahFollowUp = &jfu
		}
	}

	var refPhotoID *string
	if refStr := c.FormValue("ref_photo_id"); refStr != "" {
		refPhotoID = &refStr
	}

	req := &issue.UploadPhotoRequest{
		IssueID:        issueID,
		RefPhotoID:     refPhotoID,
		PICUserID:      picUserID,
		PhotoType:      photoType,
		Keterangan:     c.FormValue("keterangan"),
		FollowUpDate:   followUpDate,
		JumlahFollowUp: jumlahFollowUp,
	}

	chunkIndexStr := c.FormValue("chunk_index")
	totalChunksStr := c.FormValue("total_chunks")
	fileID := c.FormValue("file_id")

	if chunkIndexStr != "" && totalChunksStr != "" && fileID != "" {
		chunkIndex, _ := strconv.Atoi(chunkIndexStr)
		totalChunks, _ := strconv.Atoi(totalChunksStr)

		tempFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("upload_%s_%s", issueID, fileID))

		tempFile, err := os.OpenFile(tempFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return response.InternalServerError(c, "failed to create temp file", err.Error())
		}

		_, err = io.Copy(tempFile, file)
		tempFile.Close()

		if err != nil {
			return response.InternalServerError(c, "failed to write chunk", err.Error())
		}

		if chunkIndex == totalChunks-1 {
			// Last chunk, process it
			fullFile, err := os.Open(tempFilePath)
			if err != nil {
				return response.InternalServerError(c, "failed to open full file", err.Error())
			}
			defer fullFile.Close()
			defer os.Remove(tempFilePath)

			fileInfo, _ := fullFile.Stat()

			photo, err := h.uc.Upload(c.UserContext(), req, fullFile, fileInfo.Size(), header.Filename, contentType)
			if err != nil {
				return response.BadRequest(c, err.Error(), nil)
			}
			return response.Created(c, "photo uploaded", photo)
		} else {
			return response.OK(c, "chunk received", nil)
		}
	}

	photo, err := h.uc.Upload(c.UserContext(), req, file, header.Size, header.Filename, contentType)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.Created(c, "photo uploaded", photo)
}

func (h *IssuePhotoHandler) Update(c *fiber.Ctx) error {
	var req struct {
		Keterangan string `json:"keterangan"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid body", err.Error())
	}
	photo, err := h.uc.Update(c.UserContext(), c.Params("photo_id"), req.Keterangan)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "photo updated", photo)
}

// @Summary Delete an issue photo
// @Description Delete a specific issue photo by its ID
// @Tags issue-photos
// @Accept json
// @Produce json
// @Param photo_id path string true "Photo ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/issues/photos/{photo_id} [delete]
// @Security BearerAuth
func (h *IssuePhotoHandler) Delete(c *fiber.Ctx) error {
	if err := h.uc.Delete(c.UserContext(), c.Params("photo_id")); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "photo deleted", nil)
}
