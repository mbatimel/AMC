-- +goose Up
-- +goose StatementBegin
ALTER TABLE roles
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE roles
    ALTER COLUMN code TYPE INTEGER USING NULLIF(code, '')::INTEGER;

ALTER TABLE user_roles
    ADD COLUMN IF NOT EXISTS id UUID NOT NULL DEFAULT gen_random_uuid(),
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_code ON roles(code);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_id ON user_roles(id);

INSERT INTO roles (code, name, description)
VALUES
    (0, 'admin', 'Full administrative access'),
    (1, 'support', 'Support access'),
    (2, 'supplier', 'Supplier access'),
    (3, 'buyer', 'Buyer access')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = now();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM roles WHERE code IN (0, 1, 2, 3);
DROP INDEX IF EXISTS idx_user_roles_id;
DROP INDEX IF EXISTS idx_roles_code;
ALTER TABLE user_roles
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS id;
ALTER TABLE roles
    ALTER COLUMN code TYPE VARCHAR(255) USING code::VARCHAR,
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS description;
-- +goose StatementEnd
