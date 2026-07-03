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

func setRefreshTokenCookie(c *gin.Context, cfg *config.Config, token string) {
	maxAge := int(cfg.JWTRefreshTTL.Seconds())
	if maxAge <= 0 {
		maxAge = 0
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   cookieSecure(c),
		SameSite: http.SameSiteLaxMode,
	})
}

func clearRefreshTokenCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cookieSecure(c),
		SameSite: http.SameSiteLaxMode,
	})
}

func refreshTokenFromRequest(c *gin.Context) string {
	if token, err := c.Cookie(refreshTokenCookieName); err == nil && token != "" {
		return token
	}
	return c.Query("refresh_token")
}
