Saya sedang mengerjakan Praktikum 7 Backend Lanjut menggunakan Go Fiber.

PROJECT YANG DIGUNAKAN:
Saya melanjutkan project dari Modul 6, yaitu:

api-students

Jangan membuat project baru.
Jangan mengganti project menjadi "latihan-fiber".
Project api-students adalah project utama yang harus dimodifikasi.

SUMBER UTAMA:
Gunakan PDF Modul 7 yang saya lampirkan sebagai acuan utama.
Ikuti struktur, istilah, urutan, requirement, acceptance test, dan tugas D.1-D.5 yang ada di modul.
Jangan mengganti requirement modul dengan implementasi dari sumber lain kecuali diperlukan secara teknis.
Jika ada bagian yang tidak didukung oleh modul, jelaskan terlebih dahulu dan jangan mengarang requirement.

KONTEKS PROJECT:
Project ini sudah menyelesaikan Modul 6:
- Go Fiber
- REST API students
- PostgreSQL
- Authentication JWT
- Refresh token
- RBAC
- Permission
- Ownership check
- owner_id pada students
- middleware RequireAuth
- middleware RequirePermission
- PermissionSet
- StudentService
- StudentRepository
- logging request
- migration 004_student_permissions.sql

Project sebelumnya sudah berhasil:
go test ./...

Dan server dapat berjalan dengan:
go run .

TUJUAN PRAKTIKUM 7:

1. Perbaiki 9 bug yang sengaja ditanam pada Bagian B Modul 7.
   - 3 compile error
   - 6 behavioral error

2. Setelah semua bug diperbaiki, implementasikan:
   - validasi deklaratif menggunakan go-playground/validator
   - custom validation sesuai requirement modul
   - cursor/keyset pagination untuk entity students
   - EXPLAIN ANALYZE dan index sesuai kebutuhan modul
   - content negotiation JSON/CSV
   - centralized ErrorHandler
   - request_id pada error response/logging sesuai requirement modul

3. Kerjakan tugas:
   - D.1 Berita Acara Perbaikan 9 Bug
   - D.2 Validasi Deklaratif
   - D.3 Cursor Pagination
   - D.4 Content Negotiation + Centralized ErrorHandler
   - D.5 Analisis

PENTING:
Jangan langsung mengubah seluruh project.

Gunakan workflow:
1. Inspect project terlebih dahulu.
2. Jalankan go build ./...
3. Identifikasi compile error.
4. Perbaiki SATU bug.
5. Jalankan go build ./... atau test yang relevan.
6. Catat bug tersebut untuk D.1.
7. Lanjut bug berikutnya.
8. Setelah 3 compile error selesai, cari 6 behavioral error menggunakan acceptance test Modul 7.
9. Jangan menghapus fitur Modul 6 yang masih dibutuhkan.
10. Jangan merusak authentication, RBAC, ownership, atau repository yang sudah bekerja.

UNTUK SETIAP BUG:
Berikan format:

BUG #:
- Lokasi file:
- Fungsi:
- Jenis: Compile Error / Behavioral Error
- Gejala:
- Penyebab:
- Perbaikan:
- Kode sebelum:
- Kode sesudah:
- Cara testing:
- Expected result:
- Actual result:
- Status:

JANGAN mengklaim test berhasil kalau belum benar-benar dijalankan.

UNTUK IMPLEMENTASI D.2-D.4:
Gunakan entity STUDENTS milik project saya, bukan membuat entity users baru.

Pertahankan struktur project yang sudah ada:
app/
  handler/
  model/
  repository/
  service/
config/
database/
helper/
middleware/
route/
migrations/

Sebelum membuat file baru:
- periksa apakah fungsi/struct yang dibutuhkan sudah ada.
- gunakan kembali kode yang sudah ada jika memungkinkan.
- jangan membuat duplicate type/function.
- jangan membuat duplicate middleware.
- jangan membuat duplicate response format.

KETIKA MEMBERIKAN PERUBAHAN KODE:
Berikan FULL FILE jika perubahan menyentuh banyak bagian file.
Jangan hanya memberikan potongan yang menyebabkan saya bingung harus menaruhnya di mana.

Setelah perubahan:
- jelaskan file mana yang berubah
- jelaskan alasan perubahan
- berikan command testing

VALIDASI:
Gunakan validator/v10 sesuai requirement Modul 7.
Jangan menghapus validasi bisnis yang sudah ada apabila masih diperlukan.
Pisahkan validasi format/input dari business authorization.

CURSOR PAGINATION:
Implementasikan untuk students sesuai requirement Modul 7.
Jangan mengubah menjadi offset pagination jika modul meminta cursor/keyset pagination.
Pastikan cursor:
- dapat digunakan untuk request halaman berikutnya
- memiliki deterministic ordering
- tidak menghasilkan duplicate data antar halaman
- menangani invalid cursor sesuai error contract
- mempertahankan filter/search/sort yang diwajibkan modul

CONTENT NEGOTIATION:
Implementasikan JSON dan CSV sesuai requirement Modul 7.
Pastikan:
- JSON tetap bekerja
- CSV bekerja ketika diminta client
- Content-Type benar
- error response tetap mengikuti centralized ErrorHandler

ERROR HANDLER:
Gunakan centralized Fiber ErrorHandler sesuai requirement Modul 7.
Jangan membuat setiap handler memiliki format error yang berbeda.
Pertahankan status code yang benar:
- 400 Bad Request
- 401 Unauthorized
- 403 Forbidden
- 404 Not Found
- 409 Conflict
- 422 Unprocessable Entity
- 500 Internal Server Error
sesuai konteks dan requirement modul.

TESTING:
Setiap tahap harus diuji.
Minimal gunakan:
go build ./...
go test ./...

Dan endpoint testing menggunakan Postman/curl bila diperlukan.

JANGAN:
- membuat project baru
- mengganti database
- mengganti framework
- menghapus authentication
- menghapus RBAC Modul 6
- menghapus ownership
- mengubah ID student menjadi integer jika project menggunakan string ID
- mengarang requirement yang tidak ada di Modul 7
- mengklaim screenshot/test berhasil tanpa bukti
- mengubah file yang tidak diperlukan

HASIL AKHIR YANG SAYA INGINKAN:
1. Project api-students berhasil build.
2. go test ./... berhasil.
3. 9 bug Modul 7 sudah diperbaiki.
4. D.2 selesai.
5. D.3 selesai.
6. D.4 selesai.
7. D.5 sudah dijawab.
8. Acceptance test Modul 7 terpenuhi.
9. Semua perubahan dijelaskan.
10. Dibuatkan bahan laporan:
    - potongan kode penting
    - Berita Acara 9 bug
    - tabel matriks
    - bukti testing
    - negative testing
    - analisis D.5
    - kesimpulan.

MULAI SEKARANG:

Jangan langsung menulis kode.

Langkah pertama:
1. Periksa struktur project api-students.
2. Periksa go.mod.
3. Jalankan/analisis:
   go build ./...
4. Identifikasi compile error pertama.
5. Tampilkan hanya diagnosis Bug #1 terlebih dahulu.
6. Tunggu konfirmasi saya sebelum lanjut ke Bug #2.