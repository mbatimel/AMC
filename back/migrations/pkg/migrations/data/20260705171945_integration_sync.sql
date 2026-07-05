-- +goose Up
-- +goose StatementBegin
CREATE TABLE integration_systems (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    code VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    base_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (id)
);

CREATE TABLE integration_credentials (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    system_id UUID REFERENCES integration_systems(id),
    name VARCHAR(255),
    encrypted_payload BYTEA,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_integration_credentials_system_id ON integration_credentials(system_id);

CREATE TABLE sync_jobs (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    system_id UUID REFERENCES integration_systems(id),
    direction VARCHAR(255),
    entity_type VARCHAR(255),
    entity_id UUID,
    payload JSONB,
    status VARCHAR(255),
    attempts INTEGER,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE INDEX idx_sync_jobs_system_id ON sync_jobs(system_id);

CREATE TABLE sync_logs (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    job_id UUID REFERENCES sync_jobs(id),
    system_id UUID REFERENCES integration_systems(id),
    level VARCHAR(255),
    message TEXT,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_sync_logs_job_id ON sync_logs(job_id);
CREATE INDEX idx_sync_logs_system_id ON sync_logs(system_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sync_logs CASCADE;
DROP TABLE IF EXISTS sync_jobs CASCADE;
DROP TABLE IF EXISTS integration_credentials CASCADE;
DROP TABLE IF EXISTS integration_systems CASCADE;
-- +goose StatementEnd
