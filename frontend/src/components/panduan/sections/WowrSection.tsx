import { SectionShell } from "../SectionShell";
import { RoleBadge } from "../RoleBadge";
import { RoleScope } from "../PanduanFilterContext";
import { StatusFlow } from "../StatusFlow";
import { GuideImage } from "../GuideImage";
import { WOWR_STATUS_STYLE } from "../theme";

export function WowrSection() {
  return (
    <SectionShell
      id="wowr"
      index={6}
      title="Perintah Kerja (WO/WR)"
      lede="Jalur tindak lanjut formal untuk temuan yang butuh perbaikan lewat Work Order/Work Request."
      role={
        <div className="flex flex-wrap gap-2">
          <RoleBadge role="auditee" label="Auditee/PIC mengajukan" />
          <RoleBadge role="admin" label="Admin & Super Admin memvalidasi" />
          <RoleBadge role="auditor" label="Auditor memvalidasi" />
        </div>
      }
    >
      <div className="space-y-8">
        <StatusFlow
          steps={[
            { style: { label: "Belum Ada", text: "text-muted-foreground", bg: "bg-muted", border: "border-border" } },
            { style: WOWR_STATUS_STYLE.PendingValidation, caption: "PIC isi No. WO/WR + unggah bukti" },
            { style: WOWR_STATUS_STYLE.Verified, caption: "Auditor/Admin verifikasi" },
          ]}
          branches={["Auditor/Admin tolak → status Ditolak, PIC unggah ulang bukti."]}
        />

        <RoleScope roles={["auditee"]}>
          <div className="rounded-3xl border border-border border-l-2 border-l-rose-500 bg-card p-5">
            <h4 className="mb-2 text-sm font-semibold text-foreground">Mengajukan (Auditee/PIC)</h4>
            <p className="text-sm text-muted-foreground">
              Buka menu Perintah Kerja, isi nomor WO/WR pada temuan terkait, unggah foto bukti perbaikan, dan
              simpan — statusnya menjadi Menunggu Validasi. Deskripsi bukti bisa disunting kembali sebelum
              diverifikasi.
            </p>
          </div>
        </RoleScope>

        <RoleScope roles={["admin", "auditor"]}>
          <div className="rounded-3xl border border-border border-l-2 border-l-sky-500 bg-card p-5">
            <h4 className="mb-2 text-sm font-semibold text-foreground">Memverifikasi (Auditor/Admin)</h4>
            <p className="mb-3 text-sm text-muted-foreground">
              Tinjau bukti pada papan Perintah Kerja, lalu Verifikasi (temuan induk bisa otomatis ikut
              tertutup) atau Tolak (PIC perlu mengunggah ulang).
            </p>
            <GuideImage src="/guide/wowr-board.png" alt="Papan Perintah Kerja" caption="Papan Perintah Kerja (WO/WR)." />
          </div>
        </RoleScope>
      </div>
    </SectionShell>
  );
}
