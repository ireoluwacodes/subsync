-- +goose Up
-- Repair tenants whose created_at was zeroed by a prior update bug (stored as year 0001).
-- No immutable creation-time record exists for a tenant, so recover the best available lower
-- bound: the earliest created_at among the tenant's users, falling back to the row's own updated_at.
UPDATE tenants t
SET created_at = COALESCE(
	(
		SELECT MIN(u.created_at)
		FROM users u
		WHERE u.tenant_id = t.id
		  AND u.created_at > '1970-01-01'
	),
	t.updated_at
)
WHERE t.created_at < '1970-01-01';

-- +goose Down
-- Irreversible data repair; no rollback.
SELECT 1;
