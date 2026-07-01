-- +goose Up
CREATE TABLE plans (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id),
  name            TEXT NOT NULL,
  description     TEXT,
  amount          BIGINT NOT NULL,
  currency        TEXT NOT NULL DEFAULT 'NGN',
  interval        TEXT NOT NULL,
  interval_days   INT,
  trial_days      INT NOT NULL DEFAULT 0,
  features        JSONB NOT NULL DEFAULT '[]',
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  is_archived     BOOLEAN NOT NULL DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX plans_tenant_id ON plans(tenant_id);

-- +goose Down
DROP TABLE IF EXISTS plans;
