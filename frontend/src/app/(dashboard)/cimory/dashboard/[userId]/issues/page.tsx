"use client";

import { useState, useDeferredValue } from "react";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import {
  AlertTriangle,
  Search,
  Clock,
  CheckCircle2,
  XCircle,
  Loader2,
  ChevronRight,
  Image as ImageIcon,
  ShieldAlert,
  CircleDot,
  ChevronLeft,
  X,
} from "lucide-react";

import { Issue, IssueStatus } from "@/lib/api/issue.api";
import { filterApi, IssueFilterParams } from "@/lib/api/filter.api";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/authStore";
import { usePermissions } from "@/lib/usePermissions";
import { SearchLatencyBadge } from "@/components/ui/SearchLatencyBadge";
import { useSSE } from "@/hooks/useSSE";

/* ── Status configuration ─────────────────────────────────────────────── */
const STATUS_OPTIONS: { label: string; value: IssueStatus | "all" }[] = [
  { label: "Semua", value: "all" },
  { label: "Open", value: "Open" },
  { label: "In Progress", value: "InProgress" },
  { label: "Pending", value: "PendingValidation" },
  { label: "Closed", value: "Closed" },
  { label: "Verified", value: "Verified" },
];

const statusConfig: Record<
  IssueStatus,
  { label: string; icon: React.ElementType; color: string; bg: string; border: string }
> = {
  Open:              { label: "Open",              icon: AlertTriangle,  color: "text-red-400",    bg: "bg-red-500/10",    border: "border-red-500/30" },
  InProgress:        { label: "In Progress",       icon: CircleDot,      color: "text-blue-400",   bg: "bg-blue-500/10",   border: "border-blue-500/30" },
  PendingValidation: { label: "Pending Validation",icon: Loader2,        color: "text-amber-400",  bg: "bg-amber-500/10",  border: "border-amber-500/30" },
  Closed:            { label: "Closed",            icon: XCircle,        color: "text-zinc-400",   bg: "bg-zinc-500/10",   border: "border-zinc-500/30" },
  Verified:          { label: "Verified",          icon: CheckCircle2,   color: "text-emerald-400",bg: "bg-emerald-500/10",border: "border-emerald-500/30" },
  Overdue:           { label: "Overdue",           icon: ShieldAlert,    color: "text-orange-400", bg: "bg-orange-500/10", border: "border-orange-500/30" },
};

/* ── Issue Card ────────────────────────────────────────────────────────── */
function IssueCard({ issue }: { issue: Issue }) {
  const cfg = statusConfig[issue.issue_status] ?? statusConfig["Open"];
  const StatusIcon = cfg.icon;
  const dueDate = issue.due_date ? new Date(issue.due_date) : null;
  const isOverdue =
    dueDate &&
    dueDate < new Date() &&
    issue.issue_status !== "Closed" &&
    issue.issue_status !== "Verified";

  return (
    <Link href={`./issues/${issue.issue_id}`}>
      <div
        className={cn(
          "group relative flex items-stretch gap-0 rounded-2xl border bg-card/60 backdrop-blur-md",
          "hover:border-primary/40 hover:bg-card/80 hover:shadow-lg hover:shadow-primary/5",
          "active:scale-[0.99] transition-all duration-200 overflow-hidden",
          isOverdue ? "border-orange-500/40" : "border-border/60"
        )}
      >
        {/* Left accent bar */}
        <div className={cn("w-1 shrink-0", cfg.color.replace("text-", "bg-"))} />

        {/* Main content */}
        <div className="flex-1 min-w-0 p-4">
          {/* Top row: title + badge */}
          <div className="flex items-start justify-between gap-2 mb-2">
            <div>
              <h4 className="font-semibold text-sm leading-snug text-foreground">
                {issue.area_name || "Tanpa Area"} · {issue.kawasan_name || "Tanpa Kawasan"} · {issue.detail_kawasan_name || "Tanpa Detail"}
              </h4>
              <p className="text-xs text-muted-foreground mt-1 font-medium">
                {issue.keterangan || "Tanpa keterangan"}
              </p>
            </div>
            <span
              className={cn(
                "shrink-0 text-[10px] font-bold px-2 py-0.5 rounded-full border",
                cfg.bg,
                cfg.color,
                cfg.border
              )}
            >
              {cfg.label}
            </span>
          </div>

          {/* Meta row */}
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1 mt-3 pt-2 border-t border-border/40">
            {/* Issue ID */}
            <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground font-mono">
              <StatusIcon className={cn("h-3 w-3 shrink-0", cfg.color)} />
              ID: {issue.issue_id}
            </span>

            {/* Due date */}
            {dueDate && (
              <span
                className={cn(
                  "inline-flex items-center gap-1 text-[11px]",
                  isOverdue ? "text-orange-400 font-semibold" : "text-muted-foreground"
                )}
              >
                <Clock className="h-3 w-3 shrink-0" />
                {isOverdue ? "Overdue – " : ""}
                Tenggat: {dueDate.toLocaleDateString("id-ID", {
                  day: "numeric",
                  month: "short",
                  year: "numeric",
                })}
              </span>
            )}

            {/* Photos */}
            {issue.photos && issue.photos.length > 0 && (
              <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
                <ImageIcon className="h-3 w-3 shrink-0" />
                {issue.photos.length} foto
              </span>
            )}
          </div>
        </div>

        {/* Right chevron */}
        <div className="flex items-center pr-3 text-muted-foreground/40 group-hover:text-muted-foreground/70 transition-colors">
          <ChevronRight className="h-4 w-4" />
        </div>
      </div>
    </Link>
  );
}

