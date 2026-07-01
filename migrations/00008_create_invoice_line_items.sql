-- +goose Up
CREATE TABLE invoice_line_items (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_id      UUID NOT NULL REFERENCES invoices(id),
  tenant_id       UUID NOT NULL REFERENCES tenants(id),
  type            TEXT NOT NULL,
  description     TEXT NOT NULL,
  amount          BIGINT NOT NULL,
  currency        TEXT NOT NULL DEFAULT 'NGN',
  period_start    TIMESTAMPTZ,
  period_end      TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX line_items_invoice ON invoice_line_items(invoice_id);

-- +goose Down
DROP TABLE IF EXISTS invoice_line_items;
