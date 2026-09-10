-- +goose Up
-- +goose StatementBegin
-- Real data can already contain duplicate INNs from before this constraint
-- existed (that's the whole reason it's being added). Don't let that block
-- the rest of this deploy's migrations — skip the index with a warning
-- listing the offending INNs, so it can be added properly (see the
-- companion migration this comment points future readers to) once the
-- duplicates are resolved.
DO $$
DECLARE
    dup_inns TEXT;
BEGIN
    CREATE UNIQUE INDEX IF NOT EXISTS uq_counterparties_inn ON counterparties (inn) WHERE inn IS NOT NULL AND inn <> '';
EXCEPTION
    WHEN unique_violation THEN
        SELECT string_agg(format('%s (%s counterparties)', inn, cnt), ', ')
        INTO dup_inns
        FROM (
            SELECT inn, count(*) AS cnt
            FROM counterparties
            WHERE inn IS NOT NULL AND inn <> ''
            GROUP BY inn
            HAVING count(*) > 1
        ) d;
        RAISE WARNING 'uq_counterparties_inn NOT created: duplicate INNs already exist in counterparties (%). Resolve the duplicates, then create the index manually: CREATE UNIQUE INDEX uq_counterparties_inn ON counterparties (inn) WHERE inn IS NOT NULL AND inn <> ''''.', dup_inns;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uq_counterparties_inn;
-- +goose StatementEnd
