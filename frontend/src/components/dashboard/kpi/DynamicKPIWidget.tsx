"use client";

import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import ReactECharts from "echarts-for-react";
import { AlertTriangle, ChevronRight, Loader2, Pencil, RotateCcw } from "lucide-react";
import { analyticsApi, type AnalyticsQueryResult } from "@/lib/api/analytics.api";
import { getApiErrorStatus } from "@/lib/api/error";
import { useKPIDashboard } from "@/components/dashboard/kpi/KPIDashboardContext";
import { aggregateSingleMeasure, buildEChartsOption, formatMeasureValue } from "./buildEChartsOption";
import type { VizType } from "@/components/dashboard/types";

export interface DynamicKPIWidgetProps {
  title: string;
  measures: string[];
  dimension: string;
  /** Only set for Heatmap/Sankey (a dim1 x dim2 x 1 measure matrix). */
  dimension2?: string;
  /** Ordered drill-down levels beyond `dimension` (level 0) — clicking a
   * category mark in a compatible chart descends one level, filtered to the
   * clicked value. Mutually exclusive with dimension2. */
  drillDimensions?: string[];
  /** Pure rendering options — see buildEChartsOption.ts's ChartDisplayOptions. */
  showLabels?: boolean;
  showLabelValues?: boolean;
  showTrendLine?: boolean;
  vizType?: VizType;
  /** When set, renders a small pencil icon next to the "Custom" badge that
   * opens the builder pre-filled with this widget's config (see
   * DashboardKPI.tsx). Absent for the 6 built-in widgets — they never
   * render DynamicKPIWidget at all. */
  onEdit?: () => void;
}

/** One accumulated drill-down step: which dimension was pinned, its exact
 * key value (used to filter), and its display label (used in the
 * breadcrumb — never sent to the backend). */
interface DrillStep {
  dimensionId: string;
  value: string;
  label: string;
}

// Mirrors VisualizationBuilder's DRILLABLE_CHARTS — chart types with one
// clickable mark per category row. Kept in sync manually (small, stable
// list); duplicating rather than importing avoids a builder-UI ->
// rendering-widget dependency for a single const array.
const DRILLABLE_VIZ_TYPES: VizType[] = ["bar", "horizontal_bar", "stacked_bar", "line", "area", "pie", "donut", "treemap", "funnel"];

/**
 * Generic ECharts-rendered widget for the Custom KPI Visualization Builder.
 * Every field a user could pick (measures/dimension) was already validated
 * catalog IDs at build time in VisualizationBuilder — this component's job
 * is purely to fetch + render, and to degrade gracefully if the catalog
 * later drops one of those IDs (see `isInvalidConfig` below).
 */
