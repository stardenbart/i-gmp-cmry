import { SectionShell } from "../SectionShell";
import { RoleBadge } from "../RoleBadge";
import { RoleScope } from "../PanduanFilterContext";
import { Callout } from "../Callout";
import { StepList } from "../StepList";
import { GuideImage } from "../GuideImage";

export function DashboardSection() {
  return (
    <SectionShell
      id="dashboard"
      index={2}
      title="Dashboard"
      lede="Halaman pertama setelah masuk. Isinya kumpulan widget ringkasan yang disesuaikan dengan peran Anda."
      role={<RoleBadge role="all" label="Semua Peran — isi berbeda per peran" />}
    >
      <div className="space-y-8">
        <div className="space-y-4">
          <h3 className="text-lg font-semibold text-foreground">Widget Sesuai Peran</h3>

          <RoleScope roles={["admin"]}>
            <div className="space-y-3 rounded-3xl border border-border border-l-2 border-l-primary bg-card p-5">
              <RoleBadge role="admin" />
              <p className="text-sm text-muted-foreground">
                Kartu Statistik ringkas, Aktivitas Inspeksi, Aktivitas PIC, Status per Lokasi Auditee, Status
                Perintah Kerja/WOWR, dan Grafik Tren Kepatuhan. Super Admin juga punya filter Pilih Plant; Admin
                biasa melihat plant-nya sendiri sebagai lencana tetap. Ada juga filter Pilih Area.
              </p>
              <GuideImage src="/guide/dashboard-admin.png" alt="Dashboard Admin" caption="Dashboard Admin/Super Admin." />
            </div>
          </RoleScope>

          <RoleScope roles={["auditor"]}>
            <div className="space-y-3 rounded-3xl border border-border border-l-2 border-l-sky-500 bg-card p-5">
              <RoleBadge role="auditor" />
              <p className="text-sm text-muted-foreground">
                Kartu Statistik pribadi (Total Inspeksi, Berlangsung, Selesai, Menunggu Validasi dengan lencana
                berdenyut), banner &ldquo;Menunggu Validasi Anda&rdquo; berisi hingga 3 temuan teratas, dan
                Grafik Tren pribadi.
              </p>
              <GuideImage src="/guide/dashboard-auditor.png" alt="Dashboard Auditor" caption="Dashboard Auditor." />
            </div>
          </RoleScope>

          <RoleScope roles={["auditee"]}>
            <div className="space-y-3 rounded-3xl border border-border border-l-2 border-l-rose-500 bg-card p-5">
              <RoleBadge role="auditee" />
              <p className="text-sm text-muted-foreground">
                Kartu Statistik tugas (Total Tugas, Terbuka, Menunggu Validasi, Selesai) dan Daftar Tugas
                Terbuka diurutkan berdasarkan tenggat waktu — tugas yang lewat tenggat ditandai menyala.
              </p>
              <GuideImage src="/guide/dashboard-auditee.png" alt="Dashboard Auditee" caption="Dashboard Auditee/PIC." />
            </div>
          </RoleScope>
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Filter Tren Kepatuhan</h3>
          <p className="text-sm text-muted-foreground">
            Grafik tren (widget admin &amp; auditor) punya dua mode, dipilih lewat tombol segmen di bagian atas
            grafik:
          </p>
          <ul className="list-disc space-y-2 pl-5 text-sm text-muted-foreground marker:text-primary">
            <li>
              <strong className="font-semibold text-foreground">Kuartal</strong> — jendela tetap berisi 8
              kuartal kalender terakhir, tanpa pengaturan tambahan.
            </li>
            <li>
              <strong className="font-semibold text-foreground">Rentang Tanggal</strong> — kolom Dari/Sampai,
              ditambah pilihan &ldquo;Kelompokkan per&rdquo;: Harian, Mingguan, Bulanan, atau Tahunan.
            </li>
          </ul>
          <GuideImage src="/guide/trend-filter.png" alt="Filter tren rentang tanggal" caption="Mode Rentang Tanggal dengan Kelompokkan per." />
          <p className="text-sm text-muted-foreground">
            Pilihan mode, rentang, dan pengelompokan ikut tersimpan di tautan halaman — jadi tampilan yang sama
            bisa langsung dibagikan lewat salin-tempel URL.
          </p>
        </div>

        <RoleScope roles={["admin"]}>
          <div className="space-y-3">
            <h3 className="text-lg font-semibold text-foreground">Mengatur Ulang Tata Letak Dashboard</h3>
            <Callout variant="warn">
              Pengguna biasa <strong>tidak bisa</strong> mengubah tata letak dashboardnya sendiri. Pengaturan
              ini hanya dilakukan Admin/Super Admin, dari menu Manajemen Pengguna → pilih pengguna → tab Tata
              Letak Dashboard.
            </Callout>
            <Callout variant="info">
              Daftar widget di tab ini <strong>tidak lagi dibatasi sesuai peran pengguna yang diedit</strong> —
              Admin bisa memasang widget milik peran mana pun (Admin, Auditor, atau Auditee) ke dashboard
              siapa saja, misalnya memberi seorang Auditee salah satu widget yang biasanya cuma tampil di
              dashboard Auditor. Widget yang belum pernah ditambahkan tetap tersembunyi secara default di
              dashboard pengguna tersebut — baru muncul setelah Admin menambahkannya secara eksplisit di sini.
            </Callout>
            <StepList
              steps={[
                <><strong>Aktifkan mode edit</strong> lewat tombol &ldquo;Sesuaikan Dashboard&rdquo;.</>,
                <><strong>Pindahkan widget</strong> dengan menyeret bilah judulnya (hanya tampil dalam mode edit).</>,
                <><strong>Ubah ukuran</strong> dengan menyeret sudut kanan-bawah widget.</>,
                <><strong>Sembunyikan widget</strong> lewat tombol &ldquo;×&rdquo; di judulnya.</>,
                <><strong>Tampilkan kembali</strong> dengan menekan chip widget di tray &ldquo;Disembunyikan&rdquo;.</>,
                <><strong>Simpan</strong> lewat &ldquo;Simpan Tata Letak&rdquo;, atau <strong>Batal</strong> untuk membatalkan.</>,
              ]}
            />
          </div>
        </RoleScope>

        <div className="space-y-2">
          <h3 className="text-lg font-semibold text-foreground">Pembaruan Otomatis</h3>
          <p className="text-sm text-muted-foreground">
            Dashboard memeriksa pembaruan setiap ±30 detik secara diam-diam (tanpa memuat ulang halaman) ketika
            ada inspeksi atau temuan baru. Auditor dan Auditee juga punya tombol refresh manual di header;
            Admin mengandalkan pembaruan otomatis dan perubahan filter.
          </p>
        </div>
      </div>
    </SectionShell>
  );
}
