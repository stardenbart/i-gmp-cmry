import { SectionShell } from "../SectionShell";
import { RoleBadge } from "../RoleBadge";
import { StepList } from "../StepList";
import { Callout } from "../Callout";
import { GuideImage } from "../GuideImage";

export function MulaiSection() {
  return (
    <SectionShell
      id="memulai"
      index={1}
      title="Memulai"
      lede="Cara masuk pertama kali, memulihkan akun yang lupa kata sandi, dan mengenali bagian-bagian utama aplikasi."
      role={<RoleBadge role="all" />}
    >
      <div className="space-y-8">
        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Masuk (Login)</h3>
          <p className="text-sm text-muted-foreground">
            Buka halaman login, isi <strong className="font-semibold text-foreground">Username</strong>{" "}
            (minimal 3 karakter) dan <strong className="font-semibold text-foreground">Password</strong>{" "}
            (minimal 6 karakter), lalu tekan <strong className="font-semibold text-foreground">Masuk</strong>.
            Anda akan diarahkan langsung ke dashboard yang sesuai dengan plant dan peran akun Anda.
          </p>
          <GuideImage src="/guide/login.png" alt="Halaman login I-GMP" caption="Halaman login." />
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Lupa Password</h3>
          <p className="text-sm text-muted-foreground">
            Jika lupa kata sandi, gunakan tautan &ldquo;Lupa password?&rdquo; di halaman login. Prosesnya tiga langkah:
          </p>
          <StepList
            steps={[
              <>
                <strong>Minta kode.</strong> Masukkan alamat email yang terdaftar. Sistem mengirim kode OTP 6
                digit ke email tersebut.
              </>,
              <>
                <strong>Verifikasi &amp; buat password baru.</strong> Masukkan kode OTP (berlaku 10 menit),
                lalu password baru dan konfirmasinya. Bisa kirim ulang kode atau ganti email bila diperlukan.
              </>,
              <>
                <strong>Selesai.</strong> Muncul konfirmasi berhasil dengan tautan kembali ke halaman login.
              </>,
            ]}
          />
          <Callout>
            Demi keamanan, pesan pada langkah pertama tetap menampilkan konfirmasi umum walau emailnya tidak
            terdaftar — ini bukan berarti kode berhasil terkirim.
          </Callout>
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Mengenal Navigasi Utama</h3>
          <p className="text-sm text-muted-foreground">Tiga bagian navigasi selalu ada, wujudnya berbeda di layar besar dan kecil:</p>
          <ul className="list-disc space-y-2 pl-5 text-sm text-muted-foreground marker:text-primary">
            <li>
              <strong className="font-semibold text-foreground">Sidebar (desktop)</strong> — menu utama di kiri,
              terbagi dua kelompok: Utama (Dasbor, Panduan, Inspeksi, Temuan Inspeksi, Perintah Kerja) dan
              Administrator — kelompok kedua hanya tampil bila Anda punya akses ke minimal satu menunya.
            </li>
            <li>
              <strong className="font-semibold text-foreground">Navigasi bawah (mobile)</strong> — Home, Inspeksi,
              Temuan, WO/WR, Profil, ditata sebagai bilah tetap di bawah layar.
            </li>
            <li>
              <strong className="font-semibold text-foreground">Header (atas)</strong> — judul halaman aktif,
              tombol Panduan, pasang aplikasi (PWA), sakelar tema gelap/terang, lonceng notifikasi, dan tombol
              keluar.
            </li>
          </ul>
          <Callout>
            Menu yang Anda lihat bisa berbeda dari rekan kerja dengan peran yang sama — setiap item menu muncul
            berdasarkan hak akses (permission) yang diatur Admin, bukan cuma nama peran.
          </Callout>
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Profil &amp; Ganti Password</h3>
          <p className="text-sm text-muted-foreground">
            Halaman Profil (ikon akun di header, atau menu Profil di navigasi bawah) tersedia untuk semua peran:
          </p>
          <ul className="list-disc space-y-2 pl-5 text-sm text-muted-foreground marker:text-primary">
            <li>Melihat identitas akun: nama, username, peran, departemen, status akun.</li>
            <li>Mengubah nama lengkap dan email lewat mode Edit pada kartu &ldquo;Informasi Profil&rdquo;.</li>
            <li>
              Mengganti password lewat kartu &ldquo;Ubah Password&rdquo; (password lama, baru minimal 8 karakter,
              konfirmasi) — indikator kekuatan password tampil saat mengetik.
            </li>
            <li>Tombol Keluar untuk mengakhiri sesi.</li>
          </ul>
        </div>
      </div>
    </SectionShell>
  );
}
