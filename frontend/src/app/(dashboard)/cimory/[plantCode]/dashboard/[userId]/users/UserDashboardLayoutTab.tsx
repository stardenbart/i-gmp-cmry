"use client";

import { LayoutGrid } from "lucide-react";
import { DashboardGrid } from "@/components/dashboard/DashboardGrid";
import { adminWidgetRegistry } from "@/components/dashboard/widgets/admin/registry";
import { auditorWidgetRegistry } from "@/components/dashboard/widgets/auditor/registry";
import { auditeeWidgetRegistry } from "@/components/dashboard/widgets/auditee/registry";
import { isAdminUser, isAuditorUser } from "@/lib/useAdminGuard";

interface UserDashboardLayoutTabProps {
  userId: string;
  roleId: string;
  username?: string;
}

/**
 * Dashboard customization lives here — on the Edit User screen, Admin/Super
 * Admin only — rather than as self-service on each user's own dashboard.
 * Which registry applies mirrors exactly how that user's own dashboard page
 * picks a view (see dashboard/[userId]/page.tsx): Admin roles get the admin
 * widgets, Auditor roles get the auditor widgets, everyone else (Auditee,
 * Supervisor, Manager, Staff) gets the auditee widgets.
 */
export function UserDashboardLayoutTab({ userId, roleId, username }: UserDashboardLayoutTabProps) {
  const registry = isAdminUser(roleId)
    ? adminWidgetRegistry
    : isAuditorUser(roleId, undefined, username)
      ? auditorWidgetRegistry
      : auditeeWidgetRegistry;

  return (
    <div className="space-y-3">
      <div className="flex items-start gap-2.5 p-3 rounded-xl bg-primary/5 border border-primary/20 text-xs text-muted-foreground">
        <LayoutGrid className="h-4 w-4 text-primary shrink-0 mt-0.5" />
        <p>
          Susunan ini akan langsung berlaku di dashboard milik pengguna ini. Pengguna sendiri tidak dapat mengubah
          tata letak dashboardnya — hanya Admin/Super Admin yang dapat mengaturnya di sini.
        </p>
      </div>
      <DashboardGrid registry={registry} enabled={!!userId} target={{ userId, roleId }} editable />
    </div>
  );
}
