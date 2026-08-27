"use client";

import { createContext, useContext, useCallback, type ReactNode } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { inspectionApi } from "@/lib/api/inspection.api";
import { filterApi } from "@/lib/api/filter.api";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { usePolling } from "@/hooks/usePolling";
import { format } from "date-fns";
import { id as idLocale } from "date-fns/locale";

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
  currentYear: number;
  chartData: Array<Record<string, string | number>>;
  isTrendLoading: boolean;
}

const AuditorDashboardContext = createContext<AuditorDashboardContextValue | null>(null);

export function AuditorDashboardProvider({ children }: { children: ReactNode }) {
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();

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

  const currentYear = new Date().getFullYear();
  const { data: trendData, isLoading: isTrendLoading } = useQuery({
    queryKey: ["auditor-inspections-trend", user?.id, currentYear],
    queryFn: () => inspectionApi.getAnalyticsTrend(user?.id || "", currentYear),
    enabled: mounted && !!user,
  });

  const { data: issueSummaryData, isLoading: isIssueSummaryLoading } = useQuery({
    queryKey: ["auditor-issue-summary"],
    queryFn: () => filterApi.issues({ limit: 1 }),
    enabled: mounted,
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
    queryClient.invalidateQueries({ queryKey: ["auditor-inspections-trend"] });
    queryClient.invalidateQueries({ queryKey: ["auditor-issue-summary"] });
    queryClient.invalidateQueries({ queryKey: ["auditor-pending-issues"] });
  }, [queryClient]);

  const inspections = inspectionsData?.items || [];
  const totalInspections = inspections.length;
  const completedInspections = inspections.filter((i) => i.status === "Completed" || i.status === "Approved").length;
  const ongoingInspections = inspections.filter((i) => i.status === "Ongoing" || i.status === "Draft").length;

  const issueStatusCounts = issueSummaryData?.facets.status || {};
  const totalIssues = issueSummaryData?.total || 0;
  const pendingValidationList = pendingIssuesData?.items || [];
  const pendingValidationCount = issueStatusCounts.PendingValidation || 0;
  const openIssues = (issueStatusCounts.Open || 0) + (issueStatusCounts.InProgress || 0) + (issueStatusCounts.Overdue || 0);

  const isLoading = isInspectionsLoading || isTrendLoading || isIssueSummaryLoading || isPendingIssuesLoading;
  const isFetching = isInspectionsFetching;

  const chartData = (trendData?.data || []).map((item: Record<string, string | number>) => ({
    ...item,
    period_label: format(new Date(`${item.date}T00:00:00`), "MMM yyyy", { locale: idLocale }),
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
        currentYear,
        chartData,
        isTrendLoading,
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
