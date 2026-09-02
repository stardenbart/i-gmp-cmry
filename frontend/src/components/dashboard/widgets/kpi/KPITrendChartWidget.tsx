"use client";

import { Loader2 } from "lucide-react";
import { AreaChart } from "@/components/ui/area-chart";
import { useKPIDashboard } from "@/components/dashboard/kpi/KPIDashboardContext";
import { TrendPeriodFilter } from "@/components/dashboard/TrendPeriodFilter";
import { VisualizationSwitch } from "@/components/dashboard/kpi/VisualizationSwitch";
import type { VizType } from "@/components/dashboard/types";

export function KPITrendChartWidget({ vizType }: { vizType?: VizType }) {
  const {
    trendData,
    isTrendLoading,
    isTrendFetching,
    trendMode,
    setTrendMode,
    trendRange,
    setTrendRange,
    trendGranularity,
    setTrendGranularity,
  } = useKPIDashboard();

  const summary = trendData?.summary || { total_inspections: 0, total_issues: 0, average_compliance: 0 };
  const periodAvgNum = Number(summary.average_compliance || 0);
  const chartFormattedData = (trendData?.points || []).map((item) => ({
    ...item,
    period_label: item.label,
    average_compliance: periodAvgNum,
  }));

  return (
    <div className="space-y-4">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-foreground">Tren Tingkat Kepatuhan</h2>
          <p className="text-xs font-medium text-muted-foreground">
            {trendData?.period_label || "Pergerakan tingkat kepatuhan dan issue berdasarkan periode"}
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2.5">
          <div className="flex items-center gap-2 bg-card border border-border/70 rounded-xl px-3 py-1.5 text-xs shadow-2xs">
            <span className="text-muted-foreground font-medium">Inspeksi Periode:</span>
            <span className="font-bold font-mono text-purple-600 dark:text-purple-400 text-sm">{summary.total_inspections} Inspeksi</span>
          </div>
          <div className="flex items-center gap-2 bg-card border border-border/70 rounded-xl px-3 py-1.5 text-xs shadow-2xs">
            <span className="text-muted-foreground font-medium">Total Issue:</span>
            <span className="font-bold font-mono text-orange-600 dark:text-orange-400 text-sm">{summary.total_issues} Issue Foto</span>
          </div>
          <div className="flex items-center gap-2 bg-card border border-border/70 rounded-xl px-3 py-1.5 text-xs shadow-2xs">
            <span className="text-muted-foreground font-medium">Kepatuhan Periode:</span>
            <span className="font-bold font-mono text-emerald-600 dark:text-emerald-400 text-sm">{periodAvgNum.toFixed(1)}%</span>
          </div>

          <TrendPeriodFilter
            mode={trendMode}
            onModeChange={setTrendMode}
            range={trendRange}
            onRangeChange={setTrendRange}
            granularity={trendGranularity}
            onGranularityChange={setTrendGranularity}
            disabled={isTrendLoading}
          />
        </div>
      </div>

      <div className="bg-card rounded-xl p-4 border border-border shadow-2xs h-100 flex flex-col">
        <div className="grow w-full">
          {isTrendLoading ? (
            <div className="w-full h-full flex items-center justify-center">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : vizType && vizType !== "area" && vizType !== "list" ? (
            // Bar/line/table alt views: dual-axis + per-series units don't
            // translate cleanly to a generic chart, so these show just the
            // compliance rate — the default "area" view below keeps every
            // original series/axis exactly as before.
            <VisualizationSwitch
              vizType={vizType}
              data={chartFormattedData}
              categoryKey="period_label"
              valueKeys={[{ key: "compliance_rate", name: "Tingkat Kepatuhan (%)", color: "#2563eb" }]}
              height={340}
            />
          ) : (
            <div className="relative h-full">
              <AreaChart
                data={chartFormattedData}
                xAxisKey="period_label"
                unit="%"
                series={[
                  { dataKey: "compliance_rate", name: "Tingkat Kepatuhan (%)", color: "var(--color-primary, #2563eb)", unit: "%", yAxisId: "left" },
                  { dataKey: "average_compliance", name: "Rata-Rata Periode (%)", color: "#10b981", unit: "%", yAxisId: "left" },
                  { dataKey: "total_inspections", name: "Total Inspeksi", color: "#a855f7", unit: "", yAxisId: "right" },
                  { dataKey: "total_issues", name: "Issue per Foto", color: "#f97316", unit: "", yAxisId: "right" },
                ]}
              />
              {isTrendFetching && <div className="pointer-events-none absolute inset-0 rounded-lg bg-background/35 backdrop-blur-[1px]" />}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
