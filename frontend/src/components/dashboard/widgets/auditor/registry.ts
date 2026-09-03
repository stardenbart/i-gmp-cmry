import type { WidgetDefinition } from "@/components/dashboard/types";
import { withAuditorDashboardContext } from "@/components/dashboard/auditor/AuditorDashboardContext";
import { AuditorStatsWidget } from "./AuditorStatsWidget";
import { AuditorPendingBannerWidget } from "./AuditorPendingBannerWidget";
import { AuditorTrendChartWidget } from "./AuditorTrendChartWidget";

// Keep widget_id values in sync with backend/internal/domain/dashboard/dashboard_layout.go.
//
// Every Component is wrapped in withAuditorDashboardContext so it can also
// render on a non-Auditor dashboard (Admin can mix these into any user's
// layout via Edit User → Tata Letak Dashboard — see widgets/registry.ts's
// allMainDashboardWidgets) without crashing for lack of AuditorDashboardContext.
export const auditorWidgetRegistry: WidgetDefinition[] = [
  { id: "auditor-stats", title: "Kartu Statistik", Component: withAuditorDashboardContext(AuditorStatsWidget), defaultLayout: { x: 0, y: 0, w: 12, h: 3 } },
  { id: "auditor-pending-banner", title: "Temuan Menunggu Validasi", Component: withAuditorDashboardContext(AuditorPendingBannerWidget), defaultLayout: { x: 0, y: 3, w: 12, h: 5 } },
  { id: "auditor-trend-chart", title: "Tren Inspeksi Bulanan", Component: withAuditorDashboardContext(AuditorTrendChartWidget), defaultLayout: { x: 0, y: 8, w: 12, h: 12 } },
];
