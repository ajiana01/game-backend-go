CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(24) NOT NULL UNIQUE,
    level INTEGER NOT NULL DEFAULT 1 CHECK (level >= 1),
    exp BIGINT NOT NULL DEFAULT 0 CHECK (exp >= 0),
    gold BIGINT NOT NULL DEFAULT 0 CHECK (gold >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rewards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('login', 'battle', 'quest')),
    gold BIGINT NOT NULL CHECK (gold >= 0),
    exp BIGINT NOT NULL CHECK (exp >= 0),
    item_id VARCHAR(100),
    item_name VARCHAR(120),
    item_quantity INTEGER NOT NULL DEFAULT 0 CHECK (item_quantity >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS rewards_player_created_idx
    ON rewards (player_id, created_at DESC);
