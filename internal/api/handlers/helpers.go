package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/api/middleware"
	"github.com/ireoluwacodes/subsync/internal/domain"
)

func middlewareUser(c *gin.Context) (*domain.User, bool) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		dto.RespondError(c, domain.ErrNotFound)
		return nil, false
	}
	return user, true
}

func middlewareTenant(c *gin.Context) (*domain.Tenant, bool) {
	tenant, ok := middleware.TenantFromContext(c)
	if !ok {
		dto.RespondError(c, domain.ErrNotFound)
		return nil, false
	}
	return tenant, true
}
