"use client";

import { lazy, Suspense } from "react";
import { useMounted } from "@/lib/useMounted";
import { Loader2 } from "lucide-react";
import { isAdminUser, isAuditorUser } from "@/lib/useAdminGuard";
import { useAuthStore } from "@/stores/authStore";

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
  const user = useAuthStore((state) => state.user);

  const LoadingFallback = (
    <div className="h-[50vh] w-full flex items-center justify-center bg-background">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
    </div>
  );

  if (!mounted || !user) return LoadingFallback;

  const isAdmin = isAdminUser(user.role_id, user.role?.role_name);
  const isAuditor = isAuditorUser(user.role_id, user.role?.role_name, user.username);

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
