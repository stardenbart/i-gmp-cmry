"use client";

import { useState, useCallback, useDeferredValue } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Plus, Search, ClipboardCheck, ChevronLeft, ChevronRight, X
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
  { label: "Draft", value: "Draft" },
  { label: "Ongoing", value: "Ongoing" },
  { label: "Completed", value: "Completed" },
  { label: "Approved", value: "Approved" },
];

const statusColor: Record<string, string> = {
  Draft: "bg-zinc-500/10 text-zinc-400 border-zinc-500/30",
  Ongoing: "bg-blue-500/10 text-blue-400 border-blue-500/30",
  Completed: "bg-green-500/10 text-green-400 border-green-500/30",
  Approved: "bg-emerald-500/10 text-emerald-400 border-emerald-500/30",
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

export default function InspectionsPage() {
  const user = useAuthStore(state => state.user);
  const { isAuditor, isLoading: isGuardLoading } = useAuditorGuard();

  // Filter state
  const [q, setQ] = useState("");
  const [status, setStatus] = useState("");
  const [page, setPage] = useState(1);

  // Debounce search input
  const deferredQ = useDeferredValue(q);

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
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Inspeksi</h2>
          <p className="text-muted-foreground text-sm">
            {isLoading ? "Memuat..." : `${total} inspeksi ditemukan`}
          </p>
        </div>
        <Link href={`/cimory/dashboard/${user?.id}/inspections/create`}>
          <Button>
            <Plus className="mr-2 h-4 w-4" /> Buat Inspeksi
          </Button>
        </Link>
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

      {/* Status filter chips (with facet counts from API) */}
      <div className="-mx-4 sm:-mx-6 px-4 sm:px-6">
        <div className="flex gap-2 overflow-x-auto pb-0.5" style={{ scrollbarWidth: "none" }}>
          {STATUS_OPTIONS.map(opt => {
            const count = opt.value
              ? (facets?.status?.[opt.value] ?? 0)
              : total;
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
      <div className="grid gap-4">
        {isLoading ? (
          <InspectionSkeleton />
        ) : items.length === 0 ? (
          <Card className="p-12 flex flex-col items-center justify-center text-center bg-card/40 backdrop-blur-md border-dashed">
            <ClipboardCheck className="h-12 w-12 text-muted-foreground mb-4 opacity-50" />
            <h3 className="text-lg font-semibold">
              {hasActiveFilters ? "Tidak ada hasil" : "Belum ada inspeksi"}
            </h3>
            <p className="text-muted-foreground text-sm mt-1 max-w-sm">
              {hasActiveFilters
                ? "Coba ubah atau hapus filter untuk melihat lebih banyak data."
                : "Mulai audit dengan membuat data inspeksi baru."}
            </p>
            {hasActiveFilters ? (
              <Button variant="outline" className="mt-4" onClick={resetFilters}>
                Hapus Filter
              </Button>
            ) : (
              <Link href={`/cimory/dashboard/${user?.id}/inspections/create`} className="mt-4">
                <Button variant="outline">Buat Inspeksi Pertama</Button>
              </Link>
            )}
          </Card>
        ) : (
          items.map((inspection) => (
            <Card
              key={inspection.inspection_id}
              className="p-4 hover:border-primary/50 transition-colors bg-card/60 backdrop-blur-md"
            >
              <div className="flex flex-col sm:flex-row gap-4 sm:items-center justify-between">
                <div className="flex items-center gap-4">
                  <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center shrink-0">
                    <ClipboardCheck className="h-5 w-5 text-primary" />
                  </div>
                  <div>
                    <h4 className="font-semibold">
                      {inspection.area_name || "Tanpa Area"} · {inspection.kawasan_name || "Tanpa Kawasan"} · {inspection.detail_kawasan_name || "Tanpa Detail"}
                    </h4>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      Session ID: {inspection.inspection_id}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {new Date(inspection.created_at).toLocaleDateString("id-ID", {
                        day: "numeric",
                        month: "long",
                        year: "numeric",
                      })}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-4">
                  {(inspection.status === "Completed" || inspection.status === "Approved") && inspection.score != null && (
                    <span className="text-xs font-bold text-primary bg-primary/10 px-2.5 py-1 rounded-full border border-primary/20">
                      Skor: {Number(inspection.score).toFixed(1)}%
                    </span>
                  )}
                  <span className={cn(
                    "text-xs font-semibold px-2.5 py-1 rounded-full border",
                    statusColor[inspection.status] ?? "bg-zinc-500/10 text-zinc-400 border-zinc-500/30"
                  )}>
                    {inspection.status}
                  </span>
                  <Link href={`/cimory/dashboard/${user?.id}/inspections/${inspection.inspection_id}`}>
                    <Button variant="outline" size="sm">Detail</Button>
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
    </div>
  );
}
