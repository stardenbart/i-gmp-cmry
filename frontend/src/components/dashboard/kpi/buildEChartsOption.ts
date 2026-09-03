import type { AnalyticsQueryResult, CatalogField } from "@/lib/api/analytics.api";
import type { VizType } from "@/components/dashboard/types";

// Kept deliberately separate from DynamicKPIWidget.tsx per the approved
// plan: option-building is pure/testable and independent of React's
// lifecycle. echarts-for-react types its `option` prop as `any` (see
// node_modules/echarts-for-react/lib/types.d.ts) — that is the library's
// own looseness, not this file's; every type declared HERE is concrete, so
// nothing in this codebase ever needs to write `any` itself.
//
// "table" and "number_card" never reach this file — DynamicKPIWidget
// renders those itself (see its own component) since neither is an
// ECharts series at all.
const MEASURE_COLORS = ["#2563eb", "#10b981", "#f59e0b", "#ef4444"];
const NUMBER_FORMATTER_CACHE = new Map<number, Intl.NumberFormat>();

function numberFormatter(decimals: number): Intl.NumberFormat {
  let formatter = NUMBER_FORMATTER_CACHE.get(decimals);
  if (!formatter) {
    formatter = new Intl.NumberFormat("id-ID", { minimumFractionDigits: decimals, maximumFractionDigits: decimals });
    NUMBER_FORMATTER_CACHE.set(decimals, formatter);
  }
  return formatter;
}

export function formatMeasureValue(value: number | null | undefined, format: CatalogField["format"]): string {
  if (value === null || value === undefined || Number.isNaN(value)) return "-";
  const decimals = format?.decimals ?? 0;
  const formatted = numberFormatter(decimals).format(value);
  if (format?.style === "percent") return `${formatted}%`;
  if (format?.suffix) return `${formatted} ${format.suffix}`;
  return formatted;
}

/** Reads a row's numeric value for `measureId`, tolerating the generic
 * `Record<string, AnalyticsRowValue>` shape (never `any`). */
function numericValue(row: AnalyticsQueryResult["rows"][number], key: string): number | null {
  const raw = row[key];
  return typeof raw === "number" ? raw : null;
}

/** Reads a row's category label for a given field ("category"/"category2"),
 * falling back to its key, then to an empty string. */
function labelOf(row: AnalyticsQueryResult["rows"][number], labelKey: string, keyKey: string): string {
  if (labelKey === "category" && Array.isArray(row.path)) {
    const pathLabel = row.path
      .map((item) => item.category || item.key)
      .filter(Boolean)
      .join("\n");
    if (pathLabel) return pathLabel;
  }
  const label = row[labelKey];
  if (typeof label === "string") return label;
  if (typeof label === "number") return String(label);
  const key = row[keyKey];
  if (typeof key === "string") return key;
  if (typeof key === "number") return String(key);
  return "";
}

/** Sums a single measure's values across every returned row — used by both
 * Number Card and Gauge, which show one aggregate number rather than a
 * per-category breakdown. Pure, no React/DOM. */
export function aggregateSingleMeasure(result: AnalyticsQueryResult): { value: number; measure: CatalogField } | null {
  const measure = result.measures[0];
  if (!measure) return null;
  const value = result.rows.reduce((sum, row) => sum + (numericValue(row, measure.id) ?? 0), 0);
  return { value, measure };
}

/** Rounds `value` up to a "nice" round number (1/2/5 x 10^n) for Gauge's
 * `max` when the measure isn't a natural 0-100 percent — a display
 * heuristic, not a real business target (there is no target/min/max
 * concept in the data model today). */
export function computeNiceMax(value: number): number {
  if (value <= 0) return 10;
  const magnitude = Math.pow(10, Math.floor(Math.log10(value)));
  const normalized = value / magnitude;
  let niceNormalized: number;
  if (normalized <= 1) niceNormalized = 1;
  else if (normalized <= 2) niceNormalized = 2;
  else if (normalized <= 5) niceNormalized = 5;
  else niceNormalized = 10;
  return niceNormalized * magnitude;
}

