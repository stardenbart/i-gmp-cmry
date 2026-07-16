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

// @Summary Get all departments
// @Description Fetch a paginated list of departments
// @Tags Department
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search query"
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/master/departments [get]
// @Security BearerAuth
func (h *DepartmentHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("search"))
	if err != nil { return response.InternalServerError(c, "failed to fetch departments", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}
// @Summary Get a department by ID
// @Description Fetch a single department by its ID
// @Tags Department
// @Accept json
// @Produce json
// @Param id path string true "Department ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/master/departments/{id} [get]
// @Security BearerAuth
func (h *DepartmentHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "department not found") }
	return response.OK(c, "success", item)
}
// @Summary Create a department
// @Description Create a new department
// @Tags Department
// @Accept json
// @Produce json
// @Param request body master.Department true "Department Request"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/departments [post]
// @Security BearerAuth
func (h *DepartmentHandler) Create(c *fiber.Ctx) error  {
	var item master.Department
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if err := h.uc.Create(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "department created", item)
}
// @Summary Update a department
// @Description Update an existing department by its ID
// @Tags Department
// @Accept json
// @Produce json
// @Param id path string true "Department ID"
// @Param request body master.Department true "Department Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/departments/{id} [put]
// @Security BearerAuth
func (h *DepartmentHandler) Update(c *fiber.Ctx) error  {
	var item master.Department
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item.DepartmentID = c.Params("id")
	if err := h.uc.Update(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "department updated", item)
}
// @Summary Delete a department
// @Description Delete a department by its ID
// @Tags Department
// @Accept json
// @Produce json
// @Param id path string true "Department ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/departments/{id} [delete]
// @Security BearerAuth
func (h *DepartmentHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "department deleted", nil)
}

// ── Area Handler ──────────────────────────────────────────────────────────

type AreaHandler struct{ uc master.AreaUseCase }
func NewAreaHandler(uc master.AreaUseCase) *AreaHandler { return &AreaHandler{uc: uc} }

// @Summary Get all areas
// @Description Fetch a paginated list of areas
// @Tags Area
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search query"
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/master/area [get]
// @Security BearerAuth
func (h *AreaHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("search"))
	if err != nil { return response.InternalServerError(c, "failed to fetch areas", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}
// @Summary Get an area by ID
// @Description Fetch a single area by its ID
// @Tags Area
// @Accept json
// @Produce json
// @Param id path string true "Area ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/master/area/{id} [get]
// @Security BearerAuth
func (h *AreaHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "area not found") }
	return response.OK(c, "success", item)
}
// @Summary Create an area
// @Description Create a new area
// @Tags Area
// @Accept json
// @Produce json
// @Param request body master.Area true "Area Request"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/area [post]
// @Security BearerAuth
func (h *AreaHandler) Create(c *fiber.Ctx) error  {
	var item master.Area
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if err := h.uc.Create(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "area created", item)
}
// @Summary Update an area
// @Description Update an existing area by its ID
// @Tags Area
// @Accept json
// @Produce json
// @Param id path string true "Area ID"
// @Param request body master.Area true "Area Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/area/{id} [put]
// @Security BearerAuth
func (h *AreaHandler) Update(c *fiber.Ctx) error  {
	var item master.Area
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item.AreaID = c.Params("id")
	if err := h.uc.Update(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "area updated", item)
}
// @Summary Delete an area
// @Description Delete an area by its ID
// @Tags Area
// @Accept json
// @Produce json
// @Param id path string true "Area ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/area/{id} [delete]
// @Security BearerAuth
func (h *AreaHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "area deleted", nil)
}

// ── Kawasan Handler ───────────────────────────────────────────────────────

type KawasanHandler struct{ uc master.KawasanUseCase }
func NewKawasanHandler(uc master.KawasanUseCase) *KawasanHandler { return &KawasanHandler{uc: uc} }

// @Summary Get all kawasans
// @Description Fetch a paginated list of kawasans
// @Tags Kawasan
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param area_id query string false "Area ID filter"
// @Param search query string false "Search query"
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/master/kawasan [get]
// @Security BearerAuth
func (h *KawasanHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("area_id"), c.Query("search"))
	if err != nil { return response.InternalServerError(c, "failed to fetch kawasans", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}
// @Summary Get a kawasan by ID
// @Description Fetch a single kawasan by its ID
// @Tags Kawasan
// @Accept json
// @Produce json
// @Param id path string true "Kawasan ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/master/kawasan/{id} [get]
// @Security BearerAuth
func (h *KawasanHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "kawasan not found") }
	return response.OK(c, "success", item)
}
// @Summary Create a kawasan
// @Description Create a new kawasan
// @Tags Kawasan
// @Accept json
// @Produce json
// @Param request body master.Kawasan true "Kawasan Request"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/kawasan [post]
// @Security BearerAuth
func (h *KawasanHandler) Create(c *fiber.Ctx) error  {
	var item master.Kawasan
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if err := h.uc.Create(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "kawasan created", item)
}
// @Summary Update a kawasan
// @Description Update an existing kawasan by its ID
// @Tags Kawasan
// @Accept json
// @Produce json
// @Param id path string true "Kawasan ID"
// @Param request body master.Kawasan true "Kawasan Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/kawasan/{id} [put]
// @Security BearerAuth
func (h *KawasanHandler) Update(c *fiber.Ctx) error  {
	var item master.Kawasan
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item.KawasanID = c.Params("id")
	if err := h.uc.Update(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "kawasan updated", item)
}
// @Summary Delete a kawasan
// @Description Delete a kawasan by its ID
// @Tags Kawasan
// @Accept json
// @Produce json
// @Param id path string true "Kawasan ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/kawasan/{id} [delete]
// @Security BearerAuth
func (h *KawasanHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "kawasan deleted", nil)
}

