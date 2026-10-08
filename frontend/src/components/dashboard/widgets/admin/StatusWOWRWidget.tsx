"use client";

import Link from "next/link";
import { useAdminDashboard } from "@/components/dashboard/admin/AdminDashboardContext";
import { StackedBar } from "@/components/dashboard/StackedBar";

export function StatusWOWRWidget() {
  const { stats, user } = useAdminDashboard();

  return (
    <div className="bg-card rounded-lg p-5 border border-border shadow-sm flex flex-col justify-between h-full">
      <div>
        <div className="flex items-center justify-between mb-4 pb-3 border-b border-border">
          <h3 className="text-base font-semibold text-foreground">Status Maintenance (WO / WR)</h3>
          <span className="text-xs font-semibold px-2 py-0.5 rounded bg-success/10 text-success">
            Tingkat Verifikasi: {stats?.wowr_verified_rate || 0}%
          </span>
        </div>
        <StackedBar
          label="Status WO/WR"
          total={stats?.wowr_total || 0}
          segments={[
            { key: "verified", label: "Disetujui", value: stats?.wowr_verified || 0, className: "bg-success" },
            { key: "pending", label: "Menunggu", value: stats?.wowr_pending || 0, className: "bg-info" },
            { key: "rejected", label: "Ditolak", value: stats?.wowr_rejected || 0, className: "bg-destructive" },
            { key: "awaiting", label: "Belum Bukti", value: stats?.wowr_awaiting || 0, className: "bg-muted-foreground/50" },
          ]}
        />
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
