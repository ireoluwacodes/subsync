-- +goose Up
CREATE TABLE customers (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id),
  external_id     TEXT,
  email           TEXT NOT NULL,
  name            TEXT,
  phone           TEXT,
  metadata        JSONB NOT NULL DEFAULT '{}',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, email)
);

CREATE INDEX customers_tenant_id ON customers(tenant_id);
CREATE INDEX customers_external_id ON customers(tenant_id, external_id);

-- +goose Down
DROP TABLE IF EXISTS customers;
