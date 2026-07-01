-- +goose Up
CREATE TABLE webhook_deliveries (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(id),
  endpoint_id     UUID NOT NULL REFERENCES webhook_endpoints(id),
  event_type      TEXT NOT NULL,
  payload         JSONB NOT NULL,
  attempt_count   INT NOT NULL DEFAULT 0,
  last_status     INT,
  last_error      TEXT,
  delivered_at    TIMESTAMPTZ,
  next_retry_at   TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX webhook_deliveries_tenant ON webhook_deliveries(tenant_id);
CREATE INDEX webhook_deliveries_retry ON webhook_deliveries(next_retry_at) WHERE delivered_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS webhook_deliveries;
