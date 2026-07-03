package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrConflict          = errors.New("conflict")
	ErrInvalidTransition = errors.New("invalid transition")
	ErrValidation        = errors.New("validation failed")
	ErrNotImplemented    = errors.New("not implemented")
	ErrUnauthorized      = errors.New("unauthorized")
)
