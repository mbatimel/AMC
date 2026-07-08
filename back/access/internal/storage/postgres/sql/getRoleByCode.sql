SELECT id,
       code,
       COALESCE(name, ''),
       COALESCE(description, ''),
       created_at,
       updated_at
FROM roles
WHERE code = $1
