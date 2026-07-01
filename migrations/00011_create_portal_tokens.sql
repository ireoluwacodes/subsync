-- +goose Up
CREATE TABLE portal_tokens (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id),
  subscription_id UUID NOT NULL REFERENCES subscriptions(id),
  customer_id     UUID NOT NULL REFERENCES customers(id),
  token_hash      TEXT NOT NULL UNIQUE,
  expires_at      TIMESTAMPTZ NOT NULL,
  used_at         TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX portal_tokens_hash ON portal_tokens(token_hash);

-- +goose Down
DROP TABLE IF EXISTS portal_tokens;
