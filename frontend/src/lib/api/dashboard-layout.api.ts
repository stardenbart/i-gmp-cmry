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
  /** Ordered drill-down levels beyond `dimension` (level 0) — level 1, 2...
   * Mutually exclusive with dimension2. Absent for every widget without
   * drill-down. */
  drillDimensions?: string[];
  /** Pure rendering options — never sent to /analytics/query. See
   * buildEChartsOption.ts's ChartDisplayOptions for how each is used. */
  showLabels?: boolean;
  showLabelValues?: boolean;
  showTrendLine?: boolean;
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
    const res = await api.get<SingleItemResponse<WidgetConfig[]>>("/dashboard/layout", {
      params: { user_id: target?.userId, role_id: target?.roleId, dashboard_key: target?.dashboardKey },
    });
    return res.data.data;
  },

  save: async (widgets: WidgetConfig[], target?: LayoutTarget): Promise<WidgetConfig[]> => {
    const res = await api.put<SingleItemResponse<WidgetConfig[]>>("/dashboard/layout", widgets, {
      params: { user_id: target?.userId, role_id: target?.roleId, dashboard_key: target?.dashboardKey },
    });
    return res.data.data;
  },
};
