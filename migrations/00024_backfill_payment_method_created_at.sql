-- +goose Up
-- Repair payment_methods whose created_at was zeroed by a prior update bug (stored as year 0001).
-- No immutable creation-time record exists for a payment method, so recover the best available
-- lower bound: the earliest created_at of any subscription that references this method (primary or
-- fallback), falling back to the row's own updated_at.
UPDATE payment_methods pm
SET created_at = COALESCE(
	(
		SELECT MIN(s.created_at)
		FROM subscriptions s
		WHERE (s.payment_method_id = pm.id OR s.fallback_payment_method_id = pm.id)
		  AND s.created_at > '1970-01-01'
	),
	pm.updated_at
)
WHERE pm.created_at < '1970-01-01';

-- +goose Down
-- Irreversible data repair; no rollback.
SELECT 1;
