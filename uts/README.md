# SIAKAD Mini

## 1. Deskripsi
SIAKAD Mini adalah aplikasi sistem informasi akademik sederhana berbasis RESTful API. Proyek ini dibangun untuk memenuhi requirement UTS Praktikum Pemrograman Backend Lanjut. Aplikasi ini memfasilitasi autentikasi (JWT), manajemen data mahasiswa oleh admin, penyediaan katalog mata kuliah, hingga transaksi pengisian Kartu Rencana Studi (KRS) dengan validasi SKS otomatis berdasarkan IPK mahasiswa.

## 2. Teknologi
Aplikasi ini dibangun murni menggunakan bahasa Go dengan tumpukan pustaka pendukung berikut:
- **Go** (minimal versi 1.26.5)
- **Fiber v2** (Web Framework)
- **GORM** (ORM untuk PostgreSQL)
- **PostgreSQL** (Database Relasional)
- **golang-jwt/jwt/v5** (Autentikasi & Otorisasi Token)
- **golang.org/x/crypto/bcrypt** (Hashing Password)
- **go-playground/validator/v10** (Validasi Data DTO)

## 3. Fitur Utama
Aplikasi ini sudah mengimplementasikan beberapa mekanisme krusial:
- **Autentikasi & RBAC (Role-Based Access Control):** Mekanisme Login berbasis JWT dengan pemisahan akses yang ketat antara role `admin` dan `mahasiswa`. Terdapat juga *Rate Limiting* (Maksimal 5x gagal login per menit).
- **Manajemen Mahasiswa:** Fitur CRUD mahasiswa untuk admin, validasi input kompleks (panjang digit NIM, format Email, batasan nilai IPK), hingga fitur *Soft-Delete* yang mengunci akses login.
- **Katalog Mata Kuliah:** Menampilkan daftar mata kuliah lengkap dengan kalkulasi dinamis (SQL Group By & Subquery) untuk menghitung kursi terisi dan sisa kuota.
- **Transaksi Pengisian KRS (Enrollments):** Fitur transaksi akademik aman *Concurrency-proof* menggunakan transaksi database dan *Row Locking* (Pessimistic Locking `FOR UPDATE`). Memiliki validasi bentrok duplikasi, batas kuota, hingga perhitungan ketat sisa SKS (18, 21, atau 24 SKS) menyesuaikan riwayat IPK.
- **Pagination & Filtering:** Dukungan pencarian parameter kustom (seperti `?search=basis&semester=3`) dipadukan dengan kontrol rentang (*pagination*) yang responsif.
- **Response Envelope Standard:** Semua hasil sukses/error dikemas dalam struktur JSON seragam berserta komponen `meta` (Metadata) pagination jika diperlukan.

## 4. Arsitektur
Aplikasi ini menggunakan pola *Layered Architecture* (Arsitektur Berlapis) untuk menjamin pemisahan kepedulian *(Separation of Concerns)*:
- **Handler (`handler/`)**: Lapisan terluar. Hanya bertanggung jawab mencegat permintaan HTTP dari Fiber, memvalidasi _Request Body/Params_ lewat DTO Validator, meneruskan nilai ke lapisan Service, dan merakit JSON *Response Envelope*.
- **Service (`service/`)**: Lapisan pusat. Menyimpan inti dari *Business Logic* (aturan SKS, aturan kuota, kalkulasi). Service ini tidak memanggil sintaks SQL secara langsung melainkan berinteraksi melalui *interface* Repository.
- **Repository (`repository/`)**: Lapisan terdalam. Berisi isolasi interaksi murni menggunakan GORM (GORM Query Builder, Transaksi, *Locking*).

## 5. Prasyarat
Untuk menjalankan aplikasi ini, pastikan komputer Anda telah terinstal:
- **Go** (Versi 1.26.5 atau lebih baru)
- **PostgreSQL** Server aktif (Versi mana pun yang di-support pgx)

## 6. Instalasi dan Konfigurasi
Langkah-langkah untuk menyiapkan *environment* pengembangan:

1. **Clone repositori proyek ini**
   ```bash
   git clone <url-repository>
   cd uts
   ```

2. **Unduh seluruh dependency (Library)**
   ```bash
   go mod tidy
   ```

