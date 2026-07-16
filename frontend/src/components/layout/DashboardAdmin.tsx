"use client";

import { useState, useEffect, useCallback } from "react";
import { 
  ClipboardCheck, 
  CheckCircle, 
  AlertTriangle, 
  AlertCircle,
  Factory,
  Store,
  RefreshCw,
  Loader2,
  TrendingUp,
  Download
} from "lucide-react";
import { AreaChart } from "@/components/ui/area-chart";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api/axios";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { cn } from "@/lib/utils";

// Fetch dashboard stats (mocked/adapted for Stitch design)
const fetchDashboardStats = async () => {
  try {
    const res = await api.get("/dashboard/stats");
    return res.data.data;
  } catch {
    // Return mock data aligned with Stitch design
    return {
      total_inspections_running: 124,
      compliance_rate: 89.2,
      compliance_rate_trend: 2.4,
      total_open_issues: 45,
      issue_overdue: 12,
      issue_overdue_trend: 5,
      inspections_completed: 78,
      inspections_running: 46,
      pic_followup_completed: 234,
      pic_followup_overdue: 12,
      auditee_status: [
        { name: "Gudang Utama Jakarta", type: "Area Operasional", compliance: 68, open_issues: 8, icon: Store },
        { name: "Pabrik Perakitan Cikarang", type: "Area Produksi", compliance: 72, open_issues: 5, icon: Factory }
      ],
      compliance_trend: [
        { month: "Jan", rate: 75 },
        { month: "Feb", rate: 78 },
        { month: "Mar", rate: 82 },
        { month: "Apr", rate: 80 },
        { month: "May", rate: 85 },
        { month: "Jun", rate: 89.2 },
      ]
    };
  }
};