export function DynamicKPIWidget({
  title,
  measures,
  dimension,
  dimension2,
  drillDimensions,
  showLabels,
  showLabelValues,
  showTrendLine,
  vizType,
  onEdit,
}: DynamicKPIWidgetProps) {
  const { effectivePlant } = useKPIDashboard();

  // Sorted so two widgets picking the same measures in a different order
  // (or the same widget re-rendering with a new-but-equivalent array
  // reference) share one cache entry instead of missing needlessly.
  const sortedMeasures = useMemo(() => [...measures].sort(), [measures]);
  const chartVizType: VizType = vizType && vizType !== "list" ? vizType : "bar";

  // Accumulated drill-down path — session-only UI state, never persisted
  // (reloading a widget always starts back at level 0/`dimension`). Reset
  // whenever the widget's underlying config identity changes (e.g. its
  // measures/dimension/drillDimensions were rebuilt) so a stale filter never
  // survives into a differently-configured widget instance. Uses React's
  // "adjust state during render" pattern (comparing against a signature
  // computed from props) rather than an effect + setState, per
  // https://react.dev/learn/you-might-not-need-an-effect — no extra
  // render/commit cycle, and avoids the react-hooks/set-state-in-effect rule.
  const configSignature = `${dimension}|${(drillDimensions ?? []).join(",")}|${sortedMeasures.join(",")}`;
  const [drillPath, setDrillPath] = useState<DrillStep[]>([]);
  const [lastConfigSignature, setLastConfigSignature] = useState(configSignature);
  if (configSignature !== lastConfigSignature) {
    setLastConfigSignature(configSignature);
    setDrillPath([]);
  }

  const currentDimension = drillPath.length === 0 ? dimension : drillDimensions![drillPath.length - 1];
  const canDrillDeeper = !!drillDimensions && drillPath.length < drillDimensions.length;
  const isDrillableChart = DRILLABLE_VIZ_TYPES.includes(chartVizType);
  const filters = useMemo(() => drillPath.map((p) => ({ dimensionId: p.dimensionId, value: p.value })), [drillPath]);

  const queryKey = [
    "analytics-query",
    sortedMeasures,
    currentDimension,
    dimension2 ?? "",
    drillPath.map((p) => `${p.dimensionId}:${p.value}`).join("|"),
    chartVizType,
    effectivePlant || "",
  ] as const;

  const { data, isLoading, isError, error, refetch, isFetching } = useQuery<AnalyticsQueryResult>({
    queryKey,
    queryFn: ({ signal }) =>
      analyticsApi.runQuery({
        measures: sortedMeasures,
        dimension: currentDimension,
        dimension2,
        filters,
        plantId: effectivePlant || undefined,
        signal,
      }),
    staleTime: 30_000,
    retry: 1,
  });

  // A 400 here means the backend rejected an unknown measure/dimension id —
  // i.e. this widget's saved field(s) no longer exist in the catalog.
  // Every other status (500, network, timeout) is a transient failure with
  // a Retry button instead.
  const isInvalidConfig = isError && getApiErrorStatus(error) === 400;

  const option = useMemo(
    () => (data ? buildEChartsOption(data, chartVizType, { showLabels, showLabelValues, showTrendLine }) : null),
    [data, chartVizType, showLabels, showLabelValues, showTrendLine]
  );

  // Clicking a category mark (bar/slice/point) descends one drill level,
  // filtered to the row that mark represents. Reads straight from the fetch
  // result rather than the ECharts option — `params.name` is the category
  // LABEL ECharts shows, but the filter sent to the backend must be the raw
  // `key` (e.g. an id), never the label.
  const handleChartClick = (params: { name?: string }) => {
    if (!canDrillDeeper || !drillDimensions || !data || !params.name) return;
    const row = data.rows.find((r) => String(r.category ?? r.key ?? "") === params.name);
    if (!row) return;
    const value = row.key ?? row.category;
    if (value === null || value === undefined) return;
    setDrillPath((prev) => [...prev, { dimensionId: currentDimension, value: String(value), label: params.name! }]);
  };

  return (
    <div className="flex h-full flex-col rounded-xl border border-border bg-card p-5 shadow-sm">
      <div className="mb-3 flex items-center justify-between gap-2 border-b border-border pb-3">
        <h3 className="truncate text-base font-semibold text-foreground">{title}</h3>
        <div className="flex shrink-0 items-center gap-1.5">
          {onEdit && (
            <button
              type="button"
              onClick={onEdit}
              aria-label="Edit visualisasi"
              title="Edit visualisasi"
              className="rounded p-0.5 text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <Pencil className="h-3.5 w-3.5" />
            </button>
          )}
          <span className="rounded bg-primary/10 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-primary">Custom</span>
        </div>
      </div>

      {drillPath.length > 0 && (
        <div className="mb-2 flex flex-wrap items-center gap-0.5 text-[11px]">
          <button type="button" onClick={() => setDrillPath([])} className="font-medium text-primary hover:underline">
            Semua
          </button>
          {drillPath.map((step, idx) => {
            const isLast = idx === drillPath.length - 1;
            return (
              <span key={`${step.dimensionId}-${idx}`} className="flex items-center gap-0.5">
                <ChevronRight className="h-3 w-3 text-muted-foreground" />
                {isLast ? (
                  <span className="font-semibold text-foreground">{step.label}</span>
                ) : (
                  <button
                    type="button"
                    onClick={() => setDrillPath((prev) => prev.slice(0, idx + 1))}
                    className="font-medium text-primary hover:underline"
                  >
                    {step.label}
                  </button>
                )}
              </span>
            );
          })}
        </div>
      )}

      <div className="min-h-0 flex-1">
        {isLoading ? (
          <div className="flex h-full items-center justify-center">
            <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
          </div>
        ) : isInvalidConfig ? (
          <div className="flex h-full flex-col items-center justify-center gap-1.5 p-4 text-center">
            <AlertTriangle className="h-6 w-6 text-amber-500" />
            <p className="text-xs font-medium text-foreground">Konfigurasi widget sudah tidak tersedia</p>
            <p className="text-[11px] text-muted-foreground">
              Field yang dipakai widget ini sudah tidak ada di katalog. Hapus widget ini dan buat ulang.
            </p>
          </div>
        ) : isError ? (
          <div className="flex h-full flex-col items-center justify-center gap-2 p-4 text-center">
            <AlertTriangle className="h-6 w-6 text-red-500" />
            <p className="text-xs font-medium text-foreground">Gagal memuat data visualisasi</p>
            <button
              type="button"
              onClick={() => refetch()}
              className="mt-1 inline-flex items-center gap-1.5 rounded-lg border border-border px-2.5 py-1 text-[11px] font-medium text-foreground hover:border-primary/50 hover:text-primary"
            >
              <RotateCcw className="h-3 w-3" /> Coba Lagi
            </button>
          </div>
        ) : !data || data.rows.length === 0 ? (
          <div className="flex h-full items-center justify-center p-4 text-center text-xs italic text-muted-foreground">
            Belum ada data untuk kombinasi ini
          </div>
        ) : chartVizType === "table" ? (
          <DynamicTable result={data} />
        ) : chartVizType === "number_card" ? (
          <NumberCard result={data} />
        ) : (
          option && (
            <ReactECharts
              option={option}
              notMerge
              lazyUpdate
              opts={{ renderer: "canvas" }}
              style={{ height: "100%", width: "100%", minHeight: 200, cursor: canDrillDeeper && isDrillableChart ? "pointer" : undefined }}
              onEvents={canDrillDeeper && isDrillableChart ? { click: handleChartClick } : undefined}
            />
          )
        )}
      </div>

      {isFetching && !isLoading && (
        <div className="mt-1 flex items-center gap-1 text-[10px] text-muted-foreground">
          <Loader2 className="h-3 w-3 animate-spin" /> Memperbarui...
        </div>
      )}
    </div>
  );
}

