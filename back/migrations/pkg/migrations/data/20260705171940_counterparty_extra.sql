-- +goose Up
-- +goose StatementBegin
CREATE TABLE counterparty_addresses (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    counterparty_id UUID REFERENCES counterparties(id),
    type VARCHAR(255),
    address TEXT,
    city VARCHAR(255),
    region VARCHAR(255),
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id)
);
CREATE INDEX idx_counterparty_addresses_counterparty_id ON counterparty_addresses(counterparty_id);

CREATE TABLE counterparty_contacts (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    counterparty_id UUID REFERENCES counterparties(id),
    full_name VARCHAR(255),
    phone VARCHAR(255),
    email VARCHAR(255),
    position VARCHAR(255),
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id)
);
CREATE INDEX idx_counterparty_contacts_counterparty_id ON counterparty_contacts(counterparty_id);

CREATE TABLE counterparty_category_discounts (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    counterparty_id UUID REFERENCES counterparties(id),
    category_id UUID REFERENCES categories(id),
    discount_percent NUMERIC,
    valid_from DATE,
    valid_to DATE,
    PRIMARY KEY (id)
);
CREATE INDEX idx_counterparty_category_discounts_counterparty_id ON counterparty_category_discounts(counterparty_id);
CREATE INDEX idx_counterparty_category_discounts_category_id ON counterparty_category_discounts(category_id);

CREATE TABLE counterparty_special_prices (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    counterparty_id UUID REFERENCES counterparties(id),
    product_id UUID REFERENCES products(id),
    price NUMERIC,
    valid_from DATE,
    valid_to DATE,
    PRIMARY KEY (id)
);
CREATE INDEX idx_counterparty_special_prices_counterparty_id ON counterparty_special_prices(counterparty_id);
CREATE INDEX idx_counterparty_special_prices_product_id ON counterparty_special_prices(product_id);

CREATE TABLE volume_discounts (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    counterparty_id UUID REFERENCES counterparties(id),
    price_group_id UUID REFERENCES price_groups(id),
    min_order_amount NUMERIC,
    discount_percent NUMERIC,
    PRIMARY KEY (id)
);
CREATE INDEX idx_volume_discounts_counterparty_id ON volume_discounts(counterparty_id);
CREATE INDEX idx_volume_discounts_price_group_id ON volume_discounts(price_group_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS volume_discounts CASCADE;
DROP TABLE IF EXISTS counterparty_special_prices CASCADE;
DROP TABLE IF EXISTS counterparty_category_discounts CASCADE;
DROP TABLE IF EXISTS counterparty_contacts CASCADE;
DROP TABLE IF EXISTS counterparty_addresses CASCADE;
-- +goose StatementEnd
