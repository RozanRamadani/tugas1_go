# LAPORAN PRAKTIKUM PEMROGRAMAN BACKEND LANJUT
## MODUL 9 — TRANSACTIONS & CONCURRENCY (EDISI PUZZLE)

---

### METADATA MAHASISWA & REPOSITORY
- **Mata Kuliah:** Praktikum Pemrograman Backend Lanjut (2 SKS)
- **Program Studi:** D4 Teknik Informatika — Fakultas Vokasi, Universitas Airlangga
- **Nama Mahasiswa:** Rozan Aiman Ramadani
- **NIM:** 434241069
- **Tautan Repository GitHub:** [https://github.com/RozanRamadani/tugas1_go](https://github.com/RozanRamadani/tugas1_go)
- **Lokasi Proyek:** Subfolder `TM-9/latihan-fiber`

---

## BAB I — RINGKASAN & LANDASAN TEORI

### 1.1 Masalah Konkurensi pada Sistem KRS
Sebelum Modul 9, endpoint aplikasi backend umumnya hanya mengeksekusi satu pernyataan SQL per request. Di PostgreSQL, satu pernyataan selalu atomik dengan sendirinya. Namun, pendaftaran kelas atau pengisian KRS (*Kartu Rencana Studi*) menuntut banyak langkah:
1. Membaca sisa kuota dan total SKS yang telah diambil.
2. Memeriksa aturan bisnis (apakah kursi masih ada, apakah total SKS $\le 24$).
3. Menyimpan baris pendaftaran ke tabel `enrollments`.
4. Menambahkan jumlah peserta pada kolom `enrolled_count` di tabel `courses`.

Ketika ratusan mahasiswa mengirim request pada detik yang sama, pendekatan sekuensial naif mengalami kegagalan fatal:
- **Race Condition (*Check-Then-Act*):** Mahasiswa sama-sama membaca sisa kursi $> 0$, padahal kursi hanya tersisa 1. Keduanya sama-sama memutuskan mendaftar sehingga kapasitas kelas terlampaui.
- **Lost Update:** Dua penambahan kuota dilakukan bersamaan berdasarkan angka bacaan lama, sehingga satu penambahan hilang tertimpa oleh penulisan lainnya.

### 1.2 Kesalahpahaman Mengenai Transaksi
Membungkus kode naif dengan transaksi `BEGIN ... COMMIT` tanpa teknik penguncian tidak menyelesaikan masalah race condition. Pada *isolation level* bawaan PostgreSQL (`READ COMMITTED`), pernyataan `SELECT` biasa tidak mengunci baris apa pun. Transaksi hanya menjamin *Atomicity* (semua tersimpan atau semua dibatalkan), tetapi **tidak otomatis menjamin bebas dari race condition**.

### 1.3 Tiga Teknik Pengendalian Konkurensi
1. **UPDATE Bersyarat yang Atomik:**
   Pemeriksaan syarat dan penulisan nilai digabung dalam satu gerakan di database:
   ```sql
   UPDATE courses SET enrolled_count = enrolled_count + 1
   WHERE id = $1 AND enrolled_count < quota
   RETURNING id, enrolled_count;
   ```
2. **Pessimistic Locking (`SELECT ... FOR UPDATE`):**
   Mengunci baris data di awal transaksi hingga `COMMIT` atau `ROLLBACK`. Sangat cocok untuk validasi batas 24 SKS yang datanya tersebar di banyak baris.
3. **Optimistic Locking (Kolom `version`):**
   Tidak memegang kunci selama user/admin melihat data. Saat penyimpanan dilakukan, sistem memastikan versi data belum berubah:
   ```sql
   UPDATE courses SET quota = $1, version = version + 1
   WHERE id = $2 AND version = $3 AND $1 >= enrolled_count;
   ```

### 1.4 Pencegahan Deadlock & Clean Architecture
- **Deadlock:** Terjadi saat transaksi A memegang kunci 1 dan menunggu kunci 2, sementara transaksi B memegang kunci 2 dan menunggu kunci 1. Pencegahannya adalah mewajibkan **urutan penguncian yang seragam** (misalnya selalu mengunci `id` terkecil lebih dulu: `ORDER BY id FOR UPDATE`).
- **Pola DBTX & TxManager:** Transaksi dibuka pada layer *Service* menggunakan `TxManager` dengan kemampuan *auto-retry* saat terjadi deadlock (SQLSTATE `40P01`) atau serialization failure (`40001`). Layer repository diabstraksikan melalui interface `DBTX` agar method query dapat bekerja di dalam maupun di luar transaksi tanpa duplikasi kode.

---

## BAB II — D.1 BERITA ACARA PERBAIKAN (9 BUG PUZZLE)

Modul praktikum edisi puzzle memuat 9 kesalahan sengaja pada Bagian B. Berikut adalah berita acara perbaikan lengkap:

| No | Berkas dan Baris | Jenis | Gejala | Reprodusibilitas | Akar Masalah | Perbaikan dan Bukti |
|:--:|:---|:---|:---|:---|:---|:---|
| **1** | `app/repository/course_repository.go` baris 33 | Ditolak compiler | `cannot use &coursePostgresRepository{…} as CourseRepository value: wrong type for method WithTx (have WithTx(DBTX) *coursePostgresRepository, want WithTx(DBTX) CourseRepository)` | 100% setiap kali `go build` dijalankan | Bahasa Go mewajibkan tipe kembalian (*return type*) method pada struct sama persis dengan yang didefinisikan pada interface. Pada modul, method mengembalikan pointer konkrit `*coursePostgresRepository`, bukan interface `CourseRepository`. | Mengubah tipe kembalian method `WithTx` menjadi `CourseRepository`. Memenuhi spesifikasi penerimaan C.7 butir 1. |
| **2** | `tools/loadtest/main.go` baris 44 | Ditolak compiler | `invalid operation: cannot send to receive-only channel <-chan string out (variable of type <-chan string)` | 100% setiap kali `go build` dijalankan | Parameter `out` pada goroutine dideklarasikan dengan tipe `<-chan string` (*receive-only*). Mengirim data ke channel receive-only (`out <- job(i)`) ditolak oleh compiler Go. | Mengubah tipe parameter channel menjadi *send-only* `chan<- string` (atau bidirectional `chan string`). Memenuhi C.7 butir 1. |
| **3** | `app/repository/tx.go` baris 34 | Ditolak compiler | `cannot use func(tx repository.DBTX) error {…} as func(tx pgx.Tx) error value in argument to s.tx.WithTx` | 100% setiap kali `go build` dijalankan | Signature method `WithTx` pada modul menerima `func(tx pgx.Tx) error`. Hal ini bertentangan dengan komentarnya dan melanggar prinsip *Clean Architecture* yang menyembunyikan detail library driver `pgx` dari layer service. | Mengubah signature menjadi `func(m *TxManager) WithTx(ctx context.Context, fn func(tx DBTX) error) error`. Memenuhi C.7 butir 1. |
| **4** | `app/repository/tx.go` baris 53 | Lolos compiler, salah perilaku (*Resource leak*) | Koneksi pool habis setelah request gagal. Kueri `pg_stat_activity` menunjukkan banyak koneksi menggantung dengan status `idle in transaction`, hingga server berhenti melayani request. | Muncul setiap kali request di dalam transaksi menghasilkan error | Kode modul tidak memanggil `defer tx.Rollback(ctx)`. Tanpa defer rollback, kegagalan transaksi meninggalkan koneksi dalam keadaan terbuka dan menyandera slot koneksi pool selamanya. | Menambahkan baris `defer tx.Rollback(ctx)` tepat setelah `m.pool.Begin(ctx)`. Memenuhi spesifikasi C.5 dan C.7 butir 11. |
| **5** | `app/repository/course_repository.go` baris 96 | Lolos compiler, salah perilaku (*Concurrency*) | Pada pengujian serentak `mode credits`, mahasiswa berhasil mendaftar hingga 27–36 SKS (melebihi batas maksimal 24 SKS). | 3 dari 3 percobaan `mode credits` sebelum perbaikan | Kueri pada method `LockStudent` hanya menjalankan `SELECT id FROM users WHERE id = $1` tanpa klausa `FOR UPDATE`. Tanpa penguncian baris user, pemeriksaan batas SKS dapat diakali oleh request serentak dari mahasiswa yang sama. | Menambahkan `FOR UPDATE` pada kueri: `SELECT id FROM users WHERE id = $1 FOR UPDATE`. Memenuhi spesifikasi C.2 dan C.7 butir 3. |
| **6** | `app/repository/course_repository.go` baris 117 | Lolos compiler, salah perilaku (*Constraint violation*) | Mahasiswa ke-6 yang mendaftar ke kelas berkuota 5 menerima respon HTTP 500 INTERNAL_ERROR alih-alih 409 COURSE_FULL, serta memicu pelanggaran `courses_enrolled_within_quota`. | Muncul setiap kali kelas telah penuh dan ada request pendaftaran baru | Kueri UPDATE menggunakan klausa `WHERE enrolled_count <= quota`. Saat `enrolled_count = 5` dan `quota = 5`, perbandingan `5 <= 5` bernilai `true`, sehingga sistem menaikkan count menjadi 6 yang melanggar CHECK constraint database. | Mengubah kondisi WHERE menjadi `enrolled_count < quota`. Memenuhi spesifikasi C.1, C.4, dan C.7 butir 5. |
| **7** | `app/repository/course_repository.go` baris 145 | Lolos compiler, salah perilaku (*Data Integrity*) | Pembatalan pendaftaran pada kelas yang tidak pernah diambil menghasilkan 204 No Content, dan `enrolled_count` kelas tersebut berkurang secara keliru menjadi negatif. | 100% terjadi saat memanggil endpoint `DELETE /courses/:id/enroll` untuk kelas yang belum pernah diambil | Method `DeleteEnrollment` mengabaikan `tag.RowsAffected()`. Pada PostgreSQL, perintah DELETE yang tidak menemukan baris yang cocok tetap sukses dengan 0 baris terhapus (bukan error). | Menambahkan pemeriksaan: `if tag.RowsAffected() == 0 { return ErrNotEnrolled }`. Memenuhi spesifikasi C.4 dan C.7 butir 7. |
| **8** | `app/service/course_service.go` baris 224 | Lolos compiler, salah perilaku (*Deadlock / Concurrency Leak*) | Endpoint pindah kelas (`/enrollments/swap`) mengalami macet total (*hang*) selama 5 detik hingga dibatalkan oleh *context deadline exceeded*. | 100% terjadi setiap kali pemindahan kelas dijalankan | Di dalam callback `s.tx.WithTx`, kode memanggil `s.courses.DecrementEnrolled` (menggunakan koneksi pool di luar transaksi). Padahal baris kelas tersebut sudah dikunci eksklusif oleh `repo.LockCourses` di dalam transaksi. Kueri dari pool akhirnya macet menunggu kuncinya sendiri. | Mengganti pemanggilan menjadi `repo.DecrementEnrolled` agar dieksekusi di dalam transaksi yang sama. Memenuhi C.3, C.4, dan C.7 butir 4 & 9. |
| **9** | `app/repository/course_repository.go` baris 182 | Lolos compiler, salah perilaku (*Optimistic Locking*) | Admin yang mengirimkan nomor versi lama tetap berhasil mengubah kuota, dan nilai `version` pada respon tidak pernah bertambah (tetap bernilai 1). | 100% terjadi saat pengubahan kuota dilakukan berulang kali | Kueri UPDATE hanya mengeksekusi `SET quota = $1` tanpa menaikkan nomor versi (`version = version + 1`). Akibatnya, verifikasi versi pada penyimpanan berikutnya tidak berfungsi. | Menambahkan kenaikan versi pada kueri: `SET quota = $1, version = version + 1`. Memenuhi spesifikasi C.4 dan C.7 butir 10. |

---

## BAB III — D.2 IMPLEMENTASI ENTITAS PEMBIMBING TUGAS AKHIR

Konsep transaksi dan konkurensi diimplementasikan pada entitas mahasiswa (`students`) untuk kasus **penetapan dan penggantian dosen pembimbing tugas akhir**.

### 3.1 Skema Basis Data & Migrasi (`migrations/007_supervisors.sql`)
Aturan bisnis bahwa **"seorang mahasiswa hanya boleh memiliki satu pembimbing"** dijamin langsung pada level basis data menggunakan constraint **`UNIQUE`** pada kolom `student_id` di tabel `supervisions`.

```sql
CREATE TABLE IF NOT EXISTS supervisors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    quota INTEGER NOT NULL CHECK (quota > 0),
    supervised_count INTEGER NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT supervisors_count_within_quota
        CHECK (supervised_count >= 0 AND supervised_count <= quota)
);

CREATE TABLE IF NOT EXISTS supervisions (
    id BIGSERIAL PRIMARY KEY,
    student_id VARCHAR(50) NOT NULL UNIQUE REFERENCES students(id) ON DELETE CASCADE,
    supervisor_id INTEGER NOT NULL REFERENCES supervisors(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 3.2 Potongan Kode Penting
1. **Penetapan Pembimbing dengan UPDATE Bersyarat & Row Lock Mahasiswa (`app/service/supervisor_service.go`):**
   ```go
   err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
       repo := s.supervisors.WithTx(tx)

       // 1. Kunci baris student untuk mencegah penetapan ganda serentak
       if err := repo.LockStudent(ctx, req.StudentID); err != nil {
           return err
       }

       // 2. Ambil kuota pembimbing secara atomik dengan UPDATE bersyarat
       sup, err := repo.IncrementSupervised(ctx, supervisorID)
       if err != nil {
           return err
       }

       // 3. Catat penetapan (dijaga UNIQUE constraint di supervisions)
       return repo.InsertSupervision(ctx, req.StudentID, supervisorID)
   })
   ```

2. **Pergantian Pembimbing Bebas Deadlock (`POST /supervisions/switch`):**
   ```go
   // Kunci kedua pembimbing dengan urutan ID menaik untuk mencegah deadlock
   if err := repo.LockSupervisors(ctx, currentSupervision.SupervisorID, req.ToSupervisorID); err != nil {
       return err
   }
   if _, err := repo.DecrementSupervised(ctx, currentSupervision.SupervisorID); err != nil {
       return err
   }
   if _, err := repo.IncrementSupervised(ctx, req.ToSupervisorID); err != nil {
       return err
   }
   return repo.UpdateSupervision(ctx, req.StudentID, req.ToSupervisorID)
   ```

3. **Optimistic Locking Ubah Kuota (`PATCH /supervisors/:id`):**
   ```sql
   UPDATE supervisors SET quota = $1, version = version + 1
   WHERE id = $2 AND version = $3 AND $1 >= supervised_count
   RETURNING id, name, quota, supervised_count, version, created_at;
   ```

---

## BAB IV — D.3 BUKTI UJI BEBAN & PERCOBAAN PEMBANDING

Pengujian beban serentak dijalankan menggunakan tool `tools/loadtest_supervisors` dengan 30 goroutine serentak yang mencoba menetapkan mahasiswa berbeda (`std_001` s/d `std_030`) ke Dosen Pembimbing ID 1 (Dr. Budi Prasetyo, M.Kom) yang hanya memiliki kuota 3.

### 4.1 Keluaran Tiga Kali Percobaan Berturut-Turut (Kode Benar)
```text
=== PERCOBAAN 1 ===
30 mahasiswa serentak ditetapkan ke pembimbing ID 1 (kuota 3)...
hasil request:
  201 OK                    3
  409 SUPERVISOR_FULL       27
