import { SectionShell } from "../SectionShell";
import { RoleBadge } from "../RoleBadge";
import { Callout } from "../Callout";
import { GuideImage } from "../GuideImage";

export function AdminSection() {
  return (
    <SectionShell
      id="admin"
      index={9}
      title="Administrasi"
      lede="Pengaturan data induk, pengguna, hak akses, dan konfigurasi sistem."
      role={<RoleBadge role="admin" />}
    >
      <div className="space-y-8">
        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Data Induk (Master)</h3>
          <p className="text-sm text-muted-foreground">
            Sembilan tab data referensi yang dipakai di seluruh aplikasi: Plant, Department, Area, Kawasan,
            Detail Kawasan, Aspek Audit, Detail Audit, Uraian, dan Klasifikasi HEI — masing-masing tabel CRUD
            sederhana (tambah/ubah/hapus).
          </p>
          <Callout variant="warn">
            Tab <strong>Plant</strong> sepenuhnya tersembunyi untuk Admin biasa — hanya Super Admin yang bisa
            membuat/mengubah/menghapus Plant.
          </Callout>
          <GuideImage src="/guide/master-data.png" alt="Halaman Data Induk" caption="Data Induk (Master)." />
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Manajemen Pengguna</h3>
          <ul className="list-disc space-y-2 pl-5 text-sm text-muted-foreground marker:text-primary">
            <li>
              Tab <strong className="font-semibold text-foreground">Profil</strong> — buat/ubah/hapus akun
              pengguna, atur ulang password.
            </li>
            <li>
              Tab <strong className="font-semibold text-foreground">Permission</strong> — mengecualikan atau
              menambah hak akses untuk satu pengguna tertentu, di luar bawaan perannya.
            </li>
            <li>
              Tab <strong className="font-semibold text-foreground">Tata Letak Dashboard</strong> — atur
              susunan widget dashboard milik pengguna tersebut (lihat Bab 02).
            </li>
          </ul>
          <GuideImage src="/guide/users.png" alt="Halaman Manajemen Pengguna" caption="Manajemen Pengguna." />
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Roles &amp; Permissions</h3>
          <p className="text-sm text-muted-foreground">
            Matriks modul × hak akses per peran. Pilih satu peran, nyalakan/matikan izin per modul, lalu Simpan
            (atau Reset ke bawaan).
          </p>
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Riwayat Aktivitas</h3>
          <p className="text-sm text-muted-foreground">Dua jenis catatan yang bisa disaring dan dicari:</p>
          <ul className="list-disc space-y-2 pl-5 text-sm text-muted-foreground marker:text-primary">
            <li>
              <strong className="font-semibold text-foreground">Log Aktivitas</strong> — aksi, modul, tabel
              yang terdampak, serta nilai lama/baru.
            </li>
            <li>
              <strong className="font-semibold text-foreground">Log Masuk</strong> — alamat IP, info perangkat,
              waktu login/logout, dan status login.
            </li>
          </ul>
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Pengaturan Sistem</h3>
          <ul className="list-disc space-y-2 pl-5 text-sm text-muted-foreground marker:text-primary">
            <li>
              <strong className="font-semibold text-foreground">Umum</strong> — keamanan (maks. percobaan
              login, batas waktu sesi idle), penyimpanan (ukuran unggah maksimum, daftar IP MinIO yang
              diizinkan), dan tenggat temuan (hari jatuh tempo, hari auto-approve untuk follow-up &amp; WO/WR).
            </li>
            <li>
              <strong className="font-semibold text-foreground">Konfigurasi SMTP</strong> — pengaturan server
              email pengirim. Khusus Super Admin.
            </li>
            <li>
              <strong className="font-semibold text-foreground">Template Email</strong> — tiga templat: kode
              OTP lupa password, notifikasi penugasan temuan, konfirmasi inspeksi.
            </li>
            <li>
              <strong className="font-semibold text-foreground">API Key Power BI</strong> — kelola kunci API
              untuk integrasi Power BI.
            </li>
          </ul>
          <GuideImage src="/guide/settings.png" alt="Halaman Pengaturan Sistem" caption="Pengaturan Sistem." />
        </div>
      </div>
    </SectionShell>
  );
}