3. **Konfigurasi Environment**
   Salin berkas cetakan ke berkas _environment_ aktual:
   - Di Windows PowerShell: `cp .env.example .env`
   - Di Linux / macOS: `cp .env.example .env`

4. **Koneksi Database**
   Buka file `.env` di _text editor_ Anda. Sesuaikan variabel seperti `DB_USER`, `DB_PASSWORD`, dan `DB_NAME` dengan konfigurasi PostgreSQL lokal Anda. Pastikan nama database sudah Anda buat sebelumnya di DBMS (misal melalui `psql` atau pgAdmin).
   > **Catatan Keamanan:** Aturan `.gitignore` telah di-set agar mencegah file `.env` bocor ter-commit ke publik.

## 7. Database, Migration, dan Seeder
Setelah `.env` disesuaikan, jalankan dua skrip independen ini untuk menyiapkan skema *(schema)* dan memuat data *dummy*.

- **Jalankan Migration (Pembuatan Tabel):**
  ```bash
  go run database/migration/migration.go
  ```
- **Jalankan Seeder (Pengisian Data):**
  ```bash
  go run database/seeder/seeder.go
  ```
- *(Opsional)* Jika terjadi kesalahan, Anda dapat mereset ulang seluruh tabel dengan argumen rollback: `go run database/migration/migration.go -rollback`

## 8. Menjalankan Aplikasi
Untuk memulai _Web Server_ Fiber:
```bash
go run main.go
```
Secara default (berdasar parameter `APP_PORT` di `.env.example`), server akan mengikat lalu-lintas di `http://127.0.0.1:3000`. 
Anda dapat menguji apakah server/database menyala normal dengan mengakses URL: `GET http://localhost:3000/api/v1/health`.

## 9. Akun Demo
Sistem `seeder.go` telah menyuntikkan beberapa akun untuk pengujian instan.
- **Role Admin**
  - Email: `admin@siakad.com`
  - Password: `admin123`
- **Role Mahasiswa**
  - Tersedia 20 mahasiswa dengan email `mhs01@student.com` hingga `mhs20@student.com`.
  - Password: `mhs12345`

> **Peringatan:** Kredensial ini hanya dialokasikan untuk keperluan praktikum lokal. Seluruh _password hash_ menggunakan algoritma `bcrypt`.

## 10. API Documentation

| HTTP Method | Endpoint | Role Diizinkan | Deskripsi | Status Code |
|---|---|---|---|---|
| `POST` | `/api/v1/auth/login` | Publik | Otentikasi user & mengembalikan Token JWT. | `200, 401, 422, 429` |
| `GET` | `/api/v1/auth/me` | Admin, Mahasiswa | Mengekstrak identitas & role pengguna dari token. | `200, 401` |
| `GET` | `/api/v1/students` | Admin | Menampilkan daftar mahasiswa, mendukung *pagination*. | `200, 401, 403` |
| `POST` | `/api/v1/students` | Admin | Menambahkan data mahasiswa baru ke sistem. | `201, 403, 409, 422` |
| `GET` | `/api/v1/students/:id` | Admin, Mahasiswa | Detail mhs & rekapan KRS (*mhs hanya bisa melihat profil sendiri*). | `200, 403, 404` |
| `PUT` | `/api/v1/students/:id` | Admin | Memperbarui data mhs secara utuh (kecuali NIM). | `200, 403, 404, 422` |
| `DELETE` | `/api/v1/students/:id` | Admin | Menghapus data mhs menggunakan metode *Soft Delete*. | `204, 403, 404` |
| `GET` | `/api/v1/courses` | Admin, Mahasiswa | Daftar lengkap mata kuliah, sisa kuota, dan *pagination*. | `200, 401` |
| `POST` | `/api/v1/enrollments` | Mahasiswa | Mengajukan KRS ke suatu *course* berdasarkan sisa IPK & kuota. | `201, 403, 409, 422` |
| `DELETE` | `/api/v1/enrollments/:id` | Mahasiswa | Membatalkan KRS (*hanya boleh membatalkan milik sendiri*). | `204, 403, 404` |

*Format Response (Contoh Berhasil)*
```json
{
  "success": true,
  "message": "Pesan keberhasilan sistem",
  "data": { ... },
  "meta": { "current_page": 1, "per_page": 10, "total_data": 5, "total_page": 1 }
}
```
