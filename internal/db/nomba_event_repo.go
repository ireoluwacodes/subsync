package db

import (
	"context"

	"github.com/google/uuid"
)

type NombaEvent struct {
	ID          uuid.UUID
	EventID     string
	EventType   string
	Payload     map[string]any
	Processed   bool
	ProcessedAt *string
	Error       string
}

type NombaEventRepo struct {
	db *DB
}

func NewNombaEventRepo(db *DB) *NombaEventRepo {
	return &NombaEventRepo{db: db}
}

func (r *NombaEventRepo) Create(ctx context.Context, event *NombaEvent) error {
	return nil // stub
}

func (r *NombaEventRepo) GetByEventID(ctx context.Context, eventID string) (*NombaEvent, error) {
	return nil, nil // stub
}

func (r *NombaEventRepo) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	return nil // stub
}
