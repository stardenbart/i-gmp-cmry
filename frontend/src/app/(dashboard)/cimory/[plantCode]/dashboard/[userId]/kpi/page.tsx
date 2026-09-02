"use client";

import { lazy, Suspense } from "react";
import { Loader2 } from "lucide-react";

const DashboardPanelKPI = lazy(() =>
  import("@/components/layout/DashboardKPI").then((mod) => ({ default: mod.DashboardPanelKPI }))
);

export default function KPIDashboardPage() {
  const LoadingFallback = (
    <div className="h-[50vh] w-full flex items-center justify-center bg-background">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
    </div>
  );

  return (
    <Suspense fallback={LoadingFallback}>
      <DashboardPanelKPI />
    </Suspense>
  );
}
