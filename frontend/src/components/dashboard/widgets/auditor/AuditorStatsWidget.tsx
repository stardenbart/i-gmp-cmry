"use client";

import { Loader2 } from "lucide-react";
import Link from "next/link";
import { useAuditorDashboard } from "@/components/dashboard/auditor/AuditorDashboardContext";

export function AuditorStatsWidget() {
  const { isLoading, user, totalInspections, ongoingInspections, completedInspections, pendingValidationCount, openIssues, totalIssues } =
    useAuditorDashboard();

  return (
    <section className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-4">
      <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden group">
        <div className="absolute top-0 left-0 w-full h-1 bg-primary" />
        <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Total Inspeksi</h3>
        <p className="text-2xl font-bold text-foreground">{isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : totalInspections}</p>
      </div>

      <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
        <div className="absolute top-0 left-0 w-full h-1 bg-amber-500" />
        <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Berlangsung</h3>
        <p className="text-2xl font-bold text-foreground">{isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : ongoingInspections}</p>
      </div>

      <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
        <div className="absolute top-0 left-0 w-full h-1 bg-green-600" />
        <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Selesai</h3>
        <p className="text-2xl font-bold text-foreground">{isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : completedInspections}</p>
      </div>

      <Link href={`/cimory/${user?.plant_id || "global"}/dashboard/${user?.id || "overview"}/issues?status=PendingValidation`}>
        <div className="bg-card rounded-xl p-4 border border-purple-500/30 shadow-sm hover:shadow-md hover:border-purple-500/60 transition-all relative overflow-hidden cursor-pointer group">
          <div className="absolute top-0 left-0 w-full h-1 bg-purple-500" />
          {pendingValidationCount > 0 && (
            <span className="absolute top-3 right-3 flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-purple-400 opacity-75" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-purple-500" />
            </span>
          )}
          <h3 className="text-xs font-semibold text-purple-600 dark:text-purple-400 mb-2 uppercase tracking-wider">Menunggu Validasi</h3>
          <p className="text-2xl font-bold text-purple-600 dark:text-purple-300">
            {isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : pendingValidationCount}
          </p>
        </div>
      </Link>

      <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
        <div className="absolute top-0 left-0 w-full h-1 bg-red-500" />
        <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Temuan Terbuka</h3>
        <p className="text-2xl font-bold text-foreground">{isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : openIssues}</p>
      </div>

      <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
        <div className="absolute top-0 left-0 w-full h-1 bg-violet-500" />
        <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Total Issue (Foto)</h3>
        <p className="text-2xl font-bold text-foreground">{isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : totalIssues}</p>
      </div>
    </section>
  );
}
