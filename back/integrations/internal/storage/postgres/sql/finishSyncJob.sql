UPDATE sync_jobs
SET status = $2,
    processed_at = now()
WHERE id = $1;
