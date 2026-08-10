import { useEffect, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useAuthStore } from "@/stores/authStore";
import { api } from "@/lib/api/axios";

interface UsePollingOptions {
  onIssueUpdated?: () => void;
  intervalMs?: number;
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
  const token = useAuthStore((state) => state.token);
  const lastPollTimeRef = useRef<number>(Date.now() - 10000);
  const optionsRef = useRef(options);
  optionsRef.current = options;

  const intervalMs = options?.intervalMs ?? 5000; // 5s interval

  useEffect(() => {
    if (!token) return;

    let isMounted = true;

    const poll = async () => {
      // Pause polling if document/tab is hidden/inactive
      if (typeof document !== "undefined" && document.visibilityState !== "visible") {
        return;
      }

      try {
        const since = lastPollTimeRef.current;
        const res = await api.get<PollResponse>(`/events/poll?since=${since}`);
        
        if (!isMounted) return;

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
            queryClient.invalidateQueries({ queryKey: ["wowr-issues"] });
            queryClient.invalidateQueries({ queryKey: ["issues"] });
            queryClient.invalidateQueries({ queryKey: ["issues-filter"] });
            queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] });
            queryClient.invalidateQueries({ queryKey: ["dashboard-stats-stitch"] });
            queryClient.invalidateQueries({ queryKey: ["auditor-inspections"] });
            queryClient.invalidateQueries({ queryKey: ["auditor-inspections-trend"] });
            queryClient.invalidateQueries({ queryKey: ["auditor-global-issues"] });

            if (optionsRef.current?.onIssueUpdated) {
              optionsRef.current.onIssueUpdated();
            }
          }
        }
      } catch (err) {
        // Silent catch for network hiccups during polling
      }
    };

    // Immediate initial poll
    poll();

    const timer = setInterval(poll, intervalMs);

    return () => {
      isMounted = false;
      clearInterval(timer);
    };
  }, [token, queryClient, intervalMs]);
}
