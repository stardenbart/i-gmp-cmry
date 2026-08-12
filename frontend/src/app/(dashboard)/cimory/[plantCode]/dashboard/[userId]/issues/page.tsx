"use client";

import { useState, useDeferredValue, useMemo } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useParams } from "next/navigation";
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
  Trash2,
} from "lucide-react";
import { toast } from "sonner";

import { Issue, IssueStatus, issueApi } from "@/lib/api/issue.api";
import { filterApi, IssueFilterParams } from "@/lib/api/filter.api";
import { Button } from "@/components/ui/button";
import { cn, formatImageUrl, isEncryptedBase64 } from "@/lib/utils";
import { useAuthStore } from "@/stores/authStore";
import { usePermissions } from "@/lib/usePermissions";
import { SearchLatencyBadge } from "@/components/ui/SearchLatencyBadge";
import { usePolling } from "@/hooks/usePolling";

/* ── Status configuration ─────────────────────────────────────────────── */
const STATUS_OPTIONS: { label: string; value: IssueStatus | "all" }[] = [
  { label: "Semua", value: "all" },
  { label: "Open", value: "Open" },
  { label: "Open Overdue", value: "OpenOverdue" },
  { label: "In Progress", value: "InProgress" },
  { label: "Closed", value: "Closed" },
  { label: "Closed Overdue", value: "ClosedOverdue" },
];

const statusConfig: Record<
  IssueStatus,
  { label: string; icon: React.ElementType; color: string; bg: string; border: string }
> = {
  Open:              { label: "Open",              icon: AlertTriangle,  color: "text-blue-400",   bg: "bg-blue-500/10",   border: "border-blue-500/30" },
  InProgress:        { label: "In Progress",       icon: CircleDot,      color: "text-amber-400",  bg: "bg-amber-500/10",  border: "border-amber-500/30" },
  PendingValidation: { label: "Pending Validation",icon: Loader2,        color: "text-purple-400", bg: "bg-purple-500/10", border: "border-purple-500/30" },
  Closed:            { label: "Closed",            icon: XCircle,        color: "text-zinc-400",   bg: "bg-zinc-500/10",   border: "border-zinc-500/30" },
  Verified:          { label: "Verified",          icon: CheckCircle2,   color: "text-emerald-400",bg: "bg-emerald-500/10",border: "border-emerald-500/30" },
  Overdue:           { label: "Overdue",           icon: ShieldAlert,    color: "text-red-400",    bg: "bg-red-500/10",    border: "border-red-500/30" },
  OpenOverdue:       { label: "Open Overdue",       icon: ShieldAlert,    color: "text-red-400",    bg: "bg-red-500/10",    border: "border-red-500/30" },
  ClosedOverdue:     { label: "Closed Overdue",     icon: CheckCircle2,   color: "text-orange-400", bg: "bg-orange-500/10", border: "border-orange-500/30" },
};

