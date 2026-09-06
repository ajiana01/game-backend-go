-- Jalankan melalui SQLTools setelah memilih koneksi
-- "Portfolio Go - PostgreSQL".

SELECT
    current_database() AS database_name,
    current_user AS connected_as,
    version() AS postgres_version;

SELECT id, username, level, exp, gold, created_at, updated_at
FROM players
ORDER BY created_at DESC
LIMIT 100;

SELECT
    id,
    player_id,
    type,
    gold,
    exp,
    item_name,
    item_quantity,
    created_at
FROM rewards
ORDER BY created_at DESC
LIMIT 100;
