INSERT INTO portal_legal_docs (id, name, body, current_version, current_file_id, updated_at)
VALUES ($1, $2, $3, $4, $5, now());
