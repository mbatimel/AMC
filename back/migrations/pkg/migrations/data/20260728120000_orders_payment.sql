-- +goose Up
-- +goose StatementBegin
ALTER TABLE orders ADD COLUMN payment_status VARCHAR(255) NOT NULL DEFAULT 'not_paid';
ALTER TABLE order_status_history ADD COLUMN payment_status VARCHAR(255);

CREATE SEQUENCE orders_number_seq START WITH 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS orders_number_seq;
ALTER TABLE order_status_history DROP COLUMN IF EXISTS payment_status;
ALTER TABLE orders DROP COLUMN IF EXISTS payment_status;
-- +goose StatementEnd
