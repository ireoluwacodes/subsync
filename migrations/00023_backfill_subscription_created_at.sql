-- +goose Up
-- Repair subscriptions whose created_at was zeroed by a prior update bug (stored as year 0001).
-- Recover the timestamp from the earliest state transition (recorded at creation and never
-- mutated), falling back to current_period_start, then updated_at.
UPDATE subscriptions s
SET created_at = COALESCE(
	(
		SELECT MIN(t.created_at)
		FROM subscription_transitions t
		WHERE t.subscription_id = s.id
		  AND t.created_at > '1970-01-01'
	),
	NULLIF(s.current_period_start, '0001-01-01 00:00:00+00'),
	s.updated_at
)
WHERE s.created_at < '1970-01-01';

-- +goose Down
-- Irreversible data repair; no rollback.
SELECT 1;