/* ── Issue Card ────────────────────────────────────────────────────────── */
function IssueCard({ issue, targetUrl }: { issue: Issue; targetUrl?: string }) {
  const displayStatus = issue.computed_status || issue.issue_status;
  const cfg = statusConfig[displayStatus] ?? statusConfig["Open"];
  const StatusIcon = cfg.icon;
  const dueDate = issue.due_date ? new Date(issue.due_date) : null;
  const isOverdue =
    displayStatus === "OpenOverdue" ||
    displayStatus === "ClosedOverdue" ||
    (dueDate &&
      dueDate < new Date() &&
      issue.issue_status !== "Closed" &&
      issue.issue_status !== "Verified");

  const href = targetUrl || `issues/${issue.issue_id}`;
  const initialPhotos = issue.photos?.filter((p) => p.photo_type === "Initial") || issue.photos || [];
  const firstPhoto = initialPhotos.length > 0 ? initialPhotos[0] : (issue.photos && issue.photos.length > 0 ? issue.photos[0] : null);
  const totalPhotoCount = issue.photos?.length || 0;

  return (
    <Link href={href}>
      <div
        className={cn(
          "group relative flex items-center gap-3 sm:gap-4 rounded-2xl border bg-card/60 backdrop-blur-md p-3 sm:p-4",
          "hover:border-primary/40 hover:bg-card/80 hover:shadow-lg hover:shadow-primary/5",
          "active:scale-[0.99] transition-all duration-200 overflow-hidden",
          isOverdue ? "border-red-500/40" : "border-border/60"
        )}
      >
        {/* Left accent bar */}
        <div className={cn("absolute left-0 top-0 bottom-0 w-1.5", cfg.color.replace("text-", "bg-"))} />

        {/* Thumbnail Image Preview */}
        {firstPhoto ? (
          <div className="relative h-16 w-16 sm:h-20 sm:w-20 shrink-0 rounded-xl overflow-hidden bg-muted border border-border/80 ml-2">
            <img
              src={formatImageUrl(firstPhoto.image_url)}
              alt="Foto Temuan"
              className="h-full w-full object-cover group-hover:scale-105 transition-transform duration-300"
            />
            {totalPhotoCount > 1 && (
              <span className="absolute bottom-1 right-1 text-[9px] font-bold bg-black/80 text-white px-1.5 py-0.5 rounded-md backdrop-blur-xs">
                +{totalPhotoCount - 1}
              </span>
            )}
          </div>
        ) : (
          <div className="h-16 w-16 sm:h-20 sm:w-20 shrink-0 rounded-xl bg-muted/30 border border-dashed border-border/70 flex items-center justify-center text-muted-foreground/40 ml-2">
            <ImageIcon className="h-6 w-6" />
          </div>
        )}

        {/* Main content */}
        <div className="flex-1 min-w-0">
          {/* Top row: title + badge */}
          <div className="flex items-start justify-between gap-2 mb-1.5">
            <div>
              <h4 className="font-semibold text-xs sm:text-sm leading-snug text-foreground line-clamp-1">
                {issue.area_name || "Tanpa Area"} · {issue.kawasan_name || "Tanpa Kawasan"} · {issue.detail_kawasan_name || "Tanpa Detail"}
              </h4>
              <p className="text-xs text-muted-foreground mt-0.5 font-medium line-clamp-2">
                {isEncryptedBase64(issue.keterangan)
                  ? issue.uraian_text || "Deskripsi temuan inspeksi"
                  : issue.keterangan || "Tanpa keterangan"}
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
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1 mt-2 pt-2 border-t border-border/40">
            {/* Issue ID */}
            <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground font-mono">
              <StatusIcon className={cn("h-3 w-3 shrink-0", cfg.color)} />
              ID: {issue.issue_id}
            </span>

            {/* HEI Category Badge */}
            {(issue.hei_category || issue.hei_name) && (
              <span className="inline-flex items-center gap-1 text-[10px] font-semibold text-primary bg-primary/10 border border-primary/20 px-2 py-0.5 rounded-full">
                {issue.hei_category ? `[${issue.hei_category}] ` : ""}
                {issue.hei_name && issue.hei_name !== issue.hei_category ? issue.hei_name : ""}
              </span>
            )}

            {/* Due date */}
            {dueDate && (
              <span
                className={cn(
                  "inline-flex items-center gap-1 text-[11px]",
                  isOverdue ? "text-red-400 font-semibold" : "text-muted-foreground"
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

            {/* Follow up delay badge */}
            {issue.follow_up_delay !== undefined && issue.follow_up_delay > 0 && (
              <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-orange-400 bg-orange-500/10 border border-orange-500/20 px-1.5 py-0.5 rounded-full">
                Terlambat {issue.follow_up_delay} hari
              </span>
            )}
          </div>
        </div>

        {/* Right chevron */}
        <div className="flex items-center pr-1 text-muted-foreground/40 group-hover:text-muted-foreground/70 transition-colors">
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

import { useAppDispatch, useAppSelector } from "@/store/hooks";
import {
  setActiveStatus,
  setSearch,
  setPage,
  resetFilters,
} from "@/store/slices/issueFilterSlice";

/* ── Page ──────────────────────────────────────────────────────────────── */
export default function IssuesPage() {
  const paramsNav = useParams() as { plantCode?: string; userId?: string };
  const user = useAuthStore((state) => state.user);
  const { hasPermission, isLoading: isPermLoading } = usePermissions();
  const canAccess = hasPermission("PERM-ISS-R");

  const plantCode = paramsNav?.plantCode || user?.plant_id || "global";
  const userId = paramsNav?.userId || user?.id || "overview";

  usePolling();

  const dispatch = useAppDispatch();
  const { activeStatus, search, page } = useAppSelector((state) => state.issueFilter);

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
    enabled: canAccess,
    staleTime: 0,
    refetchOnMount: "always",
    refetchInterval: 5000,
  });

  const rawItems: Issue[] = data?.items ?? [];
  const issues: Issue[] = useMemo(() => {
    const map = new Map<string, Issue>();
    for (const item of rawItems) {
      if (item.issue_id && !map.has(item.issue_id)) {
        map.set(item.issue_id, item);
      }
    }
    return Array.from(map.values());
  }, [rawItems]);
  const facets = data?.facets;
  const total = data?.total ?? 0;
  const totalPages = data?.total_pages ?? 1;

  if (!isPermLoading && !canAccess) {
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
            onChange={(e) => dispatch(setSearch(e.target.value))}
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
              onClick={() => dispatch(setActiveStatus(opt.value))}
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
      <div className="w-full grid grid-cols-1 gap-3">
        {isLoading ? (
          <>
            <IssueSkeleton />
            <IssueSkeleton />
            <IssueSkeleton />
          </>
        ) : issues.length === 0 ? (
          <div className="w-full rounded-3xl border border-dashed border-border/70 bg-gradient-to-b from-card/80 via-card/40 to-background p-8 sm:p-12 text-center shadow-sm">
            <div className="mx-auto w-full max-w-md text-center space-y-4" style={{ width: "100%", maxWidth: "28rem", marginLeft: "auto", marginRight: "auto" }}>
              <div className="inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-500/10 border border-amber-500/20 text-amber-500 shadow-inner mx-auto mb-2">
                <AlertTriangle className="h-8 w-8 text-amber-500" />
              </div>

              <h3 className="w-full text-lg font-bold text-foreground tracking-tight text-center block">
                {search || activeStatus !== "all" ? "Tidak Ada Temuan Ditemukan" : "Belum Ada Temuan"}
              </h3>

              <p className="w-full text-sm text-muted-foreground leading-relaxed text-center block" style={{ wordBreak: "normal", overflowWrap: "break-word" }}>
                {search
                  ? `Tidak ada temuan yang cocok dengan kata kunci "${search}".`
                  : activeStatus !== "all"
                  ? `Tidak ada temuan berstatus "${activeStatus}".`
                  : "Temuan akan secara otomatis tercatat ketika ada poin inspeksi yang tidak sesuai (NG)."}
              </p>

              {(search || activeStatus !== "all") && (
                <div className="w-full flex items-center justify-center pt-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => dispatch(resetFilters())}
                    className="rounded-full px-5 h-9 text-xs font-semibold"
                  >
                    <X className="mr-1.5 h-3.5 w-3.5" /> Hapus Filter
                  </Button>
                </div>
              )}
            </div>
          </div>
        ) : (
          issues.map((issue) => (
            <IssueCard 
              key={issue.issue_id} 
              issue={issue} 
              targetUrl={`/cimory/${plantCode}/dashboard/${userId}/issues/${issue.issue_id}`} 
            />
          ))
        )}
      </div>

      {/* ── Pagination (server-driven) ── */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between pt-2">
          <p className="text-xs text-muted-foreground">Hal {page} dari {totalPages} ({total} total)</p>
          <div className="flex gap-2">
            <button
              onClick={() => dispatch(setPage(Math.max(1, page - 1)))}
              disabled={page === 1}
              className="p-1.5 rounded-lg border border-border/60 disabled:opacity-40 hover:bg-muted transition-colors"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <button
              onClick={() => dispatch(setPage(Math.min(totalPages, page + 1)))}
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
