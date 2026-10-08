"use client";

import Link from "next/link";
import { useAdminDashboard } from "@/components/dashboard/admin/AdminDashboardContext";
import { StackedBar } from "@/components/dashboard/StackedBar";

export function AktivitasPICWidget() {
  const { stats, user } = useAdminDashboard();

  const resolvedCount = stats?.issues_resolved ?? stats?.pic_followup_completed ?? 0;
  const openCount = stats?.total_open_issues || 0;
  const totalCount = openCount + resolvedCount;
  // Overdue findings are a subset of the open ones (backend: not closed AND
  // DueDate < now), so split open into on-time and overdue for the bar.
  const overdueCount = Math.min(stats?.issue_overdue ?? stats?.pic_followup_overdue ?? 0, openCount);
  const onTimeOpenCount = openCount - overdueCount;

  return (
    <div className="bg-card rounded-lg p-5 border border-border shadow-sm flex flex-col justify-between h-full">
      <div>
        <div className="flex items-center justify-between mb-4 pb-3 border-b border-border">
          <h3 className="text-base font-semibold text-foreground">Aktivitas Penanggung Jawab (PIC)</h3>
          <span className="text-xs font-semibold px-2 py-0.5 rounded bg-primary/10 text-primary">
            Total {totalCount} Temuan
          </span>
        </div>
        <StackedBar
          label="Aktivitas penanggung jawab"
          total={totalCount}
          segments={[
            { key: "resolved", label: "Diselesaikan", value: resolvedCount, className: "bg-success" },
            { key: "open", label: "Terbuka", value: onTimeOpenCount, className: "bg-warning" },
            { key: "overdue", label: "Jatuh Tempo", value: overdueCount, className: "bg-destructive" },
          ]}
        />
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
