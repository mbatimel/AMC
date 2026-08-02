-- +goose Up
-- +goose StatementBegin
-- Backfill: users created before self-signup auto-created a personal client
-- (e.g. registered without a company/INN) can end up with no active_client_id.
-- Give each of them a personal counterparty, same as a fresh self-signup gets.
DO $$
DECLARE
    orphan RECORD;
    new_client_id UUID;
BEGIN
    FOR orphan IN
        SELECT id, name, surename
        FROM users
        WHERE active_client_id IS NULL AND deleted_at IS NULL
    LOOP
        INSERT INTO counterparties (type, name, status)
        VALUES (
            'individual',
            NULLIF(TRIM(COALESCE(orphan.name, '') || ' ' || COALESCE(orphan.surename, '')), ''),
            'active'
        )
        RETURNING id INTO new_client_id;

        UPDATE users
        SET counterparty_id = COALESCE(counterparty_id, new_client_id),
            active_client_id = new_client_id,
            updated_at = now()
        WHERE id = orphan.id;

        INSERT INTO user_clients (user_id, client_id, is_default)
        VALUES (orphan.id, new_client_id, TRUE)
        ON CONFLICT (user_id, client_id) DO NOTHING;
    END LOOP;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Not reversible: we can't tell a backfilled personal client apart from one
-- a user has since started using for real (orders, other user_clients rows).
-- +goose StatementEnd
