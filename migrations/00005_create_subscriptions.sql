-- +goose Up
CREATE TABLE subscriptions (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id             UUID NOT NULL REFERENCES tenants(id),
  customer_id           UUID NOT NULL REFERENCES customers(id),
  plan_id               UUID NOT NULL REFERENCES plans(id),
  payment_method_id     UUID REFERENCES payment_methods(id),
  state                 TEXT NOT NULL DEFAULT 'trialing',
  trial_ends_at         TIMESTAMPTZ,
  current_period_start  TIMESTAMPTZ NOT NULL,
  current_period_end    TIMESTAMPTZ NOT NULL,
  next_billing_at       TIMESTAMPTZ,
  canceled_at           TIMESTAMPTZ,
  cancel_at_period_end  BOOLEAN NOT NULL DEFAULT FALSE,
  pause_starts_at       TIMESTAMPTZ,
  pause_ends_at         TIMESTAMPTZ,
  dunning_step          INT NOT NULL DEFAULT 0,
  dunning_started_at    TIMESTAMPTZ,
  metadata              JSONB NOT NULL DEFAULT '{}',
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX subscriptions_tenant ON subscriptions(tenant_id);
CREATE INDEX subscriptions_customer ON subscriptions(customer_id);
CREATE INDEX subscriptions_state ON subscriptions(state);
CREATE INDEX subscriptions_next_billing ON subscriptions(next_billing_at) WHERE state = 'active';

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
