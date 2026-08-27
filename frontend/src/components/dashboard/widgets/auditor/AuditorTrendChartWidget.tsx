"use client";

import { lazy, Suspense } from "react";
import { Loader2 } from "lucide-react";
import { useAuditorDashboard } from "@/components/dashboard/auditor/AuditorDashboardContext";

const AreaChart = lazy(() => import("@/components/ui/area-chart").then((mod) => ({ default: mod.AreaChart })));

export function AuditorTrendChartWidget() {
  const { currentYear, isTrendLoading, chartData } = useAuditorDashboard();

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold text-foreground">Tren Inspeksi dan Issue Bulanan ({currentYear})</h2>
      <div className="bg-card rounded-xl p-4 border border-border shadow-sm h-100 flex flex-col">
        <p className="text-xs font-medium text-muted-foreground mb-4">Jumlah inspeksi dan issue dari inspeksi yang Anda lakukan setiap bulan.</p>
        <div className="grow w-full">
          {isTrendLoading ? (
            <div className="w-full h-full flex items-center justify-center">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <Suspense fallback={<div className="w-full h-full flex items-center justify-center"><Loader2 className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
              <AreaChart
                data={chartData}
                xAxisKey="period_label"
                series={[
                  { dataKey: "rate", name: "Total Inspeksi", color: "var(--color-blue-500, #3b82f6)" },
                  { dataKey: "total_issues", name: "Total Issue", color: "#f97316" },
                ]}
              />
            </Suspense>
          )}
        </div>
      </div>
    </div>
  );
}
