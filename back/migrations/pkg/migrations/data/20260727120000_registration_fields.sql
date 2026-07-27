-- +goose Up
-- +goose StatementBegin
ALTER TABLE counterparties
    ADD COLUMN IF NOT EXISTS short_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS okved VARCHAR(255),
    ADD COLUMN IF NOT EXISTS tax_system VARCHAR(255),
    ADD COLUMN IF NOT EXISTS website VARCHAR(255),
    ADD COLUMN IF NOT EXISTS director_full_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS director_position VARCHAR(255),
    ADD COLUMN IF NOT EXISTS phone VARCHAR(255),
    ADD COLUMN IF NOT EXISTS additional_phone VARCHAR(255),
    ADD COLUMN IF NOT EXISTS email VARCHAR(255),
    ADD COLUMN IF NOT EXISTS bank_account VARCHAR(255),
    ADD COLUMN IF NOT EXISTS bank_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS bank_bik VARCHAR(255),
    ADD COLUMN IF NOT EXISTS correspondent_account VARCHAR(255);

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS inn VARCHAR(255),
    ADD COLUMN IF NOT EXISTS city VARCHAR(255),
    ADD COLUMN IF NOT EXISTS delivery_address TEXT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_users_counterparty_id'
    ) THEN
        ALTER TABLE users
            ADD CONSTRAINT fk_users_counterparty_id FOREIGN KEY (counterparty_id) REFERENCES counterparties(id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_users_counterparty_id ON users(counterparty_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_counterparty_id;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS fk_users_counterparty_id;

ALTER TABLE users
    DROP COLUMN IF EXISTS delivery_address,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS inn;

ALTER TABLE counterparties
    DROP COLUMN IF EXISTS correspondent_account,
    DROP COLUMN IF EXISTS bank_bik,
    DROP COLUMN IF EXISTS bank_name,
    DROP COLUMN IF EXISTS bank_account,
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS additional_phone,
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS director_position,
    DROP COLUMN IF EXISTS director_full_name,
    DROP COLUMN IF EXISTS website,
    DROP COLUMN IF EXISTS tax_system,
    DROP COLUMN IF EXISTS okved,
    DROP COLUMN IF EXISTS short_name;
-- +goose StatementEnd
