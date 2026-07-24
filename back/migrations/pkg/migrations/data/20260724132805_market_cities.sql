-- +goose Up
-- +goose StatementBegin

CREATE TABLE cities (
    id   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL
);


INSERT INTO cities (name) VALUES
    ('Москва'),
    ('Самара'),
    ('Санкт-Петербург');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