// ── DetailKawasan Handler ─────────────────────────────────────────────────

type DetailKawasanHandler struct{ uc master.DetailKawasanUseCase }
func NewDetailKawasanHandler(uc master.DetailKawasanUseCase) *DetailKawasanHandler { return &DetailKawasanHandler{uc: uc} }

// @Summary Get all detail kawasans
// @Description Fetch a paginated list of detail kawasans
// @Tags DetailKawasan
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param kawasan_id query string false "Kawasan ID filter"
// @Param search query string false "Search query"
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/master/detail-kawasan [get]
// @Security BearerAuth
func (h *DetailKawasanHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("kawasan_id"), c.Query("search"))
	if err != nil { return response.InternalServerError(c, "failed to fetch detail kawasans", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}
// @Summary Get a detail kawasan by ID
// @Description Fetch a single detail kawasan by its ID
// @Tags DetailKawasan
// @Accept json
// @Produce json
// @Param id path string true "Detail Kawasan ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/master/detail-kawasan/{id} [get]
// @Security BearerAuth
func (h *DetailKawasanHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "detail kawasan not found") }
	return response.OK(c, "success", item)
}
// @Summary Create a detail kawasan
// @Description Create a new detail kawasan
// @Tags DetailKawasan
// @Accept json
// @Produce json
// @Param request body master.DetailKawasan true "Detail Kawasan Request"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/detail-kawasan [post]
// @Security BearerAuth
func (h *DetailKawasanHandler) Create(c *fiber.Ctx) error  {
	var item master.DetailKawasan
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if err := h.uc.Create(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "detail kawasan created", item)
}
// @Summary Update a detail kawasan
// @Description Update an existing detail kawasan by its ID
// @Tags DetailKawasan
// @Accept json
// @Produce json
// @Param id path string true "Detail Kawasan ID"
// @Param request body master.DetailKawasan true "Detail Kawasan Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/detail-kawasan/{id} [put]
// @Security BearerAuth
func (h *DetailKawasanHandler) Update(c *fiber.Ctx) error  {
	var item master.DetailKawasan
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item.DetailKawasanID = c.Params("id")
	if err := h.uc.Update(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "detail kawasan updated", item)
}
// @Summary Delete a detail kawasan
// @Description Delete a detail kawasan by its ID
// @Tags DetailKawasan
// @Accept json
// @Produce json
// @Param id path string true "Detail Kawasan ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/detail-kawasan/{id} [delete]
// @Security BearerAuth
func (h *DetailKawasanHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "detail kawasan deleted", nil)
}

// ── Aspek Handler ─────────────────────────────────────────────────────────

type AspekHandler struct{ uc master.AspekUseCase }
func NewAspekHandler(uc master.AspekUseCase) *AspekHandler { return &AspekHandler{uc: uc} }

// @Summary Get all aspeks
// @Description Fetch a paginated list of aspeks
// @Tags Aspek
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param area_id query string false "Area ID filter"
// @Param search query string false "Search query"
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/master/aspek [get]
// @Security BearerAuth
func (h *AspekHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("area_id"), c.Query("search"))
	if err != nil { return response.InternalServerError(c, "failed to fetch aspeks", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}
// @Summary Get an aspek by ID
// @Description Fetch a single aspek by its ID
// @Tags Aspek
// @Accept json
// @Produce json
// @Param id path string true "Aspek ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/master/aspek/{id} [get]
// @Security BearerAuth
func (h *AspekHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "aspek not found") }
	return response.OK(c, "success", item)
}
// @Summary Create an aspek
// @Description Create a new aspek
// @Tags Aspek
// @Accept json
// @Produce json
// @Param request body master.Aspek true "Aspek Request"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/aspek [post]
// @Security BearerAuth
func (h *AspekHandler) Create(c *fiber.Ctx) error  {
	var item master.Aspek
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if err := h.uc.Create(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "aspek created", item)
}
// @Summary Update an aspek
// @Description Update an existing aspek by its ID
// @Tags Aspek
// @Accept json
// @Produce json
// @Param id path string true "Aspek ID"
// @Param request body master.Aspek true "Aspek Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/aspek/{id} [put]
// @Security BearerAuth
func (h *AspekHandler) Update(c *fiber.Ctx) error  {
	var item master.Aspek
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item.AspekID = c.Params("id")
	if err := h.uc.Update(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "aspek updated", item)
}
// @Summary Delete an aspek
// @Description Delete an aspek by its ID
// @Tags Aspek
// @Accept json
// @Produce json
// @Param id path string true "Aspek ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/aspek/{id} [delete]
// @Security BearerAuth
func (h *AspekHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "aspek deleted", nil)
}

