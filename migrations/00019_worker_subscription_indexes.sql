-- +goose Up
CREATE INDEX IF NOT EXISTS idx_subscriptions_next_billing_at ON subscriptions (next_billing_at) WHERE next_billing_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_subscriptions_trial_ends_at ON subscriptions (trial_ends_at) WHERE trial_ends_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_subscriptions_current_period_end ON subscriptions (current_period_end);
CREATE INDEX IF NOT EXISTS idx_subscriptions_pause_ends_at ON subscriptions (pause_ends_at) WHERE pause_ends_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_subscriptions_pause_ends_at;
DROP INDEX IF EXISTS idx_subscriptions_current_period_end;
DROP INDEX IF EXISTS idx_subscriptions_trial_ends_at;
DROP INDEX IF EXISTS idx_subscriptions_next_billing_at;
