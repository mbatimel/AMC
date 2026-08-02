-- +goose Up
-- +goose StatementBegin
ALTER TABLE portal_banners
    ADD COLUMN file_id UUID REFERENCES files(id);

CREATE INDEX idx_portal_banners_file_id ON portal_banners(file_id);

CREATE TABLE portal_banner_settings (
    singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
    delay_sec INT NOT NULL DEFAULT 6 CHECK (delay_sec > 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO portal_banner_settings (singleton, delay_sec)
VALUES (TRUE, 6)
ON CONFLICT (singleton) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS portal_banner_settings;
DROP INDEX IF EXISTS idx_portal_banners_file_id;
ALTER TABLE portal_banners DROP COLUMN IF EXISTS file_id;
-- +goose StatementEnd
