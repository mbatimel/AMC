-- +goose Up
-- +goose StatementBegin
CREATE TABLE admin_audit_log (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID NOT NULL,
    actor_label   VARCHAR(255) NOT NULL,
    action        TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_admin_audit_log_created_at ON admin_audit_log (created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_admin_audit_log_created_at;
DROP TABLE IF EXISTS admin_audit_log;
-- +goose StatementEnd
