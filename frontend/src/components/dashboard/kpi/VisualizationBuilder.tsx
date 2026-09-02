"use client";

import { useMemo, useState, type DragEvent } from "react";
import { createPortal } from "react-dom";
import { useQuery } from "@tanstack/react-query";
import {
  X,
  GripVertical,
  Plus,
  Loader2,
  BarChart3,
  BarChartHorizontal,
  Layers,
  LineChart,
  AreaChart,
  Radar,
  ScatterChart,
  PieChart,
  Disc3,
  LayoutGrid,
  Filter,
  Hash,
  Gauge as GaugeIcon,
  Grid3x3,
  Workflow,
  Table2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { useMounted } from "@/lib/useMounted";
import { analyticsApi, type CatalogField, type AnalyticsChartType } from "@/lib/api/analytics.api";
import { getApiErrorMessage } from "@/lib/api/error";
import type { VizType } from "@/components/dashboard/types";

export interface BuiltVisualization {
  title: string;
  measures: string[];
  dimension: string;
  /** Only set for Heatmap/Sankey (a dim1 x dim2 x 1 measure matrix). */
  dimension2?: string;
  /** Ordered drill-down levels beyond `dimension` (level 0). Mutually
   * exclusive with dimension2 — see compatibleCharts()'s `hasDrill` gate. */
  drillDimensions?: string[];
  /** Pure rendering options — see buildEChartsOption.ts's ChartDisplayOptions. */
  showLabels?: boolean;
  showLabelValues?: boolean;
  showTrendLine?: boolean;
  vizType: VizType;
}

interface VisualizationBuilderProps {
  onClose: () => void;
  onAdd: (config: BuiltVisualization) => void;
  isSubmitting?: boolean;
  /** "edit" pre-fills every field from `initial` and changes the modal's
   * title/submit label — used by the pencil icon on an existing custom
   * widget (see DynamicKPIWidget/DashboardKPI) so a widget's config can be
   * revised without deleting and rebuilding it. Defaults to "add". */
  mode?: "add" | "edit";
  /** Only meaningful with mode="edit" — this component always mounts fresh
   * per open (`{builderOpen && <VisualizationBuilder .../>}`), so every
   * piece of state below can safely lazy-init from this once, no effect
   * needed to keep it in sync. */
  initial?: BuiltVisualization;
}

// Every draggable "field" the user sees is a plain catalog entry fetched
// from the backend (GET /analytics/catalog) — there is deliberately no
// hardcoded list of measures/dimensions anywhere in this file. Dragging
// never carries anything beyond a field id + which list it came from.
type FieldKind = "measure" | "dimension";
const DRAG_MIME = "application/x-kpi-field";

interface DragPayload {
  kind: FieldKind;
  id: string;
}

type ChartGroup = "Perbandingan" | "Proporsi" | "Nilai Tunggal" | "Matriks (2 Kategori)" | "Lainnya";

const CHART_GROUP_ORDER: ChartGroup[] = ["Perbandingan", "Proporsi", "Nilai Tunggal", "Matriks (2 Kategori)", "Lainnya"];

const CHART_META: Record<AnalyticsChartType, { label: string; icon: typeof BarChart3; group: ChartGroup }> = {
  bar: { label: "Bar Chart", icon: BarChart3, group: "Perbandingan" },
  horizontal_bar: { label: "Bar Horizontal", icon: BarChartHorizontal, group: "Perbandingan" },
  stacked_bar: { label: "Bar Bertumpuk", icon: Layers, group: "Perbandingan" },
  line: { label: "Line Chart", icon: LineChart, group: "Perbandingan" },
  area: { label: "Area Chart", icon: AreaChart, group: "Perbandingan" },
  radar: { label: "Radar Chart", icon: Radar, group: "Perbandingan" },
  pie: { label: "Pie Chart", icon: PieChart, group: "Proporsi" },
  donut: { label: "Donut Chart", icon: Disc3, group: "Proporsi" },
  treemap: { label: "Treemap", icon: LayoutGrid, group: "Proporsi" },
  funnel: { label: "Funnel", icon: Filter, group: "Proporsi" },
  number_card: { label: "Kartu Angka", icon: Hash, group: "Nilai Tunggal" },
  gauge: { label: "Gauge", icon: GaugeIcon, group: "Nilai Tunggal" },
  heatmap: { label: "Heatmap", icon: Grid3x3, group: "Matriks (2 Kategori)" },
  sankey: { label: "Sankey", icon: Workflow, group: "Matriks (2 Kategori)" },
  scatter: { label: "Scatter Plot", icon: ScatterChart, group: "Lainnya" },
  table: { label: "Tabel", icon: Table2, group: "Lainnya" },
};
const ALL_CHART_TYPES: AnalyticsChartType[] = Object.keys(CHART_META) as AnalyticsChartType[];
const PIE_FAMILY: AnalyticsChartType[] = ["pie", "donut", "treemap", "funnel"];
const MATRIX_FAMILY: AnalyticsChartType[] = ["heatmap", "sankey"];
const MAX_MEASURES = 4;
const MAX_DRILL_LEVELS = 3;
// Chart types with exactly one clickable mark per category row — the only
// ones drill-down can attach a click handler to meaningfully. Excludes
// radar (clicking a spoke isn't an intuitive "drill into this category" for
// a non-technical user) and everything already excluded by other rules
// (scatter/number_card/gauge/heatmap/sankey/table).
const DRILLABLE_CHARTS: AnalyticsChartType[] = ["bar", "horizontal_bar", "stacked_bar", "line", "area", "pie", "donut", "treemap", "funnel"];
// Chart types that ALREADY always show their value with no alternate
// "no value" mode — Number Card is one big number, Gauge has its own
// `detail`, Tabel shows values in every cell, Sankey's node names are
// always-on by ECharts design. "Aktifkan Label" has nothing to toggle for
// these, so the checkbox is disabled rather than a no-op.
const NO_LABEL_TOGGLE_CHARTS: AnalyticsChartType[] = ["table", "number_card", "gauge", "sankey"];
// "Line" (garis tren) only makes sense on a bar-family chart — it traces
// the same values already drawn as bars (see buildEChartsOption.ts).
const TREND_LINE_CHARTS: AnalyticsChartType[] = ["bar", "horizontal_bar", "stacked_bar"];

function labelToggleDisabledReason(vizType: AnalyticsChartType | null): string | undefined {
  if (!vizType) return "Pilih jenis chart terlebih dahulu";
  if (NO_LABEL_TOGGLE_CHARTS.includes(vizType)) return "Nilai selalu ditampilkan untuk jenis chart ini";
  return undefined;
}

function compatibleCharts(
  measures: CatalogField[],
  dimension: CatalogField | null,
  dimension2: CatalogField | null,
  hasDrill: boolean
): AnalyticsChartType[] {
  if (measures.length === 0 || !dimension) return [];

  // Filling Kategori 2 switches into matrix mode: Heatmap/Sankey are the
  // ONLY valid types then (and only with exactly 1 measure) — every other
  // chart type here is built for a single-dimension query shape.
  if (dimension2) {
    return measures.length === 1 ? [...MATRIX_FAMILY] : [];
  }

  let charts: AnalyticsChartType[] = measures[0].compatible_charts ? [...measures[0].compatible_charts] : [...ALL_CHART_TYPES];
  for (const m of measures.slice(1)) {
    const allowed = new Set(m.compatible_charts ?? ALL_CHART_TYPES);
    charts = charts.filter((c) => allowed.has(c));
  }

  // "scatter" is deliberately absent from every measure's own
  // compatible_charts (backend catalog) — its eligibility depends on the
  // PAIR of measures selected, not on any one measure alone — so it's
  // added here explicitly rather than surviving the intersection above.
  if (measures.length === 2 && !charts.includes("scatter")) {
    charts.push("scatter");
  }

  // "Part of a whole" family only makes sense for exactly one measure
  // grouped by a categorical dimension — never for 2+ measures, never for
  // a temporal dimension.
  if (measures.length > 1 || dimension.kind !== "categorical") {
    charts = charts.filter((c) => !PIE_FAMILY.includes(c));
  }
  // Scatter needs an X and a Y — exactly 2 measures, no more, no less.
  if (measures.length !== 2) {
    charts = charts.filter((c) => c !== "scatter");
  }
  // Number Card / Gauge show one aggregate number — exactly 1 measure so
  // it's never ambiguous which measure the number represents.
  if (measures.length !== 1) {
    charts = charts.filter((c) => c !== "number_card" && c !== "gauge");
  }
  // Heatmap/Sankey only ever show up once Kategori 2 is filled (above).
  charts = charts.filter((c) => !MATRIX_FAMILY.includes(c));
  // Drill-down needs a chart with one clickable mark per category.
  if (hasDrill) {
    const drillable = new Set(DRILLABLE_CHARTS);
    charts = charts.filter((c) => drillable.has(c));
  }

  return charts;
}

/**
 * Human-readable reason a chart-type button is disabled — surfaced as its
 * `title` tooltip. Mirrors the gating rules in `compatibleCharts()` above;
 * kept as a separate pure function purely for that explanatory text, no
 * effect on which types are actually (in)compatible.
 */
function incompatibleReason(
  type: AnalyticsChartType,
  measureCount: number,
  dimension: CatalogField | null,
  dimension2: CatalogField | null,
  hasDrill: boolean,
): string {
  if (MATRIX_FAMILY.includes(type)) {
    if (!dimension2) return "Isi 'Kategori 2' terlebih dahulu untuk mengaktifkan Heatmap/Sankey";
    if (measureCount !== 1) return "Heatmap/Sankey hanya mendukung tepat 1 Nilai";
  }
  if (dimension2) return "Kosongkan 'Kategori 2' untuk memakai jenis chart ini";
  if (type === "scatter") return "Scatter Plot butuh tepat 2 Nilai (Nilai pertama = X, kedua = Y)";
  if (type === "number_card" || type === "gauge") return `${type === "gauge" ? "Gauge" : "Kartu Angka"} butuh tepat 1 Nilai`;
  if (PIE_FAMILY.includes(type)) {
    if (measureCount > 1) return "Pie/Donut/Treemap/Funnel hanya mendukung tepat 1 Nilai";
    if (dimension && dimension.kind !== "categorical") return "Pie/Donut/Treemap/Funnel butuh Kategori bertipe kategorikal, bukan tanggal/waktu";
  }
  if (hasDrill && !DRILLABLE_CHARTS.includes(type)) {
    return "Drill-down cuma didukung chart kategori (Bar/Line/Area/Pie/Donut/Treemap/Funnel)";
  }
  return "Tidak cocok untuk kombinasi field ini";
}

function FieldCard({ field, kind }: { field: CatalogField; kind: FieldKind }) {
  const handleDragStart = (event: DragEvent<HTMLDivElement>) => {
    const payload: DragPayload = { kind, id: field.id };
    event.dataTransfer.setData(DRAG_MIME, JSON.stringify(payload));
    event.dataTransfer.setData("text/plain", field.label);
    event.dataTransfer.effectAllowed = "copy";
  };
  return (
    <div
      draggable
      unselectable="on"
      onDragStart={handleDragStart}
      title={field.description}
      className="flex cursor-grab select-none items-center gap-1.5 rounded-lg border border-border bg-card px-2.5 py-1.5 text-xs font-medium text-foreground shadow-sm transition-colors active:cursor-grabbing hover:border-primary/50 hover:bg-primary/5"
    >
      <GripVertical className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
      <span className="truncate">{field.label}</span>
      {field.fans_out && <span className="shrink-0 text-[10px] text-amber-500" title="Bisa dihitung lebih dari sekali per temuan">⚠</span>}
    </div>
  );
}

function DropZone({
  label,
  hint,
  items,
  onDrop,
  onRemove,
  emptyLabel,
  acceptKind,
  disabledHint,
}: {
  label: string;
  hint: string;
  items: CatalogField[];
  onDrop: (payload: DragPayload) => void;
  onRemove: (id: string) => void;
  emptyLabel: string;
  acceptKind: FieldKind;
  /** When set, the zone is rendered dimmed/inert with this text instead of
   * the normal empty/hint text — used for mutually-exclusive zones (Kategori
   * 2 vs. Drill-down) rather than a separate disabled component. */
  disabledHint?: string;
}) {
  const [isOver, setIsOver] = useState(false);
  const disabled = !!disabledHint;
  return (
    <div className={disabled ? "opacity-50" : undefined}>
      <p className="mb-1.5 text-xs font-semibold text-foreground">{label}</p>
      <div
        onDragOver={(e) => {
          if (disabled) return;
          e.preventDefault();
          setIsOver(true);
        }}
        onDragLeave={() => setIsOver(false)}
        onDrop={(e) => {
          e.preventDefault();
          setIsOver(false);
          if (disabled) return;
          const raw = e.dataTransfer.getData(DRAG_MIME);
          if (!raw) return;
          try {
            const payload = JSON.parse(raw) as DragPayload;
            if (payload.kind === acceptKind) onDrop(payload);
          } catch {
            // ignore malformed drag payload
          }
        }}
        className={`flex min-h-[64px] flex-wrap items-center gap-2 rounded-xl border-2 border-dashed p-3 transition-colors ${
          disabled ? "cursor-not-allowed border-border/50 bg-muted/10" : isOver ? "border-primary bg-primary/5" : "border-border bg-muted/20"
        }`}
      >
        {items.length === 0 && <span className="text-xs italic text-muted-foreground">{disabledHint ?? emptyLabel}</span>}
        {items.map((item) => (
          <span
            key={item.id}
            className="inline-flex items-center gap-1.5 rounded-lg bg-primary/10 px-2.5 py-1 text-xs font-semibold text-primary"
          >
            {item.label}
            <button type="button" onClick={() => onRemove(item.id)} aria-label={`Hapus ${item.label}`} className="hover:text-red-500">
              <X className="h-3 w-3" />
            </button>
          </span>
        ))}
      </div>
      <p className="mt-1 text-[11px] text-muted-foreground">{hint}</p>
    </div>
  );
}

/**
 * Power BI-style drag-and-drop panel for building a custom KPI widget:
 * drag a Dimension into "Kategori" and one or more Measures into "Nilai",
 * pick a compatible chart type, then add it to the dashboard. The user
 * never sees a field id, SQL, or formula — only the catalog's label/
 * description text (see analyticsApi.getCatalog).
 */
export function VisualizationBuilder({ onClose, onAdd, isSubmitting = false, mode = "add", initial }: VisualizationBuilderProps) {
  const mounted = useMounted();
  const { data: catalog, isLoading, isError, error } = useQuery({
    queryKey: ["analytics-catalog"],
    queryFn: () => analyticsApi.getCatalog(),
    staleTime: 10 * 60_000,
  });

  const [dimensionId, setDimensionId] = useState<string | null>(() => initial?.dimension ?? null);
  const [dimension2Id, setDimension2Id] = useState<string | null>(() => initial?.dimension2 ?? null);
  const [drillIds, setDrillIds] = useState<string[]>(() => initial?.drillDimensions ?? []);
  const [measureIds, setMeasureIds] = useState<string[]>(() => initial?.measures ?? []);
  const [vizType, setVizType] = useState<AnalyticsChartType | null>(() => (initial?.vizType && initial.vizType !== "list" ? (initial.vizType as AnalyticsChartType) : null));
  const [title, setTitle] = useState(() => initial?.title ?? "");
  const [showLabels, setShowLabels] = useState(() => initial?.showLabels ?? false);
  const [showLabelValues, setShowLabelValues] = useState(() => initial?.showLabelValues ?? false);
  const [showTrendLine, setShowTrendLine] = useState(() => initial?.showTrendLine ?? false);

  const dimensionById = useMemo(() => new Map((catalog?.dimensions ?? []).map((d) => [d.id, d])), [catalog]);
  const measureById = useMemo(() => new Map((catalog?.measures ?? []).map((m) => [m.id, m])), [catalog]);

  const selectedDimension = dimensionId ? dimensionById.get(dimensionId) ?? null : null;
  const selectedDimension2 = dimension2Id ? dimensionById.get(dimension2Id) ?? null : null;
  const selectedDrills = drillIds.map((id) => dimensionById.get(id)).filter((d): d is CatalogField => !!d);
  const selectedMeasures = measureIds.map((id) => measureById.get(id)).filter((m): m is CatalogField => !!m);
  const hasDrill = selectedDrills.length > 0;
  const availableCharts = useMemo(
    () => compatibleCharts(selectedMeasures, selectedDimension, selectedDimension2, hasDrill),
    [selectedMeasures, selectedDimension, selectedDimension2, hasDrill]
  );
  const effectiveVizType: AnalyticsChartType | null = vizType && availableCharts.includes(vizType) ? vizType : availableCharts[0] ?? null;

  const normalizedTitle = title.trim();
  const canAdd = normalizedTitle.length > 0 && !!selectedDimension && selectedMeasures.length > 0 && !!effectiveVizType;

  const handleDropDimension = (payload: DragPayload) => setDimensionId(payload.id);
  const handleDropDimension2 = (payload: DragPayload) => {
    if (drillIds.length > 0) return; // mutually exclusive with drill-down
    setDimension2Id(payload.id);
  };
  const handleDropDrill = (payload: DragPayload) => {
    if (dimension2Id) return; // mutually exclusive with matrix mode
    if (payload.id === dimensionId) return; // drilling into the primary Kategori itself is meaningless
    setDrillIds((prev) => {
      if (prev.includes(payload.id)) return prev;
      if (prev.length >= MAX_DRILL_LEVELS) return prev;
      return [...prev, payload.id];
    });
  };
  const handleDropMeasure = (payload: DragPayload) => {
    setMeasureIds((prev) => {
      if (prev.includes(payload.id)) return prev;
      if (prev.length >= MAX_MEASURES) return prev;
      return [...prev, payload.id];
    });
  };

  const chartsByGroup = useMemo(() => {
    const groups = new Map<ChartGroup, AnalyticsChartType[]>();
    for (const type of ALL_CHART_TYPES) {
      const group = CHART_META[type].group;
      groups.set(group, [...(groups.get(group) ?? []), type]);
    }
    return groups;
  }, []);

  if (!mounted || typeof window === "undefined") return null;

  return createPortal(
    <div
      className="fixed inset-0 z-[110] flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      aria-labelledby="visualization-builder-title"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget && !isSubmitting) onClose();
      }}
    >
      <div className="flex max-h-[90vh] w-full max-w-3xl flex-col rounded-2xl border border-border/60 bg-background shadow-2xl">
        <div className="flex items-center justify-between border-b border-border px-5 py-4">
          <div>
            <h2 id="visualization-builder-title" className="text-base font-semibold text-foreground">
              {mode === "edit" ? "Edit Visualisasi" : "Tambah Visualisasi"}
            </h2>
            <p className="text-xs text-muted-foreground">
              {mode === "edit" ? "Ubah konfigurasi visualisasi ini — perubahan langsung menimpa widget yang ada." : "Seret Kategori dan Nilai untuk menyusun visualisasi Anda sendiri."}
            </p>
          </div>
          <button type="button" onClick={onClose} disabled={isSubmitting} className="rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground">
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="flex-1 space-y-5 overflow-y-auto px-5 py-4">
          {isLoading ? (
            <div className="flex items-center justify-center py-10">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : isError ? (
            <p className="py-10 text-center text-sm text-red-500">{getApiErrorMessage(error, "Gagal memuat daftar field")}</p>
          ) : (
            <>
              <div>
                <div className="mb-1.5 flex items-center justify-between gap-3">
                  <label htmlFor="custom-visualization-title" className="text-xs font-semibold text-foreground">
                    Judul visualisasi
                  </label>
                  <span className="text-[10px] tabular-nums text-muted-foreground">{title.length}/100</span>
                </div>
                <input
                  id="custom-visualization-title"
                  type="text"
                  value={title}
                  maxLength={100}
                  onChange={(event) => setTitle(event.target.value)}
                  placeholder="Contoh: Tren Temuan per Bulan"
                  className="h-10 w-full rounded-xl border border-border bg-card px-3 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-primary focus:ring-2 focus:ring-primary/20"
                />
                <p className="mt-1 text-[11px] text-muted-foreground">Gunakan nama yang menjelaskan isi dan tujuan grafik.</p>
              </div>

              <div>
                <p className="mb-1.5 text-xs font-semibold text-foreground">Field tersedia</p>
                <div className="space-y-2 rounded-xl border border-border bg-muted/10 p-3">
                  <div>
                    <p className="mb-1 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Kategori</p>
                    <div className="flex flex-wrap gap-1.5">
                      {(catalog?.dimensions ?? []).map((d) => (
                        <FieldCard key={d.id} field={d} kind="dimension" />
                      ))}
                    </div>
                  </div>
                  <div>
                    <p className="mb-1 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Nilai</p>
                    <div className="flex flex-wrap gap-1.5">
                      {(catalog?.measures ?? []).map((m) => (
                        <FieldCard key={m.id} field={m} kind="measure" />
                      ))}
                    </div>
                  </div>
                </div>
              </div>

              <div className="grid gap-4 sm:grid-cols-2">
                <DropZone
                  label="Kategori"
                  hint="Satu field untuk mengelompokkan data (sumbu kategori)."
                  items={selectedDimension ? [selectedDimension] : []}
                  onDrop={handleDropDimension}
                  onRemove={() => setDimensionId(null)}
                  emptyLabel="Seret 1 field Kategori ke sini"
                  acceptKind="dimension"
                />
                <DropZone
                  label={`Nilai (maks. ${MAX_MEASURES})`}
                  hint="Satu atau lebih angka yang ingin ditampilkan."
                  items={selectedMeasures}
                  onDrop={handleDropMeasure}
                  onRemove={(id) => setMeasureIds((prev) => prev.filter((m) => m !== id))}
                  emptyLabel="Seret 1+ field Nilai ke sini"
                  acceptKind="measure"
                />
                <DropZone
                  label="Kategori 2 (opsional)"
                  hint="Isi untuk membuat Heatmap/Sankey (matriks 2 kategori) — kosongkan untuk chart biasa."
                  items={selectedDimension2 ? [selectedDimension2] : []}
                  onDrop={handleDropDimension2}
                  onRemove={() => setDimension2Id(null)}
                  emptyLabel="Seret 1 field Kategori untuk mode Heatmap/Sankey"
                  acceptKind="dimension"
                  disabledHint={hasDrill ? "Tidak bisa dipakai bersama Drill-down" : undefined}
                />
                <DropZone
                  label={`Drill-down (opsional, maks. ${MAX_DRILL_LEVELS} level tambahan)`}
                  hint="Seret Kategori secara berurutan — klik chart nanti akan turun ke level berikutnya."
                  items={selectedDrills}
                  onDrop={handleDropDrill}
                  onRemove={(id) => setDrillIds((prev) => prev.filter((d) => d !== id))}
                  emptyLabel="Seret Kategori berurutan untuk drill-down"
                  acceptKind="dimension"
                  disabledHint={selectedDimension2 ? "Tidak bisa dipakai bersama Heatmap/Sankey (Kategori 2)" : undefined}
                />
              </div>

              <div>
                <p className="mb-1.5 text-xs font-semibold text-foreground">Jenis Visualisasi</p>
                {availableCharts.length === 0 ? (
                  <p className="text-xs italic text-muted-foreground">
                    {selectedDimension2
                      ? "Heatmap/Sankey butuh tepat 1 field Nilai — kurangi jadi 1, atau kosongkan Kategori 2 untuk chart biasa."
                      : "Pilih Kategori dan Nilai terlebih dahulu untuk melihat pilihan visualisasi."}
                  </p>
                ) : (
                  <div className="space-y-3">
                    {CHART_GROUP_ORDER.map((group) => {
                      const types = chartsByGroup.get(group) ?? [];
                      if (types.length === 0) return null;
                      return (
                        <div key={group}>
                          <p className="mb-1 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">{group}</p>
                          <div className="flex flex-wrap gap-2">
                            {types.map((type) => {
                              const meta = CHART_META[type];
                              const Icon = meta.icon;
                              const isCompatible = availableCharts.includes(type);
                              const isActive = effectiveVizType === type;
                              return (
                                <button
                                  key={type}
                                  type="button"
                                  disabled={!isCompatible}
                                  onClick={() => setVizType(type)}
                                  title={
                                    !isCompatible
                                      ? incompatibleReason(type, selectedMeasures.length, selectedDimension, selectedDimension2, hasDrill)
                                      : type === "gauge"
                                        ? "Skala otomatis, bukan target bisnis"
                                        : meta.label
                                  }
                                  className={`flex items-center gap-1.5 rounded-lg border px-3 py-1.5 text-xs font-medium transition-colors ${
                                    !isCompatible
                                      ? "cursor-not-allowed border-border/50 text-muted-foreground/50"
                                      : isActive
                                        ? "border-primary bg-primary/10 text-primary"
                                        : "border-border text-foreground hover:border-primary/50"
                                  }`}
                                >
                                  <Icon className="h-3.5 w-3.5" />
                                  {meta.label}
                                </button>
                              );
                            })}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>

              <div>
                <p className="mb-1.5 text-xs font-semibold text-foreground">Opsi Tampilan</p>
                <div className="flex flex-wrap gap-4">
                  <label
                    title={labelToggleDisabledReason(effectiveVizType) ?? "Tampilkan nilai/nama langsung di atas chart, bukan cuma lewat hover"}
                    className={`flex items-center gap-1.5 text-xs font-medium ${labelToggleDisabledReason(effectiveVizType) ? "cursor-not-allowed text-muted-foreground/50" : "cursor-pointer text-foreground"}`}
                  >
                    <input
                      type="checkbox"
                      checked={showLabels}
                      disabled={!!labelToggleDisabledReason(effectiveVizType)}
                      onChange={(e) => setShowLabels(e.target.checked)}
                      className="h-3.5 w-3.5 rounded border-border accent-primary"
                    />
                    Aktifkan Label
                  </label>
                  <label
                    title={
                      !showLabels || !effectiveVizType || !PIE_FAMILY.includes(effectiveVizType)
                        ? "Hanya berlaku untuk Pie/Donut/Treemap/Funnel dengan Label aktif"
                        : "Tampilkan angka nilai di label, bukan cuma nama kategori"
                    }
                    className={`flex items-center gap-1.5 text-xs font-medium ${
                      !showLabels || !effectiveVizType || !PIE_FAMILY.includes(effectiveVizType) ? "cursor-not-allowed text-muted-foreground/50" : "cursor-pointer text-foreground"
                    }`}
                  >
                    <input
                      type="checkbox"
                      checked={showLabelValues}
                      disabled={!showLabels || !effectiveVizType || !PIE_FAMILY.includes(effectiveVizType)}
                      onChange={(e) => setShowLabelValues(e.target.checked)}
                      className="h-3.5 w-3.5 rounded border-border accent-primary"
                    />
                    Tampilkan Nilai di Label
                  </label>
                  <label
                    title={
                      !effectiveVizType || !TREND_LINE_CHARTS.includes(effectiveVizType)
                        ? "Hanya berlaku untuk Bar Chart/Bar Horizontal/Bar Bertumpuk"
                        : "Tambah garis yang mengikuti nilai batangnya sendiri"
                    }
                    className={`flex items-center gap-1.5 text-xs font-medium ${
                      !effectiveVizType || !TREND_LINE_CHARTS.includes(effectiveVizType) ? "cursor-not-allowed text-muted-foreground/50" : "cursor-pointer text-foreground"
                    }`}
                  >
                    <input
                      type="checkbox"
                      checked={showTrendLine}
                      disabled={!effectiveVizType || !TREND_LINE_CHARTS.includes(effectiveVizType)}
                      onChange={(e) => setShowTrendLine(e.target.checked)}
                      className="h-3.5 w-3.5 rounded border-border accent-primary"
                    />
                    Tambahkan Garis Tren (Line)
                  </label>
                </div>
              </div>

              {selectedMeasures.some((m) => m.fans_out) && (
                <p className="rounded-lg border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-[11px] text-amber-700 dark:text-amber-400">
                  ⚠ Salah satu field yang dipilih bisa terhitung lebih dari sekali per temuan (lihat deskripsi field). Totalnya bisa lebih besar dari jumlah temuan unik.
                </p>
              )}
            </>
          )}
        </div>

        <div className="flex items-center justify-end gap-2 border-t border-border px-5 py-4">
          <Button variant="outline" size="sm" onClick={onClose} disabled={isSubmitting}>
            Batal
          </Button>
          <Button
            size="sm"
            className="gap-2"
            disabled={!canAdd || isSubmitting}
            isLoading={isSubmitting}
            onClick={() => {
              if (!selectedDimension || !effectiveVizType || selectedMeasures.length === 0) return;
              onAdd({
                title: normalizedTitle,
                measures: selectedMeasures.map((m) => m.id),
                dimension: selectedDimension.id,
                dimension2: selectedDimension2?.id,
                drillDimensions: selectedDrills.length ? selectedDrills.map((d) => d.id) : undefined,
                showLabels,
                showLabelValues,
                showTrendLine,
                vizType: effectiveVizType,
              });
            }}
          >
            <Plus className="h-4 w-4" />
            {mode === "edit" ? "Simpan Perubahan" : "Tambah ke Dashboard"}
          </Button>
        </div>
      </div>
    </div>,
    document.body
  );
}
