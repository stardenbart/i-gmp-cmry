import type { WidgetDefinition } from "@/components/dashboard/WidgetGrid";
import { StatsCardsWidget } from "./StatsCardsWidget";
import { AktivitasInspeksiWidget } from "./AktivitasInspeksiWidget";
import { AktivitasPICWidget } from "./AktivitasPICWidget";
import { StatusWOWRWidget } from "./StatusWOWRWidget";
import { StatusAuditeeWidget } from "./StatusAuditeeWidget";
import { ChartTrenKepatuhanWidget } from "./ChartTrenKepatuhanWidget";

// Keep widget_id values in sync with backend/internal/domain/dashboard/dashboard_layout.go's
// DefaultLayoutForRole — these are the source of truth for what each id renders.
export const adminWidgetRegistry: WidgetDefinition[] = [
  { id: "stats-cards", title: "Kartu Statistik", Component: StatsCardsWidget, className: "lg:col-span-3" },
  { id: "aktivitas-inspeksi", title: "Aktivitas Inspeksi", Component: AktivitasInspeksiWidget, className: "lg:col-span-1" },
  { id: "aktivitas-pic", title: "Aktivitas Penanggung Jawab (PIC)", Component: AktivitasPICWidget, className: "lg:col-span-1" },
  { id: "status-auditee", title: "Status Lokasi Auditee", Component: StatusAuditeeWidget, className: "lg:col-span-1" },
  { id: "status-wowr", title: "Status Maintenance (WO/WR)", Component: StatusWOWRWidget, className: "lg:col-span-3" },
  { id: "chart-tren-kepatuhan", title: "Tren Tingkat Kepatuhan", Component: ChartTrenKepatuhanWidget, className: "lg:col-span-3" },
];
