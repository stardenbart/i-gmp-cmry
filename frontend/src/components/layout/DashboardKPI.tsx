"use client";

import { useCallback, useMemo, useState } from "react";
import { Share2, Sparkles } from "lucide-react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { DashboardGrid } from "@/components/dashboard/DashboardGrid";
import { KPIDashboardProvider, useKPIDashboard } from "@/components/dashboard/kpi/KPIDashboardContext";
import { VisualizationBuilder, type BuiltVisualization } from "@/components/dashboard/kpi/VisualizationBuilder";
import { buildDynamicWidgetDefinition } from "@/components/dashboard/kpi/dynamicWidgetDefinition";
import { dashboardLayoutApi, type WidgetConfig } from "@/lib/api/dashboard-layout.api";
import { getApiErrorMessage } from "@/lib/api/error";
import { isAdminUser } from "@/lib/useAdminGuard";
import { KPIShareDialog } from "@/components/dashboard/kpi/KPIShareDialog";
import type { WidgetDefinition, VizType } from "@/components/dashboard/types";

// Must match DashboardGrid's own internal queryKey/queryFn EXACTLY (same
// dashboardKey="kpi", no target = "self") so React Query dedupes this read
// against the one DashboardGrid performs internally — one network request,
// not two. See DashboardGrid.tsx's `queryKey` construction.
const KPI_LAYOUT_QUERY_KEY = ["dashboard-layout", "self", "kpi"];

/** Reconstructs the builder's `initial` prop from a saved widget — the
 * inverse of the `custom_query`/`viz_type` shape saveVisualizationMutation
 * writes. Only ever called for a widget that already passed
 * buildDynamicWidget's `cq` guard (dimension + measures present). */
function toBuiltVisualization(cfg: WidgetConfig): BuiltVisualization | null {
  const cq = cfg.custom_query;
  if (!cq || !cq.dimension || cq.measures.length === 0) return null;
  return {
    title: cq.title ?? "",
    measures: cq.measures,
    dimension: cq.dimension,
    dimension2: cq.dimension2,
    drillDimensions: cq.drillDimensions,
    showLabels: cq.showLabels,
    showLabelValues: cq.showLabelValues,
    showTrendLine: cq.showTrendLine,
    vizType: (cfg.viz_type as VizType) ?? "bar",
  };
}

