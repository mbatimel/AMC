-- +goose Up
-- +goose StatementBegin
CREATE TABLE marketplaces (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    code VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    PRIMARY KEY (id)
);

CREATE TABLE marketplace_attribute_mappings (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    marketplace_id UUID REFERENCES marketplaces(id),
    category_id UUID REFERENCES categories(id),
    attribute_id UUID REFERENCES attributes(id),
    external_attribute_code VARCHAR(255),
    transform_rule JSONB,
    PRIMARY KEY (id)
);
CREATE INDEX idx_marketplace_attribute_mappings_marketplace_id ON marketplace_attribute_mappings(marketplace_id);
CREATE INDEX idx_marketplace_attribute_mappings_category_id ON marketplace_attribute_mappings(category_id);
CREATE INDEX idx_marketplace_attribute_mappings_attribute_id ON marketplace_attribute_mappings(attribute_id);

CREATE TABLE marketplace_products (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    marketplace_id UUID REFERENCES marketplaces(id),
    product_id UUID REFERENCES products(id),
    external_product_id VARCHAR(255),
    external_sku VARCHAR(255),
    status VARCHAR(255),
    last_price_sync_at TIMESTAMPTZ,
    last_stock_sync_at TIMESTAMPTZ,
    last_error TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_marketplace_products_marketplace_id ON marketplace_products(marketplace_id);
CREATE INDEX idx_marketplace_products_product_id ON marketplace_products(product_id);

CREATE TABLE marketplace_publication_logs (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    marketplace_product_id UUID REFERENCES marketplace_products(id),
    operation VARCHAR(255),
    status VARCHAR(255),
    request_payload JSONB,
    response_payload JSONB,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_marketplace_publication_logs_marketplace_product_id ON marketplace_publication_logs(marketplace_product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS marketplace_publication_logs CASCADE;
DROP TABLE IF EXISTS marketplace_products CASCADE;
DROP TABLE IF EXISTS marketplace_attribute_mappings CASCADE;
DROP TABLE IF EXISTS marketplaces CASCADE;
-- +goose StatementEnd
