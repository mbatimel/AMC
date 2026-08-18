UPDATE portal_signup_requests
SET status = $2, reject_reason = $3, decided_at = now()
WHERE id = $1 AND status = 'pending'
RETURNING id, company, inn, contact, email, phone, request_type, status, reject_reason, created_at, decided_at
