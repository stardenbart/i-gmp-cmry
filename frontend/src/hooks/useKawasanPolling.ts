"use client";

import { useEffect, useRef } from "react";
import { api } from "@/lib/api/axios";

export interface AspekKawasanEvent {
  type: "LOCK_ACQUIRED" | "LOCK_RELEASED" | "LOCK_EXPIRED" | "ASPEK_COMPLETED" | "KAWASAN_SYNCED";
  kawasan_id: string;
  aspek_id?: string;
  user_id?: string;
  user_name?: string;
  payload?: any;
  timestamp: number;
}

interface KawasanPollResponse {
  server_time: number;
  kawasan_id: string;
  events: AspekKawasanEvent[];
}

export function useKawasanPolling(
  kawasanId: string | null,
  onEvent?: (event: AspekKawasanEvent) => void,
  intervalMs = 3000
) {
  const lastPollTimeRef = useRef<number>(Date.now() - 10000);
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent;

  useEffect(() => {
    if (!kawasanId) return;

    let isMounted = true;

    const poll = async () => {
      // Pause polling if tab is inactive
      if (typeof document !== "undefined" && document.visibilityState !== "visible") {
        return;
      }

      try {
        const since = lastPollTimeRef.current;
        const res = await api.get<KawasanPollResponse>(
          `/inspeksi/kawasan/${kawasanId}/poll?since=${since}`
        );

        if (!isMounted) return;

        if (res.data?.server_time) {
          lastPollTimeRef.current = res.data.server_time;
        }

        const events = res.data?.events || [];
        events.forEach((ev) => {
          if (onEventRef.current) {
            onEventRef.current(ev);
          }
        });
      } catch (err) {
        // Silent catch during polling
      }
    };

    poll();
    const timer = setInterval(poll, intervalMs);

    return () => {
      isMounted = false;
      clearInterval(timer);
    };
  }, [kawasanId, intervalMs]);

  return { isConnected: true };
}
