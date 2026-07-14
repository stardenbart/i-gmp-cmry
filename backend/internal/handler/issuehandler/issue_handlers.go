package issuehandler

import (
	"github.com/gin-gonic/gin"
	"github.com/monitoring-system/backend/internal/domain/issue"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/pagination"
	"github.com/monitoring-system/backend/pkg/response"
	"github.com/monitoring-system/backend/pkg/validator"
)

// ── Issue Handler ─────────────────────────────────────────────────────────

type IssueHandler struct{ uc issue.IssueUseCase }

func NewIssueHandler(uc issue.IssueUseCase) *IssueHandler { return &IssueHandler{uc: uc} }

func (h *IssueHandler) GetAll(c *gin.Context) {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("status"), c.Query("pic_user_id"))
	if err != nil { response.InternalServerError(c, "failed to fetch issues", err.Error()); return }
	response.Paginated(c, "success", items, total, p.Page, p.Limit)
}

func (h *IssueHandler) GetByID(c *gin.Context) {
	item, err := h.uc.GetByID(c.Param("id"))
	if err != nil { response.NotFound(c, "issue not found"); return }
	response.OK(c, "success", item)
}

func (h *IssueHandler) Create(c *gin.Context) {
	var req issue.CreateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "invalid body", err.Error()); return }
	if errs := validator.Validate(&req); errs != nil { response.BadRequest(c, "validation failed", errs); return }
	actorID := middleware.GetUserID(c)
	item, err := h.uc.Create(actorID, &req)
	if err != nil { response.BadRequest(c, err.Error(), nil); return }
	response.Created(c, "issue created", item)
}

func (h *IssueHandler) Update(c *gin.Context) {
	var req issue.UpdateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "invalid body", err.Error()); return }
	actorID := middleware.GetUserID(c)
	item, err := h.uc.Update(c.Param("id"), actorID, &req)
	if err != nil { response.BadRequest(c, err.Error(), nil); return }
	response.OK(c, "issue updated", item)
}

func (h *IssueHandler) Delete(c *gin.Context) {
	actorID := middleware.GetUserID(c)
	if err := h.uc.Delete(c.Param("id"), actorID); err != nil { response.BadRequest(c, err.Error(), nil); return }
	response.OK(c, "issue deleted", nil)
}

// ── Issue Photo Handler ───────────────────────────────────────────────────

type IssuePhotoHandler struct{ uc issue.IssuePhotoUseCase }

func NewIssuePhotoHandler(uc issue.IssuePhotoUseCase) *IssuePhotoHandler { return &IssuePhotoHandler{uc: uc} }

func (h *IssuePhotoHandler) GetByIssueID(c *gin.Context) {
	photos, err := h.uc.GetByIssueID(c.Param("id"))
	if err != nil { response.InternalServerError(c, "failed to fetch photos", err.Error()); return }
	response.OK(c, "success", photos)
}

// Upload handles multipart/form-data file upload for issue photos.
func (h *IssuePhotoHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("photo")
	if err != nil { response.BadRequest(c, "file is required", err.Error()); return }
	defer file.Close()

	photoType := issue.PhotoType(c.PostForm("photo_type"))
	picUserID := middleware.GetUserID(c)
	issueID := c.Param("id")
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	req := &issue.UploadPhotoRequest{
		IssueID:   issueID,
		PICUserID: picUserID,
		PhotoType: photoType,
	}

	photo, err := h.uc.Upload(c.Request.Context(), req, file, header.Size, header.Filename, contentType)
	if err != nil { response.InternalServerError(c, "upload failed", err.Error()); return }
	response.Created(c, "photo uploaded", photo)
}

func (h *IssuePhotoHandler) Delete(c *gin.Context) {
	if err := h.uc.Delete(c.Request.Context(), c.Param("photo_id")); err != nil { response.BadRequest(c, err.Error(), nil); return }
	response.OK(c, "photo deleted", nil)
}
