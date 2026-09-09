UPDATE users
SET is_active = false,
    status = 'inactive',
    blocked_reason = NULLIF($2, ''),
    blocked_contact_name = NULLIF($3, ''),
    blocked_contact_phone = NULLIF($4, ''),
    blocked_contact_email = NULLIF($5, ''),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
