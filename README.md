# API Students - Praktikum Backend Lanjut

API mahasiswa menggunakan Go, Fiber, dan PostgreSQL.

## 1. Teknologi

- Go
- Fiber
- PostgreSQL
- pgx
- godotenv

## 2. Struktur Database

Database yang digunakan:

- Database: praktikum_backend
- Table: students

### Struktur tabel

| Kolom | Tipe | Keterangan |
|---|---|---|
| id | varchar(50) | Primary Key |
| nim | varchar(20) | Wajib unik |
| name | varchar(100) | Nama mahasiswa |
| grade | double precision | Nilai mahasiswa |
| is_active | boolean | Status mahasiswa |
| created_at | timestamptz | Waktu pembuatan data |

### Index

Tabel students memiliki:

- Primary Key pada `id`
- UNIQUE constraint pada `nim`
- Index `students_name_lower_idx` pada `LOWER(name)`

UNIQUE pada NIM digunakan agar tidak terdapat dua mahasiswa dengan NIM yang sama.

Index pada nama digunakan untuk membantu pencarian nama yang tidak membedakan huruf besar dan kecil.

## 3. Persiapan Database

Buat database:

```sql
CREATE DATABASE praktikum_backend;