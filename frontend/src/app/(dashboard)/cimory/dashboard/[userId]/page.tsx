"use client";

import { lazy, Suspense } from "react";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { Loader2 } from "lucide-react";

import { isAdminUser, isAuditorUser } from "@/lib/useAdminGuard";

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
  const user = useAuthStore((state) => state.user);
  const mounted = useMounted();

  const LoadingFallback = (
    <div className="h-[50vh] w-full flex items-center justify-center bg-background">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
    </div>
  );

  if (!mounted) return LoadingFallback;

  const isAdmin = isAdminUser(user?.role_id);
  const isAuditor = isAuditorUser(user?.role_id);

  return (
    <Suspense fallback={LoadingFallback}>
      {isAdmin ? (
        <DashboardPanelAdmin />
      ) : isAuditor ? (
        <DashboardPanelAuditor />
      ) : (
        <DashboardPanelAuditee />
      )}
    </Suspense>
  );
}