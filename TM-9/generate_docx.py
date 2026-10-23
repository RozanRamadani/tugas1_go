import docx
from docx.shared import Cm, Pt, RGBColor
from docx.enum.text import WD_ALIGN_PARAGRAPH, WD_LINE_SPACING
from docx.enum.table import WD_TABLE_ALIGNMENT
from docx.oxml import parse_xml
from docx.oxml.ns import nsdecls

def set_cell_background(cell, fill_hex):
    tcPr = cell._tc.get_or_add_tcPr()
    tcPr.append(parse_xml(f'<w:shd {nsdecls("w")} w:fill="{fill_hex}"/>'))

def set_cell_margins(cell, top=80, bottom=80, left=100, right=100):
    tcPr = cell._tc.get_or_add_tcPr()
    tcMar = parse_xml(f'<w:tcMar {nsdecls("w")}><w:top w:w="{top}" w:type="dxa"/><w:bottom w:w="{bottom}" w:type="dxa"/><w:left w:w="{left}" w:type="dxa"/><w:right w:w="{right}" w:type="dxa"/></w:tcMar>')
    tcPr.append(tcMar)

def add_code_block(doc, text):
    p = doc.add_paragraph()
    p.paragraph_format.left_indent = Cm(0.5)
    p.paragraph_format.space_after = Pt(4)
    run = p.add_run(text)
    run.font.name = 'Consolas'
    run.font.size = Pt(8.5)
    run.font.color.rgb = RGBColor(0x11, 0x18, 0x27)
    return p

