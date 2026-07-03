-- +goose Up
DELETE FROM nomba_events;

ALTER TABLE nomba_events
  ADD COLUMN tenant_id UUID NOT NULL REFERENCES tenants(id);

ALTER TABLE nomba_events DROP CONSTRAINT IF EXISTS nomba_events_event_id_key;
CREATE UNIQUE INDEX nomba_events_tenant_event ON nomba_events(tenant_id, event_id);

-- +goose Down
DROP INDEX IF EXISTS nomba_events_tenant_event;
ALTER TABLE nomba_events DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE nomba_events ADD CONSTRAINT nomba_events_event_id_key UNIQUE (event_id);
