"use client";

import { createContext, useContext, useCallback, type ComponentType, type ReactNode } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { filterApi } from "@/lib/api/filter.api";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { usePolling } from "@/hooks/usePolling";

interface AuditeeDashboardContextValue {
  mounted: boolean;
  user: ReturnType<typeof useAuthStore.getState>["user"];
  isLoading: boolean;
  isFetching: boolean;
  handleRefresh: () => void;
  totalIssues: number;
  openIssues: number;
  pendingIssues: number;
  closedIssues: number;
  activeTasks: Awaited<ReturnType<typeof filterApi.issues>>["items"];
}

const AuditeeDashboardContext = createContext<AuditeeDashboardContextValue | null>(null);

export function AuditeeDashboardProvider({ children }: { children: ReactNode }) {
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();

  usePolling({ intervalMs: 30_000, immediate: false });

  // NOTE: `issue_pic_user_id` is intentionally NOT sent here. The backend
  // already scopes non-auditor callers to their own issues automatically
  // (GetFiltered sets ScopeUserID), which matches an issue assigned
  // directly, delegated to the user, or mapped via PIC_Mapping for the
  // Kawasan — the actual ways an Auditee ends up responsible for a
  // finding. Findings created from an inspection always store
  // IssuePICUserID as the *inspector*, not the Auditee, so adding an
  // explicit `issue_pic_user_id` filter here would AND against that
  // broader OR-scope and hide every finding reached only via delegation
  // or PIC_Mapping — i.e. almost all of them.
  const {
    data: summaryData,
    isLoading: isSummaryLoading,
    isFetching: isSummaryFetching,
  } = useQuery({
    queryKey: ["auditee-issue-summary", user?.id],
    queryFn: () => filterApi.issues({ limit: 1 }),
    enabled: mounted && !!user?.id,
    staleTime: 30_000,
  });

  const {
    data: priorityData,
    isLoading: isPriorityLoading,
    isFetching: isPriorityFetching,
  } = useQuery({
    queryKey: ["auditee-priority-issues", user?.id],
    queryFn: () =>
      filterApi.issues({
        limit: 10,
        status__in: "Open,InProgress,Overdue,PendingValidation",
        sort_by: "due_date",
        sort_order: "asc",
      }),
    enabled: mounted && !!user?.id,
    staleTime: 30_000,
  });

  const handleRefresh = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ["auditee-issue-summary"] });
    queryClient.invalidateQueries({ queryKey: ["auditee-priority-issues"] });
  }, [queryClient]);

  const statusCounts = summaryData?.facets.status || {};
  const statusCount = (...statuses: string[]) => statuses.reduce((total, status) => total + (statusCounts[status] || 0), 0);
  const totalIssues = summaryData?.total || 0;
  const openIssues = statusCount("Open", "InProgress", "Overdue");
  const pendingIssues = statusCount("PendingValidation");
  const closedIssues = statusCount("Closed", "Verified");

  const isLoading = isSummaryLoading || isPriorityLoading;
  const isFetching = isSummaryFetching || isPriorityFetching;
  const activeTasks = priorityData?.items || [];

  return (
    <AuditeeDashboardContext.Provider
      value={{
        mounted,
        user,
        isLoading,
        isFetching,
        handleRefresh,
        totalIssues,
        openIssues,
        pendingIssues,
        closedIssues,
        activeTasks,
      }}
    >
      {children}
    </AuditeeDashboardContext.Provider>
  );
}

export function useAuditeeDashboard() {
  const ctx = useContext(AuditeeDashboardContext);
  if (!ctx) {
    throw new Error("useAuditeeDashboard must be used within AuditeeDashboardProvider");
  }
  return ctx;
}

/** See withAdminDashboardContext (AdminDashboardContext.tsx) for the full
 * rationale — same pattern, for auditee-family widgets. */
export function withAuditeeDashboardContext<P extends object>(Component: ComponentType<P>): ComponentType<P> {
  return function WithAuditeeDashboardContext(props: P) {
    const existing = useContext(AuditeeDashboardContext);
    if (existing) return <Component {...props} />;
    return (
      <AuditeeDashboardProvider>
        <Component {...props} />
      </AuditeeDashboardProvider>
    );
  };
}
