"use client";

import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import ReactECharts from "echarts-for-react";
import { AlertTriangle, ChevronRight, Loader2, Pencil, RotateCcw } from "lucide-react";
import { analyticsApi, type AnalyticsQueryResult } from "@/lib/api/analytics.api";
import { kpiShareApi } from "@/lib/api/kpi-share.api";
import { getApiErrorStatus } from "@/lib/api/error";
import { useKPIDashboard } from "@/components/dashboard/kpi/KPIDashboardContext";
import { aggregateSingleMeasure, buildEChartsOption, formatMeasureValue } from "./buildEChartsOption";
import { DrillDownToolbar } from "./DrillDownToolbar";
import type { VizType } from "@/components/dashboard/types";

export interface DynamicKPIWidgetProps {
  widgetId?: string;
  title: string;
  measures: string[];
  dimension: string;
  /** Only set for Heatmap/Sankey (a dim1 x dim2 x 1 measure matrix). */
  dimension2?: string;
  /** Ordered drill-down levels beyond `dimension` (level 0). Moving down
   * displays every value of the next dimension, without selecting or
   * filtering a single value from the previous level. */
  drillDimensions?: string[];
  /** Pure rendering options — see buildEChartsOption.ts's ChartDisplayOptions. */
  showLabels?: boolean;
  showLabelValues?: boolean;
  showTrendLine?: boolean;
  vizType?: VizType;
  /** Badge shown in the widget header. The builder uses "Preview" while
   * persisted dashboard widgets keep the default "Custom" label. */
  badgeLabel?: string;
  /** When set, renders a small pencil icon next to the "Custom" badge that
   * opens the builder pre-filled with this widget's config (see
   * DashboardKPI.tsx). */
  onEdit?: () => void;
}

// Mirrors VisualizationBuilder's DRILLABLE_CHARTS — chart types that step
// through one hierarchy level at a time via the DrillDownToolbar. "table"
// is deliberately absent: it shows every configured level at once as a
// static multi-column breakdown instead (see the isTable branch below and
// DynamicTable's hierarchy/path rendering), no progressive stepping. Kept
// in sync manually (small, stable list); duplicating rather than importing
// avoids a builder-UI -> rendering-widget dependency for a single const array.
const DRILLABLE_VIZ_TYPES: VizType[] = ["bar", "horizontal_bar", "stacked_bar", "line", "area", "pie", "donut", "treemap", "funnel"];

/**
 * Generic ECharts-rendered widget for the Custom KPI Visualization Builder.
 * Every field a user could pick (measures/dimension) was already validated
 * catalog IDs at build time in VisualizationBuilder — this component's job
 * is purely to fetch + render, and to degrade gracefully if the catalog
 * later drops one of those IDs (see `isInvalidConfig` below).
 */
