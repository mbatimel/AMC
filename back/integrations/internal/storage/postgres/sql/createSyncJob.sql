INSERT INTO sync_jobs (system_id, direction, entity_type, status, attempts)
VALUES ($1, 'inbound', 'onec_full_sync', 'running', 1)
RETURNING id;
