'use client'
import { useState, useCallback, useDeferredValue, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import {
  Plus, Search, ClipboardCheck, ChevronLeft, ChevronRight,
  X, PlayCircle, ArrowRight, Clock, CheckCircle2, AlertCircle, RotateCcw, Trash2, Loader2,
} from "lucide-react";
import Link from "next/link";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { inspectionApi, InspectionHeader } from "@/lib/api/inspection.api";
import { filterApi, InspectionFilterParams, InspectionFacets } from "@/lib/api/filter.api";
import { usePermissions } from "@/lib/usePermissions";
import { useAuthStore } from "@/stores/authStore";
import { cn } from "@/lib/utils";
import { SearchLatencyBadge } from "@/components/ui/SearchLatencyBadge";
import { useDebounce } from "@/hooks/useDebounce";

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
function MyTaskCard({
  inspection,
  userId,
  plantCode,
  onDelete,
}: {
  inspection: any;
  userId: string;
  plantCode?: string;
  onDelete?: (id: string, name: string) => void;
}) {
  const elapsed = (() => {
    const diff = Date.now() - new Date(inspection.created_at).getTime();
    const h = Math.floor(diff / 3_600_000);
    const m = Math.floor((diff % 3_600_000) / 60_000);
    if (h > 0) return `${h} jam ${m} menit`;
    return `${m} menit`;
  })();

  return (
    <div className="group relative overflow-hidden rounded-2xl border border-amber-500/30 bg-gradient-to-br from-amber-500/5 via-card/60 to-orange-500/5 p-4 hover:border-amber-500/60 hover:shadow-lg hover:shadow-amber-500/10 transition-all duration-300">
      {/* Animated glow dot */}
      <span className="absolute top-3.5 right-3.5 flex h-2.5 w-2.5">
        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-75" />
        <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-amber-500" />
      </span>

      <Link href={`/cimory/${plantCode || "all"}/dashboard/${userId}/inspections/${inspection.inspection_id}`}>
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
              <span className="text-label-sm text-amber-600 dark:text-amber-400 font-medium">
                Berjalan {elapsed}
              </span>
            </div>
          </div>
        </div>
      </Link>

      <div className="mt-3 pt-3 border-t border-amber-500/15 flex items-center justify-between">
        <span className="text-[10px] text-muted-foreground font-mono truncate max-w-[55%]">
          {inspection.inspection_id}
        </span>
        <div className="flex items-center gap-2">
          {onDelete && (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onDelete(inspection.inspection_id, inspection.detail_kawasan_name || inspection.inspection_id);
              }}
              className="h-7 w-7 p-0 text-destructive hover:bg-destructive/10 hover:text-destructive shrink-0"
              title="Hapus Inspeksi"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          )}
          <Link href={`/cimory/${plantCode || "all"}/dashboard/${userId}/inspections/${inspection.inspection_id}`}>
            <span className="flex items-center gap-1 text-label-sm font-semibold text-amber-600 dark:text-amber-400 group-hover:gap-2 transition-all">
              Lanjutkan <ArrowRight className="h-3 w-3" />
            </span>
          </Link>
        </div>
      </div>
    </div>
  );
}


import { useAppDispatch, useAppSelector } from "@/store/hooks";
import {
  setSearchQuery,
  setStatusFilter,
  setPage,
  resetFilters,
} from "@/store/slices/inspectionFilterSlice";

