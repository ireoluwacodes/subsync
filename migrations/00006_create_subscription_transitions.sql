-- +goose Up
CREATE TABLE subscription_transitions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  subscription_id UUID NOT NULL REFERENCES subscriptions(id),
  tenant_id       UUID NOT NULL REFERENCES tenants(id),
  from_state      TEXT NOT NULL,
  to_state        TEXT NOT NULL,
  reason          TEXT NOT NULL,
  actor           TEXT NOT NULL DEFAULT 'system',
  metadata        JSONB NOT NULL DEFAULT '{}',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX transitions_subscription ON subscription_transitions(subscription_id);
CREATE INDEX transitions_tenant ON subscription_transitions(tenant_id);

-- +goose Down
DROP TABLE IF EXISTS subscription_transitions;
