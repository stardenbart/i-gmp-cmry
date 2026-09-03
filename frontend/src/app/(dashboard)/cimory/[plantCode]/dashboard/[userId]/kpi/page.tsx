"use client";

import { lazy, Suspense } from "react";
import { Loader2 } from "lucide-react";
import { usePermissions } from "@/lib/usePermissions";

const DashboardPanelKPI = lazy(() =>
  import("@/components/layout/DashboardKPI").then((mod) => ({ default: mod.DashboardPanelKPI }))
);

export default function KPIDashboardPage() {
  const { hasPermission, isLoading: isGuardLoading } = usePermissions();
  const canViewKPI = hasPermission("PERM-KPI-R");

  const LoadingFallback = (
    <div className="h-[50vh] w-full flex items-center justify-center bg-background">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
    </div>
  );

  if (isGuardLoading) {
    return LoadingFallback;
  }

  if (!isGuardLoading && !canViewKPI) return null;

  return (
    <Suspense fallback={LoadingFallback}>
      <DashboardPanelKPI />
    </Suspense>
  );
}