// ── Detail Handler ────────────────────────────────────────────────────────

type DetailHandler struct{ uc master.DetailUseCase }
func NewDetailHandler(uc master.DetailUseCase) *DetailHandler { return &DetailHandler{uc: uc} }

// @Summary Get all details
// @Description Fetch a paginated list of details
// @Tags Detail
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param aspek_id query string false "Aspek ID filter"
// @Param search query string false "Search query"
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/master/details [get]
// @Security BearerAuth
func (h *DetailHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("aspek_id"), c.Query("search"))
	if err != nil { return response.InternalServerError(c, "failed to fetch details", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}
// @Summary Get a detail by ID
// @Description Fetch a single detail by its ID
// @Tags Detail
// @Accept json
// @Produce json
// @Param id path string true "Detail ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/master/details/{id} [get]
// @Security BearerAuth
func (h *DetailHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "detail not found") }
	return response.OK(c, "success", item)
}
// @Summary Create a detail
// @Description Create a new detail
// @Tags Detail
// @Accept json
// @Produce json
// @Param request body master.Detail true "Detail Request"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/details [post]
// @Security BearerAuth
func (h *DetailHandler) Create(c *fiber.Ctx) error  {
	var item master.Detail
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if err := h.uc.Create(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "detail created", item)
}
// @Summary Update a detail
// @Description Update an existing detail by its ID
// @Tags Detail
// @Accept json
// @Produce json
// @Param id path string true "Detail ID"
// @Param request body master.Detail true "Detail Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/details/{id} [put]
// @Security BearerAuth
func (h *DetailHandler) Update(c *fiber.Ctx) error  {
	var item master.Detail
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item.DetailID = c.Params("id")
	if err := h.uc.Update(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "detail updated", item)
}
// @Summary Delete a detail
// @Description Delete a detail by its ID
// @Tags Detail
// @Accept json
// @Produce json
// @Param id path string true "Detail ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/details/{id} [delete]
// @Security BearerAuth
func (h *DetailHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "detail deleted", nil)
}

// ── Uraian Handler ────────────────────────────────────────────────────────

type UraianHandler struct{ uc master.UraianUseCase }
func NewUraianHandler(uc master.UraianUseCase) *UraianHandler { return &UraianHandler{uc: uc} }

// @Summary Get all uraians
// @Description Fetch a paginated list of uraians
// @Tags Uraian
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param detail_id query string false "Detail ID filter"
// @Param search query string false "Search query"
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/master/urain [get]
// @Security BearerAuth
func (h *UraianHandler) GetAll(c *fiber.Ctx) error  {
	p := pagination.FromQuery(c)
	items, total, err := h.uc.GetAll(p.Page, p.Limit, c.Query("detail_id"), c.Query("search"))
	if err != nil { return response.InternalServerError(c, "failed to fetch urains", err.Error()) }
	return response.Paginated(c, "success", items, total, p.Page, p.Limit)
}
// @Summary Get a uraian by ID
// @Description Fetch a single uraian by its ID
// @Tags Uraian
// @Accept json
// @Produce json
// @Param id path string true "Uraian ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/master/urain/{id} [get]
// @Security BearerAuth
func (h *UraianHandler) GetByID(c *fiber.Ctx) error  {
	item, err := h.uc.GetByID(c.Params("id"))
	if err != nil { return response.NotFound(c, "uraian not found") }
	return response.OK(c, "success", item)
}
// @Summary Create a uraian
// @Description Create a new uraian
// @Tags Uraian
// @Accept json
// @Produce json
// @Param request body master.Uraian true "Uraian Request"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/urain [post]
// @Security BearerAuth
func (h *UraianHandler) Create(c *fiber.Ctx) error  {
	var item master.Uraian
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	if err := h.uc.Create(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.Created(c, "uraian created", item)
}
// @Summary Update a uraian
// @Description Update an existing uraian by its ID
// @Tags Uraian
// @Accept json
// @Produce json
// @Param id path string true "Uraian ID"
// @Param request body master.Uraian true "Uraian Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/urain/{id} [put]
// @Security BearerAuth
func (h *UraianHandler) Update(c *fiber.Ctx) error  {
	var item master.Uraian
	if err := c.BodyParser(&item); err != nil { return response.BadRequest(c, "invalid body", err.Error()) }
	item.UraianID = c.Params("id")
	if err := h.uc.Update(&item); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "uraian updated", item)
}
// @Summary Delete a uraian
// @Description Delete a uraian by its ID
// @Tags Uraian
// @Accept json
// @Produce json
// @Param id path string true "Uraian ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/master/urain/{id} [delete]
// @Security BearerAuth
func (h *UraianHandler) Delete(c *fiber.Ctx) error  {
	if err := h.uc.Delete(c.Params("id")); err != nil { return response.BadRequest(c, err.Error(), nil) }
	return response.OK(c, "uraian deleted", nil)
}
