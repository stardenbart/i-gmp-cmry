"use client";

import Link from "next/link";
import { useAdminDashboard } from "@/components/dashboard/admin/AdminDashboardContext";

export function AktivitasInspeksiWidget() {
  const { stats, user } = useAdminDashboard();
  const completed = stats?.inspections_completed || 0;
  const running = stats?.inspections_running || 0;
  const total = completed + running;

  return (
    <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex flex-col justify-between h-full">
      <div>
        <div className="flex items-center justify-between mb-4 pb-3 border-b border-border">
          <h3 className="text-base font-semibold text-foreground">Aktivitas Inspeksi</h3>
          <span className="text-xs font-semibold px-2 py-0.5 rounded bg-primary/10 text-primary">
            Total {total}
          </span>
        </div>
        <div className="space-y-4">
          <div>
            <div className="flex justify-between text-xs font-medium mb-1.5">
              <span className="text-muted-foreground">Inspeksi Selesai</span>
              <span className="font-bold text-foreground">
                {completed} / {total || 1}
              </span>
            </div>
            <div className="w-full bg-muted h-2.5 rounded-full overflow-hidden">
              <div
                className="bg-green-600 h-full rounded-full transition-all duration-1000"
                style={{ width: `${total > 0 ? (completed / total) * 100 : 0}%` }}
              />
            </div>
          </div>
          <div>
            <div className="flex justify-between text-xs font-medium mb-1.5">
              <span className="text-muted-foreground">Inspeksi Berlangsung</span>
              <span className="font-bold text-foreground">
                {running} / {total || 1}
              </span>
            </div>
            <div className="w-full bg-muted h-2.5 rounded-full overflow-hidden">
              <div
                className="bg-primary h-full rounded-full transition-all duration-1000"
                style={{ width: `${total > 0 ? (running / total) * 100 : 0}%` }}
              />
            </div>
          </div>
        </div>
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
