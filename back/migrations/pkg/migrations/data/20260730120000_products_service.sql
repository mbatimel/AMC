-- +goose Up
-- +goose StatementBegin
CREATE TABLE brands (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

ALTER TABLE categories
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE products
    ADD COLUMN brand_id UUID REFERENCES brands(id),
    ADD COLUMN material VARCHAR(255),
    ADD COLUMN size VARCHAR(255),
    ADD COLUMN deleted_at TIMESTAMPTZ;

ALTER TABLE product_images
    ADD COLUMN url TEXT,
    ADD COLUMN alt TEXT,
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE product_prices
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE stock_balances
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE INDEX idx_products_brand_id ON products(brand_id);
CREATE INDEX idx_products_catalog_search ON products(sku, name, gost_tu);
CREATE INDEX idx_products_material ON products(material);
CREATE INDEX idx_products_size ON products(size);
CREATE INDEX idx_products_not_deleted ON products(id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_product_images_one_main
    ON product_images(product_id)
    WHERE is_main = TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_product_images_one_main;
DROP INDEX IF EXISTS idx_products_not_deleted;
DROP INDEX IF EXISTS idx_products_size;
DROP INDEX IF EXISTS idx_products_material;
DROP INDEX IF EXISTS idx_products_catalog_search;
DROP INDEX IF EXISTS idx_products_brand_id;

ALTER TABLE stock_balances
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS created_at;

ALTER TABLE product_prices
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS created_at;

ALTER TABLE product_images
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS alt,
    DROP COLUMN IF EXISTS url;

ALTER TABLE products
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS size,
    DROP COLUMN IF EXISTS material,
    DROP COLUMN IF EXISTS brand_id;

ALTER TABLE categories
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS created_at;

DROP TABLE IF EXISTS brands;
-- +goose StatementEnd