interface AxisTooltipPoint {
  seriesName?: string;
  value?: number | string | [number, number] | null;
  axisValueLabel?: string;
  marker?: string;
  name?: string;
  data?: { name?: string; value?: number };
}

interface ChartSeriesBase {
  name: string;
  itemStyle?: { color: string };
}

/** Data-label config shared by every series type below — `show: false` is
 * always a valid value (renders nothing extra, current/legacy behavior);
 * `formatter` is only ever set alongside `show: true`. */
interface SeriesLabel {
  show: boolean;
  position?: "top" | "right";
  formatter?: (params: { name?: string; value?: number | number[] }) => string;
}

interface BarLineSeries extends ChartSeriesBase {
  type: "bar" | "line";
  data: (number | null)[];
  smooth?: boolean;
  /** Same string across every measure's series stacks them (stacked_bar). */
  stack?: string;
  /** `{}` turns a "line" series into an area chart (area). */
  areaStyle?: Record<string, never>;
  lineStyle?: { width: number };
  symbol?: "circle";
  label?: SeriesLabel;
}

interface CustomTrendRenderAPI {
  coord: (values: (string | number)[]) => number[];
  barLayout: (options: { count: number; barGap: string }) => { offsetCenter: number }[] | undefined;
}

interface CustomTrendSeries extends ChartSeriesBase {
  type: "custom";
  data: (number | null)[];
  silent: boolean;
  z: number;
  tooltip: { show: boolean };
  renderItem: (
    params: { dataIndex: number },
    api: CustomTrendRenderAPI,
  ) => Record<string, unknown> | undefined;
}

interface PieSeries extends ChartSeriesBase {
  type: "pie";
  radius: string | [string, string];
  data: { name: string; value: number }[];
  label: { show?: boolean; formatter?: (params: { name: string; percent?: number }) => string };
}

interface RadarSeriesItem {
  value: number[];
  name: string;
  itemStyle?: { color: string };
  /** Set per data item (per measure), not at the series level, so each
   * measure's label can use ITS OWN format (percent vs count vs days). */
  label?: SeriesLabel;
}

interface RadarSeries {
  type: "radar";
  data: RadarSeriesItem[];
}

interface ScatterSeries extends ChartSeriesBase {
  type: "scatter";
  data: [number, number][];
  symbolSize?: number;
  label?: SeriesLabel;
}

interface TreemapSeries {
  type: "treemap";
  name: string;
  data: { name: string; value: number }[];
  label?: { show?: boolean; formatter?: (params: { name: string; value?: number }) => string };
}

interface FunnelSeries {
  type: "funnel";
  name: string;
  sort: "descending";
  data: { name: string; value: number }[];
  label?: { show: boolean; formatter?: (params: { name: string; value?: number }) => string };
}

interface GaugeSeries {
  type: "gauge";
  min: number;
  max: number;
  progress: { show: boolean };
  detail: { formatter: (value: number) => string };
  data: { value: number; name: string }[];
}

interface HeatmapSeries {
  type: "heatmap";
  name: string;
  data: [number, number, number][];
  label?: { show: boolean; formatter?: (params: { value?: [number, number, number] }) => string };
}

interface SankeySeries {
  type: "sankey";
  data: { name: string }[];
  links: { source: string; target: string; value: number }[];
}

type ChartSeries = BarLineSeries | CustomTrendSeries | PieSeries | RadarSeries | ScatterSeries | TreemapSeries | FunnelSeries | GaugeSeries | HeatmapSeries | SankeySeries;

interface CategoryAxis {
  type: "category";
  data: string[];
  name?: string;
  axisLabel?: { interval: number; rotate: number; hideOverlap: boolean };
}

/**
 * ECharts places every ordinary line point at the category tick center, but
 * clustered bars are shifted left/right around that center. A custom series
 * can ask ECharts for the exact bar layout (`api.barLayout`) and draw both
 * its curve segment and symbol at the corresponding bar's center. This keeps
 * the overlay responsive without guessing a fixed pixel offset.
 */
