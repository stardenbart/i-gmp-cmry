import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useAuthStore } from "@/stores/authStore";

interface UseSSEOptions {
  onIssueUpdated?: () => void;
}

export function useSSE(options?: UseSSEOptions) {
  const queryClient = useQueryClient();
  const token = useAuthStore((state) => state.token);

  useEffect(() => {
    if (!token) return;

    // Use absolute URL if necessary or rely on proxy/setup
    // Assuming /api is mapped correctly or use base URL from env
    const baseURL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
    
    // EventSource doesn't support custom headers easily without polyfills.
    // Usually, we pass token in URL or rely on cookies.
    // If your backend auth middleware requires Bearer token, you might need a polyfill 
    // or adjust backend to accept token from query param `?token=...`.
    // Since we put it behind rate limiter but let's assume it accepts query tokens 
    // or we can just bypass auth for SSE (if no sensitive data in payload).
    const sseUrl = `${baseURL}/sse/events?token=${token}`;
    
    // We append token as query so backend can authenticate the connection.
    const eventSource = new EventSource(sseUrl);

    eventSource.onopen = () => {
      console.log("SSE Connection established");
    };

    eventSource.onmessage = (event) => {
      const data = event.data;
      if (data === "ISSUE_UPDATED") {
        // Invalidate both auditee and wowr queries
        queryClient.invalidateQueries({ queryKey: ["auditee-issues"] });
        queryClient.invalidateQueries({ queryKey: ["wowr-issues"] });
        queryClient.invalidateQueries({ queryKey: ["issues"] }); // For Auditor page
        
        if (options?.onIssueUpdated) {
          options.onIssueUpdated();
        }
      }
    };

    eventSource.onerror = (error) => {
      console.error("SSE Connection error:", error);
      eventSource.close();
      
      // Auto-reconnect logic is built into EventSource, 
      // but closing it on error and letting it retry or manual reconnect can be done.
      // By default EventSource retries automatically.
    };

    return () => {
      eventSource.close();
      console.log("SSE Connection closed");
    };
  }, [queryClient, token, options]);
}
