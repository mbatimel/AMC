-- +goose Up
-- +goose StatementBegin
-- Legacy scaffold table from 20260705171947_auth_misc.sql: never referenced
-- by any service. Dropped and replaced with the real promotions schema
-- (name, discount_percent, product/qty-threshold links) used by the
-- products-service promotions feature.
DROP TABLE IF EXISTS promotions CASCADE;

CREATE TABLE promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    discount_percent NUMERIC NOT NULL CHECK (discount_percent >= 0 AND discount_percent <= 100),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL CHECK (ends_at > starts_at),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE promotion_products (
    promotion_id UUID NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    min_qty INTEGER NOT NULL DEFAULT 1 CHECK (min_qty >= 1),
    PRIMARY KEY (promotion_id, product_id)
);
CREATE INDEX idx_promotion_products_product_id ON promotion_products(product_id);
CREATE INDEX idx_promotions_period ON promotions(starts_at, ends_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS promotion_products;
DROP TABLE IF EXISTS promotions;

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
