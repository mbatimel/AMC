-- +goose Up
-- +goose StatementBegin
-- Data fix: OOO "PO Volzhskiy instrument" (INN 6312081084) was created three
-- times (2026-09-02, 2026-09-08, 2026-09-09) instead of reused. This blocks
-- 20260911120000_enforce_normalized_counterparty_inn.sql, which requires the
-- normalized INN to be unique. Merge the two later duplicates into the
-- original (oldest) counterparty and repoint all references before that
-- migration runs.
DO $$
DECLARE
    canonical_id UUID := '4867c224-d673-4eaf-a327-8aaf135f5287';
    dup_ids UUID[] := ARRAY[
        '2820b18b-25f4-43a3-9146-db924dbd3b10',
        'ea41209e-0f6d-4cff-88b4-3e923627b815'
    ]::UUID[];
BEGIN
    UPDATE users SET counterparty_id = canonical_id WHERE counterparty_id = ANY(dup_ids);
    UPDATE users SET active_client_id = canonical_id WHERE active_client_id = ANY(dup_ids);

    INSERT INTO user_clients (user_id, client_id, is_default, created_at, updated_at)
    SELECT user_id, canonical_id, bool_or(is_default), min(created_at), now()
    FROM user_clients
    WHERE client_id = ANY(dup_ids)
    GROUP BY user_id
    ON CONFLICT (user_id, client_id) DO NOTHING;

    INSERT INTO favorites (user_id, client_id, product_id, created_at)
    SELECT user_id, canonical_id, product_id, min(created_at)
    FROM favorites
    WHERE client_id = ANY(dup_ids)
    GROUP BY user_id, product_id
    ON CONFLICT (user_id, client_id, product_id) DO NOTHING;

    DELETE FROM favorites WHERE client_id = ANY(dup_ids);
    DELETE FROM user_clients WHERE client_id = ANY(dup_ids);

    UPDATE ai_chat_sessions SET counterparty_id = canonical_id WHERE counterparty_id = ANY(dup_ids);
    UPDATE counterparty_addresses SET counterparty_id = canonical_id WHERE counterparty_id = ANY(dup_ids);
    UPDATE counterparty_contacts SET counterparty_id = canonical_id WHERE counterparty_id = ANY(dup_ids);
    UPDATE counterparty_category_discounts SET counterparty_id = canonical_id WHERE counterparty_id = ANY(dup_ids);
    UPDATE counterparty_special_prices SET counterparty_id = canonical_id WHERE counterparty_id = ANY(dup_ids);
    UPDATE volume_discounts SET counterparty_id = canonical_id WHERE counterparty_id = ANY(dup_ids);
    UPDATE carts SET counterparty_id = canonical_id WHERE counterparty_id = ANY(dup_ids);
    UPDATE orders SET counterparty_id = canonical_id WHERE counterparty_id = ANY(dup_ids);
    UPDATE documents SET counterparty_id = canonical_id WHERE counterparty_id = ANY(dup_ids);
    UPDATE sales_history SET counterparty_id = canonical_id WHERE counterparty_id = ANY(dup_ids);

    DELETE FROM counterparties WHERE id = ANY(dup_ids);
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Not reversible: merged duplicate counterparty rows and their identity are gone.
-- +goose StatementEnd
