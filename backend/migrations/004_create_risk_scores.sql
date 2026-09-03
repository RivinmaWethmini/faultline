-- 004_create_risk_scores.sql
CREATE TABLE IF NOT EXISTS risk_scores (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id          UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    timestamp           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    overall_score       INTEGER NOT NULL CHECK (overall_score BETWEEN 0 AND 100),
    risk_level          VARCHAR(20) NOT NULL,
    latency_anomaly     INTEGER NOT NULL DEFAULT 0,
    error_anomaly       INTEGER NOT NULL DEFAULT 0,
    timeout_anomaly     INTEGER NOT NULL DEFAULT 0,
    traffic_anomaly     INTEGER NOT NULL DEFAULT 0,
    dependency_anomaly  INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_risk_scores_service_time ON risk_scores(service_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_risk_scores_level ON risk_scores(risk_level);
