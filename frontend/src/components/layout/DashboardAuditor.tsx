"use client";

import { RefreshCw } from "lucide-react";
import { cn } from "@/lib/utils";
import { DashboardGrid } from "@/components/dashboard/DashboardGrid";
import { auditorWidgetRegistry } from "@/components/dashboard/widgets/auditor/registry";
import { AuditorDashboardProvider, useAuditorDashboard } from "@/components/dashboard/auditor/AuditorDashboardContext";

function AuditorDashboardBody() {
  const { mounted, user, isFetching, handleRefresh } = useAuditorDashboard();

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
              Selamat Datang, {user?.name || "Auditor"}
            </h1>
            <span className="bg-primary/10 text-primary text-xs font-semibold px-2.5 py-1 rounded uppercase tracking-wider">AUDITOR</span>
          </div>
          <p className="text-muted-foreground text-sm">Pengawasan Inspeksi · Ringkasan aktivitas pemeriksaan Anda</p>
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

      <DashboardGrid registry={auditorWidgetRegistry} enabled={mounted && !!user} />
    </div>
  );
}

export const DashboardPanelAuditor = () => {
  return (
    <AuditorDashboardProvider>
      <AuditorDashboardBody />
    </AuditorDashboardProvider>
  );
};
