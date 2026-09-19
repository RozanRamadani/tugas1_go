-- ---------------------------------------------------------------
-- 1. Tambah permission untuk entity students
-- ---------------------------------------------------------------
INSERT INTO permissions (name, description) VALUES
    ('student:list', 'Melihat daftar seluruh mahasiswa'),
    ('student:read:any', 'Melihat data mahasiswa mana pun'),
    ('student:create', 'Menambah data mahasiswa baru'),
    ('student:update:any', 'Mengubah data mahasiswa mana pun'),
    ('student:delete', 'Menghapus data mahasiswa')
ON CONFLICT (name) DO NOTHING;

-- ---------------------------------------------------------------
-- 2. Pasangkan permission ke role
-- ---------------------------------------------------------------
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete'),
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create')
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------
-- 3. Tambahkan kolom owner_id pada tabel students
-- ---------------------------------------------------------------
ALTER TABLE students
ADD COLUMN IF NOT EXISTS owner_id INTEGER;

-- Mengisi data lama dengan ID user admin/pertama jika ada data lama yang NULL
UPDATE students 
SET owner_id = (SELECT id FROM users ORDER BY id ASC LIMIT 1)
WHERE owner_id IS NULL AND EXISTS (SELECT 1 FROM users);

-- Baru tambahkan FOREIGN KEY constraint
ALTER TABLE students
DROP CONSTRAINT IF EXISTS students_owner_fkey;

ALTER TABLE students
ADD CONSTRAINT students_owner_fkey
FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS students_owner_idx ON students (owner_id);