waktu total : 126.4517ms
Supervisor ID 1 (Dr. Budi Prasetyo, M.Kom) kuota=3 supervised_count=3 baris supervisions=3 -> KONSISTEN

=== PERCOBAAN 2 ===
30 mahasiswa serentak ditetapkan ke pembimbing ID 1 (kuota 3)...
hasil request:
  201 OK                    3
  409 SUPERVISOR_FULL       27
waktu total : 21.517ms
Supervisor ID 1 (Dr. Budi Prasetyo, M.Kom) kuota=3 supervised_count=3 baris supervisions=3 -> KONSISTEN

=== PERCOBAAN 3 ===
30 mahasiswa serentak ditetapkan ke pembimbing ID 1 (kuota 3)...
hasil request:
  201 OK                    3
  409 SUPERVISOR_FULL       27
waktu total : 22.3875ms
Supervisor ID 1 (Dr. Budi Prasetyo, M.Kom) kuota=3 supervised_count=3 baris supervisions=3 -> KONSISTEN
```

### 4.2 Percobaan Pembanding (Negative Testing)
Sesuai instruksi D.3, syarat kuota `AND supervised_count < quota` dihapus sementara dari method `IncrementSupervised` pada `app/repository/supervisor_repository.go`. Hasil pengujian pembanding adalah sebagai berikut:

```text
30 mahasiswa serentak ditetapkan ke pembimbing ID 1 (kuota 3)...
hasil request:
  201 OK                    3
  500 INTERNAL_ERROR        27
