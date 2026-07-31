"use client";

import { useEffect, useRef, useState } from "react";

export interface AspekWSEvent {
  type: "LOCK_ACQUIRED" | "LOCK_RELEASED" | "LOCK_EXPIRED" | "ASPEK_COMPLETED" | "KAWASAN_SYNCED";
  kawasan_id: string;
  aspek_id?: string;
  user_id?: string;
  user_name?: string;
  payload?: any;
}

export function useInspeksiWebSocket(kawasanId: string | null, onEvent?: (event: AspekWSEvent) => void) {
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!kawasanId) return;

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/ws/inspeksi/${kawasanId}`;

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => setIsConnected(true);
    ws.onclose = () => setIsConnected(false);
    ws.onerror = (err) => console.error("[WebSocket] Error:", err);

    ws.onmessage = (e) => {
      try {
        const event: AspekWSEvent = JSON.parse(e.data);
        if (onEvent) onEvent(event);
      } catch (err) {
        console.error("[WebSocket] Failed to parse message:", err);
      }
    };

    return () => {
      ws.close();
    };
  }, [kawasanId, onEvent]);

  return { isConnected };
}
