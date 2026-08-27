"use client";

import { Loader2 } from "lucide-react";
import { AreaChart } from "@/components/ui/area-chart";
import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { id as idLocale } from "date-fns/locale";
import { useAdminDashboard, type TrendPeriod } from "@/components/dashboard/admin/AdminDashboardContext";

const PERIOD_OPTIONS: { id: TrendPeriod; label: string }[] = [
  { id: "1m", label: "1 Bulan" },
  { id: "3m", label: "3 Bulan" },
  { id: "6m", label: "6 Bulan" },
  { id: "quarter", label: "Per Quarter" },
  { id: "1y", label: "1 Tahun" },
];

export function ChartTrenKepatuhanWidget() {
  const { stats, isLoading, isFetching, trendPeriod, setTrendPeriod } = useAdminDashboard();

  const rawTrend = stats?.compliance_trend || [];
  const periodAvgNum =
    rawTrend.length > 0
      ? Number((rawTrend.reduce((acc: number, curr) => acc + (Number(curr.rate) || 0), 0) / rawTrend.length).toFixed(1))
      : 0;
  const currentRateNum =
    stats?.compliance_rate != null
      ? Number(Number(stats.compliance_rate).toFixed(1))
      : rawTrend.length > 0
        ? Number(Number(rawTrend[rawTrend.length - 1].rate).toFixed(1))
        : 0;

  const periodAvg = periodAvgNum.toFixed(1);
  const currentRate = currentRateNum.toFixed(1);
  const totalInspectionsCount = (stats?.inspections_completed || 0) + (stats?.inspections_running || 0);

  const chartFormattedData = rawTrend.map((item) => {
    const startDate = new Date(`${item.start_date}T00:00:00`);
    const endDate = new Date(`${item.end_date}T00:00:00`);
    let periodLabel = format(startDate, "MMM yyyy", { locale: idLocale });
    if (trendPeriod === "1m") {
      periodLabel = `${format(startDate, "d MMM", { locale: idLocale })}–${format(endDate, "d MMM yyyy", { locale: idLocale })}`;
    } else if (trendPeriod === "quarter") {
      const quarterNumber = Math.floor(startDate.getMonth() / 3) + 1;
      periodLabel = `Q${quarterNumber} ${format(startDate, "yyyy")} (${format(startDate, "MMM", { locale: idLocale })}–${format(endDate, "MMM", { locale: idLocale })})`;
    }

    return {
      ...item,
      period_label: periodLabel,
      avg: periodAvgNum,
      totalCount: totalInspectionsCount,
    };
  });

  return (
    <div className="space-y-4">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-foreground">Tren Tingkat Kepatuhan</h2>
          <p className="text-xs font-medium text-muted-foreground">
            Pergerakan tingkat kepatuhan dan jumlah issue berdasarkan rentang waktu yang dipilih
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2.5">
          <div className="flex items-center gap-2 bg-card border border-border/70 rounded-xl px-3 py-1.5 text-xs shadow-2xs">
            <span className="text-muted-foreground font-medium">Total Semua:</span>
            <span className="font-bold font-mono text-purple-600 dark:text-purple-400 text-sm">{totalInspectionsCount} Inspeksi</span>
          </div>
          <div className="flex items-center gap-2 bg-card border border-border/70 rounded-xl px-3 py-1.5 text-xs shadow-2xs">
            <span className="text-muted-foreground font-medium">Total Issue:</span>
            <span className="font-bold font-mono text-orange-600 dark:text-orange-400 text-sm">{stats?.total_issues || 0} Issue</span>
          </div>
          <div className="flex items-center gap-2 bg-card border border-border/70 rounded-xl px-3 py-1.5 text-xs shadow-2xs">
            <span className="text-muted-foreground font-medium">Total Kepatuhan Keseluruhan:</span>
            <span className="font-bold font-mono text-emerald-600 dark:text-emerald-400 text-sm">{currentRate}%</span>
          </div>
          <div className="flex items-center gap-2 bg-card border border-border/70 rounded-xl px-3 py-1.5 text-xs shadow-2xs">
            <span className="text-muted-foreground font-medium">Rata-Rata Periode:</span>
            <span className="font-bold font-mono text-primary text-sm">{periodAvg}%</span>
          </div>

          <div className="inline-flex items-center p-1 bg-muted rounded-lg border border-border/60">
            {PERIOD_OPTIONS.map((p) => (
              <button
                key={p.id}
                onClick={() => setTrendPeriod(p.id)}
                className={cn(
                  "px-3 py-1 text-xs font-semibold rounded-md transition-all",
                  trendPeriod === p.id ? "bg-background text-foreground shadow-2xs border border-border/80" : "text-muted-foreground hover:text-foreground"
                )}
              >
                {p.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="bg-card rounded-xl p-4 border border-border shadow-2xs h-100 flex flex-col">
        <div className="grow w-full">
          {isLoading || isFetching ? (
            <div className="w-full h-full flex items-center justify-center">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <AreaChart
              data={chartFormattedData}
              xAxisKey="period_label"
              unit="%"
              series={[
                { dataKey: "rate", name: "Tingkat Kepatuhan (%)", color: "var(--color-primary, #2563eb)", unit: "%", yAxisId: "left" },
                { dataKey: "avg", name: "Rata-Rata Periode (%)", color: "#10b981", unit: "%", yAxisId: "left" },
                { dataKey: "totalCount", name: "Total Semua Inspeksi", color: "#a855f7", unit: "", yAxisId: "right" },
                { dataKey: "total_issues", name: "Total Issue", color: "#f97316", unit: "", yAxisId: "right" },
              ]}
            />
          )}
        </div>
      </div>
    </div>
  );
}
