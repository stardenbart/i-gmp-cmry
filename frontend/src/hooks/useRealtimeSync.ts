"use client";

import { useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useAuthStore } from "@/stores/authStore";
import { api } from "@/lib/api/axios";

interface UseRealtimeSyncOptions {
  kawasanId?: string | null;
  onEvent?: (event: any) => void;
  onIssueUpdated?: () => void;
  pollIntervalMs?: number;
}

export function useRealtimeSync(options?: UseRealtimeSyncOptions) {
  const queryClient = useQueryClient();
  const token = useAuthStore((state) => state.token);
  const user = useAuthStore((state) => state.user);
  
  const [isConnected, setIsConnected] = useState(false);
  const [isWebSocketActive, setIsWebSocketActive] = useState(false);

  const optionsRef = useRef(options);
  optionsRef.current = options;

  const lastPollTimeRef = useRef<number>(Date.now() - 10000);
  const wsRef = useRef<WebSocket | null>(null);

  const kawasanId = options?.kawasanId;
  const pollIntervalMs = options?.pollIntervalMs ?? 5000;

  // 1. Primary WebSocket Real-time Push
  useEffect(() => {
    if (typeof window === "undefined" || !token) return;

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = process.env.NEXT_PUBLIC_WS_HOST || window.location.hostname + ":8080";
    const wsUrl = `${protocol}//${host}/api/v1/ws?token=${encodeURIComponent(token)}${
      kawasanId ? `&kawasan_id=${encodeURIComponent(kawasanId)}` : ""
    }`;

    let socket: WebSocket | null = null;
    let pingInterval: NodeJS.Timeout | null = null;

    try {
      socket = new WebSocket(wsUrl);
      wsRef.current = socket;

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
          const data = JSON.parse(event.data);
          if (data.type === "pong") return;

          if (optionsRef.current?.onEvent) {
            optionsRef.current.onEvent(data);
          }

          if (data.type === "ISSUE_UPDATED" || data.type === "INSPECTION_UPDATED") {
            queryClient.invalidateQueries({ queryKey: ["auditee-issues"] });
            queryClient.invalidateQueries({ queryKey: ["wowr-issues"] });
            queryClient.invalidateQueries({ queryKey: ["issues"] });
            queryClient.invalidateQueries({ queryKey: ["issues-filter"] });
            queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] });
            if (optionsRef.current?.onIssueUpdated) {
              optionsRef.current.onIssueUpdated();
            }
          }
        } catch (e) {}
      };

      socket.onerror = () => {
        setIsWebSocketActive(false);
      };

      socket.onclose = () => {
        setIsConnected(false);
        setIsWebSocketActive(false);
      };
    } catch (e) {
      setIsWebSocketActive(false);
    }

    return () => {
      if (pingInterval) clearInterval(pingInterval);
      if (socket) {
        socket.close();
      }
    };
  }, [token, kawasanId, queryClient]);

  // 2. Fallback HTTP Polling (Active ONLY if WebSocket is down & Tab is Visible)
  useEffect(() => {
    // If WebSocket is actively pushing events, SKIP HTTP polling 100%!
    if (isWebSocketActive || !token) return;

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
      } catch (err) {}
    };

    poll();
    const timer = setInterval(poll, pollIntervalMs);

    return () => {
      isMounted = false;
      clearInterval(timer);
    };
  }, [isWebSocketActive, token, kawasanId, pollIntervalMs]);

  return {
    isConnected,
    isWebSocketActive,
  };
}
