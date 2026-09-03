"use client";

import { useEffect, useEffectEvent, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useAuthStore } from "@/stores/authStore";
import { api } from "@/lib/api/axios";

interface UseRealtimeSyncOptions {
  kawasanId?: string | null;
  onEvent?: (event: RealtimeEvent) => void;
  onIssueUpdated?: () => void;
  pollIntervalMs?: number;
}

export interface RealtimeEvent {
  type?: string;
  [key: string]: unknown;
}

export function useRealtimeSync(options?: UseRealtimeSyncOptions) {
  const queryClient = useQueryClient();
  // A logged-in user is enough to attempt the connection — the actual
  // credential is the httpOnly access_token cookie, which the browser
  // attaches to the WebSocket upgrade request automatically (same as any
  // other same-site request), so it never needs to be read into JS or put
  // in the URL the way a query-string token would.
  const isLoggedIn = useAuthStore((state) => Boolean(state.user));

  const [isConnected, setIsConnected] = useState(false);
  const [isWebSocketActive, setIsWebSocketActive] = useState(false);

  const lastPollTimeRef = useRef<number>(0);

  const kawasanId = options?.kawasanId;
  const pollIntervalMs = options?.pollIntervalMs ?? 5000;
  const handleRealtimeEvent = useEffectEvent((data: RealtimeEvent) => {
    options?.onEvent?.(data);

    if (data.type === "ISSUE_UPDATED" || data.type === "INSPECTION_UPDATED") {
      queryClient.invalidateQueries({ queryKey: ["auditee-issues"] });
      queryClient.invalidateQueries({ queryKey: ["wowr-issues"] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      queryClient.invalidateQueries({ queryKey: ["issues-filter"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] });
      options?.onIssueUpdated?.();
    }
  });

  // 1. Primary WebSocket Real-time Push
  useEffect(() => {
    if (typeof window === "undefined" || !isLoggedIn) return;

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    let host = process.env.NEXT_PUBLIC_WS_HOST;
    if (!host) {
      const hostname = window.location.hostname;
      if (hostname.includes("ngrok") || window.location.protocol === "https:" || window.location.port === "") {
        host = window.location.host;
      } else {
        host = `${hostname}:8080`;
      }
    }
    const wsUrl = `${protocol}//${host}/api/v1/ws${kawasanId ? `?kawasan_id=${encodeURIComponent(kawasanId)}` : ""}`;

    let socket: WebSocket | null = null;
    let pingInterval: NodeJS.Timeout | null = null;

    try {
      socket = new WebSocket(wsUrl);

      socket.onopen = () => {
        setIsConnected(true);
        setIsWebSocketActive(true);

        // Ping keepalive every 15s
        pingInterval = setInterval(() => {
          if (socket?.readyState === WebSocket.OPEN) {
            socket.send("ping");
          }
        }, 15000);
      };

      socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as RealtimeEvent;
          if (data.type === "pong") return;

          handleRealtimeEvent(data);
        } catch {}
      };

      socket.onerror = () => {
        setIsWebSocketActive(false);
      };

      socket.onclose = () => {
        setIsConnected(false);
        setIsWebSocketActive(false);
      };
    } catch {}

    return () => {
      if (pingInterval) clearInterval(pingInterval);
      if (socket) {
		socket.onopen = null;
		socket.onmessage = null;
		socket.onerror = null;
		socket.onclose = null;
        socket.close();
      }
    };
  }, [isLoggedIn, kawasanId, queryClient]);

  // 2. Fallback HTTP Polling (Active ONLY if WebSocket is down & Tab is Visible)
  useEffect(() => {
    // If WebSocket is actively pushing events, SKIP HTTP polling 100%!
    if (isWebSocketActive || !isLoggedIn) return;
	if (lastPollTimeRef.current === 0) {
		lastPollTimeRef.current = Date.now() - 10000;
	}

    let isMounted = true;

    const poll = async () => {
      // Skip polling if document/tab is hidden/inactive
      if (typeof document !== "undefined" && document.visibilityState !== "visible") {
        return;
      }

      try {
        const since = lastPollTimeRef.current;
        const res = await api.get(`/inspeksi/sync?since=${since}${kawasanId ? `&kawasan_id=${kawasanId}` : ""}`);

        if (!isMounted) return;

        if (res.data?.data?.server_time) {
          lastPollTimeRef.current = res.data.data.server_time;
        }
      } catch {}
    };

    poll();
    const timer = setInterval(poll, pollIntervalMs);

    return () => {
      isMounted = false;
      clearInterval(timer);
    };
  }, [isWebSocketActive, isLoggedIn, kawasanId, pollIntervalMs]);

  return {
    isConnected,
    isWebSocketActive,
  };
}
