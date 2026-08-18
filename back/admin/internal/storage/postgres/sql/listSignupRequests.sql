SELECT id, company, inn, contact, email, phone, request_type, status, reject_reason, created_at, decided_at
FROM portal_signup_requests
WHERE $1 = '' OR status = $1
ORDER BY created_at DESC
