package middleware

import (
	"crypto/subtle"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/pkg/response"
)

// HeaderCSRFToken is the request header the frontend must echo the
// csrf_token cookie's value into for any state-changing cookie-authenticated
// request.
const HeaderCSRFToken = "X-CSRF-Token"

var csrfProtectedMethods = map[string]bool{
	fiber.MethodPost:   true,
	fiber.MethodPut:    true,
	fiber.MethodPatch:  true,
	fiber.MethodDelete: true,
}

// CSRFMiddleware enforces the double-submit-cookie pattern for browser
// (cookie-authenticated) requests.
//
// Switching auth from a Bearer header stored in localStorage to an httpOnly
// cookie removes the XSS exposure of the token, but a cookie is attached by
// the browser automatically on ANY request to this origin — including one
// triggered by a malicious page the victim merely has open in another tab.
// Requiring a second, matching value that only same-origin JavaScript can
// read (csrf_token is not httpOnly) and attach as a custom header closes
// that gap: a cross-site page can trigger the request but cannot read the
// cookie to produce a matching header.
//
// Bearer-token requests are exempt: a caller that already had to obtain and
// attach a bearer secret isn't relying on the browser's ambient
// cookie-sending behavior, so it isn't exposed to CSRF the same way.
func CSRFMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !csrfProtectedMethods[c.Method()] {
			return c.Next()
		}
		if c.Get("Authorization") != "" {
			return c.Next()
		}
		// No session cookie at all — e.g. login itself, or an anonymous
		// request that some other middleware will reject anyway. Nothing
		// for CSRF to protect yet.
		if c.Cookies(CookieAccessToken) == "" && c.Cookies(CookieRefreshToken) == "" {
			return c.Next()
		}

		cookieToken := c.Cookies(CookieCSRFToken)
		headerToken := c.Get(HeaderCSRFToken)
		if cookieToken == "" || headerToken == "" ||
			subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) != 1 {
			return response.Forbidden(c, "missing or invalid CSRF token")
		}
		return c.Next()
	}
}
