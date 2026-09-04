package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestCSRFMiddlewareAllowsPublicAuthWithStaleSession(t *testing.T) {
	paths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/forgot-password",
		"/api/v1/auth/reset-password",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			app := fiber.New()
			app.Use(CSRFMiddleware())
			app.Post(path, func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusNoContent)
			})

			req := httptest.NewRequest(fiber.MethodPost, path, nil)
			req.Header.Set("Cookie", CookieAccessToken+"=stale-access-token")
			res, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if res.StatusCode != fiber.StatusNoContent {
				t.Fatalf("expected %d, got %d", fiber.StatusNoContent, res.StatusCode)
			}
		})
	}
}

func TestCSRFMiddlewareStillProtectsAuthenticatedMutation(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFMiddleware())
	app.Post("/api/v1/issues", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/issues", nil)
	req.Header.Set("Cookie", CookieAccessToken+"=access; "+CookieCSRFToken+"=expected")
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected %d, got %d", fiber.StatusForbidden, res.StatusCode)
	}
}

func TestCSRFMiddlewareAllowsMatchingDoubleSubmitToken(t *testing.T) {
	app := fiber.New()
	app.Use(CSRFMiddleware())
	app.Post("/api/v1/issues", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/issues", nil)
	req.Header.Set("Cookie", CookieAccessToken+"=access; "+CookieCSRFToken+"=matching-token")
	req.Header.Set(HeaderCSRFToken, "matching-token")
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected %d, got %d", fiber.StatusNoContent, res.StatusCode)
	}
}
