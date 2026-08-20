"use client";

import { useQuery } from "@tanstack/react-query";
import { Lock, Unlock } from "lucide-react";
import { inspectionApi } from "@/lib/api/inspection.api";
import { useRealtimeSync } from "@/hooks/useRealtimeSync";
import { useCallback } from "react";

interface AspekStatusPanelProps {
  kawasanId: string;
  aspeks?: Array<{ aspek_id: string; aspek_name: string }>;
  currentUserId?: string;
  activeAspekId?: string;
  onSelectAspek?: (aspekId: string) => void;
}

export function AspekStatusPanel({
  kawasanId,
  aspeks = [],
  currentUserId,
  activeAspekId,
  onSelectAspek,
}: AspekStatusPanelProps) {
  const { data: statusRes, refetch } = useQuery({
    queryKey: ["kawasan_status", kawasanId],
    queryFn: () => inspectionApi.getKawasanStatus(kawasanId),
    enabled: !!kawasanId,
  });

  const handleEvent = useCallback(
    (_ev: any) => {
      refetch();
    },
    [refetch]
  );

  const { isWebSocketActive } = useRealtimeSync({
    kawasanId,
    onEvent: handleEvent,
  });

  const lockStatuses: Array<{ aspek_id: string; status: string; locked_by?: string }> =
    statusRes?.data || [];

  const statusMap = new Map<string, { status: string; lockedBy?: string }>();
  lockStatuses.forEach((s) => {
    statusMap.set(s.aspek_id, { status: s.status, lockedBy: s.locked_by });
  });

  return (
    <div className="bg-card border border-border/80 rounded-2xl p-3 sm:p-4 mb-3 sm:mb-6 shadow-xs">
      <div className="flex items-center justify-between gap-2 mb-2 sm:mb-3 border-b border-border/50 pb-2">
        <div className="flex items-center gap-1.5 min-w-0">
          <Lock className="w-3.5 h-3.5 text-primary shrink-0" />
          <h4 className="text-xs sm:text-sm font-bold text-foreground tracking-tight truncate">
            Status Lock Real-Time
          </h4>
        </div>
        <span className={`text-[10px] font-semibold px-2 py-0.5 rounded-full shrink-0 ${isWebSocketActive ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20" : "bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20"}`}>
          {isWebSocketActive ? "Live" : "Sync"}
        </span>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2 sm:gap-2.5">
        {aspeks.map((aspek) => {
          const info = statusMap.get(aspek.aspek_id);
          const isLocked = info?.status === "LOCKED";
          const lockedBy = info?.lockedBy;
          const isLockedByMe = isLocked && lockedBy === currentUserId;
          const isActiveTab = aspek.aspek_id === activeAspekId;

          let badgeText = "FREE";

          if (isLockedByMe) {
            badgeText = "Aktif Mengedit";
          } else if (isLocked) {
            badgeText = `Dikunci (${lockedBy ? lockedBy.substring(0, 6) : "Lain"})`;
          }

          return (
            <button
              key={aspek.aspek_id}
              onClick={() => onSelectAspek?.(aspek.aspek_id)}
              className={`flex items-start justify-between p-2 sm:p-2.5 rounded-xl border text-left transition-all ${
                isActiveTab
                  ? "border-primary bg-primary/10 ring-1 ring-primary/40 shadow-xs"
                  : "border-border/70 hover:bg-accent/40"
              }`}
            >
              <div className="min-w-0 flex-1 mr-1">
                <p className="text-[11px] sm:text-xs font-semibold text-foreground truncate">{aspek.aspek_name}</p>
                <div className="flex items-center gap-1 mt-0.5">
                  <span className={`text-[9px] sm:text-[10px] font-mono font-medium truncate ${isLockedByMe ? "text-emerald-600 dark:text-emerald-400 font-bold" : isLocked ? "text-red-500 font-bold" : "text-muted-foreground"}`}>
                    {badgeText}
                  </span>
                </div>
              </div>
              {isLockedByMe && (
                <span className="text-[9px] py-0.5 px-1.5 rounded-full bg-emerald-600 text-white font-bold shrink-0">
                  ●
                </span>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}
