package jobs

import (
	"fmt"

	"github.com/google/uuid"
)

func parseIDs(tenantRaw, idRaw string) (uuid.UUID, uuid.UUID, error) {
	tenantID, err := uuid.Parse(tenantRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid tenant_id: %w", err)
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid id: %w", err)
	}
	return tenantID, id, nil
}
