-- +goose Up
-- +goose StatementBegin
ALTER TABLE counterparties
    ADD COLUMN requisites_file_url TEXT,
    ADD COLUMN requisites_file_name TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE counterparties
    DROP COLUMN requisites_file_url,
    DROP COLUMN requisites_file_name;
-- +goose StatementEnd
