-- 005_create_incidents.sql
CREATE TABLE IF NOT EXISTS incidents (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id        UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    title             VARCHAR(500) NOT NULL,
    severity          VARCHAR(20) NOT NULL,
    status            VARCHAR(20) NOT NULL DEFAULT 'active',
    risk_score        INTEGER NOT NULL,
    root_cause        TEXT,
    propagation_path  JSONB,
    impacted_services JSONB,
    anomalies         JSONB,
    started_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_incidents_service ON incidents(service_id);
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
CREATE INDEX IF NOT EXISTS idx_incidents_started ON incidents(started_at DESC);
