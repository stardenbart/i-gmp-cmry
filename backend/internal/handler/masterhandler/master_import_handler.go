package masterhandler

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/masterusecase"
	"github.com/monitoring-system/backend/pkg/masterimportxlsx"
	"github.com/monitoring-system/backend/pkg/response"
)

type MasterImportHandler struct {
	uc *masterusecase.MasterImportUseCase
}

func NewMasterImportHandler(uc *masterusecase.MasterImportUseCase) *MasterImportHandler {
	return &MasterImportHandler{uc: uc}
}

func (h *MasterImportHandler) DownloadTemplate(c *fiber.Ctx) error {
	actor := importActor(c)
	content, fileName, err := h.uc.GenerateTemplate(c.UserContext(), c.Params("type"), actor, c.Query("plant_id"))
	if err != nil {
		return writeMasterImportError(c, err)
	}
	c.Set(fiber.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, fileName))
	return c.Send(content)
}

func (h *MasterImportHandler) Validate(c *fiber.Ctx) error {
	actor := importActor(c)
	if err := h.uc.CheckRateLimit(c.UserContext(), actor.UserID); err != nil {
		return writeMasterImportError(c, err)
	}
	fileName, content, uploadStatus, err := readImportFile(c)
	if err != nil {
		return c.Status(uploadStatus).JSON(response.APIResponse{Success: false, StatusCode: uploadStatus, Message: err.Error()})
	}
	preview, err := h.uc.Validate(c.UserContext(), c.Params("type"), actor, c.FormValue("plant_id"), fileName, content)
	if err != nil {
		return writeMasterImportError(c, err)
	}
	return response.OK(c, "validasi import berhasil", preview)
}

func (h *MasterImportHandler) Commit(c *fiber.Ctx) error {
	fileName, content, uploadStatus, err := readImportFile(c)
	if err != nil {
		return c.Status(uploadStatus).JSON(response.APIResponse{Success: false, StatusCode: uploadStatus, Message: err.Error()})
	}
	result, err := h.uc.Commit(
		c.UserContext(), c.Params("type"), importActor(c), c.FormValue("plant_id"),
		c.FormValue("validation_token"), fileName, content,
	)
	if err != nil {
		return writeMasterImportError(c, err)
	}
	return response.Created(c, "seluruh data master berhasil diimport", result)
}

func readImportFile(c *fiber.Ctx) (string, []byte, int, error) {
	header, err := c.FormFile("file")
	if err != nil {
		return "", nil, http.StatusBadRequest, errors.New("file import wajib diunggah")
	}
	if strings.ToLower(filepath.Ext(header.Filename)) != ".xlsx" {
		return "", nil, http.StatusUnsupportedMediaType, errors.New("hanya file .xlsx yang diperbolehkan")
	}
	if header.Size > master.ImportMaxFileSize {
		return "", nil, http.StatusRequestEntityTooLarge, fmt.Errorf("ukuran file melebihi batas %d MB", master.ImportMaxFileSize/(1024*1024))
	}
	file, err := header.Open()
	if err != nil {
		return "", nil, http.StatusBadRequest, errors.New("file import tidak dapat dibuka")
	}
	defer file.Close()
	content, err := masterimportxlsx.ReadLimited(file)
	if err != nil {
		return "", nil, http.StatusRequestEntityTooLarge, err
	}
	return header.Filename, content, http.StatusOK, nil
}

func importActor(c *fiber.Ctx) master.ImportActor {
	plantID, _ := c.Locals("userPlantID").(string)
	return master.ImportActor{
		UserID:  middleware.GetUserID(c),
		RoleID:  middleware.GetRoleID(c),
		PlantID: plantID,
	}
}

func writeMasterImportError(c *fiber.Ctx, err error) error {
	var importError *masterusecase.MasterImportError
	if !errors.As(err, &importError) {
		return response.InternalServerError(c, "proses import master data gagal", nil)
	}
	status := importError.Status
	if status == 0 {
		status = http.StatusInternalServerError
	}
	return c.Status(status).JSON(response.APIResponse{
		Success: false, StatusCode: status, Message: importError.Message,
		Data: importError.Data,
	})
}
