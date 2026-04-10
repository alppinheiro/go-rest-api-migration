CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    aggregate_type TEXT NOT NULL,
    event_type TEXT NOT NULL,
    event_version INTEGER NOT NULL,
    payload JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ux_events_aggregate_version UNIQUE (aggregate_id, event_version)
);

CREATE INDEX IF NOT EXISTS idx_events_aggregate ON events(aggregate_id, event_version);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type);