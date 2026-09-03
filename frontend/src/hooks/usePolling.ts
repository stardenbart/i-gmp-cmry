import { useEffect, useEffectEvent, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useAuthStore } from "@/stores/authStore";
import { api } from "@/lib/api/axios";

interface UsePollingOptions {
  onIssueUpdated?: () => void;
  intervalMs?: number;
  immediate?: boolean;
}

interface PollEvent {
  type: string;
  timestamp: number;
}

interface PollResponse {
  server_time: number;
  events: PollEvent[];
}

export function usePolling(options?: UsePollingOptions) {
  const queryClient = useQueryClient();
  const isLoggedIn = useAuthStore((state) => Boolean(state.user));
  const lastPollTimeRef = useRef<number>(0);
  const consecutiveFailuresRef = useRef<number>(0);
  const notifyIssueUpdated = useEffectEvent(() => {
    options?.onIssueUpdated?.();
  });

  const baseIntervalMs = options?.intervalMs ?? 10000; // 10s default interval
  const pollImmediately = options?.immediate ?? true;

  useEffect(() => {
    if (!isLoggedIn) return;
    if (lastPollTimeRef.current === 0) {
      lastPollTimeRef.current = Date.now() - 10000;
    }

    let isMounted = true;
    let timerId: NodeJS.Timeout | null = null;

    const poll = async () => {
      // Pause polling if document/tab is hidden/inactive
      if (typeof document !== "undefined" && document.visibilityState !== "visible") {
        scheduleNextPoll();
        return;
      }

      try {
        const since = lastPollTimeRef.current;
        const res = await api.get<PollResponse>(`/events/poll?since=${since}`);
        
        if (!isMounted) return;

        consecutiveFailuresRef.current = 0; // reset failures on success

        if (res.data?.server_time) {
          lastPollTimeRef.current = res.data.server_time;
        }

        const events = res.data?.events || [];
        if (events.length > 0) {
          let hasIssueUpdate = false;
          events.forEach((ev) => {
            if (ev.type === "ISSUE_UPDATED" || ev.type === "INSPECTION_UPDATED") {
              hasIssueUpdate = true;
            }
          });

          if (hasIssueUpdate) {
            queryClient.invalidateQueries({ queryKey: ["auditee-issues"] });
            queryClient.invalidateQueries({ queryKey: ["auditee-issue-summary"] });
            queryClient.invalidateQueries({ queryKey: ["auditee-priority-issues"] });
            queryClient.invalidateQueries({ queryKey: ["wowr-issues"] });
            queryClient.invalidateQueries({ queryKey: ["issues"] });
            queryClient.invalidateQueries({ queryKey: ["issues-filter"] });
            queryClient.invalidateQueries({ queryKey: ["issue"] });
            queryClient.invalidateQueries({ queryKey: ["issue-photos"] });
            queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] });
            queryClient.invalidateQueries({ queryKey: ["dashboard-stats-stitch"] });
            queryClient.invalidateQueries({ queryKey: ["auditor-inspections"] });
            queryClient.invalidateQueries({ queryKey: ["auditor-inspections-trend"] });
            queryClient.invalidateQueries({ queryKey: ["auditor-issue-summary"] });
            queryClient.invalidateQueries({ queryKey: ["auditor-pending-issues"] });

            notifyIssueUpdated();
          }
        }
      } catch {
        consecutiveFailuresRef.current += 1;
      } finally {
        if (isMounted) {
          scheduleNextPoll();
        }
      }
    };

    const scheduleNextPoll = () => {
      if (!isMounted) return;
      const backoffFactor = Math.min(5, consecutiveFailuresRef.current);
      const nextDelay = baseIntervalMs * (backoffFactor > 0 ? backoffFactor : 1);
      timerId = setTimeout(poll, nextDelay);
    };

    // Initial page data is already fetched by React Query. Dashboards can
    // defer this background synchronization to avoid competing with LCP.
    if (pollImmediately) {
      void poll();
    } else {
      scheduleNextPoll();
    }

    return () => {
      isMounted = false;
      if (timerId) clearTimeout(timerId);
    };
  }, [isLoggedIn, queryClient, baseIntervalMs, pollImmediately]);
}
