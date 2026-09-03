package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/config"
)

// authCookiePath scopes the refresh token cookie to only the auth endpoints
// that ever need to read it (/refresh to rotate, /logout to revoke). It's
// still httpOnly regardless, but a narrower Path means it's never even
// attached to the far larger set of unrelated API requests a session makes.
const authCookiePath = "/api/v1/auth"

func cookieSameSite(cfg *config.Config) string {
	switch strings.ToLower(cfg.CookieSameSite) {
	case "strict":
		return "Strict"
	case "none":
		return "None"
	default:
		return "Lax"
	}
}

// SetSessionCookies issues the three cookies that carry a login/refresh
// result to the browser: access_token and refresh_token are httpOnly (never
// readable by JS, so an XSS payload can't exfiltrate them the way it could
// from localStorage); csrf_token deliberately is not — the frontend reads it
// and echoes it back in a request header, which only same-origin JS can do
// (see CSRFMiddleware).
func SetSessionCookies(c *fiber.Ctx, cfg *config.Config, accessToken string, accessExpiresAt time.Time, refreshToken string, refreshExpiresAt time.Time, csrfToken string) {
	sameSite := cookieSameSite(cfg)

	c.Cookie(&fiber.Cookie{
		Name:     CookieAccessToken,
		Value:    accessToken,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		Expires:  accessExpiresAt,
		HTTPOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: sameSite,
	})
	c.Cookie(&fiber.Cookie{
		Name:     CookieRefreshToken,
		Value:    refreshToken,
		Path:     authCookiePath,
		Domain:   cfg.CookieDomain,
		Expires:  refreshExpiresAt,
		HTTPOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: sameSite,
	})
	c.Cookie(&fiber.Cookie{
		Name:     CookieCSRFToken,
		Value:    csrfToken,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		Expires:  refreshExpiresAt,
		HTTPOnly: false,
		Secure:   cfg.CookieSecure,
		SameSite: sameSite,
	})
}

// ClearSessionCookies expires all three session cookies. Because
// access_token/refresh_token are httpOnly, this server-side clear is the
// ONLY way to actually end a session — client JS calling
// document.cookie = "access_token=; expires=..." cannot touch them.
func ClearSessionCookies(c *fiber.Ctx, cfg *config.Config) {
	sameSite := cookieSameSite(cfg)
	expired := time.Now().Add(-24 * time.Hour)

	for _, cookie := range []struct {
		name string
		path string
	}{
		{CookieAccessToken, "/"},
		{CookieRefreshToken, authCookiePath},
		{CookieCSRFToken, "/"},
	} {
		c.Cookie(&fiber.Cookie{
			Name:     cookie.name,
			Value:    "",
			Path:     cookie.path,
			Domain:   cfg.CookieDomain,
			Expires:  expired,
			HTTPOnly: cookie.name != CookieCSRFToken,
			Secure:   cfg.CookieSecure,
			SameSite: sameSite,
		})
	}
}