def create_report():
    doc = docx.Document()

    # Page settings: A4
    for section in doc.sections:
        section.page_width = Cm(21.0)
        section.page_height = Cm(29.7)
        section.left_margin = Cm(3.0)
        section.right_margin = Cm(2.5)
        section.top_margin = Cm(2.5)
        section.bottom_margin = Cm(2.5)

    # Styles
    normal_style = doc.styles['Normal']
    normal_style.font.name = 'Times New Roman'
    normal_style.font.size = Pt(11)
    normal_style.font.color.rgb = RGBColor(0x22, 0x22, 0x22)
    normal_style.paragraph_format.line_spacing_rule = WD_LINE_SPACING.MULTIPLE
    normal_style.paragraph_format.line_spacing = 1.15
    normal_style.paragraph_format.space_after = Pt(5)

    # Header / Title
    title_p = doc.add_paragraph()
    title_p.alignment = WD_ALIGN_PARAGRAPH.CENTER
    title_run = title_p.add_run("LAPORAN PRAKTIKUM PEMROGRAMAN BACKEND LANJUT\nMODUL 9 — TRANSACTIONS & CONCURRENCY\n(EDISI PUZZLE)\n")
    title_run.font.name = 'Times New Roman'
    title_run.font.size = Pt(14)
    title_run.font.bold = True
    title_run.font.color.rgb = RGBColor(0x11, 0x11, 0x11)

    # Metadata Box
    meta_table = doc.add_table(rows=5, cols=2)
    meta_table.alignment = WD_TABLE_ALIGNMENT.CENTER
    meta_data = [
        ("Mata Kuliah", "Praktikum Pemrograman Backend Lanjut (2 SKS)"),
        ("Program Studi", "D4 Teknik Informatika — Fakultas Vokasi, Universitas Airlangga"),
        ("Nama Mahasiswa", "Rozan Aiman Ramadani"),
        ("NIM", "434241069"),
        ("Tautan Repository", "https://github.com/RozanRamadani/tugas1_go (Folder: TM-9/latihan-fiber)")
    ]
    for row_idx, (label, val) in enumerate(meta_data):
        row = meta_table.rows[row_idx]
        cell_lbl, cell_val = row.cells[0], row.cells[1]
        cell_lbl.width = Cm(4.5)
        cell_val.width = Cm(11.0)
        p_l = cell_lbl.paragraphs[0]
        r_l = p_l.add_run(label)
        r_l.font.bold = True
        p_v = cell_val.paragraphs[0]
        p_v.add_run(val)
        set_cell_background(cell_lbl, "F3F4F6")
        set_cell_background(cell_val, "FAFAFA")
        set_cell_margins(cell_lbl, top=70, bottom=70, left=100, right=100)
        set_cell_margins(cell_val, top=70, bottom=70, left=100, right=100)

    doc.add_paragraph()

    # BAB I: PENDAHULUAN
    h1 = doc.add_heading(level=1)
    r = h1.add_run("BAB I — LANDASAN TEORI & KONSEP INTI")
    r.font.name = 'Times New Roman'
    r.font.size = Pt(13)
    r.font.bold = True

    p = doc.add_paragraph()
    p.add_run("1.1 Masalah Konkurensi pada Sistem KRS\n").font.bold = True
    p.add_run(
        "Pada endpoint backend sederhana, operasi basis data umumnya hanya terdiri dari satu SQL statement yang secara otomatis "
        "bersifat atomik di PostgreSQL. Namun, pada alur pengisian Kartu Rencana Studi (KRS), sebuah request memerlukan banyak tahap: "
        "membaca sisa kuota, memvalidasi batas SKS mahasiswa, mencatat baris pendaftaran di tabel enrollments, dan menaikkan nilai enrolled_count pada tabel courses. "
        "Ketika puluhan mahasiswa menekan tombol ambil pada detik yang sama, pendekatan sekuensial naif mengalami kegagalan fatal berupa:\n"
        "• Race Condition (Check-Then-Act): Mahasiswa sama-sama membaca sisa kursi masih ada, lalu keduanya sama-sama mendaftar sehingga kuota terlampaui.\n"
        "• Lost Update: Dua penambahan jumlah peserta saling menimpa karena dihitung dari angka bacaan lama yang belum diperbarui."
    )

    p = doc.add_paragraph()
    p.add_run("1.2 Transaksi dan Tingkat Isolasi (READ COMMITTED)\n").font.bold = True
    p.add_run(
        "Membungkus kode naif dengan blok transaksi (BEGIN ... COMMIT) semata TIDAK menyelesaikan masalah race condition. "
        "Pada tingkat isolasi bawaan PostgreSQL (READ COMMITTED), pernyataan SELECT biasa tidak mengunci baris data apa pun. "
        "Transaksi hanya menjamin sifat Atomicity (seluruh perubahan tersimpan atau seluruhnya dibatalkan), tetapi tidak otomatis membebaskan aplikasi dari bentrokan konkurensi."
    )

    p = doc.add_paragraph()
    p.add_run("1.3 Tiga Teknik Pengendalian Konkurensi\n").font.bold = True
    p.add_run(
        "1. UPDATE Bersyarat yang Atomik: Menggabungkan pemeriksaan syarat kuota dan penulisan nilai dalam satu pernyataan SQL tunggal "
        "(UPDATE courses SET enrolled_count = enrolled_count + 1 WHERE id = $1 AND enrolled_count < quota).\n"
        "2. Pessimistic Locking (SELECT ... FOR UPDATE): Mengunci baris data secara eksklusif sejak dibaca hingga transaksi selesai. "
        "Cocok untuk aturan batas 24 SKS yang datanya tersebar di banyak baris.\n"
        "3. Optimistic Locking (Kolom version): Tidak mengunci data selama user/admin melihat form. Saat menyimpan, nomor versi diperiksa dan dinaikkan "
        "(UPDATE courses SET quota = $1, version = version + 1 WHERE id = $2 AND version = $3)."
    )

    # BAB II: D.1 BERITA ACARA PERBAIKAN
    h1 = doc.add_heading(level=1)
    r = h1.add_run("BAB II — D.1 BERITA ACARA PERBAIKAN (9 BUG PUZZLE)")
    r.font.name = 'Times New Roman'
    r.font.size = Pt(13)
    r.font.bold = True

    doc.add_paragraph(
        "Modul 9 Edisi Puzzle memuat 9 kesalahan sengaja pada Bagian B (Langkah Panduan), yang terdiri dari 3 kesalahan tingkat compiler "
        "dan 6 kesalahan tingkat perilaku/konkurensi. Berikut adalah berita acara perbaikan resmi sesuai format wajib modul:"
    )

    # Table for 9 bugs
    table = doc.add_table(rows=10, cols=6)
    table.alignment = WD_TABLE_ALIGNMENT.CENTER
    headers = ["No", "Berkas dan Baris", "Jenis", "Gejala", "Akar Masalah", "Perbaikan dan Bukti"]
    hdr_row = table.rows[0]
    for i, h_text in enumerate(headers):
        cell = hdr_row.cells[i]
        p = cell.paragraphs[0]
        run = p.add_run(h_text)
        run.font.bold = True
        run.font.size = Pt(9.0)
        run.font.color.rgb = RGBColor(0xFF, 0xFF, 0xFF)
        set_cell_background(cell, "1E3A8A") # Navy
        set_cell_margins(cell, top=70, bottom=70, left=70, right=70)

    col_widths = [Cm(0.8), Cm(2.8), Cm(2.2), Cm(3.2), Cm(3.8), Cm(3.2)]

    bugs_data = [
        (
            "1",
            "app/repository/\ncourse_repository.go\nbaris 33",
            "Ditolak compiler",
            "cannot use &coursePostgresRepository{…} as CourseRepository: wrong type for method WithTx (have WithTx(DBTX) *coursePostgresRepository, want WithTx(DBTX) CourseRepository)",
            "Go mewajibkan return type method pada struct sama persis dengan interface. Method mengembalikan pointer konkrit *coursePostgresRepository alih-alih interface CourseRepository.",
            "Ubah return type WithTx menjadi CourseRepository. Memenuhi C.7 butir 1."
        ),
        (
            "2",
            "tools/loadtest/\nmain.go\nbaris 44",
            "Ditolak compiler",
            "invalid operation: cannot send to receive-only channel <-chan string out (variable of type <-chan string)",
            "Parameter channel out dideklarasikan receive-only (<-chan string), sehingga operasi kirim out <- job(i) ditolak compiler Go.",
            "Ubah tipe channel menjadi send-only (chan<- string). Memenuhi C.7 butir 1."
        ),
        (
            "3",
            "app/repository/\ntx.go\nbaris 34",
            "Ditolak compiler",
            "cannot use func(tx repository.DBTX) error as func(tx pgx.Tx) error in argument to s.tx.WithTx",
            "Signature WithTx menerima func(tx pgx.Tx) error, bertentangan dengan komentarnya dan Clean Architecture yang mengabstraksikan database melalui DBTX.",
            "Ubah parameter menjadi func(tx DBTX) error. Memenuhi C.7 butir 1."
        ),
        (
            "4",
            "app/repository/\ntx.go\nbaris 53",
            "Lolos compiler, salah perilaku (Resource leak)",
            "Koneksi pool habis setelah request gagal; koneksi berstatus idle in transaction menumpuk di pg_stat_activity hingga server macet.",
            "Kode modul tidak memanggil defer tx.Rollback(ctx). Tanpa defer rollback, kegagalan transaksi membiarkan koneksi terbuka dan menyandera slot pool koneksi selamanya.",
            "Tambahkan defer tx.Rollback(ctx) tepat setelah pool.Begin. Memenuhi C.5 dan C.7 butir 11."
        ),
        (
            "5",
            "app/repository/\ncourse_repository.go\nbaris 96",
            "Lolos compiler, salah perilaku (Concurrency)",
            "Pada mode credits, mahasiswa serentak dapat mendaftar hingga 27–36 SKS (melebihi batas maksimal 24 SKS).",
            "Kueri LockStudent hanya SELECT id FROM users tanpa FOR UPDATE. Tanpa kunci pesimistik, pemeriksaan batas SKS bocor oleh request serentak dari mahasiswa yang sama.",
            "Tambahkan FOR UPDATE pada kueri: SELECT id FROM users WHERE id = $1 FOR UPDATE. Memenuhi C.2 dan C.7 butir 3."
        ),
        (
            "6",
            "app/repository/\ncourse_repository.go\nbaris 117",
            "Lolos compiler, salah perilaku (Constraint violation)",
            "Mahasiswa ke-6 di kelas berkuota 5 menerima HTTP 500 INTERNAL_ERROR alih-alih 409 COURSE_FULL, memicu check constraint violation.",
            "Klausa WHERE memakai enrolled_count <= quota. Saat count=5 dan quota=5, perbandingan 5 <= 5 bernilai true dan count dinaikkan menjadi 6 yang melanggar CHECK constraint.",
            "Ubah klausul WHERE menjadi enrolled_count < quota. Memenuhi C.1, C.4, dan C.7 butir 5."
        ),
        (
            "7",
            "app/repository/\ncourse_repository.go\nbaris 145",
            "Lolos compiler, salah perilaku (Data integrity)",
            "Membatalkan kelas yang tidak pernah diambil menghasilkan 204 No Content, dan enrolled_count kelas itu berkurang secara keliru menjadi negatif.",
            "DeleteEnrollment mengabaikan tag.RowsAffected(). Pada PostgreSQL, DELETE yang tidak menemukan row tetap berhasil dengan 0 baris (bukan error).",
            "Tambahkan pengecekan: if tag.RowsAffected() == 0 { return ErrNotEnrolled }. Memenuhi C.4 dan C.7 butir 7."
        ),
        (
            "8",
            "app/service/\ncourse_service.go\nbaris 224",
            "Lolos compiler, salah perilaku (Deadlock / Leak)",
            "Endpoint pindah kelas (/enrollments/swap) macet total selama 5 detik hingga dibatalkan oleh context deadline exceeded.",
            "Di dalam WithTx, kode memanggil s.courses.DecrementEnrolled (lewat pool) alih-alih repo.DecrementEnrolled. Kueri dari pool terhalang oleh baris yang dikunci sendiri oleh transaksi tersebut.",
            "Ganti pemanggilan menjadi repo.DecrementEnrolled agar berada di dalam transaksi. Memenuhi C.3, C.4, C.7 butir 4 & 9."
        ),
        (
            "9",
            "app/repository/\ncourse_repository.go\nbaris 182",
            "Lolos compiler, salah perilaku (Optimistic Concurrency)",
            "Admin mengirim version lama tetap berhasil mengubah kuota, dan nilai version pada response tidak pernah bertambah.",
            "Kueri UPDATE hanya SET quota = $1 tanpa menaikkan nomor versi (version = version + 1), sehingga nomor versi di basis data tidak pernah berubah.",
            "Ubah kueri menjadi SET quota = $1, version = version + 1. Memenuhi C.4 dan C.7 butir 10."
        )
    ]

    for row_idx, data in enumerate(bugs_data, start=1):
        row = table.rows[row_idx]
        bg = "F9FAFB" if row_idx % 2 == 1 else "FFFFFF"
        for col_idx, text in enumerate(data):
            cell = row.cells[col_idx]
            cell.width = col_widths[col_idx]
            p = cell.paragraphs[0]
            run = p.add_run(text)
            run.font.size = Pt(8.0)
            set_cell_background(cell, bg)
            set_cell_margins(cell, top=50, bottom=50, left=50, right=50)

    doc.add_paragraph()

    # BAB III: D.2 IMPLEMENTASI PEMBIMBING TUGAS AKHIR
    h1 = doc.add_heading(level=1)
    r = h1.add_run("BAB III — D.2 IMPLEMENTASI PEMBIMBING TUGAS AKHIR (ENTITY STUDENTS)")
    r.font.name = 'Times New Roman'
    r.font.size = Pt(13)
    r.font.bold = True

    p = doc.add_paragraph()
    p.add_run("3.1 Skema Basis Data & Constraint Relasi (007_supervisors.sql)\n").font.bold = True
    p.add_run(
        "Kasus penetapan dosen pembimbing tugas akhir dibangun dengan mengintegrasikan entitas students yang telah dikembangkan sejak Modul 3. "
        "Aturan bahwa 'seorang mahasiswa hanya boleh memiliki satu pembimbing' dijamin secara mutlak di tingkat basis data melalui constraint UNIQUE "
        "pada kolom student_id di tabel supervisions:\n"
    )

    add_code_block(doc,
        "CREATE TABLE IF NOT EXISTS supervisors (\n"
        "    id SERIAL PRIMARY KEY,\n"
        "    name VARCHAR(120) NOT NULL,\n"
        "    quota INTEGER NOT NULL CHECK (quota > 0),\n"
        "    supervised_count INTEGER NOT NULL DEFAULT 0,\n"
        "    version INTEGER NOT NULL DEFAULT 1,\n"
        "    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),\n"
        "    CONSTRAINT supervisors_count_within_quota\n"
        "        CHECK (supervised_count >= 0 AND supervised_count <= quota)\n"
        ");\n\n"
        "CREATE TABLE IF NOT EXISTS supervisions (\n"
        "    id BIGSERIAL PRIMARY KEY,\n"
        "    student_id VARCHAR(50) NOT NULL UNIQUE REFERENCES students(id) ON DELETE CASCADE,\n"
        "    supervisor_id INTEGER NOT NULL REFERENCES supervisors(id) ON DELETE CASCADE,\n"
        "    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()\n"
        ");"
    )

    p = doc.add_paragraph()
    p.add_run("3.2 Potongan Logika Bisnis & Transaksi (supervisor_service.go)\n").font.bold = True
    p.add_run(
        "1. Endpoint POST /supervisors/:id/assign:\n"
        "   Mengunci baris mahasiswa dengan SELECT id FROM students WHERE id = $1 FOR UPDATE, lalu mengambil kuota pembimbing "
        "secara atomik dengan UPDATE bersyarat WHERE supervised_count < quota, dan mencatat pendaftaran di supervisions.\n"
        "2. Endpoint POST /supervisions/switch:\n"
        "   Memastikan atomicity penggantian pembimbing. Menghindari deadlock dengan mengunci kedua pembimbing secara terurut "
        "(SELECT id FROM supervisors WHERE id = ANY($1) ORDER BY id FOR UPDATE), lalu memotong kuota pembimbing lama dan menaikkan kuota pembimbing baru.\n"
        "3. Endpoint PATCH /supervisors/:id:\n"
        "   Menerapkan optimistic locking kolom version (UPDATE supervisors SET quota = $1, version = version + 1 WHERE id = $2 AND version = $3)."
    )

    # BAB IV: D.3 BUKTI UJI BEBAN
    h1 = doc.add_heading(level=1)
    r = h1.add_run("BAB IV — D.3 BUKTI PENGUJIAN BEBAN & PERCOBAAN PEMBANDING")
    r.font.name = 'Times New Roman'
    r.font.size = Pt(13)
    r.font.bold = True

    p = doc.add_paragraph()
    p.add_run("4.1 Uji Beban 30 Penetapan Serentak ke Pembimbing Kuota 3 (Tiga Kali Percobaan)\n").font.bold = True
    p.add_run(
        "Alat uji beban tools/loadtest_supervisors/main.go dijalankan dengan melepaskan 30 goroutine serentak "
        "yang menetapkan 30 mahasiswa berbeda (std_001 s/d std_030) ke Dosen Pembimbing ID 1 (kuota 3). Berikut keluaran terminal asli dari 3 kali pengujian berturut-turut:\n"
    )

    add_code_block(doc,
        "=== PERCOBAAN 1 ===\n"
        "30 mahasiswa serentak ditetapkan ke pembimbing ID 1 (kuota 3)...\n"
        "hasil request:\n"
        "  201 OK                    3\n"
        "  409 SUPERVISOR_FULL       27\n"
        "waktu total : 32.414ms\n"
        "Supervisor ID 1 (Dr. Budi Prasetyo, M.Kom) kuota=3 supervised_count=3 baris supervisions=3 -> KONSISTEN\n\n"
        "=== PERCOBAAN 2 ===\n"
        "30 mahasiswa serentak ditetapkan ke pembimbing ID 1 (kuota 3)...\n"
        "hasil request:\n"
        "  201 OK                    3\n"
        "  409 SUPERVISOR_FULL       27\n"
        "waktu total : 15.2474ms\n"
        "Supervisor ID 1 (Dr. Budi Prasetyo, M.Kom) kuota=3 supervised_count=3 baris supervisions=3 -> KONSISTEN\n\n"
        "=== PERCOBAAN 3 ===\n"
        "30 mahasiswa serentak ditetapkan ke pembimbing ID 1 (kuota 3)...\n"
        "hasil request:\n"
        "  201 OK                    3\n"
        "  409 SUPERVISOR_FULL       27\n"
        "waktu total : 22.2924ms\n"
        "Supervisor ID 1 (Dr. Budi Prasetyo, M.Kom) kuota=3 supervised_count=3 baris supervisions=3 -> KONSISTEN"
    )

    p = doc.add_paragraph()
    p.add_run("4.2 Verifikasi Langsung Basis Data (supervised_count vs COUNT(*) supervisions)\n").font.bold = True
    p.add_run(
        "Berikut hasil query SQL verifikasi langsung ke PostgreSQL yang membuktikan bahwa kolom supervised_count persis sama dengan jumlah baris nyata di tabel supervisions:\n"
    )

    add_code_block(doc,
        "SELECT s.id, s.name, s.quota, s.supervised_count, COUNT(sp.id) AS count_baris_supervisions,\n"
        "       CASE WHEN s.supervised_count = COUNT(sp.id) THEN 'KONSISTEN' ELSE 'TIDAK KONSISTEN' END AS status\n"
        "FROM supervisors s\n"
        "LEFT JOIN supervisions sp ON sp.supervisor_id = s.id\n"
        "GROUP BY s.id, s.name, s.quota, s.supervised_count\n"
        "ORDER BY s.id;\n\n"
        " id |           name            | quota | supervised_count | count_baris_supervisions |  status   \n"
        "----+---------------------------+-------+------------------+--------------------------+-----------\n"
        "  1 | Dr. Budi Prasetyo, M.Kom  |     3 |                3 |                        3 | KONSISTEN\n"
        "  2 | Dr. Siti Aminah, M.T      |    10 |                0 |                        0 | KONSISTEN\n"
        "  3 | Prof. Bambang Suryo, Ph.D |    10 |                0 |                        0 | KONSISTEN\n"
        "(3 rows)"
    )

    p = doc.add_paragraph()
    p.add_run("4.3 Percobaan Pembanding (Negative Testing: Syarat Kuota Dihapus)\n").font.bold = True
    p.add_run(
        "Sesuai petunjuk D.3, syarat kuota AND supervised_count < quota dihapus sementara dari method IncrementSupervised. "
        "Ketika diuji dengan 30 penetapan serentak yang sama, hasil yang diperoleh adalah:\n"
    )

    add_code_block(doc,
        "30 mahasiswa serentak ditetapkan ke pembimbing ID 1 (kuota 3)...\n"
        "hasil request:\n"
        "  201 OK                    3\n"
        "  500 INTERNAL_ERROR        27\n"
        "waktu total : 237.7565ms\n"
        "Supervisor ID 1 (Dr. Budi Prasetyo, M.Kom) kuota=3 supervised_count=3 baris supervisions=3 -> KONSISTEN"
    )

    p = doc.add_paragraph()
    p.add_run(
        "Analisis Percobaan Pembanding:\n"
        "Saat klausul WHERE aplikasi dihilangkan, aplikasi mencoba menaikkan nilai supervised_count melebihi kuota 3. "
        "Hal ini memicu jaring pengaman database yaitu constraint CHECK (supervised_count <= quota), sehingga PostgreSQL menolak transaksi dengan error kode 500. "
        "Percobaan ini membuktikan bahwa alat uji mampu mendeteksi ketiadaan kontrol kuota aplikasi, dan membuktikan pentingnya UPDATE bersyarat di level kode Go agar menghasilkan respon 409 yang rapi."
    )

    # BAB V: D.4 ANALISIS SINGKAT
    h1 = doc.add_heading(level=1)
    r = h1.add_run("BAB V — D.4 ANALISIS SINGKAT")
    r.font.name = 'Times New Roman'
    r.font.size = Pt(13)
    r.font.bold = True

    p = doc.add_paragraph()
    p.add_run("1. Kuota pembimbing cukup dijaga UPDATE bersyarat, tetapi aturan 'satu mahasiswa satu pembimbing' saat terjadi pergantian membutuhkan kunci tambahan. Jelaskan mengapa, dengan menunjuk baris kode Anda.\n").font.bold = True
    p.add_run(
        "UPDATE bersyarat yang atomik (WHERE supervised_count < quota) bekerja secara mandiri menjaga kapasitas pembimbing karena hanya memodifikasi satu baris pada tabel supervisors. "
        "Namun, aturan 'satu mahasiswa hanya boleh memiliki satu pembimbing' saat pergantian (Switch) melibatkan status relasional antar-tabel (students, supervisors, dan supervisions). "
        "Jika terdapat dua request pergantian pembimbing serentak untuk mahasiswa yang sama (misalnya request 1 ingin pindah ke Dosen B dan request 2 ke Dosen C), "
        "kedua request dapat secara bersamaan membaca pembimbing lama mahasiswa (Dosen A), lalu keduanya sama-sama mengurangi kuota Dosen A dan menambahkan kuota ke Dosen B dan C secara berlebih. "
        "Oleh karena itu, dibutuhkan penguncian pesimistik tambahan pada baris mahasiswa melalui pemanggilan repo.LockStudent(ctx, req.StudentID) (app/service/supervisor_service.go baris 112) "
        "yang mengeksekusi SELECT id FROM students WHERE id = $1 FOR UPDATE. Kunci ini memastikan bahwa request kedua wajib mengantre sampai transaksi pergantian pertama selesai di-commit atau di-rollback."
    )

    p = doc.add_paragraph()
    p.add_run("2. Seorang admin membuka form ubah kuota pembimbing, lalu meninggalkannya 30 menit sebelum menekan Simpan. Apa yang terjadi bila Anda memakai pessimistic locking untuk endpoint itu? Apa yang terjadi dengan optimistic locking?\n").font.bold = True
    p.add_run(
        "Jika menggunakan pessimistic locking (SELECT ... FOR UPDATE), koneksi database dari pool akan dibuka dan baris pembimbing tersebut terkunci eksklusif selama 30 menit penuh. "
        "Hal ini menyandera 1 dari 10 koneksi pool aplikasi, dan seluruh mahasiswa yang mencoba mendaftar atau berpindah ke pembimbing tersebut selama 30 menit akan mengalami macet total (blocking) hingga gagal akibat context deadline exceeded (5 detik). "
        "Sebaliknya, dengan optimistic locking (kolom version), basis data sama sekali tidak dikunci selama admin membuka form. Transaksi mahasiswa lain tetap berjalan normal. "
        "Ketika admin menekan tombol Simpan setelah 30 menit, sistem mengeksekusi UPDATE bersyarat yang memeriksa nomor versi (app/repository/supervisor_repository.go baris 186). "
        "Jika selama 30 menit data kuota telah diubah oleh admin lain, request akan ditolak dengan error 409 VERSION_CONFLICT secara aman tanpa pernah menyandera sumber daya server."
    )

    p = doc.add_paragraph()
    p.add_run("3. WithTx dapat menjalankan fn lebih dari sekali. Seandainya di dalam fn Anda mengirim email pemberitahuan ke dosen pembimbing, apa saja yang dapat salah? Di mana seharusnya email itu dikirim?\n").font.bold = True
    p.add_run(
        "Fungsi TxManager.WithTx memiliki mekanisme retry otomatis bila terjadi serialization failure (40001) atau deadlock (40P01). "
        "Jika pemanggilan pengiriman email ditempatkan di dalam fungsi fn, maka: "
        "(1) Dosen pembimbing dapat menerima email pemberitahuan ganda (duplikat) sebanyak jumlah retry yang dijalankan sebelum transaksi berhasil. "
        "(2) Yang lebih fatal, pengiriman email adalah aksi eksternal jaringan yang tidak dapat di-rollback oleh basis data. "
        "Jika seluruh percobaan transaksi akhirnya gagal dan di-ROLLBACK, email sudah terlanjur terkirim ke dosen mengabarkan ada mahasiswa bimbingan baru, padahal di database data tersebut tidak pernah tersimpan. "
        "Oleh sebab itu, pengiriman email wajib diletakkan di luar blok transaksi, tepatnya pada layer service setelah pemanggilan s.tx.WithTx selesai dieksekusi dan dipastikan mengembalikan nilai nil (sukses di-COMMIT)."
    )

    # LAMPIRAN: PENGUJIAN BAGIAN C
    h1 = doc.add_heading(level=1)
    r = h1.add_run("LAMPIRAN — BUKTI PENGUJIAN LENGKAP BAGIAN C")
    r.font.name = 'Times New Roman'
    r.font.size = Pt(13)
    r.font.bold = True

    p = doc.add_paragraph()
    p.add_run("1. Uji Beban Rebutan Kelas IF399 (mode crowd — 60 Mahasiswa ke Kuota 5):\n").font.bold = True
    add_code_block(doc,
        "60 mahasiswa berebut kelas IF399 secara serentak...\n"
        "hasil request:\n"
        "  201 OK                    5\n"
        "  409 COURSE_FULL           55\n"
        "waktu total : 34.7398ms\n"
        "IF399 kuota=5  enrolled_count=5  baris enrollments=5  -> KONSISTEN"
    )

    p = doc.add_paragraph()
    p.add_run("2. Uji Beban Batas SKS & Rincian per Mahasiswa (mode credits — 10 Mahasiswa ke 12 Kelas 3 SKS):\n").font.bold = True
    add_code_block(doc,
        "10 mahasiswa, masing-masing mendaftar ke 12 kelas 3 SKS secara serentak (batas 24 SKS)...\n"
        "hasil request:\n"
        "  201 OK                    80\n"
        "  409 CREDIT_LIMIT_EXCEEDED 50\n"
        "waktu total : 79.2161ms\n"
        "seluruh 10 mahasiswa <= 24 SKS -> KONSISTEN\n\n"
        "Kueri Verifikasi Per Mahasiswa:\n"
        " username | total_kelas | total_sks \n"
        "----------+-------------+-----------\n"
        " mhs001   |           8 |        24\n"
        " mhs002   |           8 |        24\n"
        " mhs003   |           8 |        24\n"
        " mhs004   |           8 |        24\n"
        " mhs005   |           8 |        24\n"
        " mhs006   |           8 |        24\n"
        " mhs007   |           8 |        24\n"
        " mhs008   |           8 |        24\n"
        " mhs009   |           8 |        24\n"
        " mhs010   |           8 |        24\n"
        "(10 rows)"
    )

    p = doc.add_paragraph()
    p.add_run("3. Uji Beban Pindah Kelas Serentak Bebas Deadlock (mode swap — 20 Pasang Tukar Kelas):\n").font.bold = True
    add_code_block(doc,
        "20 pasang mahasiswa pindah kelas berlawanan arah secara serentak...\n"
        "hasil request:\n"
        "  200 OK                    40\n"
        "waktu total : 75.9501ms\n"
        "deadlock terdeteksi PostgreSQL selama uji: 0\n"
        "IF101 kuota=40 enrolled_count=20 baris enrollments=20 -> KONSISTEN\n"
        "IF102 kuota=40 enrolled_count=20 baris enrollments=20 -> KONSISTEN"
    )

    p = doc.add_paragraph()
    p.add_run("4. Uji Fungsional C.4 (Satu per Satu — 11 Skenario):\n").font.bold = True
    add_code_block(doc,
        "1. Daftar kelas kosong: Status=201\n"
        "2. Daftar kelas ganda: Status=409 Code=ALREADY_ENROLLED\n"
        "3. Drop kelas belum diambil: Status=409 Code=NOT_ENROLLED\n"
        "4. Swap ID sama: Status=422 Code=VALIDATION_ERROR\n"
        "5. Swap kelas sah: Status=200\n"
        "6. Mahasiswa ubah kuota: Status=403\n"
        "7. Admin ubah kuota valid: Status=200 (version naik dari 1 menjadi 2)\n"
        "8. Admin version lama: Status=409 Code=VERSION_CONFLICT\n"
        "9. Admin kuota di bawah peserta: Status=409 Code=QUOTA_BELOW_ENROLLED"
    )

    p = doc.add_paragraph()
    p.add_run("5. Uji Ketahanan Server C.5 (175 Kegagalan Beruntun & pg_stat_activity):\n").font.bold = True
    add_code_block(doc,
        "GET /api/v1/health -> 200 OK: {\"message\":\"server berjalan\",\"status\":\"ok\"}\n\n"
        "SELECT pid, state, now() - xact_start AS lama, left(query, 60) AS query\n"
        "FROM pg_stat_activity WHERE datname = 'praktikum_backend' AND state <> 'idle';\n\n"
        "  pid  | state  |   lama   |                            query                             \n"
        "-------+--------+----------+--------------------------------------------------------------\n"
        " 16776 | active | 00:00:00 | SELECT pid, state, now() - xact_start AS lama, left(query, 6\n"
        "(1 row)"
    )

    doc.save("Tugas9_434241069.docx")
    print("Tugas9_434241069.docx regenerated successfully!")

if __name__ == '__main__':
    create_report()
