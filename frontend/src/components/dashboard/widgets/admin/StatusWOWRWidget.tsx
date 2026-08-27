"use client";

import Link from "next/link";
import { useAdminDashboard } from "@/components/dashboard/admin/AdminDashboardContext";

export function StatusWOWRWidget() {
  const { stats, user } = useAdminDashboard();

  return (
    <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex flex-col justify-between h-full">
      <div>
        <div className="flex items-center justify-between mb-4 pb-3 border-b border-border">
          <h3 className="text-base font-semibold text-foreground">Status Maintenance (WO / WR)</h3>
          <span className="text-xs font-semibold px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-600 font-mono">
            Tingkat Verifikasi: {stats?.wowr_verified_rate || 0}%
          </span>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <div className="flex justify-between text-xs font-medium mb-1.5">
              <span className="text-muted-foreground">WO/WR Disetujui</span>
              <span className="font-bold text-emerald-600">
                {stats?.wowr_verified || 0} / {stats?.wowr_total || 0}
              </span>
            </div>
            <div className="w-full bg-muted h-2.5 rounded-full overflow-hidden">
              <div className="bg-emerald-500 h-full rounded-full transition-all duration-1000" style={{ width: `${stats?.wowr_verified_rate || 0}%` }} />
            </div>
          </div>
          <div className="flex items-center justify-between text-xs bg-muted/30 p-2.5 rounded-lg border border-border/50">
            <span className="text-purple-600 font-semibold">Menunggu: {stats?.wowr_pending || 0}</span>
            <span className="text-red-600 font-semibold">Ditolak: {stats?.wowr_rejected || 0}</span>
            <span className="text-muted-foreground font-semibold">Belum Bukti: {stats?.wowr_awaiting || 0}</span>
          </div>
        </div>
      </div>
      <Link
        href={`/cimory/${user?.plant_id || "global"}/dashboard/${user?.id}/monitoring/wowr`}
        className="mt-4 text-xs font-semibold text-center text-primary hover:underline block"
      >
        Lihat Laporan & Analytics WO/WR &rarr;
      </Link>
    </div>
  );
}
