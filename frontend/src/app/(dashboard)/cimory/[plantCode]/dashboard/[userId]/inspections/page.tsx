"use client";

import { useState, useCallback, useDeferredValue } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import {
  Plus, Search, ClipboardCheck, ChevronLeft, ChevronRight,
  X, PlayCircle, ArrowRight, Clock, CheckCircle2, AlertCircle, RotateCcw,
} from "lucide-react";
import Link from "next/link";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { InspectionHeader } from "@/lib/api/inspection.api";
import { filterApi, InspectionFilterParams, InspectionFacets } from "@/lib/api/filter.api";
import { useAuditorGuard } from "@/lib/useAdminGuard";
import { useAuthStore } from "@/stores/authStore";
import { cn } from "@/lib/utils";
import { SearchLatencyBadge } from "@/components/ui/SearchLatencyBadge";

// ── Status config ─────────────────────────────────────────────────────────
const STATUS_OPTIONS = [
  { label: "Semua", value: "" },
  { label: "Ongoing", value: "Ongoing" },
  { label: "Completed", value: "Completed" },
];

const statusColor: Record<string, string> = {
  Ongoing: "bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30",
  Completed: "bg-green-500/10 text-green-600 dark:text-green-400 border-green-500/30",
  Approved: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30",
};

const statusIcon: Record<string, React.ReactNode> = {
  Ongoing: <PlayCircle className="h-3.5 w-3.5" />,
  Completed: <CheckCircle2 className="h-3.5 w-3.5" />,
  Approved: <CheckCircle2 className="h-3.5 w-3.5" />,
};

// ── Skeleton ──────────────────────────────────────────────────────────────
function InspectionSkeleton() {
  return (
    <div className="animate-pulse rounded-2xl border border-border/60 bg-card/40 p-4 flex items-center gap-4">
      <div className="h-10 w-10 rounded-full bg-muted shrink-0" />
      <div className="flex-1 space-y-2">
        <div className="h-4 w-1/3 rounded bg-muted" />
        <div className="h-3 w-1/4 rounded bg-muted" />
      </div>
      <div className="h-6 w-20 rounded-full bg-muted" />
    </div>
  );
}

