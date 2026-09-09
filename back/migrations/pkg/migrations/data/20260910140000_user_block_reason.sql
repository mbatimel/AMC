-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS blocked_reason TEXT,
    ADD COLUMN IF NOT EXISTS blocked_contact_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS blocked_contact_phone VARCHAR(255),
    ADD COLUMN IF NOT EXISTS blocked_contact_email VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN IF EXISTS blocked_contact_email,
    DROP COLUMN IF EXISTS blocked_contact_phone,
    DROP COLUMN IF EXISTS blocked_contact_name,
    DROP COLUMN IF EXISTS blocked_reason;
-- +goose StatementEnd
