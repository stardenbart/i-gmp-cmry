import { SectionShell } from "../SectionShell";
import { RoleBadge } from "../RoleBadge";
import { Faq } from "../Faq";

const PERMISSIONS = [
  { code: "PERM-INSP-R/C", module: "Inspeksi (lihat/buat)", auditor: true, auditee: false },
  { code: "PERM-ISS-R/U", module: "Temuan Inspeksi", auditor: true, auditee: true },
  { code: "PERM-WOWR-R/U", module: "Perintah Kerja", auditor: true, auditee: true },
  { code: "PERM-GMP-R", module: "Data Inspeksi (GMP)", auditor: false, auditee: false },
  { code: "PERM-MSTR-R", module: "Data Induk", auditor: false, auditee: false },
  { code: "PERM-USR-C/R/U/D", module: "Manajemen Pengguna", auditor: false, auditee: false },
  { code: "PERM-LOG-R", module: "Riwayat Aktivitas", auditor: false, auditee: false },
  { code: "PERM-STNG-R", module: "Pengaturan Sistem", auditor: false, auditee: false },
];

const GLOSSARY: [string, string][] = [
  ["Temuan", "Catatan ketidaksesuaian (NG) hasil inspeksi yang perlu ditindaklanjuti."],
  ["Kawasan / Detail Kawasan", "Jenjang lokasi di bawah Area, dipakai untuk menentukan lokasi inspeksi secara spesifik."],
  ["Aspek / Uraian", "Struktur berjenjang checklist inspeksi: Aspek → Detail Aspek → daftar Uraian yang dinilai."],
  ["HEI", "Klasifikasi tingkat risiko/keparahan yang dilekatkan pada setiap foto bukti temuan NG."],
  ["WOWR", "Work Order / Work Request — jalur perbaikan formal dengan nomor tiket dan bukti verifikasi."],
  ["PIC", "Penanggung Jawab — sebutan lain untuk peran Auditee di aplikasi ini."],
  ["GMP", "Good Manufacturing Practice — rekap data inspeksi kepatuhan produksi."],
  ["Plant", "Pabrik/lokasi tingkat tertinggi dalam hierarki data induk."],
];

export function ReferensiSection() {
  return (
    <SectionShell
      id="referensi"
      index={10}
      title="Referensi"
      lede="Tabel kode izin, istilah, dan pertanyaan yang sering muncul."
      role={<RoleBadge role="all" />}
    >
      <div className="space-y-10">
        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Kode Izin (Permission)</h3>
          <p className="text-sm text-muted-foreground">
            Admin/Super Admin selalu punya semua akses; peran lain mengikuti bawaan di bawah kecuali diubah
            lewat Permission per pengguna (Bab 08).
          </p>
          <div className="overflow-x-auto rounded-3xl border border-border">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 text-left text-xs uppercase tracking-wide text-muted-foreground">
                <tr>
                  <th className="px-4 py-3 font-medium">Kode</th>
                  <th className="px-4 py-3 font-medium">Modul</th>
                  <th className="px-4 py-3 font-medium">Auditor</th>
                  <th className="px-4 py-3 font-medium">Auditee/PIC</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {PERMISSIONS.map((p) => (
                  <tr key={p.code}>
                    <td className="px-4 py-3 font-mono text-xs text-foreground">{p.code}</td>
                    <td className="px-4 py-3 text-muted-foreground">{p.module}</td>
                    <td className="px-4 py-3 text-muted-foreground">{p.auditor ? "✓" : "—"}</td>
                    <td className="px-4 py-3 text-muted-foreground">{p.auditee ? "✓" : "—"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Glosarium</h3>
          <dl className="grid gap-x-6 gap-y-3 sm:grid-cols-[max-content_1fr]">
            {GLOSSARY.map(([term, def]) => (
              <div key={term} className="contents">
                <dt className="font-mono text-sm font-semibold text-primary sm:whitespace-nowrap">{term}</dt>
                <dd className="mb-2 text-sm text-muted-foreground sm:mb-0">{def}</dd>
              </div>
            ))}
          </dl>
        </div>

        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Pertanyaan Umum</h3>
          <div className="space-y-2">
            <Faq q="Kenapa menu saya berbeda dengan rekan kerja yang perannya sama?">
              Setiap item menu muncul berdasarkan kode izin (permission) masing-masing akun, yang bisa diatur
              ulang per pengguna oleh Admin — bukan semata-mata nama perannya. Hubungi Admin bila merasa ada
              akses yang seharusnya Anda punya.
            </Faq>
            <Faq q="Kenapa saya tidak bisa mengatur tata letak dashboard sendiri?">
              Kustomisasi tata letak dashboard memang sengaja dibatasi hanya untuk Admin/Super Admin, dilakukan
              dari Manajemen Pengguna. Ini menjaga tampilan dashboard tetap konsisten antar pengguna dalam tim
              yang sama.
            </Faq>
            <Faq q="Temuan sudah saya unggah buktinya, tapi kenapa belum bisa ditutup?">
              Pastikan setiap foto NG awal sudah punya pasangan foto follow-up, dan bila memakai Perintah
              Kerja, nomor WO/WR sudah diisi serta berstatus Verified. Setelah itu, penutupan menjadi wewenang
              Auditor.
            </Faq>
            <Faq q="Grafik tren di dashboard menunjukkan data kosong, kenapa?">
              Cek apakah tanggal &ldquo;Dari&rdquo;/&ldquo;Sampai&rdquo; pada mode Rentang Tanggal sudah terisi
              dan masuk akal (mulai tidak melewati akhir). Coba juga beralih ke mode Kuartal untuk melihat data
              8 kuartal terakhir sebagai pembanding.
            </Faq>
            <Faq q="Lupa password dan tidak menerima email OTP?">
              Periksa folder spam, pastikan email yang dimasukkan sama dengan yang terdaftar di akun, lalu
              gunakan tombol &ldquo;Kirim ulang kode&rdquo; di layar verifikasi. Bila tetap gagal, hubungi Admin
              untuk pengecekan konfigurasi SMTP.
            </Faq>
          </div>
        </div>
      </div>
    </SectionShell>
  );
}
