import { SectionShell } from "../SectionShell";
import { RoleBadge } from "../RoleBadge";
import { GuideImage } from "../GuideImage";

export function GmpSection() {
  return (
    <SectionShell
      id="gmp"
      index={6}
      title="Data Inspeksi (GMP) & Export"
      lede="Menjelajah dan mengekspor rekap data inspeksi yang sudah selesai."
      role={<RoleBadge role="admin" />}
    >
      <div className="space-y-3">
        <p className="text-sm text-muted-foreground">
          Saring data lewat Area → Kawasan → Detail Kawasan dan rentang tanggal; Super Admin dapat menyaring
          lintas plant. Hasil saringan tampil sebagai tabel pratinjau di layar.
        </p>
        <p className="text-sm text-muted-foreground">
          Tekan &ldquo;Export Report&rdquo; untuk mengunduh rekap sebagai berkas Excel (.xlsx), diberi nama
          otomatis sesuai area dan waktu unduh.
        </p>
        <GuideImage src="/guide/gmp-data.png" alt="Halaman Data Inspeksi GMP" caption="Data Inspeksi (GMP)." />
      </div>
    </SectionShell>
  );
}
