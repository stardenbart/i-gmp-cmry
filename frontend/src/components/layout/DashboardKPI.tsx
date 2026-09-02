"use client";

import { useCallback, useMemo, useState } from "react";
import { Sparkles } from "lucide-react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { DashboardGrid } from "@/components/dashboard/DashboardGrid";
import { kpiWidgetRegistry } from "@/components/dashboard/widgets/kpi/registry";
import { KPIDashboardProvider, useKPIDashboard } from "@/components/dashboard/kpi/KPIDashboardContext";
import { DynamicKPIWidget } from "@/components/dashboard/kpi/DynamicKPIWidget";
import { VisualizationBuilder, type BuiltVisualization } from "@/components/dashboard/kpi/VisualizationBuilder";
import { dashboardLayoutApi, type WidgetConfig } from "@/lib/api/dashboard-layout.api";
import { getApiErrorMessage } from "@/lib/api/error";
import type { WidgetDefinition, VizType } from "@/components/dashboard/types";

// Must match DashboardGrid's own internal queryKey/queryFn EXACTLY (same
// dashboardKey="kpi", no target = "self") so React Query dedupes this read
// against the one DashboardGrid performs internally — one network request,
// not two. See DashboardGrid.tsx's `queryKey` construction.
const KPI_LAYOUT_QUERY_KEY = ["dashboard-layout", "self", "kpi"];

// Every ECharts-backed chart type the builder can produce — "list" is
// deliberately excluded (custom widgets have no bespoke fallback markup),
// and this list is intentionally the FULL set: unlike the 6 built-in
// widgets (each declaring its own narrow subset in registry.ts), a custom
// widget's actual compatibility was already enforced once at build time by
// VisualizationBuilder, so its header picker can safely offer everything —
// re-picking an incompatible type here just gets rejected by the backend
// (400) and surfaced as DynamicKPIWidget's "invalid configuration" state.
const DYNAMIC_WIDGET_VIZ_TYPES: VizType[] = [
  "bar",
  "horizontal_bar",
  "stacked_bar",
  "line",
  "area",
  "radar",
  "scatter",
  "pie",
  "donut",
  "treemap",
  "funnel",
  "number_card",
  "gauge",
  "heatmap",
  "sankey",
  "table",
];

function buildDynamicWidget(cfg: WidgetConfig, onEditWidget?: (cfg: WidgetConfig) => void): WidgetDefinition | null {
  const cq = cfg.custom_query;
  if (!cq || cq.version !== 1 || !cq.dimension || cq.measures.length === 0) return null;
  const title = cq.title?.trim() || "Visualisasi Kustom";
  return {
    id: cfg.widget_id,
    title,
    // DynamicKPIWidget ignores nothing here — measures/dimension(2) are
    // fixed per widget instance via this closure; only vizType still comes
    // from DashboardGrid's own Power BI-style picker, exactly like the 6
    // built-in widgets.
    Component: ({ vizType }: { vizType?: VizType }) => (
      <DynamicKPIWidget
        title={title}
        measures={cq.measures}
        dimension={cq.dimension}
        dimension2={cq.dimension2}
        drillDimensions={cq.drillDimensions}
        showLabels={cq.showLabels}
        showLabelValues={cq.showLabelValues}
        showTrendLine={cq.showTrendLine}
        vizType={vizType}
        onEdit={onEditWidget ? () => onEditWidget(cfg) : undefined}
      />
    ),
    defaultLayout: { x: 0, y: 9999, w: 6, h: 10 },
    supportedVizTypes: DYNAMIC_WIDGET_VIZ_TYPES,
    defaultVizType: cq.dimension2 ? "heatmap" : "bar",
  };
}

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
  const { mounted, user, isSuperAdmin, selectedPlant, setSelectedPlant, plantsResponse } = useKPIDashboard();
  const queryClient = useQueryClient();
  const [builderOpen, setBuilderOpen] = useState(false);
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
  const { data: savedLayout } = useQuery({
    queryKey: KPI_LAYOUT_QUERY_KEY,
    queryFn: () => dashboardLayoutApi.get({ dashboardKey: "kpi" }),
    enabled: mounted && !!user,
    staleTime: 60_000,
  });

  const dynamicWidgets = useMemo<WidgetDefinition[]>(() => {
    return (savedLayout ?? []).map((cfg) => buildDynamicWidget(cfg, openEditBuilder)).filter((w): w is WidgetDefinition => !!w);
  }, [savedLayout, openEditBuilder]);

  const registry = useMemo<WidgetDefinition[]>(() => [...kpiWidgetRegistry, ...dynamicWidgets], [dynamicWidgets]);

  const saveVisualizationMutation = useMutation({
    mutationFn: (built: BuiltVisualization) => {
      const existing = savedLayout ?? [];
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

      {/* Widget area — every user (Admin/Auditor/Auditee) customizes their
          own KPI layout directly here (unlike the main role dashboards,
          which are centrally configured via Edit User), including the
          drag-and-drop Toolbox for adding widgets back, and any custom
          visualizations built above. Saved separately from the main
          dashboard layout via dashboardKey="kpi". DashboardGrid itself
          knows nothing about custom_query — it only ever sees `registry`,
          which already has the dynamic widgets merged in. */}
      <DashboardGrid registry={registry} enabled={mounted && !!user} dashboardKey="kpi" editable toolboxMode />

      {(builderOpen || editingWidget) && (
        <VisualizationBuilder
          mode={editingWidget ? "edit" : "add"}
          initial={editingWidget ? (toBuiltVisualization(editingWidget) ?? undefined) : undefined}
          onClose={closeBuilder}
          onAdd={(built) => saveVisualizationMutation.mutate(built)}
          isSubmitting={saveVisualizationMutation.isPending}
        />
      )}
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
