"use client";

import dynamic from "next/dynamic";
import { useEffect, useRef, useState } from "react";
import { Zap } from "lucide-react";

import { isAdminUser } from "@/lib/useAdminGuard";
import { useMounted } from "@/lib/useMounted";
import { useAuthStore } from "@/stores/authStore";
import { useSettingsStore } from "@/stores/settingsStore";

const SearchLatencyModal = dynamic(
  () => import("@/components/performance/SearchLatencyModal").then((module) => module.SearchLatencyModal),
  { ssr: false }
);

interface SearchLatencyBadgeProps {
  searchQuery: string;
  isFetching: boolean;
  pageName: string;
  apiPath: string;
}

export function SearchLatencyBadge({ searchQuery, isFetching, pageName, apiPath }: SearchLatencyBadgeProps) {
  const [latencyMs, setLatencyMs] = useState<number | null>(null);
  const [isTesterOpen, setIsTesterOpen] = useState(false);
  const startTimeRef = useRef<number | null>(null);
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);
  const showButton = useSettingsStore((state) => state.showSearchLatencyButton);
  const canView = !!user && isAdminUser(user.role_id, user.role?.role_name);

  useEffect(() => {
    if (!canView || !showButton || document.hidden) {
      startTimeRef.current = null;
      return;
    }
    if (isFetching && startTimeRef.current === null) {
      startTimeRef.current = performance.now();
      return;
    }
    if (!isFetching && startTimeRef.current !== null) {
      const elapsed = Math.round(performance.now() - startTimeRef.current);
      startTimeRef.current = null;
      const frame = requestAnimationFrame(() => setLatencyMs(elapsed));
      return () => cancelAnimationFrame(frame);
    }
  }, [canView, isFetching, searchQuery, showButton]);

  if (!mounted || !showButton || !canView) return null;

  const latencyClass = latencyMs === null
    ? ""
    : latencyMs < 300
      ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
      : latencyMs < 800
        ? "border-amber-500/30 bg-amber-500/10 text-amber-600 dark:text-amber-400"
        : "border-rose-500/30 bg-rose-500/10 text-rose-600 dark:text-rose-400";

  return (
    <div className="inline-flex shrink-0 items-center gap-2">
      {latencyMs !== null && (
        <span className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-semibold ${latencyClass}`}>
          <Zap className="h-3.5 w-3.5 shrink-0" />
          {latencyMs} ms
        </span>
      )}
      <button
        type="button"
        onClick={() => setIsTesterOpen(true)}
        className="inline-flex shrink-0 items-center gap-1.5 rounded-xl border border-primary/20 bg-primary/10 px-3 py-1.5 text-xs font-semibold text-primary transition-colors hover:bg-primary/20 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
        title={`Uji Latensi Pencarian ${pageName}`}
      >
        <Zap className="h-3.5 w-3.5" />
        <span>Uji Latensi Search</span>
      </button>
      {isTesterOpen && (
        <SearchLatencyModal pageName={pageName} apiPath={apiPath} onClose={() => setIsTesterOpen(false)} />
      )}
    </div>
  );
}
