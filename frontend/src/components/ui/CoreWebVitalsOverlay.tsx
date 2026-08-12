"use client";

import { useState, useEffect } from "react";
import { useReportWebVitals } from "next/web-vitals";
import { Activity, Gauge, ChevronUp, ChevronDown, CheckCircle2, AlertTriangle, XCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/authStore";
import { useSettingsStore } from "@/stores/settingsStore";
import { useMounted } from "@/lib/useMounted";

interface MetricState {
  value: number;
  rating: "good" | "needs-improvement" | "poor";
}

export function CoreWebVitalsOverlay() {
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);
  const showCoreWebVitalsMonitor = useSettingsStore((state) => state.showCoreWebVitalsMonitor);

  const [isOpen, setIsOpen] = useState(false);
  const [metrics, setMetrics] = useState<Record<string, MetricState>>({
    LCP: { value: 0, rating: "good" },
    CLS: { value: 0, rating: "good" },
    INP: { value: 0, rating: "good" },
    FCP: { value: 0, rating: "good" },
    TTFB: { value: 0, rating: "good" },
  });

  const isSuperAdmin =
    user?.role_id === "ROLE-000" ||
    user?.role_id === "SUPERADMIN" ||
    user?.role?.role_name === "Super Admin" ||
    !user?.plant_id;

  const isAdmin =
    isSuperAdmin ||
    user?.role_id === "ROLE-001" ||
    user?.role_id === "ROLE-002" ||
    user?.role_id === "ADMIN" ||
    user?.role?.role_name === "Administrator";

  useReportWebVitals((metric) => {
    const { name, value, rating } = metric;
    setMetrics((prev) => ({
      ...prev,
      [name]: {
        value: name === "CLS" ? Number(value.toFixed(3)) : Math.round(value),
        rating: rating as "good" | "needs-improvement" | "poor",
      },
    }));
  });

  // Client-side PerformanceObserver fallback for real-time local testing
  useEffect(() => {
    if (typeof window === "undefined" || !("PerformanceObserver" in window)) return;

    try {
      const lcpObs = new PerformanceObserver((entryList) => {
        const entries = entryList.getEntries();
        const lastEntry = entries[entries.length - 1];
        if (lastEntry) {
          const value = Math.round(lastEntry.startTime);
          const rating = value <= 2500 ? "good" : value <= 4000 ? "needs-improvement" : "poor";
          setMetrics((prev) => ({ ...prev, LCP: { value, rating } }));
        }
      });
      lcpObs.observe({ type: "largest-contentful-paint", buffered: true });

      let clsValue = 0;
      const clsObs = new PerformanceObserver((entryList) => {
        for (const entry of entryList.getEntries() as any[]) {
          if (!entry.hadRecentInput) {
            clsValue += entry.value;
            const rating = clsValue <= 0.1 ? "good" : clsValue <= 0.25 ? "needs-improvement" : "poor";
            setMetrics((prev) => ({ ...prev, CLS: { value: Number(clsValue.toFixed(3)), rating } }));
          }
        }
      });
      clsObs.observe({ type: "layout-shift", buffered: true });

      return () => {
        lcpObs.disconnect();
        clsObs.disconnect();
      };
    } catch (e) {
      // Browser safety fallback
    }
  }, []);

  if (!mounted || !user || !isAdmin || !showCoreWebVitalsMonitor) {
    return null;
  }

  const getRatingBadge = (rating: "good" | "needs-improvement" | "poor") => {
    switch (rating) {
      case "good":
        return {
          bg: "bg-emerald-500/10 text-emerald-400 border-emerald-500/30",
          icon: CheckCircle2,
          label: "Good",
        };
      case "needs-improvement":
        return {
          bg: "bg-amber-500/10 text-amber-400 border-amber-500/30",
          icon: AlertTriangle,
          label: "Needs Imp.",
        };
      case "poor":
        return {
          bg: "bg-red-500/10 text-red-400 border-red-500/30",
          icon: XCircle,
          label: "Poor",
        };
    }
  };

  const getOverallStatus = () => {
    const ratings = Object.values(metrics).map((m) => m.rating);
    if (ratings.includes("poor")) return "poor";
    if (ratings.includes("needs-improvement")) return "needs-improvement";
    return "good";
  };

  const overall = getRatingBadge(getOverallStatus());
  const OverallIcon = overall.icon;

  return (
    <div className="fixed bottom-4 left-4 z-50 font-mono text-xs select-none">
      {/* Collapsed Pill */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          "flex items-center gap-2 rounded-full border px-3 py-1.5 shadow-xl backdrop-blur-xl transition-all duration-200 active:scale-95",
          "bg-zinc-950/90 text-zinc-100 border-zinc-800 hover:border-zinc-700",
          overall.bg
        )}
        title="Toggle Core Web Vitals Monitor (Admin Only)"
      >
        <Gauge className="h-3.5 w-3.5" />
        <span className="font-semibold tracking-wider">CWV</span>
        <span className="text-[10px] opacity-80">
          LCP: {metrics.LCP.value > 0 ? `${metrics.LCP.value}ms` : "-"}
        </span>
        <OverallIcon className="h-3.5 w-3.5 shrink-0" />
        {isOpen ? <ChevronDown className="h-3 w-3" /> : <ChevronUp className="h-3 w-3" />}
      </button>

      {/* Expanded Modal Box */}
      {isOpen && (
        <div className="mt-2 w-72 rounded-2xl border border-zinc-800 bg-zinc-950/95 p-4 shadow-2xl backdrop-blur-2xl text-zinc-100 space-y-3 animate-in fade-in slide-in-from-bottom-2 duration-200">
          <div className="flex items-center justify-between border-b border-zinc-800 pb-2">
            <div className="flex items-center gap-1.5">
              <Activity className="h-4 w-4 text-primary" />
              <span className="font-bold text-xs">Core Web Vitals (Admin)</span>
            </div>
            <span className={cn("px-2 py-0.5 rounded-full text-[10px] border font-bold", overall.bg)}>
              {overall.label}
            </span>
          </div>

          <div className="space-y-2">
            {/* LCP */}
            <div className="flex items-center justify-between">
              <div>
                <span className="font-bold block text-xs">LCP (Largest Contentful)</span>
                <span className="text-[9px] text-zinc-400">Target: ≤ 2500ms</span>
              </div>
              <span className={cn("px-2 py-1 rounded-md text-xs font-mono font-bold border", getRatingBadge(metrics.LCP.rating).bg)}>
                {metrics.LCP.value > 0 ? `${metrics.LCP.value} ms` : "Measuring..."}
              </span>
            </div>

            {/* CLS */}
            <div className="flex items-center justify-between border-t border-zinc-900 pt-1.5">
              <div>
                <span className="font-bold block text-xs">CLS (Cumulative Shift)</span>
                <span className="text-[9px] text-zinc-400">Target: ≤ 0.100</span>
              </div>
              <span className={cn("px-2 py-1 rounded-md text-xs font-mono font-bold border", getRatingBadge(metrics.CLS.rating).bg)}>
                {metrics.CLS.value}
              </span>
            </div>

            {/* INP / FID */}
            <div className="flex items-center justify-between border-t border-zinc-900 pt-1.5">
              <div>
                <span className="font-bold block text-xs">INP / FID (Interaction)</span>
                <span className="text-[9px] text-zinc-400">Target: ≤ 200ms</span>
              </div>
              <span className={cn("px-2 py-1 rounded-md text-xs font-mono font-bold border", getRatingBadge(metrics.INP.rating).bg)}>
                {metrics.INP.value > 0 ? `${metrics.INP.value} ms` : "Pending input"}
              </span>
            </div>

            {/* TTFB */}
            <div className="flex items-center justify-between border-t border-zinc-900 pt-1.5">
              <div>
                <span className="font-bold block text-xs">TTFB (Server Response)</span>
                <span className="text-[9px] text-zinc-400">Target: ≤ 800ms</span>
              </div>
              <span className={cn("px-2 py-1 rounded-md text-xs font-mono font-bold border", getRatingBadge(metrics.TTFB.rating).bg)}>
                {metrics.TTFB.value > 0 ? `${metrics.TTFB.value} ms` : "-"}
              </span>
            </div>
          </div>

          <div className="text-[9px] text-zinc-500 border-t border-zinc-800 pt-2 text-center">
            Mode Khusus Admin & SuperAdmin
          </div>
        </div>
      )}
    </div>
  );
}
