package masterhandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/pagination"
	"github.com/monitoring-system/backend/pkg/response"
)

// ── Department Handler ────────────────────────────────────────────────────

type DepartmentHandler struct{ uc master.DepartmentUseCase }
func NewDepartmentHandler(uc master.DepartmentUseCase) *DepartmentHandler { return &DepartmentHandler{uc: uc} }

func (h *DepartmentHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("search"))
	if err != nil { return response.InternalServerError(c, "failed to fetch departments", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}
func (h *DepartmentHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "department not found") }
	return response.OK(c, "success", item)
}
func (h *DepartmentHandler) Create(c *fiber.Ctx) error  {
	var item master.Department
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if err := h.uc.Create(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "department created", item)
}
func (h *DepartmentHandler) Update(c *fiber.Ctx) error  {
	var item master.Department
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item.DepartmentID = c.Params("id")
	if err := h.uc.Update(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "department updated", item)
}
func (h *DepartmentHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "department deleted", nil)
}

// ── Area Handler ──────────────────────────────────────────────────────────

type AreaHandler struct{ uc master.AreaUseCase }
func NewAreaHandler(uc master.AreaUseCase) *AreaHandler { return &AreaHandler{uc: uc} }

func (h *AreaHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("search"))
	if err != nil { return response.InternalServerError(c, "failed to fetch areas", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}
func (h *AreaHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "area not found") }
	return response.OK(c, "success", item)
}
func (h *AreaHandler) Create(c *fiber.Ctx) error  {
	var item master.Area
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if err := h.uc.Create(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "area created", item)
}
func (h *AreaHandler) Update(c *fiber.Ctx) error  {
	var item master.Area
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item.AreaID = c.Params("id")
	if err := h.uc.Update(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "area updated", item)
}
func (h *AreaHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "area deleted", nil)
}

// ── Kawasan Handler ───────────────────────────────────────────────────────

type KawasanHandler struct{ uc master.KawasanUseCase }
func NewKawasanHandler(uc master.KawasanUseCase) *KawasanHandler { return &KawasanHandler{uc: uc} }

func (h *KawasanHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("area_id"), c.Query("search"))
	if err != nil { return response.InternalServerError(c, "failed to fetch kawasans", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}
func (h *KawasanHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "kawasan not found") }
	return response.OK(c, "success", item)
}
func (h *KawasanHandler) Create(c *fiber.Ctx) error  {
	var item master.Kawasan
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if err := h.uc.Create(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "kawasan created", item)
}
func (h *KawasanHandler) Update(c *fiber.Ctx) error  {
	var item master.Kawasan
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item.KawasanID = c.Params("id")
	if err := h.uc.Update(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "kawasan updated", item)
}
func (h *KawasanHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "kawasan deleted", nil)
}

// ── DetailKawasan Handler ─────────────────────────────────────────────────

type DetailKawasanHandler struct{ uc master.DetailKawasanUseCase }
func NewDetailKawasanHandler(uc master.DetailKawasanUseCase) *DetailKawasanHandler { return &DetailKawasanHandler{uc: uc} }

func (h *DetailKawasanHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("kawasan_id"), c.Query("search"))
	if err != nil { return response.InternalServerError(c, "failed to fetch detail kawasans", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}
func (h *DetailKawasanHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "detail kawasan not found") }
	return response.OK(c, "success", item)
}
func (h *DetailKawasanHandler) Create(c *fiber.Ctx) error  {
	var item master.DetailKawasan
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if err := h.uc.Create(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "detail kawasan created", item)
}
func (h *DetailKawasanHandler) Update(c *fiber.Ctx) error  {
	var item master.DetailKawasan
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item.DetailKawasanID = c.Params("id")
	if err := h.uc.Update(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "detail kawasan updated", item)
}
func (h *DetailKawasanHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "detail kawasan deleted", nil)
}

// ── Aspek Handler ─────────────────────────────────────────────────────────

type AspekHandler struct{ uc master.AspekUseCase }
func NewAspekHandler(uc master.AspekUseCase) *AspekHandler { return &AspekHandler{uc: uc} }

func (h *AspekHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("area_id"), c.Query("search"))
	if err != nil { return response.InternalServerError(c, "failed to fetch aspeks", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}
func (h *AspekHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "aspek not found") }
	return response.OK(c, "success", item)
}
func (h *AspekHandler) Create(c *fiber.Ctx) error  {
	var item master.Aspek
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if err := h.uc.Create(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "aspek created", item)
}
func (h *AspekHandler) Update(c *fiber.Ctx) error  {
	var item master.Aspek
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item.AspekID = c.Params("id")
	if err := h.uc.Update(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "aspek updated", item)
}
func (h *AspekHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "aspek deleted", nil)
}

// ── Detail Handler ────────────────────────────────────────────────────────

type DetailHandler struct{ uc master.DetailUseCase }
func NewDetailHandler(uc master.DetailUseCase) *DetailHandler { return &DetailHandler{uc: uc} }

func (h *DetailHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("aspek_id"), c.Query("search"))
	if err != nil { return response.InternalServerError(c, "failed to fetch details", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}
func (h *DetailHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "detail not found") }
	return response.OK(c, "success", item)
}
func (h *DetailHandler) Create(c *fiber.Ctx) error  {
	var item master.Detail
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if err := h.uc.Create(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "detail created", item)
}
func (h *DetailHandler) Update(c *fiber.Ctx) error  {
	var item master.Detail
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item.DetailID = c.Params("id")
	if err := h.uc.Update(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "detail updated", item)
}
func (h *DetailHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "detail deleted", nil)
}

// ── Uraian Handler ────────────────────────────────────────────────────────

type UraianHandler struct{ uc master.UraianUseCase }
func NewUraianHandler(uc master.UraianUseCase) *UraianHandler { return &UraianHandler{uc: uc} }

func (h *UraianHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("detail_id"), c.Query("search"))
	if err != nil { return response.InternalServerError(c, "failed to fetch urains", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}
func (h *UraianHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "uraian not found") }
	return response.OK(c, "success", item)
}
func (h *UraianHandler) Create(c *fiber.Ctx) error  {
	var item master.Uraian
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if err := h.uc.Create(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "uraian created", item)
}
func (h *UraianHandler) Update(c *fiber.Ctx) error  {
	var item master.Uraian
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item.UraianID = c.Params("id")
	if err := h.uc.Update(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "uraian updated", item)
}
func (h *UraianHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "uraian deleted", nil)
}
