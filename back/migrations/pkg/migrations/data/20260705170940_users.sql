-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE,
    password TEXT,
    name VARCHAR(40),
    surename VARCHAR(40),
    middle_name VARCHAR(40),
    phone VARCHAR(255),
    status VARCHAR(255),
    counterparty_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE TABLE counterparties (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    one_c_guid UUID UNIQUE,
    type VARCHAR(255),
    name VARCHAR(255),
    inn VARCHAR(255),
    kpp VARCHAR(255),
    ogrn VARCHAR(255),
    legal_address TEXT,
    actual_address TEXT,
    price_group_id UUID,
    credit_limit NUMERIC,
    manager_id UUID REFERENCES users(id),
    status VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS counterparties CASCADE;
-- +goose StatementEnd
