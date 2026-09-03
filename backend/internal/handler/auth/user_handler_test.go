package auth

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/internal/middleware"
)

// fakeUserUC is a minimal authdomain.UserUseCase stand-in that just records
// what Update() was actually called with, so the tests below can assert on
// exactly what reached the business layer after the handler's field-strip.
type fakeUserUC struct {
	updateID  string
	updateReq *authdomain.UpdateUserRequest
}

func (f *fakeUserUC) GetAll(int, int, string, string, string, string) ([]authdomain.User, int64, error) {
	return nil, 0, nil
}
func (f *fakeUserUC) GetByID(string) (*authdomain.User, error) { return &authdomain.User{}, nil }
func (f *fakeUserUC) Create(*authdomain.CreateUserRequest) (*authdomain.User, error) {
	return &authdomain.User{}, nil
}
func (f *fakeUserUC) Update(id string, req *authdomain.UpdateUserRequest) (*authdomain.User, error) {
	f.updateID = id
	f.updateReq = req
	return &authdomain.User{UserID: id}, nil
}
func (f *fakeUserUC) Delete(string) error { return nil }
func (f *fakeUserUC) ChangePassword(string, *authdomain.ChangePasswordRequest) error {
	return nil
}
func (f *fakeUserUC) AdminResetPassword(string, *authdomain.AdminResetPasswordRequest) error {
	return nil
}
func (f *fakeUserUC) ForgotPassword(*authdomain.ForgotPasswordRequest) error { return nil }
func (f *fakeUserUC) ResetPasswordWithOTP(*authdomain.ResetPasswordWithOTPRequest) error {
	return nil
}

// TestUpdateStripsEscalationFieldsOnSelfUpdate guards the fix for a critical
// privilege-escalation bug: router/v1/auth_routes.go lets PUT /users/:id
// reach this handler WITHOUT the MOD-USR UPDATE permission whenever the
// caller's own id matches :id (so anyone can edit their own name/email).
// Without the strip in UserHandler.Update, that bypass let any authenticated
// user grant themselves e.g. Super Admin by simply including role_id in
// their own profile update request.
func TestUpdateStripsEscalationFieldsOnSelfUpdate(t *testing.T) {
	uc := &fakeUserUC{}
	h := NewUserHandler(uc)

	app := fiber.New()
	app.Put("/users/:id", func(c *fiber.Ctx) error {
		// Simulates AuthMiddleware having authenticated "USR-1" — the same
		// user this request targets, i.e. exactly the self-update path that
		// bypasses PermissionMiddleware in the real route.
		c.Locals(middleware.ContextKeyUserID, "USR-1")
		return c.Next()
	}, h.Update)

	body, _ := json.Marshal(map[string]string{
		"role_id":       "ROLE-000",
		"user_status":   "Active",
		"department_id": "DEPT-EVIL",
		"full_name":     "Legit Name Change",
	})
	req := httptest.NewRequest("PUT", "/users/USR-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if uc.updateReq == nil {
		t.Fatal("usecase.Update was never called")
	}
	if uc.updateReq.RoleID != "" {
		t.Fatalf("self-update must not be able to change role_id, got %q", uc.updateReq.RoleID)
	}
	if uc.updateReq.UserStatus != "" {
		t.Fatalf("self-update must not be able to change user_status, got %q", uc.updateReq.UserStatus)
	}
	if uc.updateReq.DepartmentID != "" {
		t.Fatalf("self-update must not be able to change department_id, got %q", uc.updateReq.DepartmentID)
	}
	if uc.updateReq.FullName != "Legit Name Change" {
		t.Fatal("self-update must still be able to change full_name — the whole point of the bypass")
	}
}

// TestUpdateAllowsRoleChangeForNonSelfUpdate is the control case: an admin
// updating a DIFFERENT user's id (the only way to reach this handler for a
// non-matching id, since the route requires MOD-USR UPDATE permission in
// that branch) must still be able to set role_id — the fix must not have
// accidentally broken real admin role management.
func TestUpdateAllowsRoleChangeForNonSelfUpdate(t *testing.T) {
	uc := &fakeUserUC{}
	h := NewUserHandler(uc)

	app := fiber.New()
	app.Put("/users/:id", func(c *fiber.Ctx) error {
		c.Locals(middleware.ContextKeyUserID, "USR-ADMIN")
		return c.Next()
	}, h.Update)

	body, _ := json.Marshal(map[string]string{"role_id": "ROLE-002"})
	req := httptest.NewRequest("PUT", "/users/USR-TARGET", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if uc.updateReq == nil || uc.updateReq.RoleID != "ROLE-002" {
		t.Fatal("an admin updating a DIFFERENT user's id must still be able to set role_id")
	}
}
