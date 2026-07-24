"use client";

import { useState } from "react";
import Link from "next/link";
import { 
  ClipboardCheck,  
  AlertTriangle, 
  AlertCircle,
  Factory,
  Store,
  Loader2,
  TrendingUp
} from "lucide-react";
import { AreaChart } from "@/components/ui/area-chart";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/axios";
import { areaApi } from "@/types/api/master";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { cn } from "@/lib/utils";
import { StatsCards } from "../admin/StatCard";

// Fetch dashboard stats directly from backend API
const fetchDashboardStats = async (areaId?: string, period: string = "6m") => {
  const res = await api.get("/dashboard/stats", {
    params: { area_id: areaId, period }
  });
  return res.data.data;
};

export const DashboardPanelAdmin = () => {
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);

  const [selectedArea, setSelectedArea] = useState<string>("");
  const [trendPeriod, setTrendPeriod] = useState<"1m" | "3m" | "6m" | "1y">("6m");

  const { data: areasResponse } = useQuery({
    queryKey: ["areas-master-dashboard"],
    queryFn: () => areaApi.getAll(1, 100),
    enabled: mounted,
  });

  const {
    data: stats,
    isLoading,
    isFetching,
  } = useQuery({
    queryKey: ["dashboard-stats-stitch", selectedArea, trendPeriod],
    queryFn: () => fetchDashboardStats(selectedArea, trendPeriod),
    enabled: mounted && !!user,
    staleTime: 10000,
  });

  if (!mounted) {
    return <div className="h-screen w-full bg-background" />;
  }

  return (
    <div className="w-full space-y-8 animate-in fade-in duration-500 pb-24 md:pb-6">
      
      {/* Header Section */}
      <section className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <h1 className="text-3xl md:text-4xl font-bold tracking-tight text-foreground">
              Selamat Datang, {user?.name || "Administrator"}
            </h1>
            <span className="bg-primary/10 text-primary text-xs font-semibold px-2.5 py-1 rounded uppercase tracking-wider">
              ADMINISTRATOR
            </span>
          </div>
          <p className="text-muted-foreground text-sm">
            Pengawasan Berkelanjutan · Ringkasan aktivitas dan performa seluruh peran
          </p>
        </div>
        <div className="flex gap-2 items-center">
          <select 
            value={selectedArea} 
            onChange={(e) => setSelectedArea(e.target.value)}
            className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground"
          >
            <option value="">Semua Area</option>
            {areasResponse?.data?.items?.map((area: any) => (
              <option key={area.area_id} value={area.area_id}>
                {area.area_name}
              </option>
            ))}
          </select>
        </div>
      </section>

      {/* Stat Cards (Bento Grid) */}
      <section className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatsCards stats={stats} isLoading={isLoading || isFetching} />
      </section>

      {/* Monitoring Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        
        {/* Monitoring per Role */}
        <div className="lg:col-span-2 space-y-4">
          <h2 className="text-lg font-semibold text-foreground">Pengawasan per Peran</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            
            {/* Aktivitas Inspeksi */}
            <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex flex-col justify-between">
              <div>
                <div className="flex items-center justify-between mb-4 pb-3 border-b border-border">
                  <h3 className="text-base font-semibold text-foreground">Aktivitas Inspeksi</h3>
                  <span className="text-xs font-semibold px-2 py-0.5 rounded bg-primary/10 text-primary">
                    Total {(stats?.inspections_completed || 0) + (stats?.inspections_running || 0)}
                  </span>
                </div>
                <div className="space-y-4">
                  <div>
                    <div className="flex justify-between text-xs font-medium mb-1.5">
                      <span className="text-muted-foreground">Inspeksi Selesai</span>
                      <span className="font-bold text-foreground">
                        {stats?.inspections_completed || 0} / {((stats?.inspections_completed || 0) + (stats?.inspections_running || 0)) || 1}
                      </span>
                    </div>
                    <div className="w-full bg-muted h-2.5 rounded-full overflow-hidden">
                      <div 
                        className="bg-green-600 h-full rounded-full transition-all duration-1000" 
                        style={{ 
                          width: `${(((stats?.inspections_completed || 0) + (stats?.inspections_running || 0)) > 0)
                            ? ((stats?.inspections_completed || 0) / ((stats?.inspections_completed || 0) + (stats?.inspections_running || 0)) * 100) 
                            : 0}%` 
                        }}
                      ></div>
                    </div>
                  </div>
                  <div>
                    <div className="flex justify-between text-xs font-medium mb-1.5">
                      <span className="text-muted-foreground">Inspeksi Berlangsung</span>
                      <span className="font-bold text-foreground">
                        {stats?.inspections_running || 0} / {((stats?.inspections_completed || 0) + (stats?.inspections_running || 0)) || 1}
                      </span>
                    </div>
                    <div className="w-full bg-muted h-2.5 rounded-full overflow-hidden">
                      <div 
                        className="bg-primary h-full rounded-full transition-all duration-1000" 
                        style={{ 
                          width: `${(((stats?.inspections_completed || 0) + (stats?.inspections_running || 0)) > 0)
                            ? ((stats?.inspections_running || 0) / ((stats?.inspections_completed || 0) + (stats?.inspections_running || 0)) * 100) 
                            : 0}%` 
                        }}
                      ></div>
                    </div>
                  </div>
                </div>
              </div>
              <Link 
                href={`/cimory/dashboard/${user?.id}/monitoring`}
                className="mt-6 text-xs font-semibold text-center text-muted-foreground hover:text-primary transition-colors block"
              >
                Lihat Detail Auditor &rarr;
              </Link>
            </div>

            {/* Aktivitas Penanggung Jawab (PIC) */}
            <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex flex-col justify-between">
              <div>
                <div className="flex items-center justify-between mb-4 pb-3 border-b border-border">
                  <h3 className="text-base font-semibold text-foreground">Aktivitas Penanggung Jawab (PIC)</h3>
                  <span className="text-xs font-semibold px-2 py-0.5 rounded bg-primary/10 text-primary">
                    Total {stats?.total_open_issues || 0} Temuan
                  </span>
                </div>
                <div className="space-y-4">
                  <div>
                    <div className="flex justify-between text-xs font-medium mb-1.5">
                      <span className="text-muted-foreground">Temuan Berhasil Diselesaikan</span>
                      <span className="font-bold text-foreground">
                        {((stats as any)?.issues_resolved || 0)} / {(stats?.total_open_issues || 0) + ((stats as any)?.issues_resolved || 0) || 1}
                      </span>
                    </div>
                    <div className="w-full bg-muted h-2.5 rounded-full overflow-hidden">
                      <div 
                        className="bg-green-600 h-full rounded-full transition-all duration-1000" 
                        style={{ 
                          width: `${(((stats?.total_open_issues || 0) + ((stats as any)?.issues_resolved || 0)) > 0)
                            ? (((stats as any)?.issues_resolved || 0) / ((stats?.total_open_issues || 0) + ((stats as any)?.issues_resolved || 0)) * 100)
                            : 0}%` 
                        }}
                      ></div>
                    </div>
                  </div>
                  <div>
                    <div className="flex justify-between text-xs font-medium mb-1.5">
                      <span className="text-muted-foreground">Temuan Jatuh Tempo (Overdue)</span>
                      <span className="font-bold text-red-600">
                        {stats?.issue_overdue || 0} / {(stats?.total_open_issues || 1)}
                      </span>
                    </div>
                    <div className="w-full bg-muted h-2.5 rounded-full overflow-hidden">
                      <div 
                        className="bg-red-500 h-full rounded-full transition-all duration-1000" 
                        style={{ 
                          width: `${((stats?.total_open_issues || 0) > 0)
                            ? ((stats?.issue_overdue || 0) / (stats?.total_open_issues || 1) * 100)
                            : 0}%` 
                        }}
                      ></div>
                    </div>
                  </div>
                </div>
              </div>
              <Link 
                href={`/cimory/dashboard/${user?.id}/monitoring`}
                className="mt-6 text-xs font-semibold text-center text-muted-foreground hover:text-primary transition-colors block"
              >
                Lihat Detail PIC &rarr;
              </Link>
            </div>

          </div>
        </div>

        {/* Status Lokasi Auditee */}
        <div className="space-y-4">
          <h2 className="text-lg font-semibold text-foreground">Status Auditee</h2>
          <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex flex-col justify-between">
            <div>
              <div className="flex justify-between items-center mb-4 pb-3 border-b border-border">
                <h3 className="text-base font-semibold text-foreground">Status Lokasi Auditee</h3>
                <span className="bg-primary/10 text-primary text-[10px] font-bold px-2 py-0.5 rounded uppercase tracking-wider">
                  Ringkasan Area
                </span>
              </div>
              <div className="space-y-3 max-h-[220px] overflow-y-auto pr-1">
                {(!stats?.auditee_status || stats?.auditee_status?.length === 0) ? (
                  <div className="p-6 text-center text-xs text-muted-foreground italic">Belum ada data status auditee</div>
                ) : (
                  stats?.auditee_status?.map((item: any, idx: number) => {
                    return (
                      <div key={idx} className="flex items-center justify-between p-3 bg-muted/40 rounded-lg border border-border/70 hover:bg-muted/70 transition-colors">
                        <div>
                          <p className="text-xs font-semibold text-foreground">{item.name}</p>
                          <p className="text-[11px] text-muted-foreground">{item.type}</p>
                        </div>
                        <div className="text-right">
                          <p className={cn("text-xs font-bold", item.compliance < 70 ? "text-red-500" : "text-green-600")}>
                            {item.compliance}%
                          </p>
                          <p className="text-[10px] text-muted-foreground">{item.open_issues} Temuan Terbuka</p>
                        </div>
                      </div>
                    );
                  })
                )}
              </div>
            </div>
            <Link
              href={`/cimory/dashboard/${user?.id}/gmp-data`}
              className="mt-4 text-xs font-semibold text-center text-muted-foreground hover:text-primary transition-colors block"
            >
              Lihat Data Auditee GMP &rarr;
            </Link>
          </div>
        </div>

      </div>

      {/* Chart Section */}
      <div className="space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
          <div>
            <h2 className="text-lg font-semibold text-foreground">Tren Tingkat Kepatuhan</h2>
            <p className="text-xs font-medium text-muted-foreground">
              Pergerakan tingkat kepatuhan berdasarkan rentang waktu yang dipilih
            </p>
          </div>
          
          {/* Period Selector Buttons */}
          <div className="inline-flex items-center p-1 bg-muted rounded-lg border border-border/60 self-start sm:self-auto">
            {(
              [
                { id: "1m", label: "1 Bulan" },
                { id: "3m", label: "3 Bulan" },
                { id: "6m", label: "6 Bulan" },
                { id: "1y", label: "1 Tahun" },
              ] as const
            ).map((p) => (
              <button
                key={p.id}
                onClick={() => setTrendPeriod(p.id)}
                className={cn(
                  "px-3 py-1 text-xs font-semibold rounded-md transition-all",
                  trendPeriod === p.id
                    ? "bg-background text-foreground shadow-sm border border-border/80"
                    : "text-muted-foreground hover:text-foreground"
                )}
              >
                {p.label}
              </button>
            ))}
          </div>
        </div>

        <div className="bg-card rounded-xl p-4 border border-border shadow-sm h-100 flex flex-col">
          <div className="grow w-full">
            {isLoading || isFetching ? (
              <div className="w-full h-full flex items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              </div>
            ) : (
              <AreaChart
                data={stats?.compliance_trend || []}
                xAxisKey="month"
                series={[
                  { dataKey: "rate", name: "Tingkat Kepatuhan (%)", color: "var(--color-primary)" }
                ]}
              />
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
