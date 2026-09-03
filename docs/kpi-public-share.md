# Public Share Dashboard KPI

Dashboard KPI dapat dibagikan oleh Admin dan Super Admin melalui tombol **Bagikan**. Link publik hanya memuat visualisasi kustom (komponen ringkasan bawaan tetap berada di dashboard utama), bersifat read-only, tidak membutuhkan login, terikat ke satu plant, dan maksimal berlaku 90 hari.

## Alur pengujian lokal

1. Jalankan migrasi backend agar `046_create_kpi_public_shares.sql` diterapkan.
2. Jalankan backend dan frontend seperti biasa.
3. Login sebagai Admin/Super Admin, buka Dashboard KPI, dan tambahkan minimal satu visualisasi kustom.
4. Untuk Super Admin, pilih satu plant; mode `Semua Plant` sengaja tidak dapat dibagikan.
5. Klik **Bagikan**, isi nama/judul dan masa berlaku, lalu buat link.
6. Salin link yang ditampilkan dan buka melalui private/incognito window.
7. Uji pencabutan dan rotasi dari dialog yang sama. Token lama harus langsung menghasilkan halaman tidak tersedia.

## Perilaku drill-down

- Tombol drill menambahkan kategori berikutnya dan menampilkan seluruh value anak di bawah setiap kategori induk.
- Pengguna tidak perlu memilih dropdown, batang, titik, atau satu value induk.
- Public share mengirim `drill_level`; backend mengambil nama dimension dari snapshot widget sehingga browser tidak dapat meminta dimension di luar konfigurasi.
- Hasil dengan kategori banyak memakai zoom/scroll. Jika guardrail query memotong hasil, jumlah value yang ditampilkan dan jumlah total diinformasikan secara eksplisit.

## Kontrak keamanan

- Token mentah hanya dikembalikan saat create/rotate; database menyimpan SHA-256 hash.
- Browser publik tidak boleh mengirim plant, measure, atau dimension bebas.
- Query visualisasi kustom hanya menerima `widget_id` dan drill path yang divalidasi terhadap snapshot.
- Logger aplikasi meredaksi token dari path `/public/kpi/:token`.
- Semua endpoint publik memakai rate limiter, `no-store`, `noindex`, dan `Referrer-Policy: no-referrer`.
- Link revoked/expired selalu ditolak sebelum query visualisasi dijalankan.
