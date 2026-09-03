-- 007_create_simulations.sql
CREATE TABLE IF NOT EXISTS simulations (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario       VARCHAR(100) NOT NULL,
    target_service VARCHAR(100) NOT NULL,
    parameters     JSONB,
    status         VARCHAR(20) NOT NULL DEFAULT 'pending',
    started_at     TIMESTAMPTZ,
    completed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_simulations_status ON simulations(status);
