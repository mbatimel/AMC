-- +goose Up
-- +goose StatementBegin
CREATE TABLE document_templates (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    type VARCHAR(255),
    name VARCHAR(255),
    template_body TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE TABLE orders (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    one_c_guid UUID UNIQUE,
    number VARCHAR(255) UNIQUE,
    counterparty_id UUID REFERENCES counterparties(id),
    user_id UUID REFERENCES users(id),
    manager_id UUID REFERENCES users(id),
    status VARCHAR(255),
    delivery_method VARCHAR(255),
    delivery_address_id UUID REFERENCES counterparty_addresses(id),
    contact_id UUID REFERENCES counterparty_contacts(id),
    comment TEXT,
    subtotal NUMERIC,
    discount_total NUMERIC,
    vat_total NUMERIC,
    total NUMERIC,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    confirmed_at TIMESTAMPTZ,
    synced_to_1c_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE INDEX idx_orders_counterparty_id ON orders(counterparty_id);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_manager_id ON orders(manager_id);
CREATE INDEX idx_orders_delivery_address_id ON orders(delivery_address_id);
CREATE INDEX idx_orders_contact_id ON orders(contact_id);

CREATE TABLE order_items (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id),
    product_id UUID REFERENCES products(id),
    sku VARCHAR(255),
    name VARCHAR(255),
    quantity NUMERIC,
    unit_price NUMERIC,
    discount_percent NUMERIC,
    vat_rate NUMERIC,
    line_total NUMERIC,
    PRIMARY KEY (id)
);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);

CREATE TABLE order_status_history (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id),
    old_status VARCHAR(255),
    new_status VARCHAR(255),
    changed_by UUID REFERENCES users(id),
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_order_status_history_order_id ON order_status_history(order_id);
CREATE INDEX idx_order_status_history_changed_by ON order_status_history(changed_by);

CREATE TABLE documents (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    one_c_guid UUID UNIQUE,
    order_id UUID REFERENCES orders(id),
    counterparty_id UUID REFERENCES counterparties(id),
    template_id UUID REFERENCES document_templates(id),
    type VARCHAR(255),
    number VARCHAR(255),
    status VARCHAR(255),
    file_id UUID REFERENCES files(id),
    amount NUMERIC,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE INDEX idx_documents_order_id ON documents(order_id);
CREATE INDEX idx_documents_counterparty_id ON documents(counterparty_id);
CREATE INDEX idx_documents_template_id ON documents(template_id);
CREATE INDEX idx_documents_file_id ON documents(file_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS documents CASCADE;
DROP TABLE IF EXISTS order_status_history CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS document_templates CASCADE;
-- +goose StatementEnd
