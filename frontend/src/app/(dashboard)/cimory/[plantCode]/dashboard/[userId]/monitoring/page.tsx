"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/axios";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { cn } from "@/lib/utils";
import { Loader2, UserCheck, ShieldCheck, Building2 } from "lucide-react";

import { useParams } from "next/navigation";

interface AuditorPerformance {
  user_id: string;
  full_name: string;
  completion_rate: number;
  completed: number;
  ongoing: number;
  draft: number;
}

interface PicPerformance {
  user_id: string;
  full_name: string;
  completion_rate: number;
  closed: number;
  verified: number;
  in_progress: number;
  open: number;
  overdue: number;
}

interface MonitoringStats {
  total_inspections_running?: number;
  compliance_rate?: number;
  inspections_completed?: number;
  auditee_status?: Array<{
    name: string;
    type: string;
    compliance: number;
    open_issues: number;
  }>;
}

// Fetch Auditor statistics directly from backend API
const fetchAuditorDetail = async (plantId?: string) => {
  const res = await api.get("/dashboard/auditor-detail", {
    params: plantId && plantId !== "all" && plantId !== "global" ? { plant_id: plantId } : {}
  });
  return (res.data?.data?.items || []) as AuditorPerformance[];
};

// Fetch PIC statistics directly from backend API
const fetchPICDetail = async (plantId?: string) => {
  const res = await api.get("/dashboard/pic-detail", {
    params: plantId && plantId !== "all" && plantId !== "global" ? { plant_id: plantId } : {}
  });
  return (res.data?.data?.items || []) as PicPerformance[];
};

// Fetch overall dashboard stats
const fetchDashboardStats = async (plantId?: string) => {
  const res = await api.get("/dashboard/stats", {
    params: plantId && plantId !== "all" && plantId !== "global" ? { plant_id: plantId } : {}
  });
  return res.data?.data as MonitoringStats;
};

