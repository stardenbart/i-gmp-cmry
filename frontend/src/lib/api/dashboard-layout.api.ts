import { api } from "./axios";
import type { SingleItemResponse } from "@/types/api/types";

export interface WidgetConfig {
  widget_id: string;
  visible: boolean;
  order: number;
  // Reserved for Fase 3b (react-grid-layout drag/resize) — absent today.
  x?: number;
  y?: number;
  w?: number;
  h?: number;
  /** Which visualization the widget is rendered as (e.g. "table", "bar") —
   * see components/dashboard/types.ts VizType. Absent = widget's own
   * default (usually its bespoke "list" markup). */
  viz_type?: string;
  /** Present only for user-built widgets from the Custom KPI Visualization
   * Builder (dashboard_key="kpi"). Carries only catalog field IDs — never
   * SQL/formulas — re-validated server-side both here (SaveLayout) and at
   * query time (POST /analytics/query). See backend
   * dashboarddomain.CustomQueryConfig. */
  custom_query?: CustomQueryConfig;
}

export interface CustomQueryConfig {
  version: number;
  /** User-defined label explaining what this visualization represents. */
  title?: string;
  measures: string[];
  dimension: string;
  /** Only present for Heatmap/Sankey widgets (a dim1 x dim2 x 1 measure
   * matrix). Absent for every other chart type. */
  dimension2?: string;
  /** Ordered hierarchy levels beyond `dimension` (level 0). Moving down
   * displays every value at level 1, 2, and so on. Mutually exclusive with
   * dimension2. */
  drillDimensions?: string[];
  /** Pure rendering options — never sent to /analytics/query. See
   * buildEChartsOption.ts's ChartDisplayOptions for how each is used. */
  showLabels?: boolean;
  showLabelValues?: boolean;
  showTrendLine?: boolean;
}

// ── Wire format ─────────────────────────────────────────────────────────
// The backend's CustomQueryConfig (backend/internal/domain/dashboard/
// dashboard_layout.go) uses snake_case JSON tags (`show_labels`,
// `drill_dimensions`, ...) — ordinary Go/REST convention. This file's own
// CustomQueryConfig above deliberately stays camelCase to match every other
// TypeScript type in the app (VisualizationBuilder/DashboardKPI/
// DynamicKPIWidget/DashboardGrid all read `cq.showLabels`, `cq.drillDimensions`,
// etc.) — nothing outside this file should ever see snake_case.
//
// `dimension`/`dimension2`/`measures`/`title`/`version` happen to be spelled
// identically in both conventions (no multi-word ambiguity), so a plain
// `JSON.stringify`/response pass-through silently "worked" for those. It
// silently did NOT for the 4 multi-word fields below: axios has no camelCase
// <-> snake_case translation of its own, so without this mapping a save
// request went out with keys like `showLabels`/`drillDimensions` that don't
// match ANY backend struct tag — Go's JSON decoder drops unmatched keys
// without error, so the request "succeeded" while silently discarding
// exactly these 4 fields, and a load right back afterward saw the backend's
// real `show_labels`/`drill_dimensions` keys land on this same camelCase
// type unrecognized too — always `undefined`. That combination is what made
// "I checked the boxes and saved" behave as if nothing had happened.
interface WireCustomQueryConfig {
  version: number;
  title?: string;
  measures: string[];
  dimension: string;
  dimension2?: string;
  drill_dimensions?: string[];
  show_labels?: boolean;
  show_label_values?: boolean;
  show_trend_line?: boolean;
}

type WireWidgetConfig = Omit<WidgetConfig, "custom_query"> & { custom_query?: WireCustomQueryConfig };

function toWireCustomQuery(cq: CustomQueryConfig): WireCustomQueryConfig {
  return {
    version: cq.version,
    title: cq.title,
    measures: cq.measures,
    dimension: cq.dimension,
    dimension2: cq.dimension2,
    drill_dimensions: cq.drillDimensions,
    show_labels: cq.showLabels,
    show_label_values: cq.showLabelValues,
    show_trend_line: cq.showTrendLine,
  };
}

function fromWireCustomQuery(cq: WireCustomQueryConfig): CustomQueryConfig {
  return {
    version: cq.version,
    title: cq.title,
    measures: cq.measures,
    dimension: cq.dimension,
    dimension2: cq.dimension2,
    drillDimensions: cq.drill_dimensions,
    showLabels: cq.show_labels,
    showLabelValues: cq.show_label_values,
    showTrendLine: cq.show_trend_line,
  };
}

function toWireWidget(widget: WidgetConfig): WireWidgetConfig {
  return { ...widget, custom_query: widget.custom_query ? toWireCustomQuery(widget.custom_query) : undefined };
}

function fromWireWidget(widget: WireWidgetConfig): WidgetConfig {
  return { ...widget, custom_query: widget.custom_query ? fromWireCustomQuery(widget.custom_query) : undefined };
}

export function deserializeWidgetConfigs(widgets: WireWidgetConfig[]): WidgetConfig[] {
  return widgets.map(fromWireWidget);
}

export interface LayoutTarget {
  /** Defaults to the caller themself. Targeting another user is only
   * permitted for Admin/Super Admin (enforced server-side, 403 otherwise). */
  userId?: string;
  /** Needed alongside userId so the backend can resolve that user's default
   * layout if they've never customized one — the caller's own role isn't
   * necessarily the target's role. */
  roleId?: string;
  /** Which of the caller's several independently-saved dashboards this is
   * — e.g. "main" (their role dashboard, the default) or "kpi" (the
   * KPI/analytics dashboard). Omit for "main"; see backend
   * dashboarddomain.DashboardKeyMain/DashboardKeyKPI. */
  dashboardKey?: string;
}

export const dashboardLayoutApi = {
  get: async (target?: LayoutTarget): Promise<WidgetConfig[]> => {
    const res = await api.get<SingleItemResponse<WireWidgetConfig[]>>("/dashboard/layout", {
      params: { user_id: target?.userId, role_id: target?.roleId, dashboard_key: target?.dashboardKey },
    });
    return deserializeWidgetConfigs(res.data.data);
  },

  save: async (widgets: WidgetConfig[], target?: LayoutTarget): Promise<WidgetConfig[]> => {
    const res = await api.put<SingleItemResponse<WireWidgetConfig[]>>("/dashboard/layout", widgets.map(toWireWidget), {
      params: { user_id: target?.userId, role_id: target?.roleId, dashboard_key: target?.dashboardKey },
    });
    return deserializeWidgetConfigs(res.data.data);
  },
};
