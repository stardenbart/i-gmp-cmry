import type { WidgetDefinition } from "@/components/dashboard/types";
import { AuditorStatsWidget } from "./AuditorStatsWidget";
import { AuditorPendingBannerWidget } from "./AuditorPendingBannerWidget";
import { AuditorTrendChartWidget } from "./AuditorTrendChartWidget";

// Keep widget_id values in sync with backend/internal/domain/dashboard/dashboard_layout.go.
export const auditorWidgetRegistry: WidgetDefinition[] = [
  { id: "auditor-stats", title: "Kartu Statistik", Component: AuditorStatsWidget, defaultLayout: { x: 0, y: 0, w: 12, h: 3 } },
  { id: "auditor-pending-banner", title: "Temuan Menunggu Validasi", Component: AuditorPendingBannerWidget, defaultLayout: { x: 0, y: 3, w: 12, h: 5 } },
  { id: "auditor-trend-chart", title: "Tren Inspeksi Bulanan", Component: AuditorTrendChartWidget, defaultLayout: { x: 0, y: 8, w: 12, h: 12 } },
];
