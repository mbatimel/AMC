-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    duplicate_groups TEXT;
BEGIN
    SELECT string_agg(format('%s (%s counterparties)', normalized_inn, counterparty_count), ', ')
    INTO duplicate_groups
    FROM (
        SELECT regexp_replace(inn, '[[:space:]-]', '', 'g') AS normalized_inn,
               count(*) AS counterparty_count
        FROM counterparties
        WHERE inn IS NOT NULL
          AND regexp_replace(inn, '[[:space:]-]', '', 'g') <> ''
        GROUP BY regexp_replace(inn, '[[:space:]-]', '', 'g')
        HAVING count(*) > 1
    ) duplicates;

    IF duplicate_groups IS NOT NULL THEN
        RAISE EXCEPTION 'cannot enforce normalized counterparty INN uniqueness; resolve duplicate groups first: %', duplicate_groups;
    END IF;
END $$;

UPDATE counterparties
SET inn = regexp_replace(inn, '[[:space:]-]', '', 'g'),
    updated_at = now()
WHERE inn IS NOT NULL
  AND inn IS DISTINCT FROM regexp_replace(inn, '[[:space:]-]', '', 'g');

DROP INDEX IF EXISTS uq_counterparties_inn;

CREATE UNIQUE INDEX uq_counterparties_inn
    ON counterparties ((regexp_replace(inn, '[[:space:]-]', '', 'g')))
    WHERE inn IS NOT NULL
      AND regexp_replace(inn, '[[:space:]-]', '', 'g') <> '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uq_counterparties_inn;

CREATE UNIQUE INDEX uq_counterparties_inn
    ON counterparties (inn)
    WHERE inn IS NOT NULL AND inn <> '';
-- +goose StatementEnd
