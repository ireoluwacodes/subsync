package dto

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ireoluwacodes/subsync/internal/domain"
	"github.com/ireoluwacodes/subsync/internal/nomba"
)

type APIError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type Meta struct {
	RequestID string `json:"request_id,omitempty"`
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"per_page,omitempty"`
	Total     int64  `json:"total,omitempty"`
}

type Envelope struct {
	Data  any       `json:"data"`
	Meta  Meta      `json:"meta"`
	Error *APIError `json:"error"`
}

func requestID(c *gin.Context) string {
	if id, ok := c.Get("request_id"); ok {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}

func RespondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{
		Data: data,
		Meta: Meta{RequestID: requestID(c)},
	})
}

func RespondCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{
		Data: data,
		Meta: Meta{RequestID: requestID(c)},
	})
}

func RespondError(c *gin.Context, err error) {
	apiErr, status := mapError(err)
	if status >= http.StatusInternalServerError {
		zap.L().Error("request failed",
			zap.String("request_id", requestID(c)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Error(err),
		)
	}
	c.JSON(status, Envelope{
		Meta:  Meta{RequestID: requestID(c)},
		Error: apiErr,
	})
}

func RespondWithStatus(c *gin.Context, status int, data any) {
	c.JSON(status, Envelope{
		Data: data,
		Meta: Meta{RequestID: requestID(c)},
	})
}

func mapError(err error) (*APIError, int) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return &APIError{Code: "not_found", Message: err.Error()}, http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		return &APIError{Code: "conflict", Message: err.Error()}, http.StatusConflict
	case errors.Is(err, domain.ErrInvalidTransition):
		return &APIError{Code: "transition_not_allowed", Message: err.Error()}, http.StatusUnprocessableEntity
	case errors.Is(err, domain.ErrValidation):
		return &APIError{Code: "validation_failed", Message: err.Error()}, http.StatusUnprocessableEntity
	case errors.Is(err, domain.ErrInvalidNombaCredentials):
		return &APIError{Code: "invalid_nomba_credentials", Message: err.Error()}, http.StatusUnprocessableEntity
	case errors.Is(err, domain.ErrUnauthorized):
		return &APIError{Code: "unauthorized", Message: err.Error()}, http.StatusUnauthorized
	case errors.Is(err, domain.ErrNotImplemented):
		return &APIError{Code: "not_implemented", Message: err.Error()}, http.StatusNotImplemented
	default:
		var nombaErr *nomba.HTTPError
		if errors.As(err, &nombaErr) {
			msg := nombaErr.Error()
			if nombaErr.StatusCode >= 500 {
				return &APIError{Code: "nomba_error", Message: msg}, http.StatusBadGateway
			}
			return &APIError{Code: "nomba_error", Message: msg}, http.StatusUnprocessableEntity
		}
		var bindErr *BindError
		if errors.As(err, &bindErr) {
			return &APIError{Code: "invalid_request", Message: bindErr.Error(), Details: bindErr.Details}, http.StatusBadRequest
		}
		return &APIError{Code: "internal_error", Message: "an unexpected error occurred"}, http.StatusInternalServerError
	}
}

type BindError struct {
	Message string
	Details map[string]any
}

func (e *BindError) Error() string {
	return e.Message
}

func NewBindError(msg string) *BindError {
	return &BindError{Message: msg}
}

// PaginationParams holds common list query params.
type PaginationParams struct {
	Page    int `form:"page"`
	PerPage int `form:"per_page"`
}

func (p *PaginationParams) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 100 {
		p.PerPage = 20
	}
}

// IDParam parses a UUID path param.
func IDParam(c *gin.Context, name string) (uuid.UUID, error) {
	raw := c.Param(name)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, &BindError{Message: "invalid " + name}
	}
	return id, nil
}