function buildAlignedClusterTrendSeries({
  name,
  color,
  values,
  categories,
  measureIndex,
  measureCount,
  horizontal,
}: {
  name: string;
  color: string;
  values: (number | null)[];
  categories: string[];
  measureIndex: number;
  measureCount: number;
  horizontal: boolean;
}): CustomTrendSeries {
  const pointAt = (api: CustomTrendRenderAPI, categoryIndex: number, value: number, offset: number) => {
    const point = horizontal
      ? api.coord([value, categories[categoryIndex]])
      : api.coord([categories[categoryIndex], value]);
    if (horizontal) point[1] += offset;
    else point[0] += offset;
    return point;
  };

  return {
    type: "custom",
    name,
    data: values,
    itemStyle: { color },
    silent: true,
    z: 4,
    tooltip: { show: false },
    renderItem: (params, api) => {
      const value = values[params.dataIndex];
      if (value === null || value === undefined) return undefined;
      // BaseBarSeries' default gap is 10%; using the same value here makes
      // this calculation identical to ECharts' own clustered-bar layout.
      const barLayouts = api.barLayout({ count: measureCount, barGap: "10%" });
      const offset = barLayouts?.[measureIndex]?.offsetCenter;
      if (offset === null || offset === undefined) return undefined;

      const [x, y] = pointAt(api, params.dataIndex, value, offset);
      const children: Record<string, unknown>[] = [];
      const previousValue = values[params.dataIndex - 1];
      if (params.dataIndex > 0 && previousValue !== null && previousValue !== undefined) {
        const [previousX, previousY] = pointAt(api, params.dataIndex - 1, previousValue, offset);
        if (horizontal) {
          const middleY = (previousY + y) / 2;
          children.push({
            type: "bezierCurve",
            shape: { x1: previousX, y1: previousY, x2: x, y2: y, cpx1: previousX, cpy1: middleY, cpx2: x, cpy2: middleY },
            style: { stroke: color, lineWidth: 2, fill: null },
          });
        } else {
          const middleX = (previousX + x) / 2;
          children.push({
            type: "bezierCurve",
            shape: { x1: previousX, y1: previousY, x2: x, y2: y, cpx1: middleX, cpy1: previousY, cpx2: middleX, cpy2: y },
            style: { stroke: color, lineWidth: 2, fill: null },
          });
        }
      }
      children.push({
        type: "circle",
        shape: { cx: x, cy: y, r: 4 },
        style: { fill: color, stroke: "#ffffff", lineWidth: 1 },
      });
      return { type: "group", children };
    },
  };
}
interface ValueAxis {
  type: "value";
  name?: string;
}
type ChartAxis = CategoryAxis | ValueAxis;

/** Pure rendering toggles a widget carries alongside its data — see
 * dashboard-layout.api.ts's CustomQueryConfig for the persisted shape and
 * the plan's "Poin Desain Kritis: Kompatibilitas Mundur" for why the
 * defaults below differ per chart family. */
export interface ChartDisplayOptions {
  showLabels?: boolean;
  showLabelValues?: boolean;
  showTrendLine?: boolean;
}

const PIE_FAMILY_TYPES: VizType[] = ["pie", "donut", "treemap", "funnel"];

/** Pie/Donut/Treemap/Funnel have ALWAYS shown a name label with no way to
 * turn it off (funnel/treemap hardcoded `show:true`, pie/donut had no
 * `show` field at all) — so a widget saved before this feature (`show_labels`
 * absent) must keep showing that label. Every other chart type has never
 * shown a label, so absent there means "off", matching today's behavior. */
function defaultShowLabels(vizType: VizType): boolean {
  return PIE_FAMILY_TYPES.includes(vizType);
}

export interface ChartOption {
  color: string[];
  tooltip: {
    trigger: "axis" | "item" | "none";
    formatter?: (params: AxisTooltipPoint[] | AxisTooltipPoint) => string;
  };
  legend?: { bottom: number; type: "scroll" };
  grid?: { left: number; right: number; top: number; bottom: number; containLabel: boolean };
  xAxis?: ChartAxis;
  yAxis?: ChartAxis;
  radar?: { indicator: { name: string; max: number }[] };
  visualMap?: { min: number; max: number; calculable: boolean; orient: "horizontal"; left: string; bottom: number };
  dataZoom?: Array<{
    type: "inside" | "slider";
    xAxisIndex?: number;
    yAxisIndex?: number;
    start: number;
    end: number;
    filterMode: "filter";
    showDetail?: boolean;
  }>;
  series: ChartSeries[];
}

