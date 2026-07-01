package nomba

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrRetryable = errors.New("nomba: retryable error")
	ErrTerminal  = errors.New("nomba: terminal error")
)

// HTTPError wraps a Nomba API error with the HTTP status code.
type HTTPError struct {
	StatusCode  int
	Code        string
	Description string
}

func (e *HTTPError) Error() string {
	if e.Description != "" {
		return e.Description
	}
	return fmt.Sprintf("nomba api error (http %d): %s", e.StatusCode, e.Code)
}

func (e *HTTPError) Unwrap() error {
	return ClassifyError(e.StatusCode)
}

func NewHTTPError(statusCode int, apiErr APIError) *HTTPError {
	return &HTTPError{
		StatusCode:  statusCode,
		Code:        apiErr.Code,
		Description: apiErr.Description,
	}
}

func ClassifyError(statusCode int) error {
	if statusCode == http.StatusServiceUnavailable || statusCode == http.StatusTooManyRequests {
		return ErrRetryable
	}
	return ErrTerminal
}
