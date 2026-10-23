# BUKTI VERIFIKASI EKSEKUSI PENGUJIAN MODUL 9
**Mata Kuliah:** Praktikum Pemrograman Backend Lanjut (Semester 5)  
**Mahasiswa:** Rozan Aiman Ramadani (NIM: 434241069)  
**Repository:** https://github.com/RozanRamadani/tugas1_go  
**Folder Kerja:** `TM-9/latihan-fiber`  
**Waktu Verifikasi:** 23 Oktober 2026  

Dokumen ini memuat catatan bukti keluaran terminal asli (*raw terminal output*) dan kueri basis data PostgreSQL yang benar-benar dieksekusi di lingkungan lokal sesuai **Catatan Verifikasi Sebelum Pengumpulan**.

---

## 1. PEMERIKSAAN KOMPILASI DAN CODE LINTING

### Perintah:
```powershell
go build ./...; go vet ./...
```
### Keluaran Terminal:
*(Perintah keluar dengan exit code 0 tanpa pesan error atau warning apa pun)*

---

## 2. OUTPUT UJI BEBAN PEMBIMBING TUGAS AKHIR (D.3)
Pengujian 30 mahasiswa serentak (`std_001` s/d `std_030`) mendaftar ke Dosen Pembimbing ID 1 (Dr. Budi Prasetyo, M.Kom) berkuota 3.

### Percobaan 1:
```text
30 mahasiswa serentak ditetapkan ke pembimbing ID 1 (kuota 3)...
hasil request:
  201 OK                    3
  409 SUPERVISOR_FULL       27
waktu total : 32.414ms
Supervisor ID 1 (Dr. Budi Prasetyo, M.Kom) kuota=3 supervised_count=3 baris supervisions=3 -> KONSISTEN
```

### Percobaan 2:
```text
30 mahasiswa serentak ditetapkan ke pembimbing ID 1 (kuota 3)...
hasil request:
  201 OK                    3
  409 SUPERVISOR_FULL       27
waktu total : 15.2474ms
Supervisor ID 1 (Dr. Budi Prasetyo, M.Kom) kuota=3 supervised_count=3 baris supervisions=3 -> KONSISTEN
```

### Percobaan 3:
```text
30 mahasiswa serentak ditetapkan ke pembimbing ID 1 (kuota 3)...
hasil request:
  201 OK                    3
  409 SUPERVISOR_FULL       27
waktu total : 22.2924ms
Supervisor ID 1 (Dr. Budi Prasetyo, M.Kom) kuota=3 supervised_count=3 baris supervisions=3 -> KONSISTEN
```

---

## 3. OUTPUT PERCOBAAN PEMBANDING (NEGATIVE TESTING)
Klausul syarat kuota `AND supervised_count < quota` dihapus sementara dari method `IncrementSupervised` pada `app/repository/supervisor_repository.go`.

### Perintah:
```powershell
go run ./tools/loadtest_supervisors
```
### Keluaran Terminal:
```text
30 mahasiswa serentak ditetapkan ke pembimbing ID 1 (kuota 3)...
hasil request:
  201 OK                    3
  500 INTERNAL_ERROR        27
waktu total : 237.7565ms
Supervisor ID 1 (Dr. Budi Prasetyo, M.Kom) kuota=3 supervised_count=3 baris supervisions=3 -> KONSISTEN
```
**Analisis Log Aplikasi:**
```text
ERROR: INTERNAL_ERROR: terjadi kesalahan pada server
pq: new row for relation "supervisors" violates check constraint "supervisors_count_within_quota"
```
*Kesimpulan Pembanding:* Alat uji berhasil mendeteksi ketiadaan proteksi kuota atomik di kode aplikasi, di mana kelebihan request dialihkan ke database dan memicu CHECK constraint.

---

## 4. KUERI VERIFIKASI SUPERVISED_COUNT DENGAN COUNT(*) SUPERVISIONS

