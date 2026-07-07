package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/config"
)

const refreshTokenCookieName = "refresh_token"

func cookieSecure(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}
	return c.GetHeader("X-Forwarded-Proto") == "https"
}

// cookieSameSite returns the SameSite policy for auth cookies. When the request
// is secure (HTTPS) we use SameSite=None so the cookie is sent on cross-site
// requests from a frontend hosted on a different origin. SameSite=None is only
// valid alongside Secure, so we fall back to Lax on plain HTTP (local dev).
func cookieSameSite(secure bool) http.SameSite {
	if secure {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func setRefreshTokenCookie(c *gin.Context, cfg *config.Config, token string) {
	maxAge := int(cfg.JWTRefreshTTL.Seconds())
	if maxAge <= 0 {
		maxAge = 0
	}
	secure := cookieSecure(c)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: cookieSameSite(secure),
	})
}

func clearRefreshTokenCookie(c *gin.Context) {
	secure := cookieSecure(c)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: cookieSameSite(secure),
	})
}

func refreshTokenFromRequest(c *gin.Context) string {
	if token, err := c.Cookie(refreshTokenCookieName); err == nil && token != "" {
		return token
	}
	return c.Query("refresh_token")
}
