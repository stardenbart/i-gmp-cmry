import { SectionShell } from "../SectionShell";
import { RoleBadge } from "../RoleBadge";
import { GuideImage } from "../GuideImage";

export function MonitoringSection() {
  return (
    <SectionShell
      id="monitoring"
      index={7}
      title="Monitoring"
      lede="Dasbor performa untuk memantau kinerja tim, di luar ringkasan yang sudah ada di Dashboard."
      role={<RoleBadge role="admin" />}
    >
      <div className="space-y-3">
        <ul className="list-disc space-y-2 pl-5 text-sm text-muted-foreground marker:text-primary">
          <li>
            <strong className="font-semibold text-foreground">Monitoring Auditor</strong> — tingkat
            penyelesaian, jumlah inspeksi berlangsung/draf/selesai per auditor.
          </li>
          <li>
            <strong className="font-semibold text-foreground">Monitoring PIC/Auditee</strong> — jumlah temuan
            closed, verified, in-progress, open, dan overdue per PIC.
          </li>
          <li>
            <strong className="font-semibold text-foreground">Monitoring Perintah Kerja</strong> — papan
            ringkasan WO/WR per area lengkap dengan pratinjau bukti sebelum/sesudah.
          </li>
        </ul>
        <GuideImage src="/guide/monitoring.png" alt="Halaman Monitoring" caption="Monitoring performa Auditor & PIC." />
      </div>
    </SectionShell>
  );
}
