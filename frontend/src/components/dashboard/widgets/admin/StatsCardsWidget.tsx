"use client";

import { StatsCards } from "@/components/admin/StatCard";
import { useAdminDashboard } from "@/components/dashboard/admin/AdminDashboardContext";

export function StatsCardsWidget() {
  const { stats, isLoading, isFetching } = useAdminDashboard();
  return (
    <section className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-5 gap-4">
      <StatsCards stats={stats} isLoading={isLoading || isFetching} />
    </section>
  );
}