function buildPieFamilyOption(
  result: AnalyticsQueryResult,
  vizType: "pie" | "donut" | "treemap" | "funnel",
  options?: ChartDisplayOptions
): ChartOption {
  const measure = result.measures[0];
  const format = measure?.format;
  const data = result.rows.map((row) => ({ name: labelOf(row, "category", "key"), value: measure ? numericValue(row, measure.id) ?? 0 : 0 }));
  const colors = [MEASURE_COLORS[0]];
  const tooltipLabel = (name: string) => name.replaceAll("\n", "<br/>");

  const showLabels = options?.showLabels ?? defaultShowLabels(vizType);
  const labelText = (name: string, value: number) => (options?.showLabelValues ? `${name}: ${formatMeasureValue(value, format)}` : name);

  if (vizType === "treemap") {
    return {
      color: colors,
      tooltip: { trigger: "item", formatter: (p) => `${tooltipLabel(Array.isArray(p) ? "" : (p.name ?? ""))}: ${formatMeasureValue(Array.isArray(p) ? null : p.value as number, format)}` },
      series: [
        {
          type: "treemap",
          name: measure?.label ?? "",
          data,
          label: showLabels ? { show: true, formatter: (p) => labelText(p.name, p.value ?? 0) } : { show: false },
        },
      ],
    };
  }
  if (vizType === "funnel") {
    return {
      color: colors,
      tooltip: { trigger: "item", formatter: (p) => `${tooltipLabel(Array.isArray(p) ? "" : (p.name ?? ""))}: ${formatMeasureValue(Array.isArray(p) ? null : p.value as number, format)}` },
      series: [
        {
          type: "funnel",
          name: measure?.label ?? "",
          sort: "descending",
          data,
          label: showLabels ? { show: true, formatter: (p) => labelText(p.name, p.value ?? 0) } : { show: false },
        },
      ],
    };
  }
  // pie / donut
  return {
    color: colors,
    tooltip: {
      trigger: "item",
      formatter: (params) => {
        const p = Array.isArray(params) ? params[0] : params;
        const value = typeof p?.value === "number" ? p.value : null;
        return `${(p?.name ?? "").replaceAll("\n", "<br/>")}: ${formatMeasureValue(value, format)}`;
      },
    },
    legend: { bottom: 0, type: "scroll" },
    series: [
      {
        type: "pie",
        name: measure?.label ?? "",
        radius: vizType === "donut" ? ["45%", "70%"] : "70%",
        data,
        label: showLabels ? { show: true, formatter: (p) => labelText(p.name, data.find((d) => d.name === p.name)?.value ?? 0) } : { show: false },
      },
    ],
  };
}

