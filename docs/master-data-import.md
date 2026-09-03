# Import master data

Import tersedia pada tab Aspek Audit, Detail Aspek, Uraian, dan Klasifikasi
HEI untuk pengguna dengan permission `PERM-MSTR-I`.

## Urutan import

Data audit mempunyai urutan parent-child:

1. Import Aspek.
2. Unduh mapping `AspekID` pada layar sukses.
3. Gunakan tombol **Lanjut Import Detail Aspek**. Tombol ini mengunduh template
   baru dengan sheet `REFERENSI` yang sudah memuat Aspek hasil commit.
4. Import Detail Aspek dan unduh mapping `DetailID`.
5. Gunakan tombol **Lanjut Import Uraian** agar template Uraian memakai referensi
   Detail terbaru.

File Excel yang sudah berada di komputer tidak dapat memperbarui sheet
`REFERENSI` sendiri. Waktu snapshot referensi disimpan pada metadata template dan
ditampilkan pada preview. Unduh ulang template setelah import parent berhasil.

HEI tidak bergantung pada tiga master audit tersebut dan dapat diimport secara
terpisah. Kolom `category_name` boleh berisi kategori yang sudah ada atau nama
kategori baru. Kategori pada sheet `REFERENSI` hanya contoh, bukan whitelist.
Kategori baru terbentuk secara implisit saat baris HEI pertama dengan kategori
tersebut berhasil di-commit.

## Aturan ketat

- Hanya template `.xlsx` resmi, maksimal 5 MB dan 5.000 baris.
- Formula, macro, external link, embedded object, merged cell, dan sheet tambahan
  ditolak.
- Satu baris invalid membatalkan seluruh batch; tidak ada partial import atau
  upsert.
- Duplikat diperiksa di dalam file dan database menggunakan fungsi normalisasi
  PostgreSQL yang sama dengan unique index.
- Duplikat HEI diperiksa dalam kategori yang sama berdasarkan nama HEI dan, jika
  diisi, kode HEI. Nama kategori baru tetap melalui normalisasi spasi, Unicode,
  dan kapitalisasi.
- Token validasi berlaku 20 menit, hanya untuk satu file/user/plant/type, dan
  hanya dapat dipakai sekali.
- Aspek, Detail, dan Uraian dikunci per plant saat commit. Parent dan plant scope
  diperiksa ulang setelah lock diperoleh.

## Deployment database lama

Sebelum migration `045_master_import_and_unique_constraints.sql`, jalankan query
read-only di `backend/migrations/checks/045_master_import_duplicate_audit.sql`.
Migration sengaja berhenti jika menemukan duplikat lama dan tidak pernah memilih
record yang harus dihapus secara otomatis.