export default function InspectionsPage() {
  const user = useAuthStore(state => state.user);
  const { plantCode } = useParams() as { plantCode?: string };
  const { hasPermission, isLoading: isGuardLoading } = usePermissions();
  const isAuditor = hasPermission("PERM-INSP-R");

  const queryClient = useQueryClient();
  const dispatch = useAppDispatch();
  const { q, status, page } = useAppSelector((state) => state.inspectionFilter);
  
  const [inspectionToDelete, setInspectionToDelete] = useState<{ id: string; name: string } | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleConfirmDeleteInspection = async () => {
    if (!inspectionToDelete) return;
    setIsDeleting(true);
    try {
      await inspectionApi.delete(inspectionToDelete.id);
      toast.success(`Inspeksi ${inspectionToDelete.id} dan temuan terkait berhasil dihapus`);
      queryClient.invalidateQueries({ queryKey: ["inspections-filter"] });
      queryClient.invalidateQueries({ queryKey: ["my-ongoing-inspections"] });
      queryClient.invalidateQueries({ queryKey: ["issues-filter"] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      setInspectionToDelete(null);
    } catch (err: any) {
      toast.error(err?.response?.data?.message || err?.message || "Gagal menghapus inspeksi");
    } finally {
      setIsDeleting(false);
    }
  };

  const [searchInputValue, setSearchInputValue] = useState(q || "");
  const debouncedSearchValue = useDebounce(searchInputValue, 400);

  useEffect(() => {
    dispatch(setSearchQuery(debouncedSearchValue));
  }, [debouncedSearchValue, dispatch]);

  const deferredQ = useDeferredValue(q);

  // ── Query: My ongoing tasks (all ongoing inspections in user's plant) ──
  const myTasksParams: InspectionFilterParams = {
    status: "Ongoing",
    limit: 20,
    sort_by: "created_at",
    sort_order: "desc",
  };

  const { data: myTasksData, isLoading: isMyTasksLoading } = useQuery({
    queryKey: ["my-ongoing-inspections", user?.id],
    queryFn: () => filterApi.inspections(myTasksParams),
    enabled: !isGuardLoading && !!user?.id,
    refetchInterval: 15_000, // refresh every 15s
    refetchOnMount: "always",
    staleTime: 0,
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
    enabled: !isGuardLoading,
    refetchOnMount: "always",
    staleTime: 0,
  });

  const items: InspectionHeader[] = data?.items ?? [];
  const facets: InspectionFacets | undefined = data?.facets;
  const totalPages = data?.total_pages ?? 1;
  const total = data?.total ?? 0;

  const handleResetFilters = useCallback(() => {
    setSearchInputValue("");
    dispatch(resetFilters());
  }, [dispatch]);

  const hasActiveFilters = q || status;

  if (isGuardLoading) {
    return (
      <div className="flex justify-center p-8">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!isGuardLoading && !isAuditor) return null;

  return (
    <div className="space-y-8">

      {/* ── My Tasks Section ───────────────────────────────────────────── */}
      <section>
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2.5">
            
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
              <MyTaskCard 
                key={t.inspection_id} 
                inspection={t} 
                userId={user?.id ?? ""} 
                plantCode={plantCode} 
                onDelete={(id, name) => setInspectionToDelete({ id, name })}
              />
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
              value={searchInputValue}
              onChange={e => setSearchInputValue(e.target.value)}
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
            <Button variant="ghost" size="sm" onClick={handleResetFilters} className="shrink-0">
              <X className="mr-1 h-3 w-3" /> Reset
            </Button>
          )}
        </div>

        {/* Status filter chips with scroll hint */}
        <div className="relative -mx-4 sm:-mx-6 px-4 sm:px-6">
          <div className="flex gap-2 overflow-x-auto pb-1 scrollbar-none snap-x snap-mandatory">
            {STATUS_OPTIONS.map(opt => {
              const count = opt.value ? (facets?.status?.[opt.value] ?? 0) : total;
              const isActive = status === opt.value;
              return (
                <Button
                  key={opt.value}
                  onClick={() => dispatch(setStatusFilter(opt.value))}
                  className={cn(
                    "shrink-0 flex items-center gap-1.5 rounded-full px-3.5 py-1.5 text-xs font-semibold snap-start min-h-[36px]",
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
          <div className="pointer-events-none absolute right-4 sm:right-6 top-0 bottom-1 w-6 bg-gradient-to-l from-background to-transparent sm:hidden" />
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
                      onClick={handleResetFilters}
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
                className="p-3.5 sm:p-4 hover:border-primary/40 transition-all duration-200 bg-card/60 backdrop-blur-md group rounded-2xl border-border/80"
              >
                <div className="flex flex-col sm:flex-row gap-3 sm:gap-4 sm:items-center justify-between">
                  <div className="flex items-start gap-3 min-w-0">
                    <div className={cn(
                      "h-9 w-9 sm:h-10 sm:w-10 rounded-xl flex items-center justify-center shrink-0 transition-colors mt-0.5 sm:mt-0",
                      inspection.status === "Ongoing"
                        ? "bg-amber-500/10 group-hover:bg-amber-500/20"
                        : "bg-primary/10 group-hover:bg-primary/20"
                    )}>
                      <ClipboardCheck className={cn(
                        "h-4 w-4 sm:h-5 sm:w-5",
                        inspection.status === "Ongoing" ? "text-amber-500" : "text-primary"
                      )} />
                    </div>
                    <div className="min-w-0 flex-1">
                      <h4 className="font-semibold text-xs sm:text-sm leading-snug break-words">
                        {inspection.area_name || "—"}
                        <span className="text-muted-foreground font-normal"> · </span>
                        {inspection.kawasan_name || "—"}
                        <span className="text-muted-foreground font-normal"> · </span>
                        {(inspection as any).detail_kawasan_name || "—"}
                      </h4>
                      <p className="text-[10px] sm:text-[11px] text-muted-foreground mt-0.5 font-mono truncate">
                        ID: {inspection.inspection_id}
                      </p>
                      <p className="text-[10px] sm:text-[11px] text-muted-foreground">
                        {new Date(inspection.created_at).toLocaleDateString("id-ID", {
                          day: "numeric", month: "short", year: "numeric",
                          hour: "2-digit", minute: "2-digit",
                        })}
                      </p>
                    </div>
                  </div>

                  <div className="flex flex-wrap items-center justify-between sm:justify-end gap-2 pt-2 sm:pt-0 border-t sm:border-t-0 border-border/40 sm:shrink-0">
                    <div className="flex items-center gap-1.5 flex-wrap">
                      {(inspection.status === "Completed" || inspection.status === "Approved") && inspection.score != null && (
                        <span className="text-[11px] font-bold text-primary bg-primary/10 px-2 py-0.5 rounded-full border border-primary/20">
                          Skor: {Number(inspection.score).toFixed(1)}%
                        </span>
                      )}
                      <span className={cn(
                        "flex items-center gap-1 text-[11px] font-semibold px-2.5 py-0.5 rounded-full border",
                        statusColor[inspection.status] ?? "bg-zinc-500/10 text-zinc-400 border-zinc-500/30"
                      )}>
                        {statusIcon[inspection.status]}
                        {inspection.status}
                      </span>
                    </div>

                    <div className="flex items-center gap-1.5 ml-auto sm:ml-0">
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          setInspectionToDelete({
                            id: inspection.inspection_id,
                            name: (inspection as any).detail_kawasan_name || inspection.inspection_id,
                          });
                        }}
                        className="h-8 w-8 p-0 text-destructive hover:bg-destructive/10 hover:text-destructive shrink-0"
                        title="Hapus Inspeksi"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>

                      <Link
                        href={`/cimory/${plantCode || "all"}/dashboard/${user?.id}/inspections/${inspection.inspection_id}`}
                      >
                        <Button variant="outline" size="sm" className="rounded-xl text-xs h-8 px-3">
                          {inspection.status === "Ongoing" ? "Lanjutkan" : "Detail"}
                          <ArrowRight className="ml-1 h-3 w-3" />
                        </Button>
                      </Link>
                    </div>
                  </div>
                </div>
              </Card>
            ))
          )}
        </div>

        {/* ── Modal Konfirmasi Hapus Inspeksi ── */}
        {inspectionToDelete && (
          <div 
            className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in duration-200 overflow-y-auto"
            onClick={(e) => {
              if (e.target === e.currentTarget) setInspectionToDelete(null);
            }}
          >
            <div className="relative w-[92vw] max-w-[420px] shrink-0 rounded-2xl border border-red-500/30 bg-card p-5 sm:p-6 shadow-2xl space-y-4 my-auto animate-in zoom-in-95 duration-200">
              <div className="flex items-center gap-3 text-red-400">
                <div className="h-10 w-10 rounded-full bg-red-500/10 border border-red-500/20 flex items-center justify-center shrink-0">
                  <Trash2 className="h-5 w-5 text-red-500" />
                </div>
                <div className="min-w-0 flex-1">
                  <h3 className="font-bold text-base text-foreground truncate">Hapus Inspeksi</h3>
                  <p className="text-xs text-muted-foreground font-mono truncate">ID: {inspectionToDelete.id}</p>
                </div>
              </div>

              <p className="text-sm text-muted-foreground leading-relaxed">
                Apakah Anda yakin ingin menghapus inspeksi <strong className="text-foreground">{inspectionToDelete.name}</strong>? Seluruh hasil poin dan temuan (issues) terkait akan ikut terhapus otomatis dari sistem.
              </p>

              <div className="flex items-center justify-end gap-2 pt-2 border-t border-border/40">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={isDeleting}
                  onClick={() => setInspectionToDelete(null)}
                  className="rounded-xl px-4 text-xs font-semibold"
                >
                  Batal
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  disabled={isDeleting}
                  onClick={handleConfirmDeleteInspection}
                  className="rounded-xl px-4 text-xs font-semibold bg-red-600 hover:bg-red-700 text-white"
                >
                  {isDeleting ? (
                    <>
                      <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" /> Menghapus...
                    </>
                  ) : (
                    <>
                      <Trash2 className="mr-1.5 h-3.5 w-3.5" /> Ya, Hapus
                    </>
                  )}
                </Button>
              </div>
            </div>
          </div>
        )}

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
                onClick={() => dispatch(setPage(Math.max(1, page - 1)))}
                disabled={page === 1}
              >
                <ChevronLeft className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => dispatch(setPage(Math.min(totalPages, page + 1)))}
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
