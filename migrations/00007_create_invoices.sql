-- +goose Up
CREATE TABLE invoices (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id             UUID NOT NULL REFERENCES tenants(id),
  subscription_id       UUID NOT NULL REFERENCES subscriptions(id),
  customer_id           UUID NOT NULL REFERENCES customers(id),
  status                TEXT NOT NULL DEFAULT 'draft',
  amount_due            BIGINT NOT NULL,
  amount_paid           BIGINT NOT NULL DEFAULT 0,
  currency              TEXT NOT NULL DEFAULT 'NGN',
  period_start          TIMESTAMPTZ NOT NULL,
  period_end            TIMESTAMPTZ NOT NULL,
  due_date              TIMESTAMPTZ,
  paid_at               TIMESTAMPTZ,
  voided_at             TIMESTAMPTZ,
  nomba_order_ref       TEXT UNIQUE,
  nomba_transaction_id  TEXT,
  attempt_count         INT NOT NULL DEFAULT 0,
  next_attempt_at       TIMESTAMPTZ,
  metadata              JSONB NOT NULL DEFAULT '{}',
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX invoices_tenant ON invoices(tenant_id);
CREATE INDEX invoices_subscription ON invoices(subscription_id);
CREATE INDEX invoices_status ON invoices(status);
CREATE INDEX invoices_next_attempt ON invoices(next_attempt_at) WHERE status = 'open';

-- +goose Down
DROP TABLE IF EXISTS invoices;
