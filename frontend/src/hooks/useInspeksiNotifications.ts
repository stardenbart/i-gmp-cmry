"use client";

import { useEffect, useRef } from "react";
import { toast } from "sonner";
import { useAuthStore } from "@/stores/authStore";
import { api } from "@/lib/api/axios";

export interface InspeksiNotificationEvent {
  type: "SAVE_SUCCESS" | "SAVE_FAILED" | "LOCK_EXPIRED" | "KAWASAN_DONE" | "YIELD_REQUEST";
  aspek_id?: string;
  message?: string;
  payload?: any;
  timestamp: number;
}

interface UserNotificationPollResponse {
  server_time: number;
  user_id: string;
  notifications: InspeksiNotificationEvent[];
}

export function useInspeksiNotifications(
  onYieldRequest?: (payload: any) => void,
  intervalMs = 2000
) {
  const user = useAuthStore((state) => state.user);
  const lastPollTimeRef = useRef<number>(Date.now() - 10000);
  const onYieldRequestRef = useRef(onYieldRequest);
  onYieldRequestRef.current = onYieldRequest;

  useEffect(() => {
    if (!user?.id) return;

    let isMounted = true;

    const poll = async () => {
      // Pause polling if tab is inactive
      if (typeof document !== "undefined" && document.visibilityState !== "visible") {
        return;
      }

      try {
        const since = lastPollTimeRef.current;
        const res = await api.get<UserNotificationPollResponse>(
          `/inspeksi/notifications/poll?user_id=${user.id}&since=${since}`
        );

        if (!isMounted) return;

        if (res.data?.server_time) {
          lastPollTimeRef.current = res.data.server_time;
        }

        const notifications = res.data?.notifications || [];
        notifications.forEach((n) => {
          if (n.type === "SAVE_FAILED") {
            toast.error(n.message || "Gagal menyimpan draft aspek");
          } else if (n.type === "LOCK_EXPIRED") {
            toast.warning("Lock aspek Anda telah kadaluarsa.");
          } else if (n.type === "KAWASAN_DONE") {
            toast.success(n.message || "Kawasan telah selesai dikerjakan & tersimpan!");
          } else if (n.type === "YIELD_REQUEST") {
            toast.info("User lain meminta giliran untuk mengisi aspek yang sedang Anda edit.", {
              duration: 10000,
            });
            if (onYieldRequestRef.current) {
              onYieldRequestRef.current(n.payload);
            }
          }
        });
      } catch (err) {
        // Silent catch
      }
    };

    poll();
    const timer = setInterval(poll, intervalMs);

    return () => {
      isMounted = false;
      clearInterval(timer);
    };
  }, [user?.id, intervalMs]);
}
