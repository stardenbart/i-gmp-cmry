"use client";

import { createContext, useContext, useState, type ComponentType, type ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { api } from "@/lib/api/axios";
import { areaApi, plantApi } from "@/types/api/master";
import type { Plant, PaginatedResponse } from "@/types/api";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { usePolling } from "@/hooks/usePolling";
import type { DashboardStats as StatDashboardStats } from "@/components/admin/StatCard";
import { dashboardApi, type DashboardTrendData, type TrendMode, type TrendGranularity } from "@/lib/api/dashboard.api";
import { useTrendModeState, useTrendRangeState, useTrendGranularityState, type TrendDateRange } from "@/components/dashboard/useTrendPeriodState";

export interface AdminDashboardStats extends StatDashboardStats {
  inspections_completed?: number;
  inspections_running?: number;
  issues_resolved?: number;
  pic_followup_completed?: number;
  pic_followup_overdue?: number;
  wowr_total?: number;
  wowr_verified?: number;
  wowr_pending?: number;
  wowr_rejected?: number;
  wowr_awaiting?: number;
  wowr_verified_rate?: number;
  auditee_status?: Array<{ name: string; type: string; compliance: number; open_issues: number }>;
  compliance_trend?: Array<{
    rate: number;
    total_issues: number;
    start_date: string;
    end_date: string;
    [key: string]: string | number;
  }>;
}

const fetchDashboardStats = async (areaId?: string, plantId?: string) => {
  const res = await api.get("/dashboard/stats", {
    params: { area_id: areaId, plant_id: plantId, include_trend: false },
  });
  return res.data.data as AdminDashboardStats;
};

interface AdminDashboardContextValue {
  mounted: boolean;
  user: ReturnType<typeof useAuthStore.getState>["user"];
  isSuperAdmin: boolean;
  effectivePlant: string;
  selectedPlant: string;
  setSelectedPlant: (plant: string) => void;
  selectedArea: string;
  setSelectedArea: (area: string) => void;
  trendMode: TrendMode;
  setTrendMode: (mode: TrendMode) => void;
  trendRange: TrendDateRange;
  setTrendRange: (range: TrendDateRange) => void;
  trendGranularity: TrendGranularity;
  setTrendGranularity: (granularity: TrendGranularity) => void;
  plantsResponse: PaginatedResponse<Plant> | undefined;
  filteredAreas: Array<{ area_id: string; area_name: string; plant_id?: string }>;
  stats: AdminDashboardStats | undefined;
  trendData: DashboardTrendData | undefined;
  isTrendLoading: boolean;
  isTrendFetching: boolean;
  isLoading: boolean;
  isFetching: boolean;
}

const AdminDashboardContext = createContext<AdminDashboardContextValue | null>(null);

export function AdminDashboardProvider({ children }: { children: ReactNode }) {
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);
  const params = useParams();
  const urlPlantCode = (params?.plantCode as string) || "";
  const urlPlant = urlPlantCode && urlPlantCode !== "all" && urlPlantCode !== "global" ? urlPlantCode : "";

  usePolling({ intervalMs: 30_000, immediate: false });

  const [selectedPlant, setSelectedPlant] = useState<string>("");
  const [selectedArea, setSelectedArea] = useState<string>("");
  const [trendMode, setTrendMode] = useTrendModeState();
  const [trendRange, setTrendRange] = useTrendRangeState();
  const [trendGranularity, setTrendGranularity] = useTrendGranularityState();

  const isSuperAdmin =
    user?.role_id === "ROLE-000" ||
    user?.role_id === "SUPERADMIN" ||
    user?.role?.role_name === "Super Admin" ||
    !user?.plant_id;

  const effectivePlant = isSuperAdmin
    ? selectedPlant === "all"
      ? ""
      : selectedPlant
    : user?.plant_id || urlPlant;

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

  const filteredAreas =
    areasResponse?.data?.items?.filter((area) => {
      if (!effectivePlant) return true;
      return area.plant_id === effectivePlant;
    }) || [];

  const {
    data: stats,
    isLoading,
    isFetching,
  } = useQuery({
    queryKey: ["dashboard-stats-stitch", selectedArea, effectivePlant],
    queryFn: () => fetchDashboardStats(selectedArea, effectivePlant),
    enabled: mounted && !!user,
    staleTime: 10000,
  });

  const {
    data: trendData,
    isLoading: isTrendLoading,
    isFetching: isTrendFetching,
  } = useQuery({
    queryKey: ["dashboard-trend", "admin", trendMode, trendMode === "range" ? trendRange : null, trendMode === "range" ? trendGranularity : null, selectedArea, effectivePlant],
    queryFn: () => dashboardApi.getTrend({
      period: trendMode,
      area_id: selectedArea || undefined,
      plant_id: effectivePlant || undefined,
      ...(trendMode === "range" ? { start_date: trendRange.start, end_date: trendRange.end, granularity: trendGranularity } : {}),
    }),
    enabled: mounted && !!user && (trendMode !== "range" || (!!trendRange.start && !!trendRange.end)),
    staleTime: 10_000,
    // Only keep the previous chart on screen while a same-mode refetch is in
    // flight (e.g. changing the date range or granularity). Switching modes
    // (quarter <-> range) must NOT show the other mode's stale labels/period.
    placeholderData: (previousData) => (previousData?.period === trendMode ? previousData : undefined),
  });

  return (
    <AdminDashboardContext.Provider
      value={{
        mounted,
        user,
        isSuperAdmin,
        effectivePlant,
        selectedPlant,
        setSelectedPlant,
        selectedArea,
        setSelectedArea,
        trendMode,
        setTrendMode,
        trendRange,
        setTrendRange,
        trendGranularity,
        setTrendGranularity,
        plantsResponse,
        filteredAreas,
        stats,
        trendData,
        isTrendLoading,
        isTrendFetching,
        isLoading,
        isFetching,
      }}
    >
      {children}
    </AdminDashboardContext.Provider>
  );
}

export function useAdminDashboard() {
  const ctx = useContext(AdminDashboardContext);
  if (!ctx) {
    throw new Error("useAdminDashboard must be used within AdminDashboardProvider");
  }
  return ctx;
}

/**
 * Wraps an admin-family widget so it can render on ANY role's dashboard,
 * not just Admin's own — needed since Admin's "Tata Letak Dashboard" (Edit
 * User) can now add an admin widget to an Auditor/Auditee's layout (see
 * components/dashboard/widgets/registry.ts's allMainDashboardWidgets).
 * DashboardAdmin.tsx already wraps its whole page in AdminDashboardProvider,
 * so on Admin's own dashboard `useContext` here finds it immediately and
 * this is a no-op passthrough — zero extra renders/queries for the common
 * case. Only when no AdminDashboardProvider exists yet (a foreign host
 * dashboard) does this mount a fresh one scoped to just this one widget
 * instance, so the extra data-fetching only ever happens for a user who
 * actually has an admin widget turned on for them.
 */
export function withAdminDashboardContext<P extends object>(Component: ComponentType<P>): ComponentType<P> {
  return function WithAdminDashboardContext(props: P) {
    const existing = useContext(AdminDashboardContext);
    if (existing) return <Component {...props} />;
    return (
      <AdminDashboardProvider>
        <Component {...props} />
      </AdminDashboardProvider>
    );
  };
}
