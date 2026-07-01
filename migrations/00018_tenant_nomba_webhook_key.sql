-- +goose Up
ALTER TABLE tenants
  ADD COLUMN IF NOT EXISTS nomba_webhook_signing_key_enc TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE tenants DROP COLUMN IF EXISTS nomba_webhook_signing_key_enc;
