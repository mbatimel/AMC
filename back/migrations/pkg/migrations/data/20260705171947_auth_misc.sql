-- +goose Up
-- +goose StatementBegin
CREATE TABLE roles (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    code VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    PRIMARY KEY (id)
);

CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id),
    role_id UUID REFERENCES roles(id),
    PRIMARY KEY (user_id, role_id)
);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

CREATE TABLE login_attempts (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    email VARCHAR(255),
    ip_address INET,
    is_success BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE TABLE audit_logs (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(255),
    entity_type VARCHAR(255),
    entity_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);

CREATE TABLE notifications (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    channel VARCHAR(255),
    subject VARCHAR(255),
    body TEXT,
    status VARCHAR(255),
    related_entity_type VARCHAR(255),
    related_entity_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE INDEX idx_notifications_user_id ON notifications(user_id);

CREATE TABLE feedback_requests (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    name VARCHAR(255),
    phone VARCHAR(255),
    email VARCHAR(255),
    message TEXT,
    status VARCHAR(255),
    assigned_manager_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_feedback_requests_assigned_manager_id ON feedback_requests(assigned_manager_id);

CREATE TABLE cms_pages (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    slug VARCHAR(255) UNIQUE,
    title VARCHAR(255),
    body TEXT,
    is_published BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE TABLE promotions (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    title VARCHAR(255),
    description TEXT,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS promotions CASCADE;
DROP TABLE IF EXISTS cms_pages CASCADE;
DROP TABLE IF EXISTS feedback_requests CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS login_attempts CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
-- +goose StatementEnd
