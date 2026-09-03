"use client";

import { RefreshCw } from "lucide-react";
import { cn } from "@/lib/utils";
import { DashboardGrid } from "@/components/dashboard/DashboardGrid";
import { allMainDashboardWidgets } from "@/components/dashboard/widgets/registry";
import { AuditeeDashboardProvider, useAuditeeDashboard } from "@/components/dashboard/auditee/AuditeeDashboardContext";

function AuditeeDashboardBody() {
  const { mounted, user, isFetching, handleRefresh } = useAuditeeDashboard();

  if (!mounted) {
    return <div className="h-screen w-full bg-background" />;
  }

  return (
    <div className="w-full space-y-8 animate-in fade-in duration-500 pb-24 md:pb-6">
      {/* Header Section */}
      <section className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <h1 className="text-3xl md:text-4xl font-bold tracking-tight text-foreground">
              Selamat Datang, {user?.name || "Penanggung Jawab"}
            </h1>
            <span className="bg-primary/10 text-primary text-xs font-semibold px-2.5 py-1 rounded uppercase tracking-wider">
              PENANGGUNG JAWAB (PIC)
            </span>
          </div>
          <p className="text-muted-foreground text-sm">
            Pengawasan Tugas · Pantau temuan dan tindakan perbaikan yang ditugaskan kepada Anda
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleRefresh}
            disabled={isFetching}
            type="button"
            aria-label="Segarkan data dashboard"
            title="Segarkan data"
            className="flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground hover:bg-muted rounded-lg transition-colors disabled:opacity-50"
          >
            <RefreshCw aria-hidden="true" className={cn("h-4 w-4", isFetching && "animate-spin motion-reduce:animate-none")} />
          </button>
        </div>
      </section>

      <DashboardGrid registry={allMainDashboardWidgets} enabled={mounted && !!user} />
    </div>
  );
}

export const DashboardPanelAuditee = () => {
  return (
    <AuditeeDashboardProvider>
      <AuditeeDashboardBody />
    </AuditeeDashboardProvider>
  );
};
