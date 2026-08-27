import type { WidgetDefinition } from "@/components/dashboard/WidgetGrid";
import { AuditorStatsWidget } from "./AuditorStatsWidget";
import { AuditorPendingBannerWidget } from "./AuditorPendingBannerWidget";
import { AuditorTrendChartWidget } from "./AuditorTrendChartWidget";

// Keep widget_id values in sync with backend/internal/domain/dashboard/dashboard_layout.go.
export const auditorWidgetRegistry: WidgetDefinition[] = [
  { id: "auditor-stats", title: "Kartu Statistik", Component: AuditorStatsWidget },
  { id: "auditor-pending-banner", title: "Temuan Menunggu Validasi", Component: AuditorPendingBannerWidget },
  { id: "auditor-trend-chart", title: "Tren Inspeksi Bulanan", Component: AuditorTrendChartWidget },
];
