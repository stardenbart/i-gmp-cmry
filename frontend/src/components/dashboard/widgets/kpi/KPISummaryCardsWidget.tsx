"use client";

import { StatsCards } from "@/components/admin/StatCard";
import { useKPIDashboard } from "@/components/dashboard/kpi/KPIDashboardContext";

export function KPISummaryCardsWidget() {
  const { stats, isLoading, isFetching } = useKPIDashboard();
  return (
    <section className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-5 gap-4">
      <StatsCards stats={stats} isLoading={isLoading || isFetching} />
    </section>
  );
}