/* ── Skeleton loader ───────────────────────────────────────────────────── */
function IssueSkeleton() {
  return (
    <div className="flex items-stretch gap-0 rounded-2xl border border-border/60 bg-card/40 overflow-hidden animate-pulse">
      <div className="w-1 bg-muted" />
      <div className="flex-1 p-4 space-y-2">
        <div className="flex justify-between gap-2">
          <div className="h-4 w-2/3 rounded bg-muted" />
          <div className="h-4 w-16 rounded-full bg-muted" />
        </div>
        <div className="h-3 w-1/3 rounded bg-muted" />
      </div>
    </div>
  );
}

/* ── Page ──────────────────────────────────────────────────────────────── */
export default function IssuesPage() {
  const user = useAuthStore((state) => state.user);
  const { hasPermission, isLoading: isPermLoading } = usePermissions();

  useSSE();

  const [activeStatus, setActiveStatus] = useState<IssueStatus | "all">("all");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);

  // Debounce search via React 18 useDeferredValue
  const deferredSearch = useDeferredValue(search);

  const params: IssueFilterParams = {
    page,
    limit: 20,
    sort_by: "created_at",
    sort_order: "desc",
    ...(activeStatus !== "all" && { status: activeStatus }),
    ...(deferredSearch && { q: deferredSearch }),
  };

  const { data, isLoading } = useQuery({
    queryKey: ["issues-filter", params],
    queryFn: () => filterApi.issues(params),
    enabled: hasPermission("PERM-ISS-R"),
  });

  const issues: Issue[] = data?.items ?? [];
  const facets = data?.facets;
  const total = data?.total ?? 0;
  const totalPages = data?.total_pages ?? 1;

  if (!isPermLoading && !hasPermission("PERM-ISS-R")) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] space-y-4">
        <div className="w-16 h-16 rounded-full bg-destructive/10 flex items-center justify-center">
          <ShieldAlert className="h-8 w-8 text-destructive" />
        </div>
        <h2 className="text-xl font-semibold">Akses Ditolak</h2>
        <p className="text-muted-foreground text-center max-w-md">
          Role Anda tidak memiliki izin untuk mengakses halaman Temuan (Issue).
          Silakan hubungi administrator jika Anda memerlukan akses.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-5 pb-6">

      {/* ── Page header ── */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold tracking-tight">Temuan</h2>
          <p className="text-xs text-muted-foreground mt-0.5">
            {isLoading ? "Memuat..." : `${total} temuan ditemukan`}
          </p>
        </div>
        {/* Summary pill */}
        {!isLoading && (facets?.status?.["Open"] ?? 0) > 0 && (
          <div className="flex items-center gap-1.5 rounded-full bg-red-500/10 border border-red-500/20 px-3 py-1">
            <div className="h-1.5 w-1.5 rounded-full bg-red-500 animate-pulse" />
            <span className="text-xs font-semibold text-red-400">
              {facets!.status["Open"]} Perlu Tindakan
            </span>
          </div>
        )}
      </div>

      {/* ── Search ── */}
      <div className="flex flex-col sm:flex-row gap-3 items-stretch sm:items-center">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
          <input
            type="search"
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            placeholder="Cari temuan (Label)..."
            className={cn(
              "w-full rounded-xl border border-border/60 bg-card/60 backdrop-blur-md",
              "pl-9 pr-4 py-2.5 text-sm placeholder:text-muted-foreground",
              "focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary/50",
              "transition-all"
            )}
          />
        </div>
        <SearchLatencyBadge
          searchQuery={search}
          isFetching={isLoading}
          pageName="Temuan (Issues)"
          apiPath="/issues"
        />
      </div>

      {/* ── Status filter chips — facet counts from API ── */}
      <div className="-mx-4 sm:-mx-6 px-4 sm:px-6">
        <div
          className="flex gap-2 overflow-x-auto pb-0.5"
          style={{ scrollbarWidth: "none", msOverflowStyle: "none" }}
        >
        {STATUS_OPTIONS.map((opt) => {
          // Use real facet counts from API, not local computation
          const count = opt.value === "all"
            ? total
            : (facets?.status?.[opt.value] ?? 0);
          const isActive = activeStatus === opt.value;
          return (
            <button
              key={opt.value}
              onClick={() => { setActiveStatus(opt.value); setPage(1); }}
              className={cn(
                "shrink-0 flex items-center gap-1.5 rounded-full px-3.5 py-1.5 text-xs font-semibold",
                "transition-all duration-200 border",
                isActive
                  ? "bg-primary text-primary-foreground border-primary shadow-md shadow-primary/20"
                  : "bg-card/60 text-muted-foreground border-border/60 hover:border-primary/40 hover:text-foreground"
              )}
            >
              {opt.label}
              {count > 0 && (
                <span
                  className={cn(
                    "text-[10px] font-bold px-1.5 py-0.5 rounded-full min-w-[18px] text-center",
                    isActive ? "bg-white/20 text-white" : "bg-muted text-muted-foreground"
                  )}
                >
                  {count}
                </span>
              )}
            </button>
          );
        })}
        </div>
      </div>

      {/* ── List ── */}
      <div className="grid gap-3">
        {isLoading ? (
          <>
            <IssueSkeleton />
            <IssueSkeleton />
            <IssueSkeleton />
          </>
        ) : issues.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center rounded-2xl border border-dashed border-border/60 bg-card/30">
            <div className="h-14 w-14 rounded-full bg-muted/50 flex items-center justify-center mb-4">
              <AlertTriangle className="h-7 w-7 text-muted-foreground/50" />
            </div>
            <p className="font-semibold text-sm text-foreground">
              {search || activeStatus !== "all" ? "Tidak ada hasil" : "Belum ada temuan"}
            </p>
            <p className="text-xs text-muted-foreground mt-1 max-w-[200px]">
              {search
                ? `Tidak ada temuan yang cocok dengan "${search}"`
                : activeStatus !== "all"
                ? "Tidak ada temuan dengan status ini."
                : "Temuan akan muncul setelah inspeksi dilakukan."}
            </p>
            {(search || activeStatus !== "all") && (
              <button
                onClick={() => { setSearch(""); setActiveStatus("all"); setPage(1); }}
                className="mt-4 flex items-center gap-1 text-xs text-primary hover:underline"
              >
                <X className="h-3 w-3" /> Hapus filter
              </button>
            )}
          </div>
        ) : (
          issues.map((issue) => <IssueCard key={issue.issue_id} issue={issue} />)
        )}
      </div>

      {/* ── Pagination (server-driven) ── */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between pt-2">
          <p className="text-xs text-muted-foreground">Hal {page} dari {totalPages} ({total} total)</p>
          <div className="flex gap-2">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="p-1.5 rounded-lg border border-border/60 disabled:opacity-40 hover:bg-muted transition-colors"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              className="p-1.5 rounded-lg border border-border/60 disabled:opacity-40 hover:bg-muted transition-colors"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
