package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ireoluwacodes/subsync/internal/api/dto"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				zap.L().Error("panic recovered",
					zap.String("request_id", c.GetString("request_id")),
					zap.Any("panic", r),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.Envelope{
					Meta: dto.Meta{RequestID: c.GetString("request_id")},
					Error: &dto.APIError{
						Code:    "internal_error",
						Message: "an unexpected error occurred",
					},
				})
			}
		}()
		c.Next()
	}
}
