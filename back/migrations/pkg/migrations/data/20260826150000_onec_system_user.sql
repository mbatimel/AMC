-- +goose Up
-- +goose StatementBegin
INSERT INTO users (id, email, name, status)
VALUES ('00000000-0000-0000-0000-0000000a0ec1', 'onec-integration@system.local', '1С интеграция', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-0000000a0ec1', id FROM roles WHERE code = 0
ON CONFLICT (user_id, role_id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM user_roles WHERE user_id = '00000000-0000-0000-0000-0000000a0ec1';
DELETE FROM users WHERE id = '00000000-0000-0000-0000-0000000a0ec1';
-- +goose StatementEnd