export const DashboardPanelAdmin = () => {
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();

  const {
    data: stats,
    isLoading,
    isFetching,
    dataUpdatedAt,
  } = useQuery({
    queryKey: ["dashboard-stats-stitch"],
    queryFn: fetchDashboardStats,
    enabled: mounted && !!user,
    staleTime: 10000,
  });

  const handleRefresh = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ["dashboard-stats-stitch"] });
  }, [queryClient]);

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
              Halo, {user?.name || "Admin"}
            </h1>
            <span className="bg-primary/10 text-primary text-xs font-semibold px-2 py-1 rounded uppercase tracking-wider">
              ADMIN
            </span>
          </div>
          <p className="text-muted-foreground">
            Monitoring Overview · Ringkasan aktivitas seluruh role
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleRefresh}
            disabled={isFetching}
            className="flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground hover:bg-muted rounded-lg transition-colors disabled:opacity-50"
          >
            <RefreshCw className={cn("h-4 w-4", isFetching && "animate-spin")} />
          </button>
          <button className="bg-background border border-border text-primary text-sm font-medium px-4 py-2 rounded-lg flex items-center gap-2 hover:bg-muted transition-colors shadow-sm">
            <Download className="h-4 w-4" /> Export Report
          </button>
        </div>
      </section>

      {/* Stat Cards (Bento Grid) */}
      <section className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {/* Card 1 */}
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden group">
          <div className="absolute top-0 left-0 w-full h-1 bg-primary"></div>
          <div className="flex justify-between items-start mb-4">
            <div className="p-2 bg-primary/10 text-primary rounded-lg">
              <ClipboardCheck className="h-5 w-5" />
            </div>
            <span className="text-xs font-semibold text-muted-foreground">Live</span>
          </div>
          <h3 className="text-xs font-medium text-muted-foreground mb-1 uppercase tracking-wider">Total Audit Berjalan</h3>
          <p className="text-2xl font-bold text-foreground">
            {isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : stats?.total_inspections_running}
          </p>
        </div>

        {/* Card 2 */}
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
          <div className="absolute top-0 left-0 w-full h-1 bg-green-500"></div>
          <div className="flex justify-between items-start mb-4">
            <div className="p-2 bg-green-500/10 text-green-500 rounded-lg">
              <CheckCircle className="h-5 w-5" />
            </div>
            <span className="text-xs font-semibold text-green-500 flex items-center gap-1">
              <TrendingUp className="h-3 w-3" /> {stats?.compliance_rate_trend}%
            </span>
          </div>
          <h3 className="text-xs font-medium text-muted-foreground mb-1 uppercase tracking-wider">Compliance Rate</h3>
          <p className="text-2xl font-bold text-foreground">
            {isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : `${stats?.compliance_rate}%`}
          </p>
        </div>

        {/* Card 3 */}
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
          <div className="absolute top-0 left-0 w-full h-1 bg-orange-500"></div>
          <div className="flex justify-between items-start mb-4">
            <div className="p-2 bg-orange-500/10 text-orange-500 rounded-lg">
              <AlertTriangle className="h-5 w-5" />
            </div>
            <span className="text-xs font-semibold text-muted-foreground">Action Req.</span>
          </div>
          <h3 className="text-xs font-medium text-muted-foreground mb-1 uppercase tracking-wider">Total Issue Terbuka</h3>
          <p className="text-2xl font-bold text-foreground">
            {isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : stats?.total_open_issues}
          </p>
        </div>

        {/* Card 4 */}
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
          <div className="absolute top-0 left-0 w-full h-1 bg-red-500"></div>
          <div className="flex justify-between items-start mb-4">
            <div className="p-2 bg-red-500/10 text-red-500 rounded-lg">
              <AlertCircle className="h-5 w-5" />
            </div>
            <span className="text-xs font-semibold text-red-500 flex items-center gap-1">
              <TrendingUp className="h-3 w-3" /> {stats?.issue_overdue_trend}
            </span>
          </div>
          <h3 className="text-xs font-medium text-muted-foreground mb-1 uppercase tracking-wider">Issue Overdue</h3>
          <p className="text-2xl font-bold text-foreground">
            {isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : stats?.issue_overdue}
          </p>
        </div>
      </section>

      {/* Monitoring Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        
        {/* Monitoring per Role */}
        <div className="lg:col-span-2 space-y-4">
          <h2 className="text-lg font-semibold text-foreground">Monitoring per Role</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            
            {/* Aktivitas Audit */}
            <div className="bg-card rounded-xl p-4 border border-border shadow-sm">
              <div className="flex items-center gap-2 mb-4 pb-4 border-b border-border">
                <ClipboardCheck className="text-primary h-6 w-6" />
                <h3 className="text-lg font-semibold text-foreground">Aktivitas Audit</h3>
              </div>
              <div className="space-y-4">
                <div>
                  <div className="flex justify-between text-xs font-medium mb-1">
                    <span className="text-muted-foreground">Inspeksi Selesai</span>
                    <span className="font-bold text-foreground">{stats?.inspections_completed}/{stats?.total_inspections_running}</span>
                  </div>
                  <div className="w-full bg-muted h-2 rounded-full overflow-hidden">
                    <div className="bg-green-500 h-full rounded-full transition-all duration-1000" style={{ width: `${(stats?.inspections_completed || 0) / (stats?.total_inspections_running || 1) * 100}%` }}></div>
                  </div>
                </div>
                <div>
                  <div className="flex justify-between text-xs font-medium mb-1">
                    <span className="text-muted-foreground">Inspeksi Berjalan</span>
                    <span className="font-bold text-foreground">{stats?.inspections_running}/{stats?.total_inspections_running}</span>
                  </div>
                  <div className="w-full bg-muted h-2 rounded-full overflow-hidden">
                    <div className="bg-primary h-full rounded-full transition-all duration-1000" style={{ width: `${(stats?.inspections_running || 0) / (stats?.total_inspections_running || 1) * 100}%` }}></div>
                  </div>
                </div>
              </div>
              <button className="mt-4 text-primary text-sm font-medium w-full py-2 hover:bg-primary/10 rounded transition-colors">Lihat Detail Auditor</button>
            </div>

            {/* Aktivitas PIC */}
            <div className="bg-card rounded-xl p-4 border border-border shadow-sm">
              <div className="flex items-center gap-2 mb-4 pb-4 border-b border-border">
                <AlertTriangle className="text-orange-500 h-6 w-6" />
                <h3 className="text-lg font-semibold text-foreground">Aktivitas PIC</h3>
              </div>
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-2 h-8 bg-green-500 rounded-full"></div>
                    <div>
                      <p className="text-xs font-medium text-muted-foreground">Follow-up Selesai</p>
                      <p className="text-lg font-semibold text-foreground">{stats?.pic_followup_completed}</p>
                    </div>
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-2 h-8 bg-red-500 rounded-full"></div>
                    <div>
                      <p className="text-xs font-medium text-muted-foreground">Follow-up Overdue</p>
                      <p className="text-lg font-semibold text-foreground">{stats?.pic_followup_overdue}</p>
                    </div>
                  </div>
                </div>
              </div>
              <button className="mt-4 text-primary text-sm font-medium w-full py-2 hover:bg-primary/10 rounded transition-colors">Lihat Detail PIC</button>
            </div>
          </div>
        </div>

        {/* Status Auditee */}
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm">
          <div className="flex justify-between items-center mb-4">
            <div className="flex items-center gap-2">
              <AlertCircle className="text-red-500 h-6 w-6" />
              <h3 className="text-lg font-semibold text-foreground">Status Auditee</h3>
            </div>
            <span className="bg-red-500/10 text-red-500 text-[10px] font-bold px-2 py-1 rounded-sm uppercase tracking-wider">
              Compliance &lt; 75%
            </span>
          </div>
          <div className="space-y-3">
            {stats?.auditee_status?.map((item: any, idx: number) => {
              const Icon = item.icon;
              return (
                <div key={idx} className="flex items-center justify-between p-3 bg-muted/50 rounded-lg border border-border hover:bg-muted transition-colors">
                  <div className="flex items-center gap-3">
                    <Icon className="text-muted-foreground h-5 w-5" />
                    <div>
                      <p className="text-sm font-semibold text-foreground">{item.name}</p>
                      <p className="text-xs font-medium text-muted-foreground">{item.type}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className={cn("text-sm font-bold", item.compliance < 70 ? "text-red-500" : "text-orange-500")}>
                      {item.compliance}%
                    </p>
                    <p className="text-[10px] font-medium text-muted-foreground">{item.open_issues} Issue Terbuka</p>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      </div>

      {/* Chart Section */}
      <div className="space-y-4">
        <h2 className="text-lg font-semibold text-foreground">Tren Compliance Rate</h2>
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm h-100 flex flex-col">
          <p className="text-xs font-medium text-muted-foreground mb-4">Pergerakan 6 Bulan Terakhir</p>
          <div className="grow w-full">
            {isLoading ? (
              <div className="w-full h-full flex items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              </div>
            ) : (
              <AreaChart
                data={stats?.compliance_trend || []}
                xAxisKey="month"
                series={[
                  { dataKey: "rate", name: "Compliance Rate", color: "var(--color-primary)" }
                ]}
              />
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
