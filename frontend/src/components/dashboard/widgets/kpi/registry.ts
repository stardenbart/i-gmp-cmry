import type { WidgetDefinition } from "@/components/dashboard/types";
import { KPISummaryCardsWidget } from "./KPISummaryCardsWidget";
import { KPITrendChartWidget } from "./KPITrendChartWidget";
import { KPIPICRankingWidget } from "./KPIPICRankingWidget";
import { KPIAuditorRankingWidget } from "./KPIAuditorRankingWidget";
import { KPIWOWRByAreaWidget } from "./KPIWOWRByAreaWidget";
import { KPIAreaProgressWidget } from "./KPIAreaProgressWidget";

// Keep widget_id values in sync with backend/internal/domain/dashboard/dashboard_layout.go's
// DefaultLayoutKPI. This registry is isolated from the 3 role dashboards'
// registries by DashboardKey="kpi" (see DashboardGrid's dashboardKey prop),
// so its ids don't need to be unique against theirs.
//
// supportedVizTypes/defaultVizType drive the Power BI-style "change
// visualization" picker in each widget's edit-mode header (see
// DashboardWidgetFrame) — "list" is always each widget's own existing
// bespoke markup, kept as the default so nothing changes visually until an
// Admin explicitly picks something else.
export const kpiWidgetRegistry: WidgetDefinition[] = [
  { id: "kpi-summary-cards", title: "Ringkasan KPI", Component: KPISummaryCardsWidget, defaultLayout: { x: 0, y: 0, w: 12, h: 3 } },
  {
    id: "kpi-trend-chart",
    title: "Tren Tingkat Kepatuhan",
    Component: KPITrendChartWidget,
    defaultLayout: { x: 0, y: 3, w: 12, h: 13 },
    supportedVizTypes: ["area", "line", "bar", "table"],
    defaultVizType: "area",
  },
  {
    id: "kpi-pic-ranking",
    title: "Kinerja PIC",
    Component: KPIPICRankingWidget,
    defaultLayout: { x: 0, y: 16, w: 4, h: 8 },
    supportedVizTypes: ["list", "table", "bar"],
    defaultVizType: "list",
  },
  {
    id: "kpi-auditor-ranking",
    title: "Kinerja Auditor",
    Component: KPIAuditorRankingWidget,
    defaultLayout: { x: 4, y: 16, w: 4, h: 8 },
    supportedVizTypes: ["list", "table", "bar"],
    defaultVizType: "list",
  },
  {
    id: "kpi-wowr-by-area",
    title: "Status WO/WR per Area",
    Component: KPIWOWRByAreaWidget,
    defaultLayout: { x: 8, y: 16, w: 4, h: 8 },
    supportedVizTypes: ["list", "table", "bar", "pie"],
    defaultVizType: "list",
  },
  {
    id: "kpi-area-progress",
    title: "Progress Kepatuhan per Area/Kawasan",
    Component: KPIAreaProgressWidget,
    defaultLayout: { x: 0, y: 24, w: 12, h: 8 },
    supportedVizTypes: ["list", "table", "bar"],
    defaultVizType: "list",
  },
];
