"use client";

import { useState, useCallback, lazy, Suspense } from "react";
import { 
  ClipboardCheck, 
  CheckCircle, 
  AlertTriangle, 
  RefreshCw,
  Loader2,
  ListTodo
} from "lucide-react";

// Lazy load heavy chart components
const AreaChart = lazy(() => import("@/components/ui/area-chart").then(mod => ({ default: mod.AreaChart })));
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { inspectionApi } from "@/lib/api/inspection.api";
import { filterApi } from "@/lib/api/filter.api";
import { issueApi } from "@/lib/api/issue.api";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { cn } from "@/lib/utils";
import Link from "next/link";
import { format, isPast } from "date-fns";
import { useSSE } from "@/hooks/useSSE";

export const DashboardPanelAuditor = () => {
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();

  useSSE();

  const {
    data: inspectionsData,
    isLoading: isInspectionsLoading,
    isFetching: isInspectionsFetching,
  } = useQuery({
    queryKey: ["auditor-inspections", user?.id],
    queryFn: () => filterApi.inspections({ limit: 100, inspector_id: user?.id, sort_by: "created_at", sort_order: "desc" }),
    enabled: mounted && !!user,
  });

  // Fetch Analytics Trend from the newly created backend endpoint
  const currentYear = new Date().getFullYear();
  const {
    data: trendData,
    isLoading: isTrendLoading,
  } = useQuery({
    queryKey: ["auditor-inspections-trend", user?.id, currentYear],
    queryFn: () => inspectionApi.getAnalyticsTrend(user?.id || "", currentYear),
    enabled: mounted && !!user,
  });

  // Fetch Issues (Global/Aggregate for now as per approval)
  const {
    data: issuesData,
    isLoading: isIssuesLoading,
  } = useQuery({
    queryKey: ["auditor-global-issues"],
    queryFn: () => issueApi.getAll({ limit: 1000 }),
    enabled: mounted,
  });

  const handleRefresh = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ["auditor-inspections"] });
    queryClient.invalidateQueries({ queryKey: ["auditor-inspections-trend"] });
    queryClient.invalidateQueries({ queryKey: ["auditor-global-issues"] });
  }, [queryClient]);

  if (!mounted) {
    return <div className="h-screen w-full bg-background" />;
  }

  // Calculate metrics
  const inspections = inspectionsData?.items || [];
  const totalInspections = inspections.length;
  const completedInspections = inspections.filter((i: any) => i.status === "Completed" || i.status === "Approved").length;
  const ongoingInspections = inspections.filter((i: any) => i.status === "Ongoing" || i.status === "Draft").length;

  const issues = Array.isArray(issuesData?.items)
    ? issuesData.items
    : Array.isArray(issuesData)
    ? issuesData
    : [];

  const pendingValidationList = issues.filter(
    (i: any) => i.issue_status === "PendingValidation" || i.computed_status === "PendingValidation"
  );
  const pendingValidationCount = pendingValidationList.length;

  const openIssues = issues.filter(
    (i: any) =>
      i.computed_status === "Open" ||
      i.computed_status === "OpenOverdue" ||
      i.issue_status === "Open" ||
      i.issue_status === "InProgress" ||
      i.issue_status === "Overdue"
  ).length;

  const isLoading = isInspectionsLoading || isTrendLoading || isIssuesLoading;
  const isFetching = isInspectionsFetching;

  // Format trend data for Recharts
  const chartData = trendData?.data || [];

  return (
    <div className="w-full space-y-8 animate-in fade-in duration-500 pb-24 md:pb-6">
      
      {/* Header Section */}
      <section className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <h1 className="text-3xl md:text-4xl font-bold tracking-tight text-foreground">
              Selamat Datang, {user?.name || "Auditor"}
            </h1>
            <span className="bg-primary/10 text-primary text-xs font-semibold px-2.5 py-1 rounded uppercase tracking-wider">
              AUDITOR
            </span>
          </div>
          <p className="text-muted-foreground text-sm">
            Pengawasan Inspeksi · Ringkasan aktivitas pemeriksaan Anda
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleRefresh}
            disabled={isFetching}
            className="flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground hover:bg-muted rounded-lg transition-colors disabled:opacity-50"
          >
            <RefreshCw className={cn("h-4 w-4", isFetching && "animate-spin")} />
          </button>
        </div>
      </section>

      {/* Stat Cards (Grid 5 Column) */}
      <section className="grid grid-cols-2 md:grid-cols-5 gap-4">
        {/* Card 1 */}
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden group">
          <div className="absolute top-0 left-0 w-full h-1 bg-primary"></div>
          <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Total Inspeksi</h3>
          <p className="text-2xl font-bold text-foreground">
            {isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : totalInspections}
          </p>
        </div>

        {/* Card 2 */}
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
          <div className="absolute top-0 left-0 w-full h-1 bg-amber-500"></div>
          <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Berlangsung</h3>
          <p className="text-2xl font-bold text-foreground">
            {isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : ongoingInspections}
          </p>
        </div>

        {/* Card 3 */}
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
          <div className="absolute top-0 left-0 w-full h-1 bg-green-600"></div>
          <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Selesai</h3>
          <p className="text-2xl font-bold text-foreground">
            {isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : completedInspections}
          </p>
        </div>

        {/* Card 4 - Pending Validation */}
        <Link href={`/cimory/dashboard/${user?.id}/issues?status=PendingValidation`}>
          <div className="bg-card rounded-xl p-4 border border-purple-500/30 shadow-sm hover:shadow-md hover:border-purple-500/60 transition-all relative overflow-hidden cursor-pointer group">
            <div className="absolute top-0 left-0 w-full h-1 bg-purple-500"></div>
            {pendingValidationCount > 0 && (
              <span className="absolute top-3 right-3 flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-purple-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-purple-500"></span>
              </span>
            )}
            <h3 className="text-xs font-semibold text-purple-600 dark:text-purple-400 mb-2 uppercase tracking-wider">Menunggu Validasi</h3>
            <p className="text-2xl font-bold text-purple-600 dark:text-purple-300">
              {isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : pendingValidationCount}
            </p>
          </div>
        </Link>

        {/* Card 5 */}
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
          <div className="absolute top-0 left-0 w-full h-1 bg-red-500"></div>
          <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Temuan Terbuka</h3>
          <p className="text-2xl font-bold text-foreground">
            {isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : openIssues}
          </p>
        </div>
      </section>

      {/* ── Pending Validation Action Banner ────────────────────────────── */}
      {pendingValidationCount > 0 && (
        <section className="rounded-2xl border border-purple-500/30 bg-gradient-to-r from-purple-500/10 via-card to-purple-500/5 p-5 shadow-sm space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-xl bg-purple-500/20 flex items-center justify-center text-purple-500 shrink-0">
                <ListTodo className="h-5 w-5" />
              </div>
              <div>
                <h3 className="font-bold text-base text-foreground flex items-center gap-2">
                  <span>Temuan Membutuhkan Validasi Anda</span>
                  <span className="bg-purple-500 text-white text-xs font-extrabold px-2 py-0.5 rounded-full">
                    {pendingValidationCount}
                  </span>
                </h3>
                <p className="text-xs text-muted-foreground">
                  Auditee telah mengajukan bukti perbaikan/follow-up. Mohon lakukan verifikasi untuk menyetujui atau meminta revisi.
                </p>
              </div>
            </div>
            <Link href={`/cimory/dashboard/${user?.id}/issues?status=PendingValidation`}>
              <span className="text-xs font-semibold text-purple-600 hover:text-purple-700 dark:text-purple-400 hover:underline">
                Lihat Semua ({pendingValidationCount}) →
              </span>
            </Link>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 pt-1">
            {pendingValidationList.slice(0, 3).map((issue: any) => (
              <Link key={issue.issue_id} href={`/cimory/dashboard/${user?.id}/issues/${issue.issue_id}`}>
                <div className="p-3.5 rounded-xl border border-purple-500/20 bg-card hover:border-purple-500/50 hover:shadow-md transition-all cursor-pointer space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-[11px] font-mono text-purple-600 dark:text-purple-400 font-semibold">
                      ID: {issue.issue_id}
                    </span>
                    <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-purple-500/15 text-purple-600 border border-purple-500/30">
                      Pending Validation
                    </span>
                  </div>
                  <p className="text-xs font-semibold text-foreground line-clamp-1">
                    {issue.area_name || "—"} · {issue.kawasan_name || "—"} · {issue.detail_kawasan_name || "—"}
                  </p>
                  <p className="text-xs text-muted-foreground line-clamp-1">
                    {issue.keterangan || "Tanpa keterangan"}
                  </p>
                </div>
              </Link>
            ))}
          </div>
        </section>
      )}

      {/* Chart Section */}
      <div className="space-y-4">
        <h2 className="text-lg font-semibold text-foreground">Tren Inspeksi Bulanan ({currentYear})</h2>
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm h-100 flex flex-col">
          <p className="text-xs font-medium text-muted-foreground mb-4">Frekuensi pemeriksaan yang Anda lakukan setiap bulan.</p>
          <div className="grow w-full">
            {isTrendLoading ? (
              <div className="w-full h-full flex items-center justify-center">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              </div>
            ) : (
              <Suspense fallback={<div className="w-full h-full flex items-center justify-center"><Loader2 className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
                <AreaChart
                  data={chartData}
                  xAxisKey="month"
                  series={[
                    { dataKey: "rate", name: "Total Inspeksi", color: "var(--color-blue-500, #3b82f6)" }
                  ]}
                />
              </Suspense>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
