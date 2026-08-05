"use client";

import { useCallback } from "react";
import { 
  ClipboardCheck, 
  CheckCircle, 
  AlertTriangle, 
  RefreshCw,
  Loader2,
  Clock,
  ArrowRight
} from "lucide-react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { issueApi, Issue } from "@/lib/api/issue.api";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { cn } from "@/lib/utils";
import Link from "next/link";
import { format, isPast } from "date-fns";
import { usePolling } from "@/hooks/usePolling";

export const DashboardPanelAuditee = () => {
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();

  // Enable real-time Server-Sent Events for issue updates
  usePolling();

  // Fetch Issues specific to this auditee (PIC)
  const {
    data: issuesData,
    isLoading: isIssuesLoading,
    isFetching: isIssuesFetching,
  } = useQuery({
    queryKey: ["auditee-issues", user?.id],
    queryFn: () => issueApi.getAll({ limit: 1000, pic_user_id: user?.id }),
    enabled: mounted && !!user?.id,
  });

  const handleRefresh = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ["auditee-issues"] });
  }, [queryClient]);

  if (!mounted) {
    return <div className="h-screen w-full bg-background" />;
  }

  // Calculate metrics
  const issues: Issue[] = Array.isArray(issuesData?.items)
    ? issuesData.items
    : Array.isArray(issuesData)
    ? issuesData
    : [];
  const totalIssues = issues.length;
  
  const openIssues = issues.filter((i) => i.computed_status === "Open" || i.computed_status === "OpenOverdue" || i.issue_status === "Open" || i.issue_status === "InProgress" || i.issue_status === "Overdue").length;
  const pendingIssues = issues.filter((i) => i.issue_status === "PendingValidation").length;
  const closedIssues = issues.filter((i) => i.computed_status === "Closed" || i.computed_status === "ClosedOverdue" || i.issue_status === "Closed" || i.issue_status === "Verified").length;

  const isLoading = isIssuesLoading;
  const isFetching = isIssuesFetching;

  // Sorting issues by due date (closest first), and filtering for open ones
  const activeTasks = issues
    .filter(i => i.issue_status !== "Closed" && i.issue_status !== "Verified")
    .sort((a, b) => {
      if (!a.due_date) return 1;
      if (!b.due_date) return -1;
      return new Date(a.due_date).getTime() - new Date(b.due_date).getTime();
    })
    .slice(0, 10); // show top 10

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
            className="flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground hover:bg-muted rounded-lg transition-colors disabled:opacity-50"
          >
            <RefreshCw className={cn("h-4 w-4", isFetching && "animate-spin")} />
          </button>
        </div>
      </section>

      {/* Stat Cards (Bento Grid) */}
      <section className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {/* Card 1 */}
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden group">
          <div className="absolute top-0 left-0 w-full h-1 bg-gray-500"></div>
          <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Total Tugas</h3>
          <p className="text-2xl font-bold text-foreground">
            {isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : totalIssues}
          </p>
        </div>

        {/* Card 2 */}
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
          <div className="absolute top-0 left-0 w-full h-1 bg-red-500"></div>
          <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Tugas Terbuka</h3>
          <p className="text-2xl font-bold text-foreground">
            {isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : openIssues}
          </p>
        </div>

        {/* Card 3 */}
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
          <div className="absolute top-0 left-0 w-full h-1 bg-amber-500"></div>
          <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Menunggu Validasi</h3>
          <p className="text-2xl font-bold text-foreground">
            {isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : pendingIssues}
          </p>
        </div>

        {/* Card 4 */}
        <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden">
          <div className="absolute top-0 left-0 w-full h-1 bg-green-600"></div>
          <h3 className="text-xs font-semibold text-muted-foreground mb-2 uppercase tracking-wider">Tugas Selesai</h3>
          <p className="text-2xl font-bold text-foreground">
            {isLoading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : closedIssues}
          </p>
        </div>
      </section>

      {/* Task List Section */}
      <div className="space-y-4">
        <h2 className="text-lg font-semibold text-foreground">Daftar Temuan Prioritas</h2>
        <div className="bg-card rounded-xl border border-border shadow-sm overflow-hidden">
          {isLoading ? (
            <div className="w-full h-40 flex items-center justify-center">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : activeTasks.length === 0 ? (
            <div className="w-full h-40 flex flex-col items-center justify-center text-muted-foreground">
              <CheckCircle className="h-8 w-8 mb-2 opacity-50" />
              <p className="text-sm">Bagus! Tidak ada tugas terbuka saat ini.</p>
            </div>
          ) : (
            <div className="divide-y divide-border">
              {activeTasks.map((task) => {
                const overdue = task.due_date ? isPast(new Date(task.due_date)) : false;
                
                return (
                  <div key={task.issue_id} className="p-4 hover:bg-muted/50 transition-colors flex items-center justify-between gap-4">
                    <div className="flex flex-col">
                      <span className="font-semibold text-foreground line-clamp-1">
                        {task.keterangan || "Issue tanpa keterangan"}
                      </span>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground mt-1">
                        <span className={cn(
                          "px-2 py-0.5 rounded-full font-medium text-[10px] uppercase",
                          task.issue_status === "PendingValidation" ? "bg-amber-500/10 text-amber-600" : "bg-red-500/10 text-red-600"
                        )}>
                          {task.issue_status}
                        </span>
                        <span>•</span>
                        <span className={cn(overdue && "text-red-500 font-semibold")}>
                          Tenggat: {task.due_date ? format(new Date(task.due_date), "dd MMM yyyy") : "Tidak ditentukan"}
                        </span>
                      </div>
                    </div>
                    <Link
                      href={`/cimory/${user?.plant_id || 'global'}/dashboard/${user?.id || 'overview'}/issues/${task.issue_id}`}
                      className="p-2 bg-primary/10 text-primary rounded-lg hover:bg-primary/20 transition-colors shrink-0"
                    >
                      <ArrowRight className="h-4 w-4" />
                    </Link>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

    </div>
  );
};
