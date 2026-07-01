-- +goose Up
ALTER TABLE tenants ADD COLUMN api_key_prefix TEXT;

UPDATE tenants SET api_key_prefix = 'legacy0000' WHERE api_key_prefix IS NULL;

ALTER TABLE tenants ALTER COLUMN api_key_prefix SET NOT NULL;

CREATE INDEX tenants_api_key_prefix ON tenants(api_key_prefix);

-- +goose Down
DROP INDEX IF EXISTS tenants_api_key_prefix;
ALTER TABLE tenants DROP COLUMN IF EXISTS api_key_prefix;
