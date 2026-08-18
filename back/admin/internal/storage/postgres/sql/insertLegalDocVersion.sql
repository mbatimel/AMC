INSERT INTO portal_legal_doc_versions (doc_id, version, summary, author, file_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at;
