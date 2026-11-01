-- Migration 007: Supervisors dan Supervisions (Pembimbing Tugas Akhir)
-- psql -U postgres -d praktikum_backend -f migrations/007_supervisors.sql

CREATE TABLE IF NOT EXISTS supervisors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    quota INTEGER NOT NULL CHECK (quota > 0),
    supervised_count INTEGER NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Jaring pengaman: supervised_count tidak boleh negatif atau melebihi kuota
    CONSTRAINT supervisors_count_within_quota
        CHECK (supervised_count >= 0 AND supervised_count <= quota)
);

CREATE TABLE IF NOT EXISTS supervisions (
    id BIGSERIAL PRIMARY KEY,
    -- Seorang mahasiswa hanya boleh memiliki SATU pembimbing (dijaga UNIQUE constraint di DB)
    student_id VARCHAR(50) NOT NULL UNIQUE REFERENCES students(id) ON DELETE CASCADE,
    supervisor_id INTEGER NOT NULL REFERENCES supervisors(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS supervisions_supervisor_idx ON supervisions (supervisor_id);

-- Seed data Dosen Pembimbing
INSERT INTO supervisors (id, name, quota) VALUES
    (1, 'Dr. Budi Prasetyo, M.Kom', 3),     -- Kuota 3 untuk target load test D.3
    (2, 'Dr. Siti Aminah, M.T', 10),
    (3, 'Prof. Bambang Suryo, Ph.D', 10)
ON CONFLICT (id) DO NOTHING;

-- Seed data 40 Mahasiswa untuk Uji Beban D.3
INSERT INTO students (id, nim, name, grade, is_active)
SELECT 
    'std_' || LPAD(g::TEXT, 3, '0'),
    '112233' || LPAD(g::TEXT, 3, '0'),
    'Mahasiswa Uji ' || LPAD(g::TEXT, 3, '0'),
    3.50,
    true
FROM generate_series(1, 40) AS g
ON CONFLICT (id) DO NOTHING;

-- Permission baru untuk manajemen pembimbingan
INSERT INTO permissions (name, description) VALUES
    ('supervision:manage', 'Menetapkan dan mengelola pembimbing tugas akhir')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'supervision:manage')
ON CONFLICT DO NOTHING;
