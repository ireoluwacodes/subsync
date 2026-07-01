-- +goose Up
ALTER TABLE tenants
  ADD COLUMN IF NOT EXISTS nomba_client_id TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS nomba_client_secret_enc TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS nomba_env TEXT NOT NULL DEFAULT 'sandbox';

ALTER TABLE tenants ALTER COLUMN nomba_sub_account_id DROP NOT NULL;

ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_nomba_sub_account_id_key;

-- +goose Down
ALTER TABLE tenants
  DROP COLUMN IF EXISTS nomba_client_id,
  DROP COLUMN IF EXISTS nomba_client_secret_enc,
  DROP COLUMN IF EXISTS nomba_env;

ALTER TABLE tenants ALTER COLUMN nomba_sub_account_id SET NOT NULL;
ALTER TABLE tenants ADD CONSTRAINT tenants_nomba_sub_account_id_key UNIQUE (nomba_sub_account_id);
