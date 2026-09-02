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
          Buka menu &ldquo;Export&rdquo;, lalu pilih &ldquo;Laporan Template&rdquo; untuk format laporan resmi atau
          &ldquo;Tabel seperti di Web&rdquo; untuk susunan kolom yang sama dengan tabel Data GMP. Keduanya
          diunduh sebagai berkas Excel (.xlsx) dan mengikuti seluruh filter serta pencarian aktif.
        </p>
        <GuideImage src="/guide/gmp-data.png" alt="Halaman Data Inspeksi GMP" caption="Data Inspeksi (GMP)." />
      </div>
    </SectionShell>
  );
}
