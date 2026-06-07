-- 0001_init.sql : početna ShopHub šema (ekvivalent GORM AutoMigrate).
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id                     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email                  TEXT NOT NULL UNIQUE,
    password_hash          TEXT NOT NULL,
    display_name           TEXT NOT NULL,
    refresh_token_version  INTEGER NOT NULL DEFAULT 0,
    created_at             TIMESTAMPTZ,
    updated_at             TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS shops (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    availability  VARCHAR(16) NOT NULL DEFAULT 'standard',
    wallet_addr   TEXT NOT NULL,
    database_type VARCHAR(16) NOT NULL DEFAULT 'postgres',
    image         TEXT NOT NULL,
    created_at    TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_shops_owner_id ON shops(owner_id);
