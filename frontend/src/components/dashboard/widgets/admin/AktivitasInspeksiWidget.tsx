"use client";

import Link from "next/link";
import { useAdminDashboard } from "@/components/dashboard/admin/AdminDashboardContext";
import { StackedBar } from "@/components/dashboard/StackedBar";

export function AktivitasInspeksiWidget() {
  const { stats, user } = useAdminDashboard();
  const completed = stats?.inspections_completed || 0;
  const running = stats?.inspections_running || 0;
  const total = completed + running;

  return (
    <div className="bg-card rounded-lg p-5 border border-border shadow-sm flex flex-col justify-between h-full">
      <div>
        <div className="flex items-center justify-between mb-4 pb-3 border-b border-border">
          <h3 className="text-base font-semibold text-foreground">Aktivitas Inspeksi</h3>
          <span className="text-xs font-semibold px-2 py-0.5 rounded bg-primary/10 text-primary">
            Total {total}
          </span>
        </div>
        <StackedBar
          label="Aktivitas inspeksi"
          total={total}
          segments={[
            { key: "completed", label: "Selesai", value: completed, className: "bg-success" },
            { key: "running", label: "Berlangsung", value: running, className: "bg-info" },
          ]}
        />
      </div>
      <Link
        href={`/cimory/${user?.plant_id || "global"}/dashboard/${user?.id}/monitoring`}
        className="mt-6 text-xs font-semibold text-center text-muted-foreground hover:text-primary transition-colors block"
      >
        Lihat Detail Auditor &rarr;
      </Link>
    </div>
  );
}
