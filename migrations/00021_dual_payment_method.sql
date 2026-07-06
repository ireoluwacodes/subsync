-- +goose Up
ALTER TABLE subscriptions
  ADD COLUMN fallback_payment_method_id UUID REFERENCES payment_methods(id);

ALTER TABLE payment_methods
  ADD COLUMN mandate_status TEXT;

CREATE INDEX subscriptions_fallback_pm ON subscriptions(fallback_payment_method_id)
  WHERE fallback_payment_method_id IS NOT NULL;

CREATE INDEX payment_methods_pending_mandate ON payment_methods(mandate_status)
  WHERE mandate_status = 'pending';

-- +goose Down
DROP INDEX IF EXISTS payment_methods_pending_mandate;
DROP INDEX IF EXISTS subscriptions_fallback_pm;
ALTER TABLE payment_methods DROP COLUMN IF EXISTS mandate_status;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS fallback_payment_method_id;