/** Big-number "Card" visual — sums the single selected measure across
 * every returned row (there's no per-category breakdown to show). Not an
 * ECharts series at all, same bypass pattern as DynamicTable/"table". */
function NumberCard({ result }: { result: AnalyticsQueryResult }) {
  const aggregate = aggregateSingleMeasure(result);
  if (!aggregate) {
    return <div className="flex h-full items-center justify-center text-xs italic text-muted-foreground">Belum ada data</div>;
  }
  return (
    <div className="flex h-full flex-col items-center justify-center gap-1 text-center">
      <span className="text-4xl font-bold tabular-nums text-foreground">{formatMeasureValue(aggregate.value, aggregate.measure.format)}</span>
      <span className="text-xs font-medium text-muted-foreground">{aggregate.measure.label}</span>
    </div>
  );
}

function DynamicTable({ result }: { result: AnalyticsQueryResult }) {
  return (
    <div className="h-full overflow-auto rounded-lg border border-border">
      <table className="w-full text-left text-xs">
        <thead className="sticky top-0 bg-muted/60">
          <tr>
            <th className="px-3 py-2 font-semibold text-muted-foreground">{result.dimension.label}</th>
            {result.measures.map((m) => (
              <th key={m.id} className="px-3 py-2 font-semibold text-muted-foreground">
                {m.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-border">
          {result.rows.map((row, idx) => (
            <tr key={idx} className="hover:bg-muted/30">
              <td className="px-3 py-2 font-medium text-foreground">{String(row.category ?? "")}</td>
              {result.measures.map((m) => {
                const raw = row[m.id];
                return (
                  <td key={m.id} className="px-3 py-2 text-foreground">
                    {formatMeasureValue(typeof raw === "number" ? raw : null, m.format)}
                  </td>
                );
              })}
            </tr>
          ))}
          {result.rows.length === 0 && (
            <tr>
              <td colSpan={result.measures.length + 1} className="px-3 py-6 text-center italic text-muted-foreground">
                Belum ada data
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}
