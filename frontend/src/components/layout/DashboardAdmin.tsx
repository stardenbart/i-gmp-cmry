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

import { usePolling } from "@/hooks/usePolling";
import { plantApi } from "@/types/api/master";

import { useParams } from "next/navigation";

// Fetch dashboard stats directly from backend API
const fetchDashboardStats = async (areaId?: string, period: string = "6m", plantId?: string) => {
  const res = await api.get("/dashboard/stats", {
    params: { area_id: areaId, period, plant_id: plantId }
  });
  return res.data.data;
};

export const DashboardPanelAdmin = () => {
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);
  const params = useParams();
  const urlPlantCode = (params?.plantCode as string) || "";
  const urlPlant = (urlPlantCode && urlPlantCode !== "all" && urlPlantCode !== "global") ? urlPlantCode : "";

  usePolling();

  const [selectedPlant, setSelectedPlant] = useState<string>("");
  const [selectedArea, setSelectedArea] = useState<string>("");
  const [trendPeriod, setTrendPeriod] = useState<"1m" | "3m" | "6m" | "1y">("6m");

  const isSuperAdmin =
    user?.role_id === "ROLE-000" ||
    user?.role_id === "SUPERADMIN" ||
    user?.role?.role_name === "Super Admin" ||
    !user?.plant_id;

  const effectivePlant = isSuperAdmin
    ? (selectedPlant === "all" ? "" : selectedPlant)
    : (user?.plant_id || urlPlant);

  const { data: plantsResponse } = useQuery({
    queryKey: ["plants-master-dashboard"],
    queryFn: () => plantApi.getAll(1, 100),
    enabled: mounted && isSuperAdmin,
  });

  const { data: areasResponse } = useQuery({
    queryKey: ["areas-master-dashboard", effectivePlant],
    queryFn: () => areaApi.getAll(1, 100),
    enabled: mounted,
  });

  // Filter area items based on plant
  const filteredAreas = areasResponse?.data?.items?.filter((area: any) => {
    if (!effectivePlant) return true;
    return area.plant_id === effectivePlant;
  }) || [];

  const {
    data: stats,
    isLoading,
    isFetching,
  } = useQuery({
    queryKey: ["dashboard-stats-stitch", selectedArea, trendPeriod, effectivePlant],
    queryFn: () => fetchDashboardStats(selectedArea, trendPeriod, effectivePlant),
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
              {isSuperAdmin ? "SUPER ADMIN" : "ADMINISTRATOR"}
            </span>
          </div>
          <p className="text-muted-foreground text-sm">
            Pengawasan Berkelanjutan · Ringkasan aktivitas dan performa seluruh peran
          </p>
        </div>
        <div className="flex flex-wrap gap-2 items-center">
          {/* Plant Selector Dropdown for SuperAdmin */}
          {isSuperAdmin ? (
            <select
              value={selectedPlant || "all"}
              onChange={(e) => {
                setSelectedPlant(e.target.value);
                setSelectedArea("");
              }}
              className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground font-medium focus:ring-2 focus:ring-primary/40 focus:outline-none"
            >
              <option value="all">Semua Plant</option>
              {plantsResponse?.data?.items?.map((plant: any) => (
                <option key={plant.plant_id} value={plant.plant_id}>
                  {plant.plant_name}
                </option>
              ))}
            </select>
          ) : user?.plant_id ? (
            <div className="px-3 py-1.5 border border-primary/30 bg-primary/5 text-primary rounded-lg text-xs font-semibold uppercase tracking-wider flex items-center gap-1.5">
              <Factory className="h-3.5 w-3.5" />
              Plant: {user.plant_id}
            </div>
          ) : null}

          {/* Area Selector Dropdown */}
          <select 
            value={selectedArea} 
            onChange={(e) => setSelectedArea(e.target.value)}
            className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground font-medium focus:ring-2 focus:ring-primary/40 focus:outline-none"
          >
            <option value="">Semua Area</option>
            {filteredAreas.map((area: any) => (
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
                href={`/cimory/${user?.plant_id || 'global'}/dashboard/${user?.id}/monitoring`}
                className="mt-6 text-xs font-semibold text-center text-muted-foreground hover:text-primary transition-colors block"
              >
                Lihat Detail Auditor &rarr;
              </Link>
            </div>

            {/* Aktivitas Penanggung Jawab (PIC) */}
            {(() => {
              const resolvedCount = (stats as any)?.issues_resolved ?? (stats as any)?.pic_followup_completed ?? 0;
              const openCount = stats?.total_open_issues || 0;
              const totalCount = openCount + resolvedCount;
              const resolvedRate = totalCount > 0 ? Math.round((resolvedCount / totalCount) * 100) : 0;
              const overdueCount = stats?.issue_overdue ?? (stats as any)?.pic_followup_overdue ?? 0;
              const overdueRate = totalCount > 0 ? Math.round((overdueCount / totalCount) * 100) : 0;

              return (
                <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex flex-col justify-between">
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
                          <div 
                            className="bg-green-600 h-full rounded-full transition-all duration-1000" 
                            style={{ width: `${resolvedRate}%` }}
                          ></div>
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
                          <div 
                            className="bg-red-500 h-full rounded-full transition-all duration-1000" 
                            style={{ width: `${overdueRate}%` }}
                          ></div>
                        </div>
                      </div>
                    </div>
                  </div>
                  <Link 
                    href={`/cimory/${user?.plant_id || 'global'}/dashboard/${user?.id}/monitoring`}
                    className="mt-6 text-xs font-semibold text-center text-muted-foreground hover:text-primary transition-colors block"
                  >
                    Lihat Detail PIC &rarr;
                  </Link>
                </div>
              );
            })()}

            {/* Status Maintenance WO / WR */}
            <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex flex-col justify-between col-span-1 md:col-span-2">
              <div>
                <div className="flex items-center justify-between mb-4 pb-3 border-b border-border">
                  <h3 className="text-base font-semibold text-foreground">Status Maintenance (WO / WR)</h3>
                  <span className="text-xs font-semibold px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-600 font-mono">
                    Tingkat Verifikasi: {stats?.wowr_verified_rate || 0}%
                  </span>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <div className="flex justify-between text-xs font-medium mb-1.5">
                      <span className="text-muted-foreground">WO/WR Disetujui</span>
                      <span className="font-bold text-emerald-600">
                        {stats?.wowr_verified || 0} / {stats?.wowr_total || 0}
                      </span>
                    </div>
                    <div className="w-full bg-muted h-2.5 rounded-full overflow-hidden">
                      <div 
                        className="bg-emerald-500 h-full rounded-full transition-all duration-1000" 
                        style={{ width: `${stats?.wowr_verified_rate || 0}%` }}
                      ></div>
                    </div>
                  </div>
                  <div className="flex items-center justify-between text-xs bg-muted/30 p-2.5 rounded-lg border border-border/50">
                    <span className="text-purple-600 font-semibold">Menunggu: {stats?.wowr_pending || 0}</span>
                    <span className="text-red-600 font-semibold">Ditolak: {stats?.wowr_rejected || 0}</span>
                    <span className="text-muted-foreground font-semibold">Belum Bukti: {stats?.wowr_awaiting || 0}</span>
                  </div>
                </div>
              </div>
              <Link 
                href={`/cimory/${user?.plant_id || 'global'}/dashboard/${user?.id}/monitoring/wowr`}
                className="mt-4 text-xs font-semibold text-center text-primary hover:underline block"
              >
                Lihat Laporan & Analytics WO/WR &rarr;
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
              href={`/cimory/${user?.plant_id || 'global'}/dashboard/${user?.id}/gmp-data`}
              className="mt-4 text-xs font-semibold text-center text-muted-foreground hover:text-primary transition-colors block"
            >
              Lihat Data Auditee GMP &rarr;
            </Link>
          </div>
        </div>

      </div>

      {/* Chart Section */}
      {(() => {
        const rawTrend = stats?.compliance_trend || [];
        const periodAvgNum = rawTrend.length > 0
          ? Number((rawTrend.reduce((acc: number, curr: any) => acc + (Number(curr.rate) || 0), 0) / rawTrend.length).toFixed(1))
          : 0;
        const currentRateNum = stats?.compliance_rate != null
          ? Number(Number(stats.compliance_rate).toFixed(1))
          : (rawTrend.length > 0 ? Number(Number(rawTrend[rawTrend.length - 1].rate).toFixed(1)) : 0);

        const periodAvg = periodAvgNum.toFixed(1);
        const currentRate = currentRateNum.toFixed(1);

        const totalInspectionsCount = (stats?.inspections_completed || 0) + (stats?.inspections_running || 0);

        const chartFormattedData = rawTrend.map((item: any) => ({
          ...item,
          avg: periodAvgNum,
          totalCount: totalInspectionsCount
        }));

        return (
          <div className="space-y-4">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-3">
              <div>
                <h2 className="text-lg font-semibold text-foreground">Tren Tingkat Kepatuhan</h2>
                <p className="text-xs font-medium text-muted-foreground">
                  Pergerakan tingkat kepatuhan berdasarkan rentang waktu yang dipilih
                </p>
              </div>
              
              <div className="flex flex-wrap items-center gap-2.5">
                {/* Overall Summary Badges */}
                <div className="flex items-center gap-2 bg-card border border-border/70 rounded-xl px-3 py-1.5 text-xs shadow-2xs">
                  <span className="text-muted-foreground font-medium">Total Semua:</span>
                  <span className="font-bold font-mono text-purple-600 dark:text-purple-400 text-sm">{totalInspectionsCount} Inspeksi</span>
                </div>
                <div className="flex items-center gap-2 bg-card border border-border/70 rounded-xl px-3 py-1.5 text-xs shadow-2xs">
                  <span className="text-muted-foreground font-medium">Total Kepatuhan Keseluruhan:</span>
                  <span className="font-bold font-mono text-emerald-600 dark:text-emerald-400 text-sm">{currentRate}%</span>
                </div>
                <div className="flex items-center gap-2 bg-card border border-border/70 rounded-xl px-3 py-1.5 text-xs shadow-2xs">
                  <span className="text-muted-foreground font-medium">Rata-Rata Periode:</span>
                  <span className="font-bold font-mono text-primary text-sm">{periodAvg}%</span>
                </div>

                {/* Period Selector Buttons */}
                <div className="inline-flex items-center p-1 bg-muted rounded-lg border border-border/60">
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
                          ? "bg-background text-foreground shadow-2xs border border-border/80"
                          : "text-muted-foreground hover:text-foreground"
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
                    xAxisKey="month"
                    unit="%"
                    series={[
                      { dataKey: "rate", name: "Tingkat Kepatuhan (%)", color: "var(--color-primary, #2563eb)", unit: "%" },
                      { dataKey: "avg", name: "Rata-Rata Periode (%)", color: "#10b981", unit: "%" },
                      { dataKey: "totalCount", name: "Total Semua Inspeksi", color: "#a855f7", unit: "" },
                    ]}
                  />
                )}
              </div>
            </div>
          </div>
        );
      })()}
    </div>
  );
};