function buildComparisonOption(
  result: AnalyticsQueryResult,
  vizType: "bar" | "horizontal_bar" | "stacked_bar" | "line" | "area" | "radar",
  options?: ChartDisplayOptions
): ChartOption {
  const categories = result.rows.map((row) => labelOf(row, "category", "key"));
  const colors = result.measures.map((_, idx) => MEASURE_COLORS[idx % MEASURE_COLORS.length]);
  const showLabels = options?.showLabels ?? false; // no chart type here ever showed labels by default before this feature

  if (vizType === "radar") {
    const maxValue = Math.max(
      1,
      ...result.measures.flatMap((m) => result.rows.map((row) => numericValue(row, m.id) ?? 0))
    );
    const indicator = categories.map((name) => ({ name, max: Math.ceil(maxValue * 1.2) }));
    const data: RadarSeriesItem[] = result.measures.map((m, idx) => ({
      name: m.label,
      itemStyle: { color: colors[idx] },
      value: result.rows.map((row) => numericValue(row, m.id) ?? 0),
      ...(showLabels
        ? { label: { show: true, formatter: (p: { value?: number | number[] }) => formatMeasureValue(typeof p.value === "number" ? p.value : null, m.format) } }
        : {}),
    }));
    return {
      color: colors,
      tooltip: { trigger: "item" },
      legend: { bottom: 0, type: "scroll" },
      radar: { indicator },
      series: [{ type: "radar", data }],
    };
  }

  const seriesType: "bar" | "line" = vizType === "line" || vizType === "area" ? "line" : "bar";
  const labelFor = (measure: (typeof result.measures)[number]): SeriesLabel | undefined =>
    showLabels
      ? {
          show: true,
          position: vizType === "horizontal_bar" ? "right" : "top",
          formatter: (p) => formatMeasureValue(typeof p.value === "number" ? p.value : null, measure.format),
        }
      : undefined;

  const series: ChartSeries[] = result.measures.map((measure, idx) => ({
    type: seriesType,
    name: measure.label,
    smooth: seriesType === "line",
    itemStyle: { color: colors[idx] },
    ...(vizType === "stacked_bar" ? { stack: "total" } : {}),
    ...(vizType === "area" ? { areaStyle: {} } : {}),
    ...(labelFor(measure) ? { label: labelFor(measure) } : {}),
    data: result.rows.map((row) => numericValue(row, measure.id)),
  }));

  // "Line" display option — a trend line that traces the SAME values
  // already drawn as bars (not a different measure rendered as a line).
  // Bar (clustered): one extra line series PER measure, sharing that
  // measure's exact name/color so ECharts merges bar+line into a single
  // legend entry (toggling one hides both, per the approved plan).
  // Stacked bar: one aggregate "Total (Tren)" line tracing the summed
  // stack height per category — a distinct series, so it legitimately
  // gets its own legend entry.
  if (options?.showTrendLine && (vizType === "bar" || vizType === "horizontal_bar")) {
    for (const [idx, measure] of result.measures.entries()) {
      const values = result.rows.map((row) => numericValue(row, measure.id));
      if (result.measures.length === 1) {
        // A single bar is already centered on its category tick, so the
        // native line series is exact and keeps ECharts' built-in animation.
        series.push({
          type: "line",
          name: measure.label,
          smooth: true,
          symbol: "circle",
          lineStyle: { width: 2 },
          itemStyle: { color: colors[idx] },
          data: values,
        });
      } else {
        series.push(
          buildAlignedClusterTrendSeries({
            name: measure.label,
            color: colors[idx],
            values,
            categories,
            measureIndex: idx,
            measureCount: result.measures.length,
            horizontal: vizType === "horizontal_bar",
          }),
        );
      }
    }
  } else if (options?.showTrendLine && vizType === "stacked_bar") {
    series.push({
      type: "line",
      name: "Total (Tren)",
      smooth: true,
      symbol: "circle",
      lineStyle: { width: 2 },
      itemStyle: { color: "#374151" },
      data: result.rows.map((row) => result.measures.reduce((sum, m) => sum + (numericValue(row, m.id) ?? 0), 0)),
    });
  }

  const tooltipFormatter = (params: AxisTooltipPoint[] | AxisTooltipPoint) => {
    const points = Array.isArray(params) ? params : [params];
    // A trend-line series shares its exact name with its bar counterpart
    // (see showTrendLine above — deliberate, so the legend merges them into
    // one entry) but an axis-trigger tooltip collects every SERIES at that
    // category independently, so bar+line would otherwise print the same
    // measure/value twice. Keep only the first point per series name.
    const seen = new Set<string>();
    const uniquePoints = points.filter((p) => {
      const name = p.seriesName ?? "";
      if (seen.has(name)) return false;
      seen.add(name);
      return true;
    });
    const lines = uniquePoints.map((p) => {
      const measure = result.measures.find((m) => m.label === p.seriesName);
      const value = typeof p.value === "number" ? p.value : null;
      return `${p.marker ?? ""}${p.seriesName ?? ""}: ${formatMeasureValue(value, measure?.format)}`;
    });
    return [(points[0]?.axisValueLabel ?? "").replaceAll("\n", "<br/>"), ...lines].join("<br/>");
  };

  const categoryAxis: CategoryAxis = {
    type: "category",
    data: categories,
    axisLabel: { interval: 0, rotate: result.hierarchy?.length ? 0 : categories.length > 6 ? 30 : 0, hideOverlap: true },
  };
  const valueAxis: ValueAxis = { type: "value" };

  const zoomEnd = categories.length > 12 ? Math.max(5, (12 / categories.length) * 100) : 100;
  const dataZoom: ChartOption["dataZoom"] = categories.length > 12
    ? vizType === "horizontal_bar"
      ? [
          { type: "inside", yAxisIndex: 0, start: 0, end: zoomEnd, filterMode: "filter" },
          { type: "slider", yAxisIndex: 0, start: 0, end: zoomEnd, filterMode: "filter", showDetail: false },
        ]
      : [
          { type: "inside", xAxisIndex: 0, start: 0, end: zoomEnd, filterMode: "filter" },
          { type: "slider", xAxisIndex: 0, start: 0, end: zoomEnd, filterMode: "filter", showDetail: false },
        ]
    : undefined;

  return {
    color: colors,
    tooltip: { trigger: "axis", formatter: tooltipFormatter },
    legend: series.length > 1 ? { bottom: 0, type: "scroll" } : undefined,
    grid: {
      left: 48,
      right: vizType === "horizontal_bar" && dataZoom ? 40 : 16,
      top: 24,
      bottom: vizType !== "horizontal_bar" && dataZoom ? 72 : series.length > 1 ? 48 : 32,
      containLabel: true,
    },
    xAxis: vizType === "horizontal_bar" ? valueAxis : categoryAxis,
    yAxis: vizType === "horizontal_bar" ? categoryAxis : valueAxis,
    dataZoom,
    series,
  };
}

