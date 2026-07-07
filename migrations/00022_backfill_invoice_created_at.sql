-- +goose Up
-- Repair invoices whose created_at was zeroed by a prior update bug (stored as year 0001).
-- Recover the timestamp from the earliest line item (set at invoice creation, never mutated),
-- falling back to updated_at when no usable line item timestamp exists.
UPDATE invoices i
SET created_at = COALESCE(
	(
		SELECT MIN(li.created_at)
		FROM invoice_line_items li
		WHERE li.invoice_id = i.id
		  AND li.created_at > '1970-01-01'
	),
	i.updated_at
)
WHERE i.created_at < '1970-01-01';

-- +goose Down
-- Irreversible data repair; no rollback.
SELECT 1;
