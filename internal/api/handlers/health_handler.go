package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ireoluwacodes/subsync/internal/api/dto"
	"github.com/ireoluwacodes/subsync/internal/api/openapi"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/queue"
)

type HealthHandler struct {
	DB    *db.DB
	Queue *queue.Queue
}

func NewHealthHandler(database *db.DB, q *queue.Queue) *HealthHandler {
	return &HealthHandler{DB: database, Queue: q}
}

func (h *HealthHandler) Liveness(c *gin.Context) {
	dto.RespondOK(c, gin.H{"status": "ok"})
}

func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.DB.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, dto.Envelope{
			Meta: dto.Meta{RequestID: c.GetString("request_id")},
			Error: &dto.APIError{
				Code:    "service_unavailable",
				Message: "database unavailable",
			},
		})
		return
	}

	if err := h.Queue.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, dto.Envelope{
			Meta: dto.Meta{RequestID: c.GetString("request_id")},
			Error: &dto.APIError{
				Code:    "service_unavailable",
				Message: "redis unavailable",
			},
		})
		return
	}

	dto.RespondOK(c, gin.H{"status": "ready"})
}

func (h *HealthHandler) OpenAPI(c *gin.Context) {
	c.Data(http.StatusOK, "application/json", openapi.Document)
}

func NotImplemented(c *gin.Context) {
	dto.RespondError(c, domain.ErrNotImplemented)
}
