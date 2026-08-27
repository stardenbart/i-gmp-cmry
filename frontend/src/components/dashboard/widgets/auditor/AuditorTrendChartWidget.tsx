"use client";

import { lazy, Suspense } from "react";
import { Loader2 } from "lucide-react";
import { useAuditorDashboard } from "@/components/dashboard/auditor/AuditorDashboardContext";
import { TrendPeriodFilter } from "@/components/dashboard/TrendPeriodFilter";

const AreaChart = lazy(() => import("@/components/ui/area-chart").then((mod) => ({ default: mod.AreaChart })));

export function AuditorTrendChartWidget() {
  const {
    trendMode,
    setTrendMode,
    trendRange,
    setTrendRange,
    trendGranularity,
    setTrendGranularity,
    trendData,
    isTrendLoading,
    isTrendFetching,
    chartData,
  } = useAuditorDashboard();
  const summary = trendData?.summary || { total_inspections: 0, total_issues: 0, average_compliance: 0 };
  const formattedData = chartData.map((item) => ({
    ...item,
    average_compliance: summary.average_compliance,
  }));

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h2 className="text-lg font-semibold text-foreground">Tren Inspeksi dan Issue Auditor</h2>
          <p className="text-xs font-medium text-muted-foreground">{trendData?.period_label || "Pilih periode tren"}</p>
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
      <div className="grid grid-cols-3 gap-2 text-xs">
        <div className="rounded-xl border bg-card px-3 py-2"><span className="block text-muted-foreground">Inspeksi</span><strong>{summary.total_inspections}</strong></div>
        <div className="rounded-xl border bg-card px-3 py-2"><span className="block text-muted-foreground">Issue Foto</span><strong>{summary.total_issues}</strong></div>
        <div className="rounded-xl border bg-card px-3 py-2"><span className="block text-muted-foreground">Kepatuhan</span><strong>{Number(summary.average_compliance).toFixed(1)}%</strong></div>
      </div>
      <div className="bg-card rounded-xl p-4 border border-border shadow-sm h-100 flex flex-col">
        <p className="text-xs font-medium text-muted-foreground mb-4">Jumlah inspeksi dan foto temuan dari inspeksi yang Anda lakukan pada periode terpilih.</p>
        <div className="grow w-full">
          {isTrendLoading ? (
            <div className="w-full h-full flex items-center justify-center">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <Suspense fallback={<div className="w-full h-full flex items-center justify-center"><Loader2 className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
              <div className="relative h-full">
                <AreaChart
                  data={formattedData}
                  xAxisKey="period_label"
                  series={[
                    { dataKey: "compliance_rate", name: "Tingkat Kepatuhan (%)", color: "var(--color-primary, #2563eb)", unit: "%", yAxisId: "left" },
                    { dataKey: "average_compliance", name: "Rata-Rata Periode (%)", color: "#10b981", unit: "%", yAxisId: "left" },
                    { dataKey: "total_inspections", name: "Total Inspeksi", color: "#a855f7", yAxisId: "right" },
                    { dataKey: "total_issues", name: "Issue per Foto", color: "#f97316", yAxisId: "right" },
                  ]}
                />
                {isTrendFetching && <div className="pointer-events-none absolute inset-0 rounded-lg bg-background/35 backdrop-blur-[1px]" />}
              </div>
            </Suspense>
          )}
        </div>
      </div>
    </div>
  );
}
