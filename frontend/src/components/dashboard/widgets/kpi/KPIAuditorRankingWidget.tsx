"use client";

import Link from "next/link";
import { cn } from "@/lib/utils";
import { useKPIDashboard } from "@/components/dashboard/kpi/KPIDashboardContext";
import { VisualizationSwitch } from "@/components/dashboard/kpi/VisualizationSwitch";
import type { VizType } from "@/components/dashboard/types";

export function KPIAuditorRankingWidget({ vizType }: { vizType?: VizType }) {
  const { auditorDetail, user } = useKPIDashboard();
  const ranked = [...auditorDetail].sort((a, b) => b.total - a.total).slice(0, 10);

  if (vizType && vizType !== "list") {
    return (
      <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex h-full flex-col">
        <h3 className="mb-3 text-base font-semibold text-foreground">Kinerja Auditor</h3>
        <div className="min-h-0 flex-1">
          <VisualizationSwitch
            vizType={vizType}
            data={ranked as unknown as Record<string, string | number>[]}
            categoryKey="full_name"
            valueKeys={[{ key: "completion_rate", name: "Completion Rate (%)", color: "#2563eb" }]}
          />
        </div>
      </div>
    );
  }

  return (
    <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex flex-col justify-between h-full">
      <div>
        <div className="flex items-center justify-between mb-4 pb-3 border-b border-border">
          <h3 className="text-base font-semibold text-foreground">Kinerja Auditor</h3>
          <span className="text-xs font-semibold px-2 py-0.5 rounded bg-primary/10 text-primary">Top {ranked.length}</span>
        </div>
        <div className="space-y-2 max-h-[260px] overflow-y-auto pr-1">
          {ranked.length === 0 ? (
            <div className="p-6 text-center text-xs text-muted-foreground italic">Belum ada data auditor</div>
          ) : (
            ranked.map((auditor, idx) => (
              <div key={auditor.user_id} className="flex items-center justify-between p-3 bg-muted/40 rounded-lg border border-border/70 hover:bg-muted/70 transition-colors">
                <div className="flex items-center gap-2.5 min-w-0">
                  <span className="shrink-0 flex h-6 w-6 items-center justify-center rounded-full bg-primary/10 text-[11px] font-bold text-primary">{idx + 1}</span>
                  <div className="min-w-0">
                    <p className="text-xs font-semibold text-foreground truncate">{auditor.full_name}</p>
                    <p className="text-[10px] text-muted-foreground">{auditor.total} inspeksi · {auditor.ongoing} berjalan</p>
                  </div>
                </div>
                <p className={cn("shrink-0 text-xs font-bold", auditor.completion_rate < 70 ? "text-red-500" : "text-green-600")}>
                  {Math.round(auditor.completion_rate)}%
                </p>
              </div>
            ))
          )}
        </div>
      </div>
      <Link
        href={`/cimory/${user?.plant_id || "global"}/dashboard/${user?.id}/monitoring`}
        className="mt-4 text-xs font-semibold text-center text-muted-foreground hover:text-primary transition-colors block"
      >
        Lihat Semua Auditor &rarr;
      </Link>
    </div>
  );
}
