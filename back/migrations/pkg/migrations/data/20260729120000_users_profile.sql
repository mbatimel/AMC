-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS active_client_id UUID REFERENCES counterparties(id),
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower_unique
    ON users (LOWER(email))
    WHERE email IS NOT NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_unique
    ON users (phone)
    WHERE phone IS NOT NULL AND phone <> '' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_users_active_client_id ON users(active_client_id);
CREATE INDEX IF NOT EXISTS idx_users_status_active ON users(status, is_active) WHERE deleted_at IS NULL;

CREATE TABLE user_clients (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES counterparties(id) ON DELETE CASCADE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, client_id)
);

CREATE INDEX idx_user_clients_client_id ON user_clients(client_id);

INSERT INTO user_clients (user_id, client_id, is_default)
SELECT id, counterparty_id, TRUE
FROM users
WHERE counterparty_id IS NOT NULL
ON CONFLICT (user_id, client_id) DO NOTHING;

UPDATE users
SET active_client_id = counterparty_id
WHERE active_client_id IS NULL
  AND counterparty_id IS NOT NULL;

ALTER TABLE counterparties
    ADD COLUMN IF NOT EXISTS credit_used NUMERIC NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS payment_terms VARCHAR(255),
    ADD COLUMN IF NOT EXISTS sales_contact VARCHAR(255),
    ADD COLUMN IF NOT EXISTS contact_channel VARCHAR(255);

CREATE TABLE favorites (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES counterparties(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, client_id, product_id),
    CONSTRAINT fk_favorites_user_client
        FOREIGN KEY (user_id, client_id)
        REFERENCES user_clients(user_id, client_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_favorites_client_id ON favorites(client_id);
CREATE INDEX idx_favorites_product_id ON favorites(product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS favorites;

ALTER TABLE counterparties
    DROP COLUMN IF EXISTS contact_channel,
    DROP COLUMN IF EXISTS sales_contact,
    DROP COLUMN IF EXISTS payment_terms,
    DROP COLUMN IF EXISTS credit_used;

DROP TABLE IF EXISTS user_clients;

DROP INDEX IF EXISTS idx_users_status_active;
DROP INDEX IF EXISTS idx_users_active_client_id;
DROP INDEX IF EXISTS idx_users_phone_unique;
DROP INDEX IF EXISTS idx_users_email_lower_unique;

ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);

ALTER TABLE users
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS active_client_id,
    DROP COLUMN IF EXISTS is_active;
-- +goose StatementEnd
