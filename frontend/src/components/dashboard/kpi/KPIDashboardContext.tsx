"use client";

import { createContext, useContext, useState, type ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { api } from "@/lib/api/axios";
import { issueApi } from "@/lib/api/issue.api";
import { plantApi } from "@/types/api/master";
import type { Plant, PaginatedResponse } from "@/types/api";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { usePolling } from "@/hooks/usePolling";
import type { AdminDashboardStats } from "@/components/dashboard/admin/AdminDashboardContext";
import { dashboardApi, type DashboardTrendData, type TrendMode, type TrendGranularity } from "@/lib/api/dashboard.api";
import { useTrendModeState, useTrendRangeState, useTrendGranularityState, type TrendDateRange } from "@/components/dashboard/useTrendPeriodState";

// Same shape monitoring/page.tsx uses for these two endpoints (see
// GET /dashboard/auditor-detail and GET /dashboard/pic-detail).
export interface KPIAuditorPerformance {
  user_id: string;
  full_name: string;
  completion_rate: number;
  total: number;
  completed: number;
  ongoing: number;
  draft: number;
}

export interface KPIPicPerformance {
  user_id: string;
  full_name: string;
  completion_rate: number;
  total: number;
  closed: number;
  verified: number;
  in_progress: number;
  open: number;
  overdue: number;
}

export interface KPIWOWRByArea {
  area_name: string;
  total: number;
  verified: number;
  pending: number;
  rejected: number;
  awaiting: number;
}

export interface KPIWOWRReport {
  summary: {
    total: number;
    verified: number;
    pending: number;
    rejected: number;
    awaiting: number;
    verified_rate: number;
    pending_rate: number;
    rejected_rate: number;
    awaiting_rate: number;
  };
  by_area: KPIWOWRByArea[];
}

const fetchKPIStats = async (plantId?: string) => {
  const res = await api.get("/dashboard/stats", {
    params: { plant_id: plantId, include_trend: false },
  });
  return res.data.data as AdminDashboardStats;
};

const fetchAuditorDetail = async (plantId?: string) => {
  const res = await api.get("/dashboard/auditor-detail", { params: plantId ? { plant_id: plantId } : {} });
  return (res.data?.data?.items || []) as KPIAuditorPerformance[];
};

const fetchPICDetail = async (plantId?: string) => {
  const res = await api.get("/dashboard/pic-detail", { params: plantId ? { plant_id: plantId } : {} });
  return (res.data?.data?.items || []) as KPIPicPerformance[];
};

interface KPIDashboardContextValue {
  mounted: boolean;
  user: ReturnType<typeof useAuthStore.getState>["user"];
  isSuperAdmin: boolean;
  selectedPlant: string;
  setSelectedPlant: (plant: string) => void;
  /** Resolved plant filter actually applied to every /dashboard/* query
   * this session (SuperAdmin's picked plant, or a non-SuperAdmin's own
   * plant) — exposed so the Custom KPI Visualization Builder's widgets can
   * apply the exact same scope to /analytics/query without recomputing it. */
  effectivePlant: string;
  plantsResponse: PaginatedResponse<Plant> | undefined;
  trendMode: TrendMode;
  setTrendMode: (mode: TrendMode) => void;
  trendRange: TrendDateRange;
  setTrendRange: (range: TrendDateRange) => void;
  trendGranularity: TrendGranularity;
  setTrendGranularity: (granularity: TrendGranularity) => void;
  stats: AdminDashboardStats | undefined;
  trendData: DashboardTrendData | undefined;
  auditorDetail: KPIAuditorPerformance[];
  picDetail: KPIPicPerformance[];
  wowrReport: KPIWOWRReport | undefined;
  isLoading: boolean;
  isFetching: boolean;
  isTrendLoading: boolean;
  isTrendFetching: boolean;
}

const KPIDashboardContext = createContext<KPIDashboardContextValue | null>(null);

export function KPIDashboardProvider({ children }: { children: ReactNode }) {
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);
  const params = useParams();
  const urlPlantCode = (params?.plantCode as string) || "";
  const urlPlant = urlPlantCode && urlPlantCode !== "all" && urlPlantCode !== "global" ? urlPlantCode : "";

  usePolling({ intervalMs: 30_000, immediate: false });

  const [selectedPlant, setSelectedPlant] = useState<string>("");
  const [trendMode, setTrendMode] = useTrendModeState();
  const [trendRange, setTrendRange] = useTrendRangeState();
  const [trendGranularity, setTrendGranularity] = useTrendGranularityState();

  const isSuperAdmin =
    user?.role_id === "ROLE-000" ||
    user?.role_id === "SUPERADMIN" ||
    user?.role?.role_name === "Super Admin" ||
    !user?.plant_id;

  // Non-SuperAdmin never gets a plant filter — the backend already scopes
  // every /dashboard/* endpoint to their own plant/area (getAllowedAreas),
  // so there's nothing useful for them to pick here.
  const effectivePlant = isSuperAdmin ? (selectedPlant === "all" ? "" : selectedPlant) : user?.plant_id || urlPlant;

  const { data: plantsResponse } = useQuery({
    queryKey: ["plants-master-kpi"],
    queryFn: () => plantApi.getAll(1, 100),
    enabled: mounted && isSuperAdmin,
  });

  const {
    data: stats,
    isLoading,
    isFetching,
  } = useQuery({
    queryKey: ["dashboard-stats-kpi", effectivePlant],
    queryFn: () => fetchKPIStats(effectivePlant),
    enabled: mounted && !!user,
    staleTime: 10_000,
  });

  const {
    data: trendData,
    isLoading: isTrendLoading,
    isFetching: isTrendFetching,
  } = useQuery({
    queryKey: ["dashboard-trend", "kpi", trendMode, trendMode === "range" ? trendRange : null, trendMode === "range" ? trendGranularity : null, effectivePlant],
    queryFn: () =>
      dashboardApi.getTrend({
        period: trendMode,
        plant_id: effectivePlant || undefined,
        ...(trendMode === "range" ? { start_date: trendRange.start, end_date: trendRange.end, granularity: trendGranularity } : {}),
      }),
    enabled: mounted && !!user && (trendMode !== "range" || (!!trendRange.start && !!trendRange.end)),
    staleTime: 10_000,
    placeholderData: (previousData) => (previousData?.period === trendMode ? previousData : undefined),
  });

  const { data: auditorDetail = [] } = useQuery({
    queryKey: ["kpi-auditor-detail", effectivePlant],
    queryFn: () => fetchAuditorDetail(effectivePlant || undefined),
    enabled: mounted && !!user,
    staleTime: 30_000,
  });

  const { data: picDetail = [] } = useQuery({
    queryKey: ["kpi-pic-detail", effectivePlant],
    queryFn: () => fetchPICDetail(effectivePlant || undefined),
    enabled: mounted && !!user,
    staleTime: 30_000,
  });

  const { data: wowrReport } = useQuery({
    queryKey: ["kpi-wowr-report", effectivePlant],
    queryFn: () => issueApi.getWOWRReport({ plant_id: effectivePlant || undefined }) as Promise<KPIWOWRReport>,
    enabled: mounted && !!user,
    staleTime: 30_000,
  });

  return (
    <KPIDashboardContext.Provider
      value={{
        mounted,
        user,
        isSuperAdmin,
        selectedPlant,
        setSelectedPlant,
        effectivePlant,
        plantsResponse,
        trendMode,
        setTrendMode,
        trendRange,
        setTrendRange,
        trendGranularity,
        setTrendGranularity,
        stats,
        trendData,
        auditorDetail,
        picDetail,
        wowrReport,
        isLoading,
        isFetching,
        isTrendLoading,
        isTrendFetching,
      }}
    >
      {children}
    </KPIDashboardContext.Provider>
  );
}

export function useKPIDashboard() {
  const ctx = useContext(KPIDashboardContext);
  if (!ctx) {
    throw new Error("useKPIDashboard must be used within KPIDashboardProvider");
  }
  return ctx;
}