function buildScatterOption(result: AnalyticsQueryResult, options?: ChartDisplayOptions): ChartOption {
  const [measureX, measureY] = result.measures;
  if (!measureX || !measureY) {
    return { color: [], tooltip: { trigger: "none" }, series: [] };
  }
  const data: [number, number][] = result.rows.map((row) => [numericValue(row, measureX.id) ?? 0, numericValue(row, measureY.id) ?? 0]);
  const label: SeriesLabel | undefined = options?.showLabels
    ? {
        show: true,
        formatter: (p) => (Array.isArray(p.value) ? `${formatMeasureValue(p.value[0], measureX.format)}, ${formatMeasureValue(p.value[1], measureY.format)}` : ""),
      }
    : undefined;
  return {
    color: [MEASURE_COLORS[0]],
    tooltip: {
      trigger: "item",
      formatter: (params) => {
        const p = Array.isArray(params) ? params[0] : params;
        const value = Array.isArray(p?.value) ? p.value : null;
        if (!value) return "";
        return `${measureX.label}: ${formatMeasureValue(value[0], measureX.format)}<br/>${measureY.label}: ${formatMeasureValue(value[1], measureY.format)}`;
      },
    },
    grid: { left: 56, right: 24, top: 24, bottom: 48, containLabel: true },
    xAxis: { type: "value", name: measureX.label },
    yAxis: { type: "value", name: measureY.label },
    series: [{ type: "scatter", name: `${measureX.label} × ${measureY.label}`, data, symbolSize: 12, itemStyle: { color: MEASURE_COLORS[0] }, ...(label ? { label } : {}) }],
  };
}

function buildGaugeOption(result: AnalyticsQueryResult): ChartOption {
  const aggregate = aggregateSingleMeasure(result);
  if (!aggregate) {
    return { color: [], tooltip: { trigger: "none" }, series: [] };
  }
  const { value, measure } = aggregate;
  const max = measure.value_type === "percent" ? 100 : computeNiceMax(value);
  return {
    color: [MEASURE_COLORS[0]],
    tooltip: { trigger: "item" },
    series: [
      {
        type: "gauge",
        min: 0,
        max,
        progress: { show: true },
        detail: { formatter: (v: number) => formatMeasureValue(v, measure.format) },
        data: [{ value, name: measure.label }],
      },
    ],
  };
}

