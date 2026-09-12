-- +goose Up
-- +goose StatementBegin
ALTER TABLE products
    ADD COLUMN wb_status VARCHAR(50),
    ADD COLUMN ozon_status VARCHAR(50),
    ADD COLUMN ym_status VARCHAR(50);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products
    DROP COLUMN IF EXISTS ym_status,
    DROP COLUMN IF EXISTS ozon_status,
    DROP COLUMN IF EXISTS wb_status;
-- +goose StatementEnd
