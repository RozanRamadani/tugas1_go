CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Username unik tanpa membedakan huruf besar dan kecil
CREATE UNIQUE INDEX IF NOT EXISTS users_username_lower_key
ON users (LOWER(username));

-- Index untuk membantu pencarian email
CREATE INDEX IF NOT EXISTS users_email_lower_idx
ON users (LOWER(email));