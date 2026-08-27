"use client";

import { createContext, useContext, useCallback, type ReactNode } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { filterApi } from "@/lib/api/filter.api";
import { api } from "@/lib/api/axios";
import { dashboardApi, type DashboardTrendData, type TrendMode, type TrendGranularity } from "@/lib/api/dashboard.api";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { usePolling } from "@/hooks/usePolling";
import { useTrendModeState, useTrendRangeState, useTrendGranularityState, type TrendDateRange } from "@/components/dashboard/useTrendPeriodState";

interface AuditorDashboardContextValue {
  mounted: boolean;
  user: ReturnType<typeof useAuthStore.getState>["user"];
  isLoading: boolean;
  isFetching: boolean;
  handleRefresh: () => void;
  totalInspections: number;
  completedInspections: number;
  ongoingInspections: number;
  totalIssues: number;
  pendingValidationCount: number;
  pendingValidationList: Awaited<ReturnType<typeof filterApi.issues>>["items"];
  openIssues: number;
  trendMode: TrendMode;
  setTrendMode: (mode: TrendMode) => void;
  trendRange: TrendDateRange;
  setTrendRange: (range: TrendDateRange) => void;
  trendGranularity: TrendGranularity;
  setTrendGranularity: (granularity: TrendGranularity) => void;
  trendData: DashboardTrendData | undefined;
  chartData: Array<Record<string, string | number>>;
  isTrendLoading: boolean;
  isTrendFetching: boolean;
}

const AuditorDashboardContext = createContext<AuditorDashboardContextValue | null>(null);

export function AuditorDashboardProvider({ children }: { children: ReactNode }) {
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();
  const [trendMode, setTrendMode] = useTrendModeState();
  const [trendRange, setTrendRange] = useTrendRangeState();
  const [trendGranularity, setTrendGranularity] = useTrendGranularityState();

  usePolling({ intervalMs: 30_000, immediate: false });

  const {
    data: inspectionsData,
    isLoading: isInspectionsLoading,
    isFetching: isInspectionsFetching,
  } = useQuery({
    queryKey: ["auditor-inspections", user?.id],
    queryFn: () => filterApi.inspections({ limit: 100, inspector_id: user?.id, sort_by: "created_at", sort_order: "desc" }),
    enabled: mounted && !!user,
  });

  const {
    data: trendData,
    isLoading: isTrendLoading,
    isFetching: isTrendFetching,
  } = useQuery({
    queryKey: ["dashboard-trend", "auditor", user?.id, trendMode, trendMode === "range" ? trendRange : null, trendMode === "range" ? trendGranularity : null],
    queryFn: () => dashboardApi.getTrend({
      period: trendMode,
      inspector_id: user?.id,
      ...(trendMode === "range" ? { start_date: trendRange.start, end_date: trendRange.end, granularity: trendGranularity } : {}),
    }),
    enabled: mounted && !!user && (trendMode !== "range" || (!!trendRange.start && !!trendRange.end)),
    staleTime: 10_000,
    placeholderData: (previousData) => previousData,
  });

  const { data: issueSummaryData, isLoading: isIssueSummaryLoading } = useQuery({
    queryKey: ["auditor-issue-summary"],
    queryFn: () => filterApi.issues({ limit: 1 }),
    enabled: mounted,
    staleTime: 30_000,
  });

  const { data: dashboardStats, isLoading: isDashboardStatsLoading } = useQuery({
    queryKey: ["auditor-dashboard-stats", user?.id],
    queryFn: async () => {
      const response = await api.get("/dashboard/stats", { params: { include_trend: false } });
      return response.data.data as { total_issues: number };
    },
    enabled: mounted && !!user,
    staleTime: 30_000,
  });

  const { data: pendingIssuesData, isLoading: isPendingIssuesLoading } = useQuery({
    queryKey: ["auditor-pending-issues"],
    queryFn: () =>
      filterApi.issues({
        status: "PendingValidation",
        limit: 3,
        sort_by: "due_date",
        sort_order: "asc",
      }),
    enabled: mounted,
    staleTime: 30_000,
  });

  const handleRefresh = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ["auditor-inspections"] });
    queryClient.invalidateQueries({ queryKey: ["dashboard-trend", "auditor"] });
    queryClient.invalidateQueries({ queryKey: ["auditor-issue-summary"] });
    queryClient.invalidateQueries({ queryKey: ["auditor-dashboard-stats"] });
    queryClient.invalidateQueries({ queryKey: ["auditor-pending-issues"] });
  }, [queryClient]);

  const inspections = inspectionsData?.items || [];
  const totalInspections = inspections.length;
  const completedInspections = inspections.filter((i) => i.status === "Completed" || i.status === "Approved").length;
  const ongoingInspections = inspections.filter((i) => i.status === "Ongoing" || i.status === "Draft").length;

  const issueStatusCounts = issueSummaryData?.facets.status || {};
  const totalIssues = dashboardStats?.total_issues || 0;
  const pendingValidationList = pendingIssuesData?.items || [];
  const pendingValidationCount = issueStatusCounts.PendingValidation || 0;
  const openIssues = (issueStatusCounts.Open || 0) + (issueStatusCounts.InProgress || 0) + (issueStatusCounts.Overdue || 0);

  const isLoading = isInspectionsLoading || isTrendLoading || isIssueSummaryLoading || isPendingIssuesLoading || isDashboardStatsLoading;
  const isFetching = isInspectionsFetching || isTrendFetching;

  const chartData = (trendData?.points || []).map((item) => ({
    ...item,
    period_label: item.label,
  }));

  return (
    <AuditorDashboardContext.Provider
      value={{
        mounted,
        user,
        isLoading,
        isFetching,
        handleRefresh,
        totalInspections,
        completedInspections,
        ongoingInspections,
        totalIssues,
        pendingValidationCount,
        pendingValidationList,
        openIssues,
        trendMode,
        setTrendMode,
        trendRange,
        setTrendRange,
        trendGranularity,
        setTrendGranularity,
        trendData,
        chartData,
        isTrendLoading,
        isTrendFetching,
      }}
    >
      {children}
    </AuditorDashboardContext.Provider>
  );
}

export function useAuditorDashboard() {
  const ctx = useContext(AuditorDashboardContext);
  if (!ctx) {
    throw new Error("useAuditorDashboard must be used within AuditorDashboardProvider");
  }
  return ctx;
}
