"use client";

import { createContext, useContext, useState, type ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { api } from "@/lib/api/axios";
import { areaApi, plantApi } from "@/types/api/master";
import type { Plant, PaginatedResponse } from "@/types/api";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { usePolling } from "@/hooks/usePolling";
import type { DashboardStats as StatDashboardStats } from "@/components/admin/StatCard";

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

export type TrendPeriod = "1m" | "3m" | "6m" | "quarter" | "1y";

const fetchDashboardStats = async (areaId?: string, period: string = "6m", plantId?: string) => {
  const res = await api.get("/dashboard/stats", {
    params: { area_id: areaId, period, plant_id: plantId },
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
  trendPeriod: TrendPeriod;
  setTrendPeriod: (period: TrendPeriod) => void;
  plantsResponse: PaginatedResponse<Plant> | undefined;
  filteredAreas: Array<{ area_id: string; area_name: string; plant_id?: string }>;
  stats: AdminDashboardStats | undefined;
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
  const [trendPeriod, setTrendPeriod] = useState<TrendPeriod>("6m");

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
    queryKey: ["dashboard-stats-stitch", selectedArea, trendPeriod, effectivePlant],
    queryFn: () => fetchDashboardStats(selectedArea, trendPeriod, effectivePlant),
    enabled: mounted && !!user,
    staleTime: 10000,
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
        trendPeriod,
        setTrendPeriod,
        plantsResponse,
        filteredAreas,
        stats,
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