function KPIDashboardBody() {
  const {
    mounted, user, isSuperAdmin, selectedPlant, setSelectedPlant, plantsResponse,
    effectivePlant, trendMode, trendRange, trendGranularity,
  } = useKPIDashboard();
  const queryClient = useQueryClient();
  const [builderOpen, setBuilderOpen] = useState(false);
  const [shareDialogOpen, setShareDialogOpen] = useState(false);
  const canShare = !!user && isAdminUser(user.role_id, user.role?.role_name);
  // Set only while editing an EXISTING custom widget (via its pencil icon)
  // — mutually exclusive with `builderOpen` (add-new) in practice, but kept
  // as its own state so the builder's `mode`/`initial` props stay simple
  // derivations rather than overloading `builderOpen`'s meaning.
  const [editingWidget, setEditingWidget] = useState<WidgetConfig | null>(null);
  const closeBuilder = useCallback(() => {
    setBuilderOpen(false);
    setEditingWidget(null);
  }, []);
  const openNewBuilder = useCallback(() => {
    setEditingWidget(null);
    setBuilderOpen(true);
  }, []);
  const openEditBuilder = useCallback((widget: WidgetConfig) => {
    setBuilderOpen(false);
    setEditingWidget(widget);
  }, []);

  // Read-only mirror of the layout DashboardGrid itself loads/saves — never
  // written to directly here except via the same PUT /dashboard/layout
  // dashboardLayoutApi.save call DashboardGrid's own "Simpan" button uses,
  // so DashboardGrid.tsx needs no changes to pick up widgets added here.
  const { data: savedLayout, isLoading: isLayoutLoading } = useQuery({
    queryKey: KPI_LAYOUT_QUERY_KEY,
    queryFn: () => dashboardLayoutApi.get({ dashboardKey: "kpi" }),
    enabled: mounted && !!user,
    staleTime: 60_000,
  });

  const dynamicWidgets = useMemo<WidgetDefinition[]>(() => {
    return (savedLayout ?? []).map((cfg) => buildDynamicWidgetDefinition(cfg, openEditBuilder)).filter((w): w is WidgetDefinition => !!w);
  }, [savedLayout, openEditBuilder]);

  const saveVisualizationMutation = useMutation({
    mutationFn: (built: BuiltVisualization) => {
      // Read at mutation time so an edit never writes an older x/y/w/h
      // snapshot captured before the user last saved/rearranged the grid.
      // KPI sekarang khusus untuk visualisasi kustom. Widget KPI bawaan
      // versi lama sengaja tidak ikut ditulis kembali agar layout pengguna
      // dibersihkan tanpa menghapus visualisasi yang mereka buat sendiri.
      const existing = (queryClient.getQueryData<WidgetConfig[]>(KPI_LAYOUT_QUERY_KEY) ?? savedLayout ?? [])
        .filter((widget) => !!widget.custom_query);
      const customQuery: NonNullable<WidgetConfig["custom_query"]> = {
        version: 1,
        title: built.title,
        measures: built.measures,
        dimension: built.dimension,
        // NOTE: dimension2/drillDimensions were previously dropped here —
        // a Heatmap/Sankey or drill-down widget built via "Tambah ke
        // Dashboard" would silently persist as a plain 1-dimension query.
        // Fixed alongside adding drillDimensions support.
        dimension2: built.dimension2,
        drillDimensions: built.drillDimensions,
        showLabels: built.showLabels,
        showLabelValues: built.showLabelValues,
        showTrendLine: built.showTrendLine,
      };

      if (editingWidget) {
        if (!existing.some((widget) => widget.widget_id === editingWidget.widget_id)) {
          throw new Error("Visualisasi yang diedit tidak ditemukan pada tata letak terbaru");
        }
        // Update in place — id/position/size/order/visibility untouched,
        // only the visual definition itself changes.
        const updated = existing.map((w) => (w.widget_id === editingWidget.widget_id ? { ...w, viz_type: built.vizType, custom_query: customQuery } : w));
        return dashboardLayoutApi.save(updated, { dashboardKey: "kpi" });
      }

      const newWidget: WidgetConfig = {
        widget_id: (typeof crypto !== "undefined" && crypto.randomUUID ? crypto.randomUUID() : `custom-${Date.now()}-${Math.random().toString(36).slice(2)}`),
        visible: true,
        order: existing.length,
        x: 0,
        y: 9999, // placed after everything else; user can drag it during edit mode
        w: 6,
        h: 10,
        viz_type: built.vizType,
        custom_query: customQuery,
      };
      return dashboardLayoutApi.save([...existing, newWidget], { dashboardKey: "kpi" });
    },
    onSuccess: (savedWidgets) => {
      // Apply the server's canonical response immediately. Waiting only for
      // an invalidation/refetch can leave the old title/config visible and
      // make a subsequent edit appear to have lost the user's changes.
      queryClient.setQueryData(KPI_LAYOUT_QUERY_KEY, savedWidgets);
      queryClient.invalidateQueries({ queryKey: KPI_LAYOUT_QUERY_KEY });
      toast.success(editingWidget ? "Perubahan visualisasi berhasil disimpan" : "Visualisasi berhasil ditambahkan ke dashboard");
      closeBuilder();
    },
    onError: (error) => {
      toast.error(getApiErrorMessage(error, editingWidget ? "Gagal menyimpan perubahan" : "Gagal menambahkan visualisasi"));
    },
  });

  if (!mounted) {
    return <div className="h-screen w-full bg-background" />;
  }

  return (
    <div className="w-full space-y-8 animate-in fade-in duration-500 pb-24 md:pb-6">
      {/* Header Section */}
      <section className="flex items-start justify-between gap-3 sm:items-center sm:gap-4">
        <div className="min-w-0">
          <div className="mb-1 flex min-w-0 items-center gap-2 sm:gap-3">
            <h1 className="truncate text-2xl font-bold tracking-tight text-foreground sm:text-3xl md:text-4xl">Dashboard KPI</h1>
            <span className="hidden shrink-0 rounded bg-primary/10 px-2.5 py-1 text-xs font-semibold uppercase tracking-wider text-primary sm:inline-flex">
              Analitik
            </span>
          </div>
          <p className="hidden text-sm text-muted-foreground sm:block">
            Ringkasan performa & kepatuhan lintas Area — susun ulang sesuai kebutuhan Anda.
          </p>
        </div>
        <div className="flex shrink-0 flex-wrap items-center justify-end gap-2">
          {canShare && (
            <Button size="sm" variant="outline" className="gap-1.5 whitespace-nowrap px-3 sm:gap-2" onClick={() => setShareDialogOpen(true)}>
              <Share2 className="h-4 w-4" />
              Bagikan
            </Button>
          )}
          <Button size="sm" variant="outline" className="gap-1.5 whitespace-nowrap px-3 sm:gap-2" onClick={openNewBuilder}>
            <Sparkles className="h-4 w-4" />
            Tambah Visualisasi
          </Button>
          {isSuperAdmin ? (
            <select
              aria-label="Pilih Plant"
              value={selectedPlant || "all"}
              onChange={(e) => setSelectedPlant(e.target.value)}
              className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground font-medium focus:ring-2 focus:ring-primary/40 focus:outline-none min-h-[44px]"
            >
              <option value="all">Semua Plant</option>
              {plantsResponse?.data?.items?.map((plant) => (
                <option key={plant.plant_id} value={plant.plant_id}>
                  {plant.plant_name}
                </option>
              ))}
            </select>
          ) : null}
        </div>
      </section>

      {!isLayoutLoading && dynamicWidgets.length === 0 ? (
        <section className="flex min-h-72 flex-col items-center justify-center rounded-2xl border border-dashed border-border bg-card/60 px-6 py-12 text-center">
          <span className="flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/10 text-primary">
            <Sparkles className="h-6 w-6" />
          </span>
          <h2 className="mt-4 text-lg font-semibold text-foreground">Belum ada visualisasi kustom</h2>
          <p className="mt-2 max-w-lg text-sm leading-relaxed text-muted-foreground">
            Dashboard KPI kini hanya berisi visualisasi yang Anda susun sendiri. Ringkasan bawaan tetap tersedia di dashboard utama.
          </p>
          <Button className="mt-5 gap-2" onClick={openNewBuilder}>
            <Sparkles className="h-4 w-4" />
            Buat Visualisasi Pertama
          </Button>
        </section>
      ) : (
        <DashboardGrid registry={dynamicWidgets} enabled={mounted && !!user} dashboardKey="kpi" editable toolboxMode />
      )}

      {(builderOpen || editingWidget) && (
        <VisualizationBuilder
          mode={editingWidget ? "edit" : "add"}
          initial={editingWidget ? (toBuiltVisualization(editingWidget) ?? undefined) : undefined}
          onClose={closeBuilder}
          onAdd={(built) => saveVisualizationMutation.mutate(built)}
          isSubmitting={saveVisualizationMutation.isPending}
        />
      )}
      <KPIShareDialog
        open={shareDialogOpen}
        onClose={() => setShareDialogOpen(false)}
        plantId={effectivePlant}
        plants={plantsResponse?.data?.items}
        filter={{ period: trendMode, start_date: trendRange.start, end_date: trendRange.end, granularity: trendGranularity }}
      />
    </div>
  );
}

export const DashboardPanelKPI = () => {
  return (
    <KPIDashboardProvider>
      <KPIDashboardBody />
    </KPIDashboardProvider>
  );
};
