import type { WidgetDefinition } from "@/components/dashboard/types";
import { withAdminDashboardContext } from "@/components/dashboard/admin/AdminDashboardContext";
import { StatsCardsWidget } from "./StatsCardsWidget";
import { AktivitasInspeksiWidget } from "./AktivitasInspeksiWidget";
import { AktivitasPICWidget } from "./AktivitasPICWidget";
import { StatusWOWRWidget } from "./StatusWOWRWidget";
import { StatusAuditeeWidget } from "./StatusAuditeeWidget";
import { ChartTrenKepatuhanWidget } from "./ChartTrenKepatuhanWidget";

// Keep widget_id values in sync with backend/internal/domain/dashboard/dashboard_layout.go's
// DefaultLayoutForRole — these are the source of truth for what each id renders.
// defaultLayout only matters the first time a user opens this dashboard,
// before they've customized anything (12-column grid).
//
// Every Component is wrapped in withAdminDashboardContext so it can also
// render on a non-Admin dashboard (Admin can mix these into any user's
// layout via Edit User → Tata Letak Dashboard — see widgets/registry.ts's
// allMainDashboardWidgets) without crashing for lack of AdminDashboardContext.
export const adminWidgetRegistry: WidgetDefinition[] = [
  { id: "stats-cards", title: "Kartu Statistik", Component: withAdminDashboardContext(StatsCardsWidget), defaultLayout: { x: 0, y: 0, w: 12, h: 3 } },
  { id: "aktivitas-inspeksi", title: "Aktivitas Inspeksi", Component: withAdminDashboardContext(AktivitasInspeksiWidget), defaultLayout: { x: 0, y: 3, w: 4, h: 6 } },
  { id: "aktivitas-pic", title: "Aktivitas Penanggung Jawab (PIC)", Component: withAdminDashboardContext(AktivitasPICWidget), defaultLayout: { x: 4, y: 3, w: 4, h: 6 } },
  { id: "status-auditee", title: "Status Lokasi Auditee", Component: withAdminDashboardContext(StatusAuditeeWidget), defaultLayout: { x: 8, y: 3, w: 4, h: 6 } },
  { id: "status-wowr", title: "Status Maintenance (WO/WR)", Component: withAdminDashboardContext(StatusWOWRWidget), defaultLayout: { x: 0, y: 9, w: 12, h: 5 } },
  { id: "chart-tren-kepatuhan", title: "Tren Tingkat Kepatuhan", Component: withAdminDashboardContext(ChartTrenKepatuhanWidget), defaultLayout: { x: 0, y: 14, w: 12, h: 13 } },
];
