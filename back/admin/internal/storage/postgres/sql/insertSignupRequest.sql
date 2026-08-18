INSERT INTO portal_signup_requests (id, company, inn, contact, email, phone, request_type, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending')
RETURNING id, company, inn, contact, email, phone, request_type, status, reject_reason, created_at, decided_at
