package issuehandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/issue"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/pagination"
	"github.com/monitoring-system/backend/pkg/response"
	"github.com/monitoring-system/backend/pkg/validator"
)

// ── Issue Handler ─────────────────────────────────────────────────────────

type IssueHandler struct{ uc issue.IssueUseCase }

func NewIssueHandler(uc issue.IssueUseCase) *IssueHandler { return &IssueHandler{uc: uc} }

func (h *IssueHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("status"), c.Query("pic_user_id"))
	if err != nil { return response.InternalServerError(c, "failed to fetch issues", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}

func (h *IssueHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "issue not found") }
	return response.OK(c, "success", item)
}

func (h *IssueHandler) Create(c *fiber.Ctx) error  {
	var req issue.CreateIssueRequest
	if err := c.BodyParser(&req); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if errs := validator.Validate(&req); errs != nil { return response.BadRequest(c, "validation failed", errs) }
	actorID := middleware.GetUserID(c)
	item, err := h.uc.Create(actorID, &req)
	if err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "issue created", item)
}

func (h *IssueHandler) Update(c *fiber.Ctx) error  {
	var req issue.UpdateIssueRequest
	if err := c.BodyParser(&req); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	actorID := middleware.GetUserID(c)
	item, err := h.uc.Update(c.Params("id"), actorID, &req)
	if err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "issue updated", item)
}

func (h *IssueHandler) Delete(c *fiber.Ctx) error  {
	actorID := middleware.GetUserID(c)
	if err := h.uc.Delete(c.Params("id"), actorID); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "issue deleted", nil)
}

// ── Issue Photo Handler ───────────────────────────────────────────────────

type IssuePhotoHandler struct{ uc issue.IssuePhotoUseCase }

func NewIssuePhotoHandler(uc issue.IssuePhotoUseCase) *IssuePhotoHandler { return &IssuePhotoHandler{uc: uc} }

func (h *IssuePhotoHandler) GetByIssueID(c *fiber.Ctx) error  {
	photos, err := h.uc.GetByIssueID(c.Params("id"))
	if err != nil { return response.InternalServerError(c, "failed to fetch photos", err.Error()) }
	return response.OK(c, "success", photos)
}

// Upload handles multipart/form-data file upload for issue photos.
func (h *IssuePhotoHandler) Upload(c *fiber.Ctx) error  {
	header, err := c.FormFile("photo")
	if err != nil { return response.BadRequest(c, "file is required", err.Error()) }
	
	file, err := header.Open()
	if err != nil { return response.InternalServerError(c, "failed to open file", err.Error()) }
	defer file.Close()

	photoType := issue.PhotoType(c.FormValue("photo_type"))
	picUserID := middleware.GetUserID(c)
	issueID := c.Params("id")
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	req := &issue.UploadPhotoRequest{
		IssueID:   issueID,
		PICUserID: picUserID,
		PhotoType: photoType,
	}

	photo, err := h.uc.Upload(c.UserContext(), req, file, header.Size, header.Filename, contentType)
	if err != nil { return response.InternalServerError(c, "upload failed", err.Error()) }
	return response.Created(c, "photo uploaded", photo)
}

func (h *IssuePhotoHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.UserContext(), c.Params("photo_id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "photo deleted", nil)
}
