package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Email        string
	PasswordHash string
	Name         string
	TokenVersion int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	BumpTokenVersion(ctx context.Context, userID uuid.UUID) error
}

type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type PasswordResetRepository interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	GetValidByTokenHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	GetLatestValidByUserID(ctx context.Context, userID uuid.UUID) (*PasswordResetToken, error)
	InvalidateUnusedForUser(ctx context.Context, userID uuid.UUID) error
	UpdateTokenHash(ctx context.Context, id uuid.UUID, tokenHash string, expiresAt time.Time) error
	MarkUsed(ctx context.Context, id uuid.UUID) error
}
