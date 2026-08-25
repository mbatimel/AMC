UPDATE sync_jobs
SET status = $2,
    last_error = $3,
    processed_at = now()
WHERE id = $1;
