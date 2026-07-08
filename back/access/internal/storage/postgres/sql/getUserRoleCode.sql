SELECT roles.code
FROM user_roles
JOIN roles ON roles.id = user_roles.role_id
WHERE user_roles.user_id = $1
ORDER BY roles.code ASC
LIMIT 1
