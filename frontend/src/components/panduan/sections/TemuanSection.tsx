import { SectionShell } from "../SectionShell";
import { RoleBadge } from "../RoleBadge";
import { RoleScope } from "../PanduanFilterContext";
import { Callout } from "../Callout";
import { StepList } from "../StepList";
import { StatusFlow } from "../StatusFlow";
import { Pill } from "../Pill";
import { ISSUE_STATUS_STYLE } from "../theme";
import { GuideImage } from "../GuideImage";

export function TemuanSection() {
  return (
    <SectionShell
      id="temuan"
      index={4}
      title="Temuan Inspeksi"
      lede="Setiap penilaian NG menjadi satu Temuan yang harus ditindaklanjuti hingga tuntas."
      role={<RoleBadge role="all" label="Melibatkan semua peran" />}
    >
      <div className="space-y-8">
        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-foreground">Alur Status</h3>
          <StatusFlow
            steps={[
              { style: ISSUE_STATUS_STYLE.Open },
              { style: ISSUE_STATUS_STYLE.InProgress, caption: "PIC mulai kerjakan" },
              { style: ISSUE_STATUS_STYLE.PendingValidation, caption: "PIC ajukan validasi" },
              { style: ISSUE_STATUS_STYLE.Closed, caption: "Auditor setujui" },
            ]}
            branches={["Auditor tolak → temuan kembali ke status In Progress agar PIC melengkapi buktinya."]}
          />
          <p className="text-xs leading-relaxed text-muted-foreground">
            Status <Pill style={ISSUE_STATUS_STYLE.Overdue} className="mx-1 align-middle" /> muncul otomatis
            begitu tenggat waktu terlewati. Status{" "}
            <Pill style={ISSUE_STATUS_STYLE.Verified} className="mx-1 align-middle" /> adalah varian Closed
            khusus saat penutupan dilakukan melalui verifikasi Perintah Kerja (WO/WR) — lihat Bab 05.
          </p>
        </div>

        <div className="overflow-x-auto rounded-3xl border border-border">
          <table className="w-full text-sm">
            <caption className="sr-only">Rangkuman status Temuan</caption>
            <thead className="bg-muted/50 text-left text-xs uppercase tracking-wide text-muted-foreground">
              <tr>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3 font-medium">Arti</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              <tr>
                <td className="px-4 py-3"><Pill style={ISSUE_STATUS_STYLE.Open} /></td>
                <td className="px-4 py-3 text-muted-foreground">Baru dibuat dari hasil NG, belum dikerjakan.</td>
              </tr>
              <tr>
                <td className="px-4 py-3"><Pill style={ISSUE_STATUS_STYLE.InProgress} /></td>
                <td className="px-4 py-3 text-muted-foreground">Sedang ditindaklanjuti oleh Auditee/PIC.</td>
              </tr>
              <tr>
                <td className="px-4 py-3"><Pill style={ISSUE_STATUS_STYLE.PendingValidation} /></td>
                <td className="px-4 py-3 text-muted-foreground">Bukti sudah diunggah, menunggu keputusan Auditor.</td>
              </tr>
              <tr>
                <td className="px-4 py-3"><Pill style={ISSUE_STATUS_STYLE.Verified} /></td>
                <td className="px-4 py-3 text-muted-foreground">Disetujui Auditor dan dianggap tuntas.</td>
              </tr>
              <tr>
                <td className="px-4 py-3"><Pill style={ISSUE_STATUS_STYLE.Overdue} /></td>
                <td className="px-4 py-3 text-muted-foreground">Varian status di atas yang melewati tenggat waktu penyelesaian.</td>
              </tr>
            </tbody>
          </table>
        </div>

        <RoleScope roles={["auditee"]}>
          <div className="rounded-3xl border border-border border-l-2 border-l-rose-500 bg-card p-5">
            <h4 className="mb-3 text-sm font-semibold text-foreground">Alur untuk Auditee / PIC</h4>
            <StepList
              steps={[
                <><strong>Mulai Kerjakan</strong> — pada temuan berstatus Open/Overdue.</>,
                <><strong>Unggah bukti perbaikan</strong> (foto follow-up) beserta keterangan.</>,
                <><strong>Ajukan Validasi</strong> — temuan pindah ke antrean tinjauan Auditor.</>,
              ]}
            />
          </div>
        </RoleScope>

        <RoleScope roles={["admin", "auditor"]}>
          <div className="rounded-3xl border border-border border-l-2 border-l-sky-500 bg-card p-5">
            <h4 className="mb-3 text-sm font-semibold text-foreground">Alur untuk Auditor</h4>
            <p className="mb-2 text-sm text-muted-foreground">
              Temuan yang menunggu keputusan tampil sebagai banner &ldquo;Menunggu Validasi&rdquo;. Pilihannya:
            </p>
            <ul className="mb-3 list-disc space-y-1.5 pl-5 text-sm text-muted-foreground marker:text-primary">
              <li><strong className="font-semibold text-foreground">Setujui — Tutup Temuan</strong></li>
              <li><strong className="font-semibold text-foreground">Tolak</strong> (temuan kembali ke status In Progress).</li>
            </ul>
            <GuideImage src="/guide/temuan-detail.png" alt="Detail temuan menunggu validasi" caption="Banner Menunggu Validasi pada detail Temuan." />
          </div>
        </RoleScope>

        <Callout>
          Syarat penutupan: setiap foto NG awal harus punya pasangan foto follow-up. Bila temuan memakai
          Perintah Kerja (WO/WR), nomor WO/WR wajib diisi dan berstatus Verified sebelum temuan bisa ditutup.
        </Callout>
      </div>
    </SectionShell>
  );
}
