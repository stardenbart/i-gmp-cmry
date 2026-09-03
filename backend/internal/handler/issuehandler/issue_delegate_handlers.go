package issuehandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/issue"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/response"
)

type IssueDelegateHandler struct {
	uc      issue.IssueDelegateUseCase
	issueUC issue.IssueUseCase
}

func NewIssueDelegateHandler(uc issue.IssueDelegateUseCase, issueUC issue.IssueUseCase) *IssueDelegateHandler {
	return &IssueDelegateHandler{uc: uc, issueUC: issueUC}
}

// @Summary      Add Issue Delegate
// @Description  Add a user as a delegate to an issue to allow them to follow up
// @Tags         Issues
// @Accept       json
// @Produce      json
// @Param        id path string true "Issue ID"
// @Param        body body interface{} true "Delegate Request"
// @Success      201 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /issues/{id}/delegates [post]
func (h *IssueDelegateHandler) AddDelegate(c *fiber.Ctx) error {
	issueID := c.Params("id")
	if _, err := checkIssueAccess(c, h.issueUC, issueID); err != nil {
		return err
	}
	actorID := middleware.GetUserID(c)

	var req issue.AddIssueDelegateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "invalid request body", err.Error())
	}

	if err := h.uc.AddDelegate(issueID, actorID, &req); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}

	return response.Created(c, "delegate added successfully", nil)
}

// @Summary      Remove Issue Delegate
// @Description  Remove a user from issue delegates
// @Tags         Issues
// @Accept       json
// @Produce      json
// @Param        id path string true "Issue ID"
// @Param        user_id path string true "Delegate User ID"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /issues/{id}/delegates/{user_id} [delete]
func (h *IssueDelegateHandler) RemoveDelegate(c *fiber.Ctx) error {
	issueID := c.Params("id")
	if _, err := checkIssueAccess(c, h.issueUC, issueID); err != nil {
		return err
	}
	delegateUserID := c.Params("user_id")
	actorID := middleware.GetUserID(c)

	if err := h.uc.RemoveDelegate(issueID, delegateUserID, actorID); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}

	return response.OK(c, "delegate removed successfully", nil)
}

// @Summary      Get Issue Delegates
// @Description  List all delegates for an issue
// @Tags         Issues
// @Accept       json
// @Produce      json
// @Param        id path string true "Issue ID"
// @Success      200 {object} response.APIResponse
// @Failure      400 {object} response.APIResponse
// @Security     BearerAuth
// @Router       /issues/{id}/delegates [get]
func (h *IssueDelegateHandler) GetDelegates(c *fiber.Ctx) error {
	issueID := c.Params("id")
	if _, err := checkIssueAccess(c, h.issueUC, issueID); err != nil {
		return err
	}
	delegates, err := h.uc.GetDelegatesByIssue(issueID)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}
	return response.OK(c, "success", delegates)
}
