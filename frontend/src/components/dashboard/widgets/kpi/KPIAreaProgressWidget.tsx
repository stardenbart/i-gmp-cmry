"use client";

import Link from "next/link";
import { cn } from "@/lib/utils";
import { useKPIDashboard } from "@/components/dashboard/kpi/KPIDashboardContext";
import { VisualizationSwitch } from "@/components/dashboard/kpi/VisualizationSwitch";
import type { VizType } from "@/components/dashboard/types";

export function KPIAreaProgressWidget({ vizType }: { vizType?: VizType }) {
  const { stats, user } = useKPIDashboard();

  if (vizType && vizType !== "list") {
    return (
      <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex h-full flex-col">
        <h3 className="mb-3 text-base font-semibold text-foreground">Progress Kepatuhan per Area/Kawasan</h3>
        <div className="min-h-0 flex-1">
          <VisualizationSwitch
            vizType={vizType}
            data={(stats?.auditee_status || []) as unknown as Record<string, string | number>[]}
            categoryKey="name"
            valueKeys={[{ key: "compliance", name: "Kepatuhan (%)", color: "#10b981" }]}
          />
        </div>
      </div>
    );
  }

  return (
    <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex flex-col justify-between h-full">
      <div>
        <div className="flex justify-between items-center mb-4 pb-3 border-b border-border">
          <h3 className="text-base font-semibold text-foreground">Progress Kepatuhan per Area/Kawasan</h3>
          <span className="bg-primary/10 text-primary text-[10px] font-bold px-2 py-0.5 rounded uppercase tracking-wider">
            Ringkasan Lokasi
          </span>
        </div>
        <div className="space-y-3 max-h-[260px] overflow-y-auto pr-1">
          {!stats?.auditee_status || stats.auditee_status.length === 0 ? (
            <div className="p-6 text-center text-xs text-muted-foreground italic">Belum ada data progress</div>
          ) : (
            stats.auditee_status.map((item, idx) => (
              <div key={idx} className="flex items-center justify-between p-3 bg-muted/40 rounded-lg border border-border/70 hover:bg-muted/70 transition-colors">
                <div className="min-w-0">
                  <p className="text-xs font-semibold text-foreground truncate">{item.name}</p>
                  <p className="text-[11px] text-muted-foreground">{item.type}</p>
                </div>
                <div className="text-right shrink-0">
                  <p className={cn("text-xs font-bold", item.compliance < 70 ? "text-red-500" : "text-green-600")}>
                    {item.compliance}%
                  </p>
                  <p className="text-[10px] text-muted-foreground">{item.open_issues} Temuan Terbuka</p>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
      <Link
        href={`/cimory/${user?.plant_id || "global"}/dashboard/${user?.id}/gmp-data`}
        className="mt-4 text-xs font-semibold text-center text-muted-foreground hover:text-primary transition-colors block"
      >
        Lihat Data GMP &rarr;
      </Link>
    </div>
  );
}