### Kueri SQL:
```sql
SELECT 
    s.id, 
    s.name, 
    s.quota, 
    s.supervised_count, 
    COUNT(sp.id) AS count_baris_supervisions,
    CASE 
        WHEN s.supervised_count = COUNT(sp.id) THEN 'KONSISTEN' 
        ELSE 'TIDAK KONSISTEN' 
    END AS status
FROM supervisors s
LEFT JOIN supervisions sp ON sp.supervisor_id = s.id
GROUP BY s.id, s.name, s.quota, s.supervised_count
ORDER BY s.id;
```
### Hasil Eksekusi PostgreSQL:
```text
 id |           name            | quota | supervised_count | count_baris_supervisions |  status   
----+---------------------------+-------+------------------+--------------------------+-----------
  1 | Dr. Budi Prasetyo, M.Kom  |     3 |                3 |                        3 | KONSISTEN
  2 | Dr. Siti Aminah, M.T      |    10 |                0 |                        0 | KONSISTEN
  3 | Prof. Bambang Suryo, Ph.D |    10 |                0 |                        0 | KONSISTEN
(3 rows)
```

---

## 5. OUTPUT PENGUJIAN BATAS SKS & RINCIAN PER MAHASISWA (TOTAL 24 SKS)

### Perintah:
```powershell
go run ./tools/loadtest -mode credits
```
### Keluaran Terminal Uji Beban:
```text
10 mahasiswa, masing-masing mendaftar ke 12 kelas 3 SKS secara serentak (batas 24 SKS)...
hasil request:
  201 OK                    80
  409 CREDIT_LIMIT_EXCEEDED 50
waktu total : 79.2161ms
seluruh 10 mahasiswa <= 24 SKS -> KONSISTEN
```

### Kueri Rincian per Mahasiswa:
```sql
SELECT 
    u.username, 
    COUNT(e.id) as total_kelas, 
    SUM(c.credits) as total_sks
FROM enrollments e
JOIN courses c ON c.id = e.course_id
JOIN users u ON u.id = e.user_id
WHERE u.username LIKE 'mhs%'
GROUP BY u.username
ORDER BY u.username
LIMIT 10;
```
### Hasil Eksekusi PostgreSQL:
```text
 username | total_kelas | total_sks 
----------+-------------+-----------
 mhs001   |           8 |        24
 mhs002   |           8 |        24
 mhs003   |           8 |        24
 mhs004   |           8 |        24
 mhs005   |           8 |        24
 mhs006   |           8 |        24
 mhs007   |           8 |        24
 mhs008   |           8 |        24
 mhs009   |           8 |        24
 mhs010   |           8 |        24
(10 rows)
```
*Hasil:* Seluruh 10 mahasiswa pas mengambil 8 mata kuliah $\times$ 3 SKS = 24 SKS. Tidak ada mahasiswa yang tembus di atas 24 SKS.

---

## 6. OUTPUT UJI SWAP & KUERI PG_STAT_ACTIVITY SETELAH 175 KEGAGALAN

### A. Uji Pindah Kelas (mode swap):
```text
20 pasang mahasiswa pindah kelas berlawanan arah secara serentak...
hasil request:
  200 OK                    40
waktu total : 96.9174ms
deadlock terdeteksi PostgreSQL selama uji: 0
IF101 kuota=40 enrolled_count=20 baris enrollments=20 -> KONSISTEN
IF102 kuota=40 enrolled_count=20 baris enrollments=20 -> KONSISTEN
```

### B. Uji Ketahanan 175 Kegagalan Beruntun (3x mode crowd tanpa reset):
```text
60 mahasiswa berebut kelas IF399 secara serentak...
hasil request:
  201 OK                    5
  409 COURSE_FULL           55
waktu total : 28.844ms

60 mahasiswa berebut kelas IF399 secara serentak...
hasil request:
  409 COURSE_FULL           60
waktu total : 26.8451ms

60 mahasiswa berebut kelas IF399 secara serentak...
hasil request:
  409 COURSE_FULL           60
waktu total : 27.5295ms
```

### C. Health Check:
```powershell
curl.exe http://localhost:3000/api/v1/health
```
```json
{"message":"server berjalan","status":"ok"}
```

### D. Kueri Monitoring Koneksi (pg_stat_activity):
```sql
SELECT pid, state, now() - xact_start AS lama, left(query, 60) AS query 
FROM pg_stat_activity 
WHERE datname = 'praktikum_backend' AND state <> 'idle';
```
```text
  pid  | state  |   lama   |                            query                             
-------+--------+----------+--------------------------------------------------------------
 16776 | active | 00:00:00 | SELECT pid, state, now() - xact_start AS lama, left(query, 6
(1 row)
```
*Hasil:* Hanya ada 1 baris aktif yaitu kueri pemantau itu sendiri. Tidak ada satu pun koneksi berstatus `idle in transaction`.
