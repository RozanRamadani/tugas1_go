-- Migration 009: Perubahan Idempotency Key Notifikasi untuk Mendukung Notifikasi Massal (Fan-Out)
-- psql -U postgres -d praktikum_backend -f migrations/009_quota_notifications.sql

-- 1. Hapus constraint lama yang hanya mengunci outbox_id
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_outbox_unique;

-- 2. Tambahkan composite unique constraint (outbox_id, user_id)
-- Ini memungkinkan satu outbox event (seperti perubahan kuota kelas) dikirimkan
-- ke banyak peserta kelas, dengan jaminan setiap peserta hanya menerima tepat 1 email
-- dan retry tidak akan mengirimkan email berulang ke peserta yang sudah sukses.
ALTER TABLE notifications ADD CONSTRAINT notifications_outbox_user_unique UNIQUE (outbox_id, user_id);