export function DynamicKPIWidget({
  widgetId,
  title,
  measures,
  dimension,
  dimension2,
  drillDimensions,
  showLabels,
  showLabelValues,
  showTrendLine,
  vizType,
  badgeLabel = "Custom",
  onEdit,
}: DynamicKPIWidgetProps) {
  const { effectivePlant, isPublic, publicShareToken, analyticsCatalog } = useKPIDashboard();

  // Sorted so two widgets picking the same measures in a different order
  // (or the same widget re-rendering with a new-but-equivalent array
  // reference) share one cache entry instead of missing needlessly.
  const sortedMeasures = useMemo(() => [...measures].sort(), [measures]);
  const chartVizType: VizType = vizType && vizType !== "list" ? vizType : "bar";

  // Current hierarchy level — session-only UI state, never persisted
  // (reloading a widget always starts back at level 0/`dimension`). Reset
  // whenever the widget's underlying config identity changes (e.g. its
  // measures/dimension/drillDimensions were rebuilt) so a stale filter never
  // survives into a differently-configured widget instance. Uses React's
  // "adjust state during render" pattern (comparing against a signature
  // computed from props) rather than an effect + setState, per
  // https://react.dev/learn/you-might-not-need-an-effect — no extra
  // render/commit cycle, and avoids the react-hooks/set-state-in-effect rule.
  const configSignature = `${dimension}|${dimension2 ?? ""}|${(drillDimensions ?? []).join(",")}|${sortedMeasures.join(",")}`;
  const [drillLevel, setDrillLevel] = useState(0);
  const [lastConfigSignature, setLastConfigSignature] = useState(configSignature);
  if (configSignature !== lastConfigSignature) {
    setLastConfigSignature(configSignature);
    setDrillLevel(0);
  }

  const hierarchy = useMemo(() => [dimension, ...(drillDimensions ?? [])], [dimension, drillDimensions]);
  // Table never drills progressively — it always queries every configured
  // level at once and renders them as stacked breadcrumb columns (see
  // DynamicTable below), so it always runs as if fully drilled-in, with no
  // toolbar/breadcrumb-bar and no drillLevel state of its own.
  const isTable = chartVizType === "table";
  const effectiveDrillLevel = isTable ? hierarchy.length - 1 : drillLevel;
  const currentDimension = hierarchy[effectiveDrillLevel] ?? dimension;
  const canDrillDeeper = !isTable && drillLevel < hierarchy.length - 1;
  const isDrillableChart = !isTable && DRILLABLE_VIZ_TYPES.includes(chartVizType);
  const nextDimensionId = canDrillDeeper ? hierarchy[drillLevel + 1] : undefined;

  const { data: fetchedCatalog } = useQuery({
    queryKey: ["analytics-catalog"],
    queryFn: () => analyticsApi.getCatalog(),
    enabled: !isPublic && !!drillDimensions?.length,
    staleTime: 10 * 60_000,
  });
  const catalog = analyticsCatalog ?? fetchedCatalog;

  const queryKey = [
    "analytics-query",
    isPublic ? "public" : "private",
    isPublic ? widgetId : "",
    sortedMeasures,
    currentDimension,
    dimension2 ?? "",
    effectiveDrillLevel,
    effectivePlant || "",
  ] as const;

  const { data, isLoading, isError, error, refetch, isFetching } = useQuery<AnalyticsQueryResult>({
    queryKey,
    queryFn: ({ signal }) =>
      isPublic && publicShareToken && widgetId
        ? kpiShareApi.runCustomQuery(publicShareToken, widgetId, effectiveDrillLevel, signal)
        : analyticsApi.runQuery({
            measures: sortedMeasures,
            dimension: currentDimension,
            hierarchyDimensions: hierarchy.slice(0, effectiveDrillLevel + 1),
            dimension2,
            plantId: effectivePlant || undefined,
            signal,
          }),
    enabled: !isPublic || (!!publicShareToken && !!widgetId),
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

  const nextDimensionLabel = nextDimensionId
    ? catalog?.dimensions.find((dimensionItem) => dimensionItem.id === nextDimensionId)?.label ?? nextDimensionId
    : undefined;

  const dimensionLabel = (id: string) => catalog?.dimensions.find((item) => item.id === id)?.label ?? id;

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden rounded-xl border border-border bg-card p-5 shadow-sm">
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
          <span className="rounded bg-primary/10 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-primary">{badgeLabel}</span>
        </div>
      </div>

      {drillLevel > 0 && (
        <div className="mb-2 flex flex-wrap items-center gap-0.5 text-[11px]">
          {hierarchy.slice(0, drillLevel + 1).map((dimensionId, idx) => {
            const isLast = idx === drillLevel;
            return (
              <span key={`${dimensionId}-${idx}`} className="flex items-center gap-0.5">
                {idx > 0 && <ChevronRight className="h-3 w-3 text-muted-foreground" />}
                {isLast ? (
                  <span className="font-semibold text-foreground">{dimensionLabel(dimensionId)}</span>
                ) : (
                  <button
                    type="button"
                    onClick={() => setDrillLevel(idx)}
                    className="font-medium text-primary hover:underline"
                  >
                    {dimensionLabel(dimensionId)}
                  </button>
                )}
              </span>
            );
          })}
        </div>
      )}

      {!!drillDimensions?.length && isDrillableChart && data && data.rows.length > 0 && (
        <DrillDownToolbar
          currentDimensionLabel={data.dimension.label}
          nextDimensionLabel={nextDimensionLabel}
          onDrillDown={() => setDrillLevel((level) => Math.min(level + 1, hierarchy.length - 1))}
          onDrillUp={() => setDrillLevel((level) => Math.max(0, level - 1))}
          onReset={() => setDrillLevel(0)}
          canDrillUp={drillLevel > 0}
          isBusy={isFetching}
        />
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
              style={{ height: "100%", width: "100%", minHeight: 120 }}
            />
          )
        )}
      </div>

      {isFetching && !isLoading && (
        <div className="mt-1 flex items-center gap-1 text-[10px] text-muted-foreground">
          <Loader2 className="h-3 w-3 animate-spin" /> Memperbarui...
        </div>
      )}
      {data?.meta?.truncated && (
        <div className="mt-1 flex items-center gap-1 text-[10px] font-medium text-amber-600">
          <AlertTriangle className="h-3 w-3" /> Menampilkan {data.meta.returned_rows} dari {data.meta.total_rows} value teratas.
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
  const categoryColumns = result.hierarchy?.length
    ? result.hierarchy
    : [result.dimension];

  return (
    <div className="h-full overflow-auto rounded-lg border border-border">
      <table className="min-w-full text-left text-xs">
        <thead className="sticky top-0 bg-muted/60">
          <tr>
            {categoryColumns.map((dimensionItem) => (
              <th
                key={dimensionItem.id}
                className="whitespace-nowrap px-3 py-2 font-semibold text-muted-foreground"
              >
                {dimensionItem.label}
              </th>
            ))}
            {result.measures.map((m) => (
              <th key={m.id} className="whitespace-nowrap px-3 py-2 font-semibold text-muted-foreground">
                {m.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-border">
          {result.rows.map((row, idx) => {
            const hierarchyPath = Array.isArray(row.path) ? row.path : [];

            return (
              <tr key={idx} className="hover:bg-muted/30">
                {categoryColumns.map((dimensionItem, columnIndex) => {
                  const pathItem = hierarchyPath[columnIndex];
                  const legacyValue = columnIndex === categoryColumns.length - 1
                    ? row.category ?? row.key
                    : null;

                  return (
                    <td
                      key={`${dimensionItem.id}-${columnIndex}`}
                      className="whitespace-nowrap px-3 py-2 font-medium text-foreground"
                    >
                      {pathItem?.category || pathItem?.key || String(legacyValue ?? "-")}
                    </td>
                  );
                })}
                {result.measures.map((m) => {
                  const raw = row[m.id];
                  return (
                    <td key={m.id} className="whitespace-nowrap px-3 py-2 text-foreground">
                      {formatMeasureValue(typeof raw === "number" ? raw : null, m.format)}
                    </td>
                  );
                })}
              </tr>
            );
          })}
          {result.rows.length === 0 && (
            <tr>
              <td
                colSpan={result.measures.length + categoryColumns.length}
                className="px-3 py-6 text-center italic text-muted-foreground"
              >
                Belum ada data
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}
