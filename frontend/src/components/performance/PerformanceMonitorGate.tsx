"use client";

import dynamic from "next/dynamic";

import { isAdminUser } from "@/lib/useAdminGuard";
import { useMounted } from "@/lib/useMounted";
import { useAuthStore } from "@/stores/authStore";
import { useSettingsStore } from "@/stores/settingsStore";

const CoreWebVitalsOverlay = dynamic(
  () => import("@/components/ui/CoreWebVitalsOverlay").then((module) => module.CoreWebVitalsOverlay),
  { ssr: false }
);

export function PerformanceMonitorGate() {
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);
  const enabled = useSettingsStore((state) => state.showCoreWebVitalsMonitor);
  const isAdmin = !!user && isAdminUser(user.role_id, user.role?.role_name);

  if (!mounted || !enabled || !isAdmin) return null;
  return <CoreWebVitalsOverlay />;
}
