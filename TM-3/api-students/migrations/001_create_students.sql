CREATE TABLE IF NOT EXISTS students (
    id VARCHAR(50) PRIMARY KEY,
    nim VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    grade DOUBLE PRECISION NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- NIM harus unik.
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_key
ON students (nim);

-- Index tambahan untuk membantu pencarian nama.
CREATE INDEX IF NOT EXISTS students_name_lower_idx
ON students (LOWER(name));