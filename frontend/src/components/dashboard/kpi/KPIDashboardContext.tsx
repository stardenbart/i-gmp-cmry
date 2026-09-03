"use client";

import { createContext, useContext, useState, type ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { plantApi } from "@/types/api/master";
import type { Plant, PaginatedResponse } from "@/types/api";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { usePolling } from "@/hooks/usePolling";
import type { TrendMode, TrendGranularity } from "@/lib/api/dashboard.api";
import type { AnalyticsCatalog } from "@/lib/api/analytics.api";
import type { PublicKPIBootstrap } from "@/lib/api/kpi-share.api";
import { useTrendModeState, useTrendRangeState, useTrendGranularityState, type TrendDateRange } from "@/components/dashboard/useTrendPeriodState";

interface KPIDashboardContextValue {
  mounted: boolean;
  user: ReturnType<typeof useAuthStore.getState>["user"];
  isPublic: boolean;
  canChangePeriod: boolean;
  publicShareToken?: string;
  analyticsCatalog?: AnalyticsCatalog;
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

  return (
    <KPIDashboardContext.Provider
      value={{
        mounted,
        user,
        isPublic: false,
        canChangePeriod: true,
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
      }}
    >
      {children}
    </KPIDashboardContext.Provider>
  );
}

export function PublicKPIDashboardProvider({
  token,
  bootstrap,
  children,
}: {
  token: string;
  bootstrap: PublicKPIBootstrap;
  children: ReactNode;
}) {
  const [trendMode, setTrendMode] = useState<TrendMode>(bootstrap.filter.period);
  const [trendRange, setTrendRange] = useState<TrendDateRange>({
    start: bootstrap.filter.start_date || "",
    end: bootstrap.filter.end_date || "",
  });
  const [trendGranularity, setTrendGranularity] = useState<TrendGranularity>(bootstrap.filter.granularity || "day");
  return (
    <KPIDashboardContext.Provider
      value={{
        mounted: true,
        user: null,
        isPublic: true,
        canChangePeriod: bootstrap.allow_period_change,
        publicShareToken: token,
        analyticsCatalog: bootstrap.catalog,
        isSuperAdmin: false,
        selectedPlant: bootstrap.plant_id,
        setSelectedPlant: () => undefined,
        effectivePlant: bootstrap.plant_id,
        plantsResponse: undefined,
        trendMode,
        setTrendMode: bootstrap.allow_period_change ? setTrendMode : () => undefined,
        trendRange,
        setTrendRange: bootstrap.allow_period_change ? setTrendRange : () => undefined,
        trendGranularity,
        setTrendGranularity: bootstrap.allow_period_change ? setTrendGranularity : () => undefined,
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
