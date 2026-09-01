-- +goose Up
-- +goose StatementBegin
-- Registrations made after user_clients was introduced could still be written
-- by auth without that link. counterparty_id is an explicit, unambiguous
-- relationship, so it is safe to restore the link and make it active only
-- when no active client has been selected yet.
UPDATE user_clients uc
SET is_default = FALSE,
    updated_at = now()
FROM users u
WHERE u.id = uc.user_id
  AND u.counterparty_id IS NOT NULL
  AND (u.active_client_id IS NULL OR u.active_client_id = u.counterparty_id)
  AND uc.client_id <> u.counterparty_id
  AND uc.is_default;

INSERT INTO user_clients (user_id, client_id, is_default)
SELECT
    id,
    counterparty_id,
    active_client_id IS NULL OR active_client_id = counterparty_id
FROM users
WHERE counterparty_id IS NOT NULL
  AND deleted_at IS NULL
ON CONFLICT (user_id, client_id) DO UPDATE
SET is_default = EXCLUDED.is_default,
    updated_at = now();

UPDATE users
SET active_client_id = counterparty_id,
    updated_at = now()
WHERE counterparty_id IS NOT NULL
  AND active_client_id IS NULL
  AND deleted_at IS NULL;

-- The director and phone stored on the explicitly linked counterparty are the
-- only profile values that can be recovered without inventing user data.
WITH recoverable AS (
    SELECT
        u.id,
        CASE
            WHEN NOT EXISTS (
                SELECT 1
                FROM users other
                WHERE other.id <> u.id
                  AND other.deleted_at IS NULL
                  AND NULLIF(other.phone, '') = NULLIF(trim(c.phone), '')
            ) THEN NULLIF(trim(c.phone), '')
        END AS phone,
        regexp_split_to_array(NULLIF(trim(c.director_full_name), ''), E'\\s+') AS fio
    FROM users u
    JOIN counterparties c ON c.id = u.counterparty_id
    WHERE u.deleted_at IS NULL
)
UPDATE users u
SET surename = COALESCE(NULLIF(u.surename, ''), recoverable.fio[1]),
    name = COALESCE(NULLIF(u.name, ''), recoverable.fio[2]),
    middle_name = COALESCE(NULLIF(u.middle_name, ''), NULLIF(array_to_string(recoverable.fio[3:], ' '), '')),
    phone = COALESCE(NULLIF(u.phone, ''), recoverable.phone),
    updated_at = now()
FROM recoverable
WHERE recoverable.id = u.id
  AND (
      NULLIF(u.surename, '') IS NULL OR
      NULLIF(u.name, '') IS NULL OR
      NULLIF(u.middle_name, '') IS NULL OR
      NULLIF(u.phone, '') IS NULL
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Data backfill is intentionally not reversible: restored links and real
-- registration data may already be in use by the time a rollback is needed.
-- +goose StatementEnd
