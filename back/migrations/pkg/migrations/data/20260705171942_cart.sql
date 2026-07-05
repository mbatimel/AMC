-- +goose Up
-- +goose StatementBegin
CREATE TABLE carts (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    counterparty_id UUID REFERENCES counterparties(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_carts_user_id ON carts(user_id);
CREATE INDEX idx_carts_counterparty_id ON carts(counterparty_id);

CREATE TABLE cart_items (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    cart_id UUID REFERENCES carts(id),
    product_id UUID REFERENCES products(id),
    quantity NUMERIC,
    price NUMERIC,
    PRIMARY KEY (id)
);
CREATE INDEX idx_cart_items_cart_id ON cart_items(cart_id);
CREATE INDEX idx_cart_items_product_id ON cart_items(product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cart_items CASCADE;
DROP TABLE IF EXISTS carts CASCADE;
-- +goose StatementEnd
