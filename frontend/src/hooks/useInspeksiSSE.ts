"use client";

import { useEffect } from "react";
import { toast } from "sonner";
import { useAuthStore } from "@/stores/authStore";

export interface InspeksiSSEEvent {
  type: "SAVE_SUCCESS" | "SAVE_FAILED" | "LOCK_EXPIRED" | "KAWASAN_DONE" | "YIELD_REQUEST";
  aspek_id?: string;
  message?: string;
  payload?: any;
}

export function useInspeksiSSE(onYieldRequest?: (payload: any) => void) {
  const user = useAuthStore((state) => state.user);

  useEffect(() => {
    if (!user?.id) return;

    const sseUrl = `/api/v1/sse/inspeksi/notifications?user_id=${user.id}`;
    const eventSource = new EventSource(sseUrl);

    eventSource.addEventListener("SAVE_SUCCESS", (e: MessageEvent) => {
      // Quiet save success feedback or subtle indicator
    });

    eventSource.addEventListener("SAVE_FAILED", (e: MessageEvent) => {
      const data = JSON.parse(e.data);
      toast.error(data.message || "Gagal menyimpan draft aspek");
    });

    eventSource.addEventListener("LOCK_EXPIRED", (e: MessageEvent) => {
      toast.warning("Lock aspek Anda telah kadaluarsa.");
    });

    eventSource.addEventListener("KAWASAN_DONE", (e: MessageEvent) => {
      const data = JSON.parse(e.data);
      toast.success(data.message || "Kawasan telah selesai dikerjakan & tersimpan!");
    });

    eventSource.addEventListener("YIELD_REQUEST", (e: MessageEvent) => {
      try {
        const data = JSON.parse(e.data);
        toast.info("User lain meminta giliran untuk mengisi aspek yang sedang Anda edit.", {
          duration: 10000,
        });
        if (onYieldRequest) onYieldRequest(data.payload);
      } catch (err) {
        console.error("Failed to parse YIELD_REQUEST:", err);
      }
    });

    return () => {
      eventSource.close();
    };
  }, [user?.id, onYieldRequest]);
}
