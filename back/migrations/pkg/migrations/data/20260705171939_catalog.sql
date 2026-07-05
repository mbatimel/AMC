-- +goose Up
-- +goose StatementBegin
-- завершение отложенных FK на users/counterparties (см. 20260705170940_users.sql)
ALTER TABLE users ADD CONSTRAINT fk_users_counterparty_id FOREIGN KEY (counterparty_id) REFERENCES counterparties(id);
CREATE INDEX idx_users_counterparty_id ON users(counterparty_id);
CREATE INDEX idx_counterparties_manager_id ON counterparties(manager_id);

CREATE TABLE price_groups (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    one_c_guid UUID UNIQUE,
    code VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    PRIMARY KEY (id)
);

ALTER TABLE counterparties ADD CONSTRAINT fk_counterparties_price_group_id FOREIGN KEY (price_group_id) REFERENCES price_groups(id);
CREATE INDEX idx_counterparties_price_group_id ON counterparties(price_group_id);

CREATE TABLE units (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    code VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    PRIMARY KEY (id)
);

CREATE TABLE categories (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    parent_id UUID REFERENCES categories(id),
    one_c_guid UUID UNIQUE,
    name VARCHAR(255),
    slug VARCHAR(255) UNIQUE,
    description TEXT,
    sort_order INTEGER,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (id)
);
CREATE INDEX idx_categories_parent_id ON categories(parent_id);

CREATE TABLE files (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    storage_key TEXT,
    original_name VARCHAR(255),
    mime_type VARCHAR(255),
    size_bytes BIGINT,
    checksum VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE TABLE attributes (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    code VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    data_type VARCHAR(255),
    PRIMARY KEY (id)
);

CREATE TABLE products (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    one_c_guid UUID UNIQUE,
    category_id UUID REFERENCES categories(id),
    unit_id UUID REFERENCES units(id),
    sku VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    slug VARCHAR(255) UNIQUE,
    gost_tu VARCHAR(255),
    description_short TEXT,
    description_full TEXT,
    package_multiple NUMERIC,
    vat_rate NUMERIC,
    allow_backorder BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_unit_id ON products(unit_id);

CREATE TABLE product_attribute_values (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    attribute_id UUID REFERENCES attributes(id),
    value_text TEXT,
    value_number NUMERIC,
    value_bool BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id)
);
CREATE INDEX idx_product_attribute_values_product_id ON product_attribute_values(product_id);
CREATE INDEX idx_product_attribute_values_attribute_id ON product_attribute_values(attribute_id);

CREATE TABLE product_images (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    file_id UUID REFERENCES files(id),
    sort_order INTEGER,
    is_main BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id)
);
CREATE INDEX idx_product_images_product_id ON product_images(product_id);
CREATE INDEX idx_product_images_file_id ON product_images(file_id);

CREATE TABLE product_prices (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    price_group_id UUID REFERENCES price_groups(id),
    price_type VARCHAR(255),
    price NUMERIC,
    discount_percent NUMERIC NOT NULL DEFAULT 0,
    final_price NUMERIC GENERATED ALWAYS AS (ROUND(price * (1 - discount_percent / 100), 2)) STORED,
    currency CHAR(3),
    valid_from TIMESTAMPTZ,
    synced_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE INDEX idx_product_prices_product_id ON product_prices(product_id);
CREATE INDEX idx_product_prices_price_group_id ON product_prices(price_group_id);

CREATE TABLE certificates (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    title VARCHAR(255),
    file_id UUID REFERENCES files(id),
    product_id UUID REFERENCES products(id),
    category_id UUID REFERENCES categories(id),
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id)
);
CREATE INDEX idx_certificates_file_id ON certificates(file_id);
CREATE INDEX idx_certificates_product_id ON certificates(product_id);
CREATE INDEX idx_certificates_category_id ON certificates(category_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS certificates CASCADE;
DROP TABLE IF EXISTS product_prices CASCADE;
DROP TABLE IF EXISTS product_images CASCADE;
DROP TABLE IF EXISTS product_attribute_values CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS attributes CASCADE;
DROP TABLE IF EXISTS files CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS units CASCADE;
DROP INDEX IF EXISTS idx_counterparties_price_group_id;
ALTER TABLE counterparties DROP CONSTRAINT IF EXISTS fk_counterparties_price_group_id;
DROP TABLE IF EXISTS price_groups CASCADE;
DROP INDEX IF EXISTS idx_counterparties_manager_id;
DROP INDEX IF EXISTS idx_users_counterparty_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_counterparty_id;
-- +goose StatementEnd
