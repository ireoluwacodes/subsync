-- +goose Up
CREATE TABLE tenants (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name                  TEXT NOT NULL,
  email                 TEXT NOT NULL UNIQUE,
  nomba_account_id      TEXT NOT NULL,
  nomba_sub_account_id  TEXT NOT NULL UNIQUE,
  api_key_hash          TEXT NOT NULL,
  webhook_secret        TEXT NOT NULL,
  dunning_config        JSONB NOT NULL DEFAULT '{
    "steps": [
      {"delay_days": 1,  "action": "retry"},
      {"delay_days": 3,  "action": "retry_and_notify"},
      {"delay_days": 7,  "action": "mandate_fallback"},
      {"delay_days": 14, "action": "cancel"}
    ]
  }',
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS tenants;
