"use client";

import Link from "next/link";
import { useKPIDashboard } from "@/components/dashboard/kpi/KPIDashboardContext";
import { VisualizationSwitch } from "@/components/dashboard/kpi/VisualizationSwitch";
import type { VizType } from "@/components/dashboard/types";

export function KPIWOWRByAreaWidget({ vizType }: { vizType?: VizType }) {
  const { wowrReport, user } = useKPIDashboard();
  const byArea = wowrReport?.by_area || [];

  if (vizType && vizType !== "list") {
    return (
      <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex h-full flex-col">
        <h3 className="mb-3 text-base font-semibold text-foreground">Status WO/WR per Area</h3>
        <div className="min-h-0 flex-1">
          <VisualizationSwitch
            vizType={vizType}
            data={byArea as unknown as Record<string, string | number>[]}
            categoryKey="area_name"
            valueKeys={[
              { key: "verified", name: "Verified", color: "#10b981" },
              { key: "pending", name: "Pending", color: "#a855f7" },
              { key: "rejected", name: "Rejected", color: "#ef4444" },
              { key: "awaiting", name: "Belum Bukti", color: "#71717a" },
            ]}
          />
        </div>
      </div>
    );
  }

  return (
    <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex flex-col justify-between h-full">
      <div>
        <div className="flex items-center justify-between mb-4 pb-3 border-b border-border">
          <h3 className="text-base font-semibold text-foreground">Status WO/WR per Area</h3>
          <span className="text-xs font-semibold px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-600 font-mono">
            Total: {wowrReport?.summary.total || 0}
          </span>
        </div>
        <div className="space-y-2.5 max-h-[260px] overflow-y-auto pr-1">
          {byArea.length === 0 ? (
            <div className="p-6 text-center text-xs text-muted-foreground italic">Belum ada data WO/WR</div>
          ) : (
            byArea.map((area) => (
              <div key={area.area_name} className="p-3 bg-muted/40 rounded-lg border border-border/70">
                <div className="flex items-center justify-between mb-1.5">
                  <p className="text-xs font-semibold text-foreground truncate">{area.area_name}</p>
                  <p className="text-[10px] text-muted-foreground">{area.total} item</p>
                </div>
                <div className="flex flex-wrap gap-x-3 gap-y-0.5 text-[10px] font-semibold">
                  <span className="text-emerald-600">Verified: {area.verified}</span>
                  <span className="text-purple-600">Pending: {area.pending}</span>
                  <span className="text-red-600">Rejected: {area.rejected}</span>
                  <span className="text-muted-foreground">Belum Bukti: {area.awaiting}</span>
                </div>
              </div>
            ))
          )}
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
