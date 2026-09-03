-- 003_create_metrics.sql
CREATE TABLE IF NOT EXISTS metrics (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id          UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    timestamp           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    response_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    request_rate        DOUBLE PRECISION NOT NULL DEFAULT 0,
    error_rate          DOUBLE PRECISION NOT NULL DEFAULT 0,
    timeout_rate        DOUBLE PRECISION NOT NULL DEFAULT 0,
    cpu_usage           DOUBLE PRECISION NOT NULL DEFAULT 0,
    memory_usage        DOUBLE PRECISION NOT NULL DEFAULT 0,
    dep_latency_ms      DOUBLE PRECISION NOT NULL DEFAULT 0,
    dep_error_rate      DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_metrics_service_time ON metrics(service_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp DESC);