function buildHeatmapOption(result: AnalyticsQueryResult, options?: ChartDisplayOptions): ChartOption {
  const measure = result.measures[0];
  if (!measure) return { color: [], tooltip: { trigger: "none" }, series: [] };

  const xCategories: string[] = [];
  const yCategories: string[] = [];
  const xIndex = new Map<string, number>();
  const yIndex = new Map<string, number>();
  for (const row of result.rows) {
    const x = labelOf(row, "category", "key");
    const y = labelOf(row, "category2", "key2");
    if (!xIndex.has(x)) {
      xIndex.set(x, xCategories.length);
      xCategories.push(x);
    }
    if (!yIndex.has(y)) {
      yIndex.set(y, yCategories.length);
      yCategories.push(y);
    }
  }
  const data: [number, number, number][] = result.rows.map((row) => {
    const x = xIndex.get(labelOf(row, "category", "key")) ?? 0;
    const y = yIndex.get(labelOf(row, "category2", "key2")) ?? 0;
    return [x, y, numericValue(row, measure.id) ?? 0];
  });
  const values = data.map((d) => d[2]);
  const min = values.length ? Math.min(...values) : 0;
  const max = values.length ? Math.max(...values) : 1;

  return {
    color: [],
    tooltip: {
      trigger: "item",
      formatter: (params) => {
        const p = Array.isArray(params) ? params[0] : params;
        const cell = Array.isArray(p?.value) ? p.value : null;
        if (!cell) return "";
        return `${xCategories[cell[0]] ?? ""} / ${yCategories[cell[1]] ?? ""}: ${formatMeasureValue(cell[1] !== undefined ? data.find((d) => d[0] === cell[0] && d[1] === cell[1])?.[2] ?? null : null, measure.format)}`;
      },
    },
    grid: { left: 80, right: 24, top: 24, bottom: 64, containLabel: true },
    xAxis: { type: "category", data: xCategories, axisLabel: { interval: 0, rotate: xCategories.length > 6 ? 30 : 0, hideOverlap: true } },
    yAxis: { type: "category", data: yCategories },
    visualMap: { min, max, calculable: true, orient: "horizontal", left: "center", bottom: 0 },
    series: [
      {
        type: "heatmap",
        name: measure.label,
        data,
        label: options?.showLabels
          ? { show: true, formatter: (p) => formatMeasureValue(p.value ? p.value[2] : null, measure.format) }
          : { show: false },
      },
    ],
  };
}

function buildSankeyOption(result: AnalyticsQueryResult): ChartOption {
  const measure = result.measures[0];
  if (!measure) return { color: [], tooltip: { trigger: "none" }, series: [] };

  const nodeNames = new Set<string>();
  const links: { source: string; target: string; value: number }[] = [];
  for (const row of result.rows) {
    const source = labelOf(row, "category", "key");
    const target = labelOf(row, "category2", "key2");
    nodeNames.add(source);
    nodeNames.add(target);
    const value = numericValue(row, measure.id);
    if (value !== null && value > 0) {
      links.push({ source, target, value });
    }
  }

  return {
    color: [MEASURE_COLORS[0]],
    tooltip: {
      trigger: "item",
      formatter: (params) => {
        const p = Array.isArray(params) ? params[0] : params;
        if (p?.data && "name" in p.data === false) return "";
        return `${p?.name ?? ""}`;
      },
    },
    series: [{ type: "sankey", data: [...nodeNames].map((name) => ({ name })), links }],
  };
}

/** Builds a fully-typed ECharts option from one analytics query result. Pure
 * function: no DOM/React access, safe to unit test and to memoize on
 * (result, vizType) alone. Never called for "table"/"number_card" — those
 * bypass ECharts entirely (see DynamicKPIWidget). */
export function buildEChartsOption(result: AnalyticsQueryResult, vizType: VizType, options?: ChartDisplayOptions): ChartOption {
  switch (vizType) {
    case "pie":
    case "donut":
    case "treemap":
    case "funnel":
      return buildPieFamilyOption(result, vizType, options);
    case "radar":
    case "horizontal_bar":
    case "stacked_bar":
    case "line":
    case "area":
      return buildComparisonOption(result, vizType, options);
    case "scatter":
      return buildScatterOption(result, options);
    case "gauge":
      return buildGaugeOption(result);
    case "heatmap":
      return buildHeatmapOption(result, options);
    case "sankey":
      return buildSankeyOption(result);
    case "bar":
    default:
      return buildComparisonOption(result, "bar", options);
  }
}
