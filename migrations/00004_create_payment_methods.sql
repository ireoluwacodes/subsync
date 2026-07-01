-- +goose Up
CREATE TABLE payment_methods (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id),
  customer_id     UUID NOT NULL REFERENCES customers(id),
  type            TEXT NOT NULL,
  token_key       TEXT,
  mandate_id      TEXT,
  card_last4      TEXT,
  card_brand      TEXT,
  card_expiry     TEXT,
  is_default      BOOLEAN NOT NULL DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX payment_methods_customer ON payment_methods(customer_id);
CREATE INDEX payment_methods_tenant ON payment_methods(tenant_id);

-- +goose Down
DROP TABLE IF EXISTS payment_methods;