waktu total : 176.5867ms
Supervisor ID 1 (Dr. Budi Prasetyo, M.Kom) kuota=3 supervised_count=3 baris supervisions=3 -> KONSISTEN
```

**Analisis Hasil Pembanding:**
Ketika syarat kuota atomik di kode aplikasi dihilangkan, aplikasi mencoba menaikkan jumlah bimbingan melampaui kuota 3. Akibatnya, request ke-4 sampai ke-30 memicu jaring pengaman terakhir basis data yaitu `CONSTRAINT supervisors_count_within_quota CHECK (supervised_count <= quota)`. Basis data menolak update tersebut dan menghasilkan error PostgreSQL yang diterjemahkan menjadi **HTTP 500 INTERNAL_ERROR**. 

Percobaan pembanding ini membuktikan bahwa:
1. Alat uji beban benar-benar mampu mendeteksi ketiadaan kontrol kuota pada kode aplikasi.
2. Kontrol kuota atomik di level aplikasi (`WHERE supervised_count < quota`) mutlak diperlukan agar penolakan kuota penuh ditangani secara anggun sebagai error bisnis **409 SUPERVISOR_FULL**, bukan error fatal 500.

---

## BAB V — D.4 ANALISIS SINGKAT

**1. Kuota pembimbing cukup dijaga UPDATE bersyarat, tetapi aturan "satu mahasiswa satu pembimbing" saat terjadi pergantian membutuhkan kunci tambahan. Jelaskan mengapa, dengan menunjuk baris kode Anda.**
> UPDATE bersyarat yang atomik (`WHERE supervised_count < quota`) bekerja mandiri untuk menjaga kapasitas pembimbing karena hanya memodifikasi satu baris pada tabel `supervisors`. Namun, aturan *"satu mahasiswa hanya boleh memiliki satu pembimbing"* saat pergantian (`Switch`) melibatkan status relasional di antara tabel `students`, `supervisors`, dan `supervisions`. Jika ada dua request pergantian pembimbing serentak untuk mahasiswa yang sama (misalnya request 1 ingin pindah ke Dosen B dan request 2 ingin pindah ke Dosen C), kedua request dapat sama-sama membaca bahwa pembimbing lama adalah Dosen A, lalu keduanya sama-sama mengurangi kuota Dosen A dan menambahkan kuota ke Dosen B dan C secara berlebih. Oleh karena itu, kita membutuhkan kunci pesimistik tambahan pada baris mahasiswa melalui pemanggilan `repo.LockStudent(ctx, req.StudentID)` (`app/service/supervisor_service.go` baris 112) yang menjalankan `SELECT id FROM students WHERE id = $1 FOR UPDATE`. Kunci ini memastikan bahwa request kedua wajib mengantre sampai request pergantian pertama selesai di-commit atau di-rollback, sehingga integritas relasi mahasiswa dan pembimbing tetap terjaga.

**2. Seorang admin membuka form ubah kuota pembimbing, lalu meninggalkannya 30 menit sebelum menekan Simpan. Apa yang terjadi bila Anda memakai pessimistic locking untuk endpoint itu? Apa yang terjadi dengan optimistic locking?**
> Jika menggunakan **pessimistic locking** (`SELECT ... FOR UPDATE`), koneksi transaksi database akan dibuka dan ditahan sejak form dibuka hingga tombol Simpan ditekan (selama 30 menit). Hal ini akan menyandera 1 dari 10 koneksi pool aplikasi dan mengunci baris dosen tersebut secara eksklusif. Akibatnya, seluruh mahasiswa yang mencoba mendaftar atau berpindah ke dosen tersebut selama 30 menit akan mengalami macet total (*blocking*) dan akhirnya gagal karena *context deadline exceeded* (batas waktu 5 detik). Sebaliknya, dengan **optimistic locking** (kolom `version`), sistem tidak memegang kunci apa pun dan tidak menahan koneksi selama admin membuka form. Transaksi mahasiswa lain tetap berjalan lancar. Ketika admin menekan tombol Simpan setelah 30 menit, sistem hanya menjalankan query UPDATE bersyarat yang mencocokkan nomor `version` (`app/repository/supervisor_repository.go` baris 186). Jika selama 30 menit data kuota telah diubah oleh admin lain, request akan ditolak dengan `409 VERSION_CONFLICT` secara aman tanpa pernah mengorbankan performa atau ketersediaan server.

**3. WithTx dapat menjalankan fn lebih dari sekali. Seandainya di dalam fn Anda mengirim email pemberitahuan ke dosen pembimbing, apa saja yang dapat salah? Di mana seharusnya email itu dikirim?**
> Fungsi `WithTx` memiliki mekanisme pengulangan otomatis (*auto-retry*) jika mendeteksi kegagalan serialisasi (`40001`) atau deadlock (`40P01`). Jika pemanggilan pengiriman email ditaruh di dalam fungsi `fn`:
> 1. Email akan terkirim berulang kali (*duplicate emails*) sebanyak jumlah percobaan yang dijalankan sebelum transaksi berhasil.
> 2. Yang lebih fatal, pengiriman email adalah operasi eksternal jaringan yang **tidak bisa di-rollback** oleh database. Jika seluruh percobaan transaksi pada akhirnya gagal dan dibatalkan (ROLLBACK), email sudah terlanjur terkirim ke dosen yang mengabarkan adanya mahasiswa bimbingan baru, padahal di database data tersebut tidak pernah tersimpan.
> 
> Karena itu, pengiriman email harus ditempatkan **di luar fungsi transaksi**, yaitu di layer service **setelah** pemanggilan `s.tx.WithTx` selesai dieksekusi dan dipastikan mengembalikan nilai `nil` (transaksi telah berhasil di-COMMIT secara permanen ke database).

---

## LAMPIRAN — HASIL PENGUJIAN LENGKAP BAGIAN C

### 1. Uji Beban Rebutan Kelas IF399 (`mode crowd`)
- **Sebelum Perbaikan (Versi Naif):**
  ```text
  hasil request:
    201 OK                    56
    409 COURSE_FULL           4
  IF399 kuota=5  enrolled_count=5  baris enrollments=56  -> TIDAK KONSISTEN
  ```
- **Sesudah Perbaikan (Konsisten 3x berturut-turut):**
  ```text
  hasil request:
    201 OK                    5
    409 COURSE_FULL           55
  waktu total : 39.74ms
  IF399 kuota=5  enrolled_count=5  baris enrollments=5  -> KONSISTEN
  ```

### 2. Uji Beban Batas SKS (`mode credits`)
- **Hasil:**
  ```text
  10 mahasiswa mendaftar ke 12 kelas 3 SKS serentak...
  hasil request:
    201 OK                    80
    409 CREDIT_LIMIT_EXCEEDED 40
  waktu total : 78.36ms
  seluruh 10 mahasiswa <= 24 SKS -> KONSISTEN
  ```

### 3. Uji Beban Pindah Kelas Serentak Bebas Deadlock (`mode swap`)
- **Hasil:**
  ```text
  20 pasang mahasiswa pindah kelas berlawanan arah...
  hasil request:
    200 OK                    40
  waktu total : 131.97ms
  deadlock terdeteksi PostgreSQL selama uji: 0
  IF101 kuota=40 enrolled_count=20 baris enrollments=20 -> KONSISTEN
  IF102 kuota=40 enrolled_count=20 baris enrollments=20 -> KONSISTEN
  ```

### 4. Uji Fungsional C.4 (Satu per Satu)
```text
1. Daftar kelas kosong: Status=201
2. Daftar kelas ganda: Status=409 Code=ALREADY_ENROLLED
3. Drop kelas belum diambil: Status=409 Code=NOT_ENROLLED
4. Swap ID sama: Status=422 Code=VALIDATION_ERROR
5. Swap kelas sah: Status=200
6. Mahasiswa ubah kuota: Status=403
7. Admin ubah kuota valid: Status=200 (version naik dari 1 menjadi 2)
8. Admin version lama: Status=409 Code=VERSION_CONFLICT
9. Admin kuota di bawah peserta: Status=409 Code=QUOTA_BELOW_ENROLLED
```

### 5. Uji Ketahanan Server C.5 (175 Kegagalan Beruntun)
- `GET /health` respon: `200 OK` (0.0005s).
- Monitoring koneksi basis data: `SELECT pid, state FROM pg_stat_activity WHERE state <> 'idle'` hanya menampilkan 1 baris (kueri itu sendiri). Tidak ada satupun koneksi berstatus `idle in transaction`.
