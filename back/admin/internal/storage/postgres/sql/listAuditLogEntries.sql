SELECT id, actor_user_id, actor_label, action, created_at
FROM admin_audit_log
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
