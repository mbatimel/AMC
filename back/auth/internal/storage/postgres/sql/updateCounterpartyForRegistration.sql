UPDATE counterparties
SET short_name = COALESCE($2, short_name),
    inn = COALESCE($3, inn),
    director_full_name = COALESCE($4, director_full_name),
    phone = COALESCE($5, phone),
    email = COALESCE(NULLIF($6, ''), email),
    requisites_file_url = NULLIF($7, ''),
    requisites_file_name = NULLIF($8, ''),
    updated_at = now()
WHERE id = $1
