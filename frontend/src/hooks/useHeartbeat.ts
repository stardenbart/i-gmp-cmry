"use client";

import { useEffect, useRef } from "react";
import { api } from "@/lib/api/axios";
import { toast } from "sonner";

interface UseHeartbeatOptions {
  kawasanId: string;
  aspekId: string;
  lockToken: string | null;
  enabled: boolean;
  onLockExpired?: () => void;
}

export function useHeartbeat({ kawasanId, aspekId, lockToken, enabled, onLockExpired }: UseHeartbeatOptions) {
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    if (!enabled || !lockToken || !kawasanId || !aspekId) {
      if (timerRef.current) clearInterval(timerRef.current);
      return;
    }

    const sendHeartbeat = async () => {
      // Skip heartbeat if tab is inactive
      if (typeof document !== "undefined" && document.visibilityState !== "visible") {
        return;
      }

      try {
        await api.put(
          `/inspeksi/${kawasanId}/${aspekId}/lock/heartbeat`,
          {},
          { headers: { "X-Lock-Token": lockToken } }
        );
      } catch (err: any) {
        if (err?.response?.status === 409 || err?.response?.status === 401) {
          if (onLockExpired) onLockExpired();
        }
      }
    };

    // Send heartbeat every 10 seconds
    timerRef.current = setInterval(sendHeartbeat, 10_000);

    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [kawasanId, aspekId, lockToken, enabled, onLockExpired]);
}
