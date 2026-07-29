SELECT EXISTS (
    SELECT 1
    FROM counterparties
    WHERE id = $1
)
