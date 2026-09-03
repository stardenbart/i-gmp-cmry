"use client";

import { LayoutGrid } from "lucide-react";
import { DashboardGrid } from "@/components/dashboard/DashboardGrid";
import { allMainDashboardWidgets } from "@/components/dashboard/widgets/registry";

interface UserDashboardLayoutTabProps {
  userId: string;
  roleId: string;
  username?: string;
}

/**
 * Dashboard customization lives here — on the Edit User screen, Admin/Super
 * Admin only — rather than as self-service on each user's own dashboard.
 * Every main-dashboard widget across every role is offered here regardless
 * of the target user's own role (allMainDashboardWidgets) — a widget only
 * ever ends up VISIBLE on that user's real dashboard once Admin explicitly
 * adds it here and saves (see DashboardGrid.tsx's resolveWidgets: a widget
 * with no saved config defaults to hidden, not shown), so mixing widgets
 * across roles is entirely opt-in per user, controlled from this one place.
 */
export function UserDashboardLayoutTab({ userId, roleId }: UserDashboardLayoutTabProps) {
  return (
    <div className="space-y-3">
      <div className="flex items-start gap-2.5 p-3 rounded-xl bg-primary/5 border border-primary/20 text-xs text-muted-foreground">
        <LayoutGrid className="h-4 w-4 text-primary shrink-0 mt-0.5" />
        <p>
          Susunan ini akan langsung berlaku di dashboard milik pengguna ini. Pengguna sendiri tidak dapat mengubah
          tata letak dashboardnya — hanya Admin/Super Admin yang dapat mengaturnya di sini. Kotak di bawah adalah
          placeholder posisi widget (bukan data live pengguna tersebut).
        </p>
      </div>
      <DashboardGrid registry={allMainDashboardWidgets} enabled={!!userId} target={{ userId, roleId }} editable previewOnly />
    </div>
  );
}