export default function MonitoringPage() {
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);
  const params = useParams();
  const plantCode = (params?.plantCode as string) || "";

  const [activeTab, setActiveTab] = useState<"audit" | "pic" | "auditee">("audit");

  const { data: auditors = [], isLoading: isAuditorLoading } = useQuery({
    queryKey: ["monitoring-auditor-detail", plantCode],
    queryFn: () => fetchAuditorDetail(plantCode),
    enabled: mounted && !!user,
  });

  const { data: pics = [], isLoading: isPICLoading } = useQuery({
    queryKey: ["monitoring-pic-detail", plantCode],
    queryFn: () => fetchPICDetail(plantCode),
    enabled: mounted && !!user,
  });

  const { data: stats, isLoading: isStatsLoading } = useQuery({
    queryKey: ["monitoring-global-stats", plantCode],
    queryFn: () => fetchDashboardStats(plantCode),
    enabled: mounted && !!user,
  });

  if (!mounted) {
    return <div className="h-screen w-full bg-background" />;
  }

  const getInitials = (name: string) => {
    if (!name) return "U";
    const parts = name.trim().split(" ");
    if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
    return name.slice(0, 2).toUpperCase();
  };

  return (
    <div className="w-full space-y-6 animate-in fade-in duration-500 pb-24 md:pb-6">
      
      {/* Header Section */}
      <section className="flex flex-col gap-1 mb-2">
        <h2 className="text-3xl font-bold tracking-tight text-primary">Monitoring Operational</h2>
        <p className="text-muted-foreground text-sm">
          Ringkasan performa real-time aktivitas inspeksi, kinerja auditor, dan penanganan isu PIC.
        </p>
      </section>

      {/* Tabs Navigation */}
      <nav className="flex border-b border-border w-full overflow-x-auto hide-scrollbar sticky top-[64px] bg-background z-30 pt-2 pb-0">
        <button 
          className={cn(
            "flex-1 py-2.5 px-4 text-center font-semibold border-b-2 transition-colors flex items-center justify-center gap-2 text-sm",
            activeTab === "audit" ? "text-primary border-primary" : "text-muted-foreground border-transparent hover:text-primary"
          )}
          onClick={() => setActiveTab("audit")}
        >
          <ShieldCheck className="h-4 w-4" /> Performance Auditor ({auditors.length})
        </button>
        <button 
          className={cn(
            "flex-1 py-2.5 px-4 text-center font-semibold border-b-2 transition-colors flex items-center justify-center gap-2 text-sm",
            activeTab === "pic" ? "text-primary border-primary" : "text-muted-foreground border-transparent hover:text-primary"
          )}
          onClick={() => setActiveTab("pic")}
        >
          <UserCheck className="h-4 w-4" /> Kinerja PIC ({pics.length})
        </button>
        <button 
          className={cn(
            "flex-1 py-2.5 px-4 text-center font-semibold border-b-2 transition-colors flex items-center justify-center gap-2 text-sm",
            activeTab === "auditee" ? "text-primary border-primary" : "text-muted-foreground border-transparent hover:text-primary"
          )}
          onClick={() => setActiveTab("auditee")}
        >
          <Building2 className="h-4 w-4" /> Compliance Auditee
        </button>
      </nav>

      {/* TAB CONTENT: AUDIT (AUDITOR PERFORMANCE) */}
      {activeTab === "audit" && (
        <div className="flex flex-col gap-4 mt-2 animate-in slide-in-from-bottom-2">
          {/* Bento Stats Row */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="bg-card border border-border rounded-xl p-4 shadow-sm flex flex-col gap-1">
              <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Inspeksi Berjalan</span>
              <span className="text-3xl font-bold text-primary">{isStatsLoading ? "..." : (stats?.total_inspections_running ?? 0)}</span>
            </div>
            <div className="bg-card border border-border rounded-lg p-4 shadow-sm flex flex-col gap-1">
              <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Compliance Rate</span>
              <span className="text-3xl font-bold text-green-500">{isStatsLoading ? "..." : `${Number(stats?.compliance_rate || 0).toFixed(1)}%`}</span>
            </div>
            <div className="bg-card border border-border rounded-lg p-4 shadow-sm flex flex-col gap-1">
              <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Inspeksi Selesai</span>
              <span className="text-3xl font-bold text-green-600">{isStatsLoading ? "..." : (stats?.inspections_completed ?? 0)}</span>
            </div>
            <div className="bg-card border border-border rounded-lg p-4 shadow-sm flex flex-col gap-1">
              <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Auditor Aktif</span>
              <span className="text-3xl font-bold text-blue-500">{auditors.length}</span>
            </div>
          </div>
          
          <h3 className="text-lg font-bold text-foreground mt-2 border-b border-border pb-2">Daftar Kinerja Auditor</h3>
          
          {isAuditorLoading ? (
            <div className="h-48 w-full flex items-center justify-center">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : auditors.length === 0 ? (
            <div className="p-8 text-center text-muted-foreground bg-card rounded-xl border border-border">
              Belum ada riwayat aktivitas inspektur auditor
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {auditors.map((auditor, idx: number) => (
                <div key={auditor.user_id || idx} className="bg-card border border-border rounded-xl shadow-sm overflow-hidden hover:shadow-md transition-shadow flex flex-col group">
                  <div className="p-4 flex justify-between items-start border-b border-border bg-muted/30 group-hover:bg-muted/50 transition-colors">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center text-primary font-bold text-sm">
                        {getInitials(auditor.full_name)}
                      </div>
                      <div className="flex flex-col">
                        <span className="font-semibold text-foreground text-sm">{auditor.full_name}</span>
                        <span className="text-xs text-muted-foreground font-mono">{auditor.user_id}</span>
                      </div>
                    </div>
                    <div className="flex flex-col items-end">
                      <span className="font-bold text-green-600 text-lg">{Number(auditor.completion_rate || 0).toFixed(0)}%</span>
                      <span className="text-[10px] font-semibold text-muted-foreground uppercase">Selesai</span>
                    </div>
                  </div>
                  <div className="p-4 flex flex-col gap-2 grow bg-card">
                    <div className="flex justify-between items-center w-full">
                      <div className="flex flex-col items-center flex-1">
                        <span className="text-lg font-bold text-green-600">{auditor.completed}</span>
                        <span className="text-[10px] font-semibold text-muted-foreground uppercase">Selesai</span>
                      </div>
                      <div className="w-px h-8 bg-border"></div>
                      <div className="flex flex-col items-center flex-1">
                        <span className="text-lg font-bold text-primary">{auditor.ongoing}</span>
                        <span className="text-[10px] font-semibold text-muted-foreground uppercase">Berjalan</span>
                      </div>
                      <div className="w-px h-8 bg-border"></div>
                      <div className="flex flex-col items-center flex-1">
                        <span className="text-lg font-bold text-muted-foreground">{auditor.draft}</span>
                        <span className="text-[10px] font-semibold text-muted-foreground uppercase">Draft</span>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* TAB CONTENT: PIC PERFORMANCE */}
      {activeTab === "pic" && (
        <div className="flex flex-col gap-4 mt-2 animate-in slide-in-from-bottom-2">
          <h3 className="text-lg font-bold text-foreground border-b border-border pb-2">Kinerja Follow-Up Person In Charge (PIC)</h3>
          
          {isPICLoading ? (
            <div className="h-48 w-full flex items-center justify-center">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : pics.length === 0 ? (
            <div className="p-8 text-center text-muted-foreground bg-card rounded-xl border border-border">
              Belum ada riwayat penanganan isu oleh PIC
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {pics.map((pic, idx: number) => (
                <div key={pic.user_id || idx} className="bg-card border border-border rounded-xl shadow-sm overflow-hidden hover:shadow-md transition-shadow flex flex-col">
                  <div className="p-4 flex justify-between items-start border-b border-border bg-muted/30">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-full bg-orange-500/10 flex items-center justify-center text-orange-600 font-bold text-sm">
                        {getInitials(pic.full_name)}
                      </div>
                      <div className="flex flex-col">
                        <span className="font-semibold text-foreground text-sm">{pic.full_name}</span>
                        <span className="text-xs text-muted-foreground font-mono">{pic.user_id}</span>
                      </div>
                    </div>
                    <div className="flex flex-col items-end">
                      <span className="font-bold text-primary text-lg">{Number(pic.completion_rate || 0).toFixed(0)}%</span>
                      <span className="text-[10px] font-semibold text-muted-foreground uppercase">Terselesaikan</span>
                    </div>
                  </div>
                  <div className="p-4 grid grid-cols-3 gap-2 text-center bg-card">
                    <div>
                      <span className="text-base font-bold text-green-600">{pic.closed + pic.verified}</span>
                      <span className="text-[10px] font-semibold text-muted-foreground block uppercase">Selesai</span>
                    </div>
                    <div>
                      <span className="text-base font-bold text-orange-500">{pic.in_progress + pic.open}</span>
                      <span className="text-[10px] font-semibold text-muted-foreground block uppercase">Proses</span>
                    </div>
                    <div>
                      <span className="text-base font-bold text-red-500">{pic.overdue}</span>
                      <span className="text-[10px] font-semibold text-muted-foreground block uppercase">Overdue</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* TAB CONTENT: AUDITEE STATUS */}
      {activeTab === "auditee" && (
        <div className="flex flex-col gap-4 mt-2 animate-in slide-in-from-bottom-2">
          <h3 className="text-lg font-bold text-foreground border-b border-border pb-2">Compliance Area & Auditee</h3>
          {(!stats?.auditee_status || stats?.auditee_status?.length === 0) ? (
            <div className="p-8 text-center text-muted-foreground bg-card rounded-xl border border-border">
              Belum ada data status compliance area
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {stats?.auditee_status?.map((item, idx: number) => (
                <div key={idx} className="flex items-center justify-between p-4 bg-card rounded-xl border border-border shadow-sm">
                  <div className="flex items-center gap-3">
                    <Building2 className="text-primary h-6 w-6" />
                    <div>
                      <p className="text-base font-semibold text-foreground">{item.name}</p>
                      <p className="text-xs text-muted-foreground">{item.type}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className={cn("text-lg font-bold", item.compliance < 70 ? "text-red-500" : "text-green-500")}>
                      {item.compliance}%
                    </p>
                    <p className="text-xs font-medium text-muted-foreground">{item.open_issues} Issue Terbuka</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
