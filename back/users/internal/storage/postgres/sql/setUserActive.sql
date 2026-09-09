UPDATE users
SET is_active = $2,
    status = $3,
    blocked_reason = CASE WHEN $2 THEN NULL ELSE blocked_reason END,
    blocked_contact_name = CASE WHEN $2 THEN NULL ELSE blocked_contact_name END,
    blocked_contact_phone = CASE WHEN $2 THEN NULL ELSE blocked_contact_phone END,
    blocked_contact_email = CASE WHEN $2 THEN NULL ELSE blocked_contact_email END,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
