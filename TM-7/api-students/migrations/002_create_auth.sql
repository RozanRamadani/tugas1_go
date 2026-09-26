-- ============================================================
-- AUTHENTICATION
-- ============================================================

-- Tabel users
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,

    username VARCHAR(50) NOT NULL UNIQUE,

    email VARCHAR(100) NOT NULL UNIQUE,

    password TEXT NOT NULL,

    role VARCHAR(20) NOT NULL DEFAULT 'user',

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tambahkan column role jika tabel users sudah ada dari praktikum sebelumnya
ALTER TABLE users
ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';

-- Role dibatasi agar tidak sembarang nilai masuk.
ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
ADD CONSTRAINT users_role_check
CHECK (role IN ('user', 'admin'));

-- Index username untuk login.
CREATE INDEX IF NOT EXISTS users_username_idx
ON users (username);


-- ============================================================
-- REFRESH TOKENS
-- ============================================================

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,

    user_id INTEGER NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    token_hash CHAR(64) NOT NULL UNIQUE,

    expires_at TIMESTAMPTZ NOT NULL,

    revoked_at TIMESTAMPTZ NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Mempercepat pencarian token milik user.
CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx
ON refresh_tokens (user_id);

-- Mempercepat pengecekan token yang masih aktif.
CREATE INDEX IF NOT EXISTS refresh_tokens_active_idx
ON refresh_tokens (token_hash, revoked_at, expires_at);