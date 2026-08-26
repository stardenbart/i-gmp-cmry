"use client";

import { lazy, Suspense } from "react";
import { useMounted } from "@/lib/useMounted";
import { Loader2 } from "lucide-react";

import { usePermissions } from "@/lib/usePermissions";

// Code splitting with lazy
const DashboardPanelAdmin = lazy(() => 
  import("@/components/layout/DashboardAdmin").then((mod) => ({ default: mod.DashboardPanelAdmin }))
);
const DashboardPanelAuditor = lazy(() => 
  import("@/components/layout/DashboardAuditor").then((mod) => ({ default: mod.DashboardPanelAuditor }))
);
const DashboardPanelAuditee = lazy(() => 
  import("@/components/layout/DashboardAuditee").then((mod) => ({ default: mod.DashboardPanelAuditee }))
);

export default function DashboardPage() {
  const mounted = useMounted();
  const { hasPermission, isLoading: isPermLoading } = usePermissions();

  const LoadingFallback = (
    <div className="h-[50vh] w-full flex items-center justify-center bg-background">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
    </div>
  );

  if (!mounted || isPermLoading) return LoadingFallback;

  const canManageSystem = hasPermission("PERM-USR-R") || hasPermission("PERM-MSTR-R");
  const canCreateInspection = hasPermission("PERM-INSP-C") || hasPermission("PERM-INSP-W");

  return (
    <Suspense fallback={LoadingFallback}>
      {canManageSystem ? (
        <DashboardPanelAdmin />
      ) : canCreateInspection ? (
        <DashboardPanelAuditor />
      ) : (
        <DashboardPanelAuditee />
      )}
    </Suspense>
  );
}
