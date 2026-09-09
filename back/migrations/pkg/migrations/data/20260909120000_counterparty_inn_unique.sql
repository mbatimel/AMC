-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS uq_counterparties_inn ON counterparties (inn) WHERE inn IS NOT NULL AND inn <> '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uq_counterparties_inn;
-- +goose StatementEnd