// ── My Task Card ──────────────────────────────────────────────────────────
function MyTaskCard({ inspection, userId, plantCode }: { inspection: any; userId: string; plantCode?: string }) {
  const elapsed = (() => {
    const diff = Date.now() - new Date(inspection.created_at).getTime();
    const h = Math.floor(diff / 3_600_000);
    const m = Math.floor((diff % 3_600_000) / 60_000);
    if (h > 0) return `${h} jam ${m} menit`;
    return `${m} menit`;
  })();

  return (
    <Link href={`/cimory/${plantCode || "all"}/dashboard/${userId}/inspections/${inspection.inspection_id}`}>
      <div className="group relative overflow-hidden rounded-2xl border border-amber-500/30 bg-gradient-to-br from-amber-500/5 via-card/60 to-orange-500/5 p-4 hover:border-amber-500/60 hover:shadow-lg hover:shadow-amber-500/10 transition-all duration-300 cursor-pointer">
        {/* Animated glow dot */}
        <span className="absolute top-3.5 right-3.5 flex h-2.5 w-2.5">
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-75" />
          <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-amber-500" />
        </span>

        <div className="flex items-start gap-3">
          <div className="h-10 w-10 rounded-xl bg-amber-500/15 flex items-center justify-center shrink-0 group-hover:bg-amber-500/25 transition-colors">
            <PlayCircle className="h-5 w-5 text-amber-500" />
          </div>
          <div className="flex-1 min-w-0">
            <p className="font-semibold text-sm text-foreground leading-snug truncate">
              {inspection.detail_kawasan_name || "—"}
            </p>
            <p className="text-xs text-muted-foreground mt-0.5 truncate">
              {inspection.area_name} · {inspection.kawasan_name}
            </p>
            <div className="flex items-center gap-1.5 mt-2">
              <Clock className="h-3 w-3 text-amber-500/70" />
              <span className="text-[11px] text-amber-600 dark:text-amber-400 font-medium">
                Berjalan {elapsed}
              </span>
            </div>
          </div>
        </div>

        <div className="mt-3 pt-3 border-t border-amber-500/15 flex items-center justify-between">
          <span className="text-[10px] text-muted-foreground font-mono truncate max-w-[70%]">
            {inspection.inspection_id}
          </span>
          <span className="flex items-center gap-1 text-label-sm font-semibold text-amber-600 dark:text-amber-400 group-hover:gap-2 transition-all">
            Lanjutkan <ArrowRight className="h-3 w-3" />
          </span>
        </div>
      </div>
    </Link>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────
export default function InspectionsPage() {
  const user = useAuthStore(state => state.user);
  const { plantCode } = useParams() as { plantCode?: string };
  const { isAuditor, isLoading: isGuardLoading } = useAuditorGuard();

  // Filter state
  const [q, setQ] = useState("");
  const [status, setStatus] = useState("");
  const [page, setPage] = useState(1);
  const deferredQ = useDeferredValue(q);

  // ── Query: My ongoing tasks ──────────────────────────────────────────────
  const myTasksParams: InspectionFilterParams = {
    inspector_id: user?.id,
    status: "Ongoing",
    limit: 20,
    sort_by: "created_at",
    sort_order: "desc",
  };

  const { data: myTasksData, isLoading: isMyTasksLoading } = useQuery({
    queryKey: ["my-ongoing-inspections", user?.id],
    queryFn: () => filterApi.inspections(myTasksParams),
    enabled: !isGuardLoading && isAuditor && !!user?.id,
    refetchInterval: 30_000, // refresh every 30s
  });

  const myTasks: any[] = myTasksData?.items ?? [];

  // ── Query: All inspections (filtered) ────────────────────────────────────
  const params: InspectionFilterParams = {
    page,
    limit: 20,
    ...(deferredQ && { q: deferredQ }),
    ...(status && { status }),
    sort_by: "created_at",
    sort_order: "desc",
  };

  const { data, isLoading } = useQuery({
    queryKey: ["inspections-filter", params],
    queryFn: () => filterApi.inspections(params),
    enabled: !isGuardLoading && isAuditor,
  });

  const items: InspectionHeader[] = data?.items ?? [];
  const facets: InspectionFacets | undefined = data?.facets;
  const totalPages = data?.total_pages ?? 1;
  const total = data?.total ?? 0;

  const resetFilters = useCallback(() => {
    setQ("");
    setStatus("");
    setPage(1);
  }, []);

  const hasActiveFilters = q || status;

  if (isGuardLoading) {
    return (
      <div className="flex justify-center p-8">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!isAuditor) return null;

  return (
    <div className="space-y-8">

      {/* ── My Tasks Section ───────────────────────────────────────────── */}
      <section>
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2.5">
            <div className="h-8 w-8 rounded-lg bg-amber-500/15 flex items-center justify-center">
              <PlayCircle className="h-4 w-4 text-amber-500" />
            </div>
            <div>
              <h2 className="text-base font-bold tracking-tight">Tugas Saya</h2>
              <p className="text-xs text-muted-foreground">Inspeksi yang sedang berjalan</p>
            </div>
            {myTasks.length > 0 && (
              <span className="ml-1 inline-flex items-center justify-center h-5 min-w-5 px-1.5 rounded-full bg-amber-500 text-white text-[10px] font-bold">
                {myTasks.length}
              </span>
            )}
          </div>
          <Link href={`/cimory/${plantCode || "all"}/dashboard/${user?.id}/inspections/create`}>
            <Button size="sm" className="rounded-xl">
              <Plus className="mr-1.5 h-3.5 w-3.5" /> Buat Inspeksi
            </Button>
          </Link>
        </div>

        {isMyTasksLoading ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {[1, 2].map(i => (
              <div key={i} className="animate-pulse rounded-2xl border border-border/60 bg-card/40 p-4 h-28" />
            ))}
          </div>
        ) : myTasks.length === 0 ? (
          <div className="flex items-center gap-3 rounded-2xl border border-dashed border-border/60 bg-card/30 px-5 py-4">
            <AlertCircle className="h-5 w-5 text-muted-foreground/50 shrink-0" />
            <div>
              <p className="text-sm font-medium text-muted-foreground">Tidak ada inspeksi aktif</p>
              <p className="text-xs text-muted-foreground/60 mt-0.5">
                Klik "Buat Inspeksi" untuk memulai tugas audit baru.
              </p>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {myTasks.map(t => (
              <MyTaskCard key={t.inspection_id} inspection={t} userId={user?.id ?? ""} plantCode={plantCode} />
            ))}
          </div>
        )}
      </section>

      {/* ── Divider ────────────────────────────────────────────────────── */}
      <div className="relative">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-border/40" />
        </div>
        <div className="relative flex justify-start">
          <span className="bg-background pr-3 text-xs text-muted-foreground font-medium">
            Semua Inspeksi
          </span>
        </div>
      </div>

      {/* ── All Inspections Section ─────────────────────────────────────── */}
      <section className="space-y-5">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <p className="text-sm text-muted-foreground">
            {isLoading ? "Memuat..." : `${total} inspeksi ditemukan`}
          </p>
        </div>

        {/* Search + Reset */}
        <div className="flex flex-col sm:flex-row gap-3 items-stretch sm:items-center">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
            <Input
              placeholder="Cari berdasarkan Session ID..."
              value={q}
              onChange={e => { setQ(e.target.value); setPage(1); }}
              className="pl-9 bg-background/50 border-border/50"
            />
          </div>
          <SearchLatencyBadge
            searchQuery={q}
            isFetching={isLoading}
            pageName="Inspeksi Digital"
            apiPath="/inspections"
          />
          {hasActiveFilters && (
            <Button variant="ghost" size="sm" onClick={resetFilters} className="shrink-0">
              <X className="mr-1 h-3 w-3" /> Reset
            </Button>
          )}
        </div>

        {/* Status filter chips */}
        <div className="-mx-4 sm:-mx-6 px-4 sm:px-6">
          <div className="flex gap-2 overflow-x-auto pb-0.5" style={{ scrollbarWidth: "none" }}>
            {STATUS_OPTIONS.map(opt => {
              const count = opt.value ? (facets?.status?.[opt.value] ?? 0) : total;
              const isActive = status === opt.value;
              return (
                <Button
                  key={opt.value}
                  onClick={() => { setStatus(opt.value); setPage(1); }}
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
                    <span className={cn(
                      "text-[10px] font-bold px-1.5 py-0.5 rounded-full min-w-4.5 text-center",
                      isActive ? "bg-white/20 text-white" : "bg-muted text-muted-foreground"
                    )}>
                      {count}
                    </span>
                  )}
                </Button>
              );
            })}
          </div>
        </div>

        {/* List */}
        <div className="w-full grid grid-cols-1 gap-3">
          {isLoading ? (
            <>
              <InspectionSkeleton />
              <InspectionSkeleton />
              <InspectionSkeleton />
            </>
          ) : items.length === 0 ? (
            <div className="w-full rounded-3xl border border-dashed border-border/70 bg-gradient-to-b from-card/80 via-card/40 to-background p-8 sm:p-12 text-center shadow-sm">
              <div className="mx-auto w-full max-w-md text-center space-y-4" style={{ width: "100%", maxWidth: "28rem", marginLeft: "auto", marginRight: "auto" }}>
                <div className="inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-500/10 border border-amber-500/20 text-amber-500 shadow-inner mx-auto mb-2">
                  {status === "Ongoing" ? (
                    <PlayCircle className="h-8 w-8 text-amber-500" />
                  ) : status === "Completed" ? (
                    <CheckCircle2 className="h-8 w-8 text-green-500" />
                  ) : (
                    <ClipboardCheck className="h-8 w-8 text-primary" />
                  )}
                </div>

                <h3 className="w-full text-lg font-bold text-foreground tracking-tight text-center block">
                  {status === "Ongoing"
                    ? "Tidak Ada Inspeksi Berjalan"
                    : status === "Completed"
                    ? "Belum Ada Inspeksi Selesai"
                    : q
                    ? "Hasil Pencarian Tidak Ditemukan"
                    : hasActiveFilters
                    ? "Data Tidak Ditemukan"
                    : "Belum Ada Inspeksi"}
                </h3>

                <p className="w-full text-sm text-muted-foreground leading-relaxed text-center block" style={{ wordBreak: "normal", overflowWrap: "break-word" }}>
                  {status === "Ongoing"
                    ? "Semua inspeksi pada area ini telah diselesaikan atau belum ada tugas audit baru yang dimulai."
                    : status === "Completed"
                    ? "Belum ada data inspeksi yang berstatus Completed pada filter ini."
                    : q
                    ? `Tidak ada inspeksi yang cocok dengan kata kunci "${q}".`
                    : hasActiveFilters
                    ? "Coba sesuaikan atau hapus filter untuk melihat seluruh data inspeksi."
                    : "Mulai aktivitas audit digital Anda dengan membuat data inspeksi pertama."}
                </p>

                <div className="w-full flex flex-wrap items-center justify-center gap-3 pt-2">
                  {hasActiveFilters && (
                    <Button
                      variant="outline"
                      onClick={resetFilters}
                      className="rounded-full px-5 h-9 text-xs font-semibold hover:bg-muted"
                    >
                      <RotateCcw className="mr-1.5 h-3.5 w-3.5" /> Reset Filter
                    </Button>
                  )}
                  <Link href={`/cimory/${plantCode || "all"}/dashboard/${user?.id}/inspections/create`}>
                    <Button className="rounded-full px-5 h-9 text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90">
                      <Plus className="mr-1.5 h-3.5 w-3.5" /> Buat Inspeksi
                    </Button>
                  </Link>
                </div>
              </div>
            </div>
          ) : (
            items.map((inspection) => (
              <Card
                key={inspection.inspection_id}
                className="p-4 hover:border-primary/40 transition-all duration-200 bg-card/60 backdrop-blur-md group"
              >
                <div className="flex flex-col sm:flex-row gap-4 sm:items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className={cn(
                      "h-10 w-10 rounded-xl flex items-center justify-center shrink-0 transition-colors",
                      inspection.status === "Ongoing"
                        ? "bg-amber-500/10 group-hover:bg-amber-500/20"
                        : "bg-primary/10 group-hover:bg-primary/20"
                    )}>
                      <ClipboardCheck className={cn(
                        "h-5 w-5",
                        inspection.status === "Ongoing" ? "text-amber-500" : "text-primary"
                      )} />
                    </div>
                    <div>
                      <h4 className="font-semibold text-sm leading-snug">
                        {inspection.area_name || "—"}
                        <span className="text-muted-foreground font-normal"> · </span>
                        {inspection.kawasan_name || "—"}
                        <span className="text-muted-foreground font-normal"> · </span>
                        {(inspection as any).detail_kawasan_name || "—"}
                      </h4>
                      <p className="text-[11px] text-muted-foreground mt-0.5 font-mono">
                        {inspection.inspection_id}
                      </p>
                      <p className="text-[11px] text-muted-foreground">
                        {new Date(inspection.created_at).toLocaleDateString("id-ID", {
                          day: "numeric", month: "long", year: "numeric",
                          hour: "2-digit", minute: "2-digit",
                        })}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3 sm:shrink-0">
                    {(inspection.status === "Completed" || inspection.status === "Approved") && inspection.score != null && (
                      <span className="text-xs font-bold text-primary bg-primary/10 px-2.5 py-1 rounded-full border border-primary/20">
                        Skor: {Number(inspection.score).toFixed(1)}%
                      </span>
                    )}
                    <span className={cn(
                      "flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-full border",
                      statusColor[inspection.status] ?? "bg-zinc-500/10 text-zinc-400 border-zinc-500/30"
                    )}>
                      {statusIcon[inspection.status]}
                      {inspection.status}
                    </span>
                    <Link href={`/cimory/${plantCode || "all"}/dashboard/${user?.id}/inspections/${inspection.inspection_id}`}>
                      <Button variant="outline" size="sm" className="rounded-xl text-xs">
                        {inspection.status === "Ongoing" ? "Lanjutkan" : "Detail"}
                      </Button>
                    </Link>
                  </div>
                </div>
              </Card>
            ))
          )}
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between pt-2">
            <p className="text-sm text-muted-foreground">
              Hal {page} dari {totalPages} ({total} total)
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
              >
                <ChevronLeft className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
              >
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </section>
    </div>
  );
}
