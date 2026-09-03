-- 002_create_dependencies.sql
CREATE TABLE IF NOT EXISTS dependencies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id       UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    target_id       UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    dependency_type VARCHAR(50) NOT NULL DEFAULT 'sync',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(source_id, target_id)
);

CREATE INDEX IF NOT EXISTS idx_dependencies_source ON dependencies(source_id);
CREATE INDEX IF NOT EXISTS idx_dependencies_target ON dependencies(target_id);
