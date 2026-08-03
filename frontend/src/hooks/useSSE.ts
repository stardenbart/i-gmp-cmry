import { useEffect, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useAuthStore } from "@/stores/authStore";

interface UseSSEOptions {
  onIssueUpdated?: () => void;
}

export function useSSE(options?: UseSSEOptions) {
  const queryClient = useQueryClient();
  const token = useAuthStore((state) => state.token);
  const optionsRef = useRef(options);
  optionsRef.current = options;

  useEffect(() => {
    if (!token) return;

    // Use absolute URL if necessary or rely on proxy/setup
    // Assuming /api is mapped correctly or use base URL from env
    const getBaseURL = () => {
      if (process.env.NEXT_PUBLIC_API_URL) return process.env.NEXT_PUBLIC_API_URL;
      if (typeof window !== "undefined") return "/api/v1";
      return "http://localhost:8080/api/v1";
    };
    const sseUrl = `${getBaseURL()}/sse/events?token=${token}`;
    
    // We append token as query so backend can authenticate the connection.
    const eventSource = new EventSource(sseUrl);

    eventSource.onopen = () => {
      console.log("SSE Connection established");
    };

    eventSource.onmessage = (event) => {
      const data = event.data;
      if (data === "ISSUE_UPDATED" || data === "INSPECTION_UPDATED") {
        // Invalidate both auditee, wowr, auditor, filter, and dashboard queries
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
    };

    eventSource.onerror = (error) => {
      console.error("SSE Connection error:", error);
      eventSource.close();
    };

    return () => {
      eventSource.close();
      console.log("SSE Connection closed");
    };
  }, [queryClient, token]);
}
