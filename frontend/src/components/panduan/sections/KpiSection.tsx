import { SectionShell } from "../SectionShell";
import { RoleBadge } from "../RoleBadge";
import { RoleScope } from "../PanduanFilterContext";
import { Callout } from "../Callout";
import { StepList } from "../StepList";

export function KpiSection() {
  return (
    <SectionShell
      id="kpi"
      index={3}
      title="Dashboard KPI"
      lede="Kanvas kosong untuk membangun visualisasi analitik sendiri — beda dari Dashboard utama yang widgetnya sudah baku per peran."
      role={<RoleBadge role="admin" label="Tergantung izin — default hanya Admin/Super Admin" />}
    >
      <div className="space-y-8">
        <div className="space-y-3">
          <Callout variant="info">
            Diakses lewat menu &ldquo;Dashboard KPI&rdquo; di sidebar — menu ini hanya tampil kalau Admin/Super
            Admin sudah memberi izin &ldquo;View Dashboard KPI&rdquo;. Aturnya dari menu Manajemen Pengguna →
            pilih pengguna → tab Hak Akses Khusus.
          </Callout>
          <p className="text-sm text-muted-foreground">
            Berbeda dari Dashboard utama (Bab 02) yang widgetnya sudah baku per peran, Dashboard KPI adalah
            kanvas kosong — Anda membangun sendiri visualisasi apa pun dari data audit yang tersedia, lalu
            mengatur tata letaknya sendiri.
          </p>
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Membuat Visualisasi Baru</h3>
          <StepList
            steps={[
              <>Tekan tombol <strong>&ldquo;Tambah Visualisasi&rdquo;</strong>.</>,
              <>Seret atau klik satu atau beberapa field <strong>Kategori</strong> — menambah lebih dari satu
                kategori mengaktifkan drill-down bertingkat.</>,
              <>Seret atau klik satu atau beberapa field <strong>Nilai</strong> — maksimal 4 untuk kebanyakan
                jenis visualisasi, khusus <strong>Tabel</strong> bisa sampai 30.</>,
              <>Pilih <strong>jenis visualisasi</strong> — hanya jenis yang cocok dengan kombinasi Kategori
                &amp; Nilai yang dipilih yang bisa diaktifkan.</>,
              <>Atur opsi tampilan bila perlu: <strong>Aktifkan Label</strong>, <strong>Tampilkan Nilai di
                Label</strong>, atau <strong>Tambahkan Garis Tren</strong> (khusus jenis Bar).</>,
              <>Isi <strong>judul</strong>, lalu tekan <strong>&ldquo;Tambahkan&rdquo;</strong>.</>,
            ]}
          />
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Jenis Visualisasi</h3>
          <p className="text-sm text-muted-foreground">16 jenis tersedia, dikelompokkan sama seperti pemilih di builder:</p>
          <ul className="grid grid-cols-1 gap-x-6 gap-y-1.5 text-sm text-muted-foreground sm:grid-cols-2">
            <li><strong className="font-semibold text-foreground">Perbandingan</strong> — Bar, Bar Horizontal, Bar Bertumpuk, Line, Area, Radar</li>
            <li><strong className="font-semibold text-foreground">Proporsi</strong> — Pie, Donut, Treemap, Funnel</li>
            <li><strong className="font-semibold text-foreground">Nilai Tunggal</strong> — Kartu Angka, Gauge</li>
            <li><strong className="font-semibold text-foreground">Matriks (2 Kategori)</strong> — Heatmap, Sankey</li>
            <li><strong className="font-semibold text-foreground">Lainnya</strong> — Scatter, Tabel</li>
          </ul>
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Kategori Bertingkat &amp; Drill-Down</h3>
          <p className="text-sm text-muted-foreground">
            Untuk jenis chart biasa (Bar/Line/Area/Pie/dst.), menambah Kategori kedua dan seterusnya
            mengaktifkan tombol naik/turun level di pojok widget — satu level kategori ditampilkan sekaligus,
            digeser satu per satu.
          </p>
          <Callout variant="info">
            Khusus jenis <strong>Tabel</strong>, seluruh level Kategori langsung ditampilkan bersamaan sebagai
            kolom bertingkat dalam satu tabel — tidak perlu klik apa pun untuk melihat level berikutnya, dan
            Tabel jugalah satu-satunya jenis yang boleh memakai lebih dari 4 Nilai sekaligus.
          </Callout>
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Mengedit &amp; Menata Ulang Widget</h3>
          <p className="text-sm text-muted-foreground">
            Tekan ikon pensil pada widget untuk membuka ulang builder dengan konfigurasi yang sama dan mengubahnya.
            Tekan &ldquo;Sesuaikan Dashboard&rdquo; untuk memindah atau mengubah ukuran widget, sama seperti di
            Dashboard utama.
          </p>
        </div>

        <RoleScope roles={["admin"]}>
          <div className="space-y-3">
            <h3 className="text-lg font-semibold text-foreground">Berbagi Publik (KPI Share)</h3>
            <Callout variant="warn">
              Membuat, merotasi, dan mencabut link publik hanya bisa dilakukan Admin/Super Admin.
            </Callout>
            <StepList
              steps={[
                <>Tekan tombol <strong>&ldquo;Bagikan&rdquo;</strong> di halaman Dashboard KPI.</>,
                <>Isi <strong>Nama link</strong> (label internal), <strong>Judul yang tampil</strong> (dilihat
                  pengunjung publik), dan <strong>Masa berlaku</strong>.</>,
                <>Centang <strong>&ldquo;Izinkan pengunjung mengubah periode&rdquo;</strong> bila pengunjung
                  publik boleh mengganti rentang tanggal sendiri.</>,
                <>Tekan <strong>&ldquo;Buat Link Publik&rdquo;</strong> — token hanya ditampilkan sekali, salin
                  segera.</>,
                <><strong>Rotasi</strong> membuat token baru dan langsung membatalkan token lama; <strong>Cabut</strong> menonaktifkan link untuk selamanya. Keduanya bisa dilakukan kapan saja dari daftar link yang sama.</>,
              ]}
            />
          </div>
        </RoleScope>
      </div>
    </SectionShell>
  );
}
