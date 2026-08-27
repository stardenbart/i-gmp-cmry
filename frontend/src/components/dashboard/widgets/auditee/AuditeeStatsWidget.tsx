"use client";

import { Loader2 } from "lucide-react";
import { useAuditeeDashboard } from "@/components/dashboard/auditee/AuditeeDashboardContext";

export function AuditeeStatsWidget() {
  const { isLoading, totalIssues, openIssues, pendingIssues, closedIssues } = useAuditeeDashboard();

  return (
    <section className="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden group">
        <div className="absolute top-0 left-0 w-full h-1 bg-gray-500" />
        <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Total Tugas</h3>
        <p className="text-2xl font-bold text-foreground">{isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : totalIssues}</p>
      </div>

      <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
        <div className="absolute top-0 left-0 w-full h-1 bg-red-500" />
        <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Tugas Terbuka</h3>
        <p className="text-2xl font-bold text-foreground">{isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : openIssues}</p>
      </div>

      <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
        <div className="absolute top-0 left-0 w-full h-1 bg-amber-500" />
        <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Menunggu Validasi</h3>
        <p className="text-2xl font-bold text-foreground">{isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : pendingIssues}</p>
      </div>

      <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
        <div className="absolute top-0 left-0 w-full h-1 bg-green-600" />
        <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Tugas Selesai</h3>
        <p className="text-2xl font-bold text-foreground">{isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : closedIssues}</p>
      </div>
    </section>
  );
}
