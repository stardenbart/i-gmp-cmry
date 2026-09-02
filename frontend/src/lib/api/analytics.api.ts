import { api } from "./axios";
import type { SingleItemResponse } from "@/types/api/types";

// Mirrors backend/internal/usecase/analyticsusecase.CatalogFieldDTO. This
// type NEVER carries SQL (SelectExpr/KeyExpr/LabelExpr/JoinSQL) — the
// backend catalog endpoint deliberately never sends it, so there is nothing
// to type here even if we wanted to.
export type AnalyticsValueType = "count" | "percent" | "days" | "float";
export type AnalyticsFieldKind = "categorical" | "temporal";
export type AnalyticsChartType =
  | "bar"
  | "horizontal_bar"
  | "stacked_bar"
  | "line"
  | "area"
  | "radar"
  | "scatter"
  | "pie"
  | "donut"
  | "treemap"
  | "funnel"
  | "number_card"
  | "gauge"
  | "heatmap"
  | "sankey"
  | "table";

export interface AnalyticsFormat {
  style: "number" | "percent" | "duration" | "currency";
  decimals: number;
  suffix?: string;
  unit?: string;
}

export interface CatalogField {
  id: string;
  label: string;
  description: string;
  value_type?: AnalyticsValueType;
  format?: AnalyticsFormat;
  kind?: AnalyticsFieldKind;
  compatible_charts?: AnalyticsChartType[];
  fans_out?: boolean;
}

export interface AnalyticsCatalog {
  dimensions: CatalogField[];
  measures: CatalogField[];
}

export interface AnalyticsQueryParams {
  measures: string[];
  dimension: string;
  /** Second grouping dimension — only used by Heatmap/Sankey (a dim1 x
   * dim2 x 1 measure matrix). Omit for every other chart type. */
  dimension2?: string;
  /** Ad-hoc "dimension = value" equality filters — the primitive behind
   * drill-down (see DynamicKPIWidget): each accumulated click pins one
   * ancestor level's exact key value here while `dimension` moves on to the
   * next hierarchy level. Independent of dimension2/matrix mode. */
  filters?: { dimensionId: string; value: string }[];
  areaId?: string;
  /** SuperAdmin's plant filter — sent as a URL query param (`?plant_id=`),
   * mirroring every other /dashboard/* endpoint, since the backend reads
   * it via c.Query() regardless of this being a POST. */
  plantId?: string;
  startDate?: string;
  endDate?: string;
  /** Forwarded to axios so a superseded request (config changed while the
   * previous one was in flight) actually cancels the underlying HTTP/DB
   * work, not just gets its response discarded. */
  signal?: AbortSignal;
}

export type AnalyticsRowValue = string | number | null;

export interface AnalyticsQueryResult {
  dimension: CatalogField;
  /** Present only when a dimension2 was requested (Heatmap/Sankey) — rows
   * then carry both "category"/"key" (dim1) and "category2"/"key2" (dim2). */
  dimension2?: CatalogField;
  measures: CatalogField[];
  rows: Record<string, AnalyticsRowValue>[];
}

export const analyticsApi = {
  getCatalog: async (): Promise<AnalyticsCatalog> => {
    const res = await api.get<SingleItemResponse<AnalyticsCatalog>>("/analytics/catalog");
    return res.data.data;
  },

  runQuery: async (params: AnalyticsQueryParams): Promise<AnalyticsQueryResult> => {
    const res = await api.post<SingleItemResponse<AnalyticsQueryResult>>(
      "/analytics/query",
      {
        measures: params.measures,
        dimension: params.dimension,
        dimension2: params.dimension2,
        filters: params.filters?.map((f) => ({ dimension_id: f.dimensionId, value: f.value })),
        area_id: params.areaId,
        start_date: params.startDate,
        end_date: params.endDate,
      },
      { params: { plant_id: params.plantId }, signal: params.signal }
    );
    return res.data.data;
  },
};
