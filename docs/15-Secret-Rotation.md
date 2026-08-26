# Rotasi Secret dan Credential

Dokumen ini dijalankan dari host deployment melalui VPN/SSH. Jangan menaruh
nilai secret pada command history, tiket, chat, atau repository.

## Persiapan

1. Jadwalkan maintenance dan hentikan backend/frontend.
2. Backup PostgreSQL dalam format custom dan catat checksum SHA-256.
3. Backup file `.env` server ke password manager, bukan ke repository.
4. Generate nilai baru untuk `DB_PASSWORD`, `MINIO_ACCESS_KEY`,
   `MINIO_SECRET_KEY`, `JWT_SECRET`, `SETTING_ENCRYPTION_KEY`, dan seluruh
   `SEED_*_PASSWORD`. Gunakan minimal 32 karakter acak; key AES harus berupa
   Base64 dari tepat 32 byte.
5. Simpan key AES lama dan baru sementara dalam file berizin `0600`. Hapus file
   rotasi setelah deployment terverifikasi.

## Urutan Eksekusi

1. Build image baru, tetapi jangan mulai backend.
2. Jalankan migrasi schema melalui binary `./migrate`.
3. Jalankan `./rotate-encryption-key` dengan environment
   `OLD_SETTING_ENCRYPTION_KEY` dan `NEW_SETTING_ENCRYPTION_KEY`. Tanpa
   `--apply`, command hanya memvalidasi dan menghitung kandidat.
4. Jika dry-run bersih, jalankan ulang dengan `--apply`. Seluruh update AES
   berada dalam satu transaksi dan rollback otomatis jika ada ciphertext yang
   tidak dapat dibaca.
5. Ubah password role PostgreSQL menggunakan sesi administrator aktif.
   Mengganti `POSTGRES_PASSWORD` saja tidak mengubah password pada volume lama.
6. Perbarui `.env` server dengan seluruh credential baru dan permission `0600`.
7. Recreate PostgreSQL dan MinIO tanpa menghapus named volume.
8. Jalankan `./rotate-seeded-passwords` dengan seluruh `SEED_*_PASSWORD` baru.
9. Mulai backend/frontend. Rotasi JWT sengaja membatalkan semua sesi lama.

## Verifikasi

- Health PostgreSQL dan MinIO harus hijau.
- `/health` backend mengembalikan HTTP 200.
- Password seeded lama ditolak dan password baru diterima.
- JWT yang dibuat sebelum maintenance ditolak.
- Issue lama dapat dibuka dan keterangannya terbaca.
- Upload, download, dan delete bukti MinIO berhasil.
- Setting SMTP terenkripsi tetap dapat dipakai tanpa pernah ditampilkan API.
- `docker compose config`/`podman compose config` tidak menampilkan placeholder.

## Rollback

Jika verifikasi gagal, hentikan backend, pulihkan dump PostgreSQL dan file env
lama, kembalikan credential PostgreSQL/MinIO lama, lalu recreate layanan. Jangan
menjalankan backend dengan database lama dan key AES baru secara bersamaan.

## Pembersihan Git History

Lakukan hanya setelah rotasi production berhasil dan credential lama tidak lagi
berlaku. Buat Git bundle backup, jalankan `git filter-repo` terhadap seluruh
secret lama pada semua refs, scan ulang, lalu force-push branch dan tag.
Seluruh clone lama wajib dibuang dan dibuat ulang agar history lama tidak masuk
kembali.
