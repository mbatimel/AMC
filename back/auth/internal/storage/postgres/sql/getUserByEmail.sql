SELECT id, email, password, status, is_active,
       blocked_reason, blocked_contact_name, blocked_contact_phone, blocked_contact_email
FROM users
WHERE email = $1
  AND deleted_at IS NULL
