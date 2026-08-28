INSERT INTO integration_systems (code, name, is_active)
VALUES ($1, $2, TRUE)
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
RETURNING id;
