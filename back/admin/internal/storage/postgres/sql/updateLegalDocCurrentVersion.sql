UPDATE portal_legal_docs
SET current_version = $2,
    current_file_id = $3,
    updated_at = now()
WHERE id = $1;
