import type { AnalyticsChartType, CatalogField } from "@/lib/api/analytics.api";

export const ALL_CHART_TYPES: AnalyticsChartType[] = [
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

export const PIE_FAMILY: AnalyticsChartType[] = ["pie", "donut", "treemap", "funnel"];
export const MATRIX_FAMILY: AnalyticsChartType[] = ["heatmap", "sankey"];
export const DRILLABLE_CHARTS: AnalyticsChartType[] = [
  "bar",
  "horizontal_bar",
  "stacked_bar",
  "line",
  "area",
  "pie",
  "donut",
  "treemap",
  "funnel",
  // Table supports multiple Kategori too — it just never appeared in this
  // list, which silently forced a fallback to Bar the moment a 2nd Kategori
  // was added. Unlike every other chart here, Table doesn't step through
  // levels one at a time via a drill-down toolbar: it always queries every
  // configured level at once and renders them as stacked breadcrumb columns
  // (DynamicTable's result.hierarchy/row.path) — so it's deliberately
  // ABSENT from DynamicKPIWidget.tsx's DRILLABLE_VIZ_TYPES, which is a
  // different list gating the progressive toolbar interaction only.
  "table",
];
export const NO_LABEL_TOGGLE_CHARTS: AnalyticsChartType[] = ["table", "number_card", "gauge", "sankey"];
export const TREND_LINE_CHARTS: AnalyticsChartType[] = ["bar", "horizontal_bar", "stacked_bar"];
export const MAX_MEASURES = 4;
// Table can render arbitrarily many measure columns (DynamicTable maps
// over result.measures with no fixed-size assumption, wrapped in
// overflow-auto) — every other chart type stays capped at MAX_MEASURES,
// where >4 series genuinely becomes unreadable. 30 is a generous ceiling
// relative to the ~17 measures in today's catalog, not a real business
// limit — see maxMeasuresFor.
export const MAX_MEASURES_TABLE = 30;
export const MAX_CATEGORIES = 4;

export function isMatrixChart(vizType: AnalyticsChartType | null): boolean {
  return !!vizType && MATRIX_FAMILY.includes(vizType);
}

export function maxMeasuresFor(vizType: AnalyticsChartType | null): number {
  return vizType === "table" ? MAX_MEASURES_TABLE : MAX_MEASURES;
}

export function labelToggleDisabledReason(vizType: AnalyticsChartType | null): string | undefined {
  if (!vizType) return "Pilih jenis chart terlebih dahulu";
  if (NO_LABEL_TOGGLE_CHARTS.includes(vizType)) return "Nilai selalu ditampilkan untuk jenis chart ini";
  return undefined;
}

/**
 * Resolves compatible charts from one ordered category list. The first item
 * is always the primary dimension. With regular charts, every later item is
 * a drill level; with Heatmap/Sankey, the second item becomes dimension2.
 */
export function compatibleCharts(measures: CatalogField[], primaryCategory: CatalogField | null, categoryCount: number): AnalyticsChartType[] {
  if (measures.length === 0 || !primaryCategory || categoryCount === 0) return [];

  let charts: AnalyticsChartType[] = measures[0].compatible_charts ? [...measures[0].compatible_charts] : [...ALL_CHART_TYPES];
  for (const measure of measures.slice(1)) {
    const allowed = new Set(measure.compatible_charts ?? ALL_CHART_TYPES);
    charts = charts.filter((chart) => allowed.has(chart));
  }

  if (measures.length === 2 && !charts.includes("scatter")) charts.push("scatter");
  if (measures.length > 1 || primaryCategory.kind !== "categorical") {
    charts = charts.filter((chart) => !PIE_FAMILY.includes(chart));
  }
  if (measures.length !== 2) charts = charts.filter((chart) => chart !== "scatter");
  if (measures.length !== 1) charts = charts.filter((chart) => chart !== "number_card" && chart !== "gauge");

  // Matrix charts have their own two-category query shape, so remove them
  // from the normal compatibility intersection and add them only when the
  // ordered category list has exactly two entries and exactly one measure.
  charts = charts.filter((chart) => !MATRIX_FAMILY.includes(chart));
  if (categoryCount > 1) {
    const drillable = new Set(DRILLABLE_CHARTS);
    charts = charts.filter((chart) => drillable.has(chart));
  }
  if (categoryCount === 2 && measures.length === 1) charts.push(...MATRIX_FAMILY);

  // Past the standard cap, only Table can actually render that many
  // series — every other chart type becomes incompatible until measures
  // are trimmed back down (see MAX_MEASURES/maxMeasuresFor).
  if (measures.length > MAX_MEASURES) {
    charts = charts.filter((chart) => chart === "table");
  }

  return charts;
}

export function incompatibleReason(
  type: AnalyticsChartType,
  measureCount: number,
  primaryCategory: CatalogField | null,
  categoryCount: number,
): string {
  if (!primaryCategory || categoryCount === 0) return "Tambahkan minimal satu Kategori terlebih dahulu";
  if (type !== "table" && measureCount > MAX_MEASURES) {
    return `Lebih dari ${MAX_MEASURES} Nilai hanya didukung oleh Tabel`;
  }
  if (MATRIX_FAMILY.includes(type)) {
    if (categoryCount !== 2) return "Heatmap/Sankey membutuhkan tepat 2 Kategori";
    if (measureCount !== 1) return "Heatmap/Sankey hanya mendukung tepat 1 Nilai";
  }
  if (categoryCount > 1 && !DRILLABLE_CHARTS.includes(type) && !MATRIX_FAMILY.includes(type)) {
    return "Jenis chart ini tidak mendukung kategori bertingkat";
  }
  if (type === "scatter") return "Scatter Plot membutuhkan tepat 2 Nilai dan tidak mendukung drill-down";
  if (type === "number_card" || type === "gauge") {
    return `${type === "gauge" ? "Gauge" : "Kartu Angka"} membutuhkan tepat 1 Nilai dan 1 Kategori`;
  }
  if (PIE_FAMILY.includes(type)) {
    if (measureCount > 1) return "Pie/Donut/Treemap/Funnel hanya mendukung tepat 1 Nilai";
    if (primaryCategory.kind !== "categorical") return "Jenis chart ini membutuhkan Kategori kategorikal, bukan tanggal/waktu";
  }
  return "Tidak cocok untuk kombinasi Kategori dan Nilai ini";
}

export interface ResolvedCategories {
  dimensionId: string | null;
  dimension2Id: string | null;
  drillIds: string[];
}

export function resolveCategories(categoryIds: string[], vizType: AnalyticsChartType | null): ResolvedCategories {
  const [dimensionId = null, ...rest] = categoryIds;
  if (isMatrixChart(vizType)) {
    return { dimensionId, dimension2Id: rest[0] ?? null, drillIds: [] };
  }
  return { dimensionId, dimension2Id: null, drillIds: rest };
}

export function categoriesFromSavedConfig(dimension?: string, dimension2?: string, drillDimensions?: string[]): string[] {
  if (!dimension) return [];
  if (dimension2) return [dimension, dimension2];
  return [dimension, ...(drillDimensions ?? [])];
}
