"use client";

import Link from "next/link";
import { useAdminDashboard } from "@/components/dashboard/admin/AdminDashboardContext";

export function AktivitasPICWidget() {
  const { stats, user } = useAdminDashboard();

  const resolvedCount = stats?.issues_resolved ?? stats?.pic_followup_completed ?? 0;
  const openCount = stats?.total_open_issues || 0;
  const totalCount = openCount + resolvedCount;
  const resolvedRate = totalCount > 0 ? Math.round((resolvedCount / totalCount) * 100) : 0;
  const overdueCount = stats?.issue_overdue ?? stats?.pic_followup_overdue ?? 0;
  const overdueRate = totalCount > 0 ? Math.round((overdueCount / totalCount) * 100) : 0;

  return (
    <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex flex-col justify-between h-full">
      <div>
        <div className="flex items-center justify-between mb-4 pb-3 border-b border-border">
          <h3 className="text-base font-semibold text-foreground">Aktivitas Penanggung Jawab (PIC)</h3>
          <span className="text-xs font-semibold px-2 py-0.5 rounded bg-primary/10 text-primary">
            Total {totalCount} Temuan
          </span>
        </div>
        <div className="space-y-4">
          <div>
            <div className="flex justify-between text-xs font-medium mb-1.5">
              <span className="text-muted-foreground">Temuan Berhasil Diselesaikan</span>
              <span className="font-bold text-foreground">
                {resolvedCount} / {totalCount}
              </span>
            </div>
            <div className="w-full bg-muted h-2.5 rounded-full overflow-hidden">
              <div className="bg-green-600 h-full rounded-full transition-all duration-1000" style={{ width: `${resolvedRate}%` }} />
            </div>
          </div>
          <div>
            <div className="flex justify-between text-xs font-medium mb-1.5">
              <span className="text-muted-foreground">Temuan Jatuh Tempo (Overdue)</span>
              <span className="font-bold text-red-600">
                {overdueCount} / {totalCount}
              </span>
            </div>
            <div className="w-full bg-muted h-2.5 rounded-full overflow-hidden">
              <div className="bg-red-500 h-full rounded-full transition-all duration-1000" style={{ width: `${overdueRate}%` }} />
            </div>
          </div>
        </div>
      </div>
      <Link
        href={`/cimory/${user?.plant_id || "global"}/dashboard/${user?.id}/monitoring`}
        className="mt-6 text-xs font-semibold text-center text-muted-foreground hover:text-primary transition-colors block"
      >
        Lihat Detail PIC &rarr;
      </Link>
    </div>
  );
}
