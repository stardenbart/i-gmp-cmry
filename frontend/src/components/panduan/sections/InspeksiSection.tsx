import { SectionShell } from "../SectionShell";
import { RoleBadge } from "../RoleBadge";
import { Callout } from "../Callout";
import { GuideImage } from "../GuideImage";

export function InspeksiSection() {
  return (
    <SectionShell
      id="inspeksi"
      index={4}
      title="Inspeksi"
      lede="Mengisi checklist audit di lapangan — sumber utama lahirnya Temuan."
      role={
        <div className="flex flex-wrap gap-2">
          <RoleBadge role="admin" />
          <RoleBadge role="auditor" />
        </div>
      }
    >
      <div className="space-y-8">
        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Membuat Inspeksi Baru</h3>
          <p className="text-sm text-muted-foreground">
            Dari menu Inspeksi → Buat Baru, pilih lokasi secara berjenjang: Area → Kawasan → Detail Kawasan,
            lalu tekan mulai untuk membuka checklist.
          </p>
          <GuideImage src="/guide/inspeksi-list.png" alt="Daftar inspeksi" caption="Daftar Inspeksi." />
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Mengisi Checklist</h3>
          <p className="text-sm text-muted-foreground">
            Checklist tersusun tiga tingkat: tab Aspek → sub-tab Detail Aspek → daftar Uraian. Setiap uraian
            dinilai <span className="font-semibold text-emerald-500">OK</span> atau{" "}
            <span className="font-semibold text-red-500">NG</span>.
          </p>
          <Callout variant="warn">
            Setiap uraian yang dinilai <strong>NG wajib</strong> disertai minimal satu foto bukti, keterangan
            tertulis, dan klasifikasi HEI — tanpa ketiganya, penilaian tidak bisa disimpan.
          </Callout>
          <p className="text-sm text-muted-foreground">
            Bila ada auditor lain yang sedang mengisi Aspek yang sama, aplikasi menampilkan penanda &ldquo;sedang
            dikerjakan&rdquo; secara langsung supaya tidak terjadi tabrakan data.
          </p>
          <GuideImage src="/guide/inspeksi-checklist.png" alt="Checklist inspeksi" caption="Mengisi checklist Aspek/Uraian." />
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Menyelesaikan &amp; Membuka Kembali</h3>
          <p className="text-sm text-muted-foreground">
            Tekan &ldquo;Selesaikan Audit&rdquo; saat semua item sudah dinilai — status inspeksi berubah menjadi
            Completed, dan setiap penilaian NG otomatis menjadi satu Temuan baru (lihat Bab 04).
          </p>
          <p className="text-sm text-muted-foreground">
            Admin atau Auditor pada plant yang sama bisa membuka kembali inspeksi yang sudah selesai lewat
            tombol &ldquo;Edit Inspeksi&rdquo;, mengembalikan statusnya ke Ongoing.
          </p>
        </div>
      </div>
    </SectionShell>
  );
}
