"use client";

import { useQuery } from "@tanstack/react-query";
import { Lock, Unlock } from "lucide-react";
import { inspectionApi } from "@/lib/api/inspection.api";
import { useKawasanPolling, AspekKawasanEvent } from "@/hooks/useKawasanPolling";
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
    refetchInterval: 5000, // Fallback poll every 5s
  });

  const handleKawasanEvent = useCallback(
    (_ev: AspekKawasanEvent) => {
      refetch();
    },
    [refetch]
  );

  useKawasanPolling(kawasanId, handleKawasanEvent, 3000);

  const lockStatuses: Array<{ aspek_id: string; status: string; locked_by?: string }> =
    statusRes?.data || [];

  const statusMap = new Map<string, { status: string; lockedBy?: string }>();
  lockStatuses.forEach((s) => {
    statusMap.set(s.aspek_id, { status: s.status, lockedBy: s.locked_by });
  });

  return (
    <div className="bg-card border rounded-lg p-4 mb-6 shadow-sm">
      <div className="flex items-center justify-between mb-3">
        <h4 className="text-sm font-semibold flex items-center gap-2 text-foreground">
          Status Real-Time Aspek Audit (Distributed Lock)
        </h4>
        <span className="text-xs text-muted-foreground">
          Diperbarui secara otomatis via Event Store
        </span>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2.5">
        {aspeks.map((aspek) => {
          const info = statusMap.get(aspek.aspek_id);
          const isLocked = info?.status === "LOCKED";
          const lockedBy = info?.lockedBy;
          const isLockedByMe = isLocked && lockedBy === currentUserId;
          const isActiveTab = aspek.aspek_id === activeAspekId;

          let badgeVariant: "default" | "secondary" | "destructive" | "outline" = "outline";
          let badgeText = "FREE";
        

          if (isLockedByMe) {
            badgeVariant = "default";
            badgeText = "Anda Sedang Mengedit";
          } else if (isLocked) {
            badgeVariant = "destructive";
            badgeText = `Dikunci oleh ${lockedBy ? lockedBy.substring(0, 8) : "Auditor Lain"}`;
          }

          return (
            <button
              key={aspek.aspek_id}
              onClick={() => onSelectAspek?.(aspek.aspek_id)}
              className={`flex items-start justify-between p-2.5 rounded-md border text-left transition-all ${
                isActiveTab
                  ? "border-primary bg-primary/5 ring-1 ring-primary"
                  : "border-border hover:bg-accent/50"
              }`}
            >
              <div className="min-w-0 flex-1 mr-2">
                <p className="text-xs font-medium text-foreground truncate">{aspek.aspek_name}</p>
                <div className="flex items-center gap-1 mt-1">
                 
                  <span className="text-[11px] font-mono text-muted-foreground truncate">
                    {badgeText}
                  </span>
                </div>
              </div>
              {isLockedByMe && (
                <span className="text-[10px] py-0.5 px-1.5 rounded bg-emerald-600 text-white font-medium shrink-0">
                  Aktif
                </span>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}
