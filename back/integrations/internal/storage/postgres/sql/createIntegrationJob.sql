INSERT INTO sync_jobs (system_id, direction, entity_type, status, attempts)
VALUES ($1, $2, $3, 'running', 1)
RETURNING id;
