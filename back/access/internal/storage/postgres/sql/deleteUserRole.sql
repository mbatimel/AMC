DELETE FROM user_roles
WHERE user_id = $1 AND role_id = $2
