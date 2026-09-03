import type { WidgetDefinition } from "@/components/dashboard/types";
import { withAuditeeDashboardContext } from "@/components/dashboard/auditee/AuditeeDashboardContext";
import { AuditeeStatsWidget } from "./AuditeeStatsWidget";
import { AuditeeTaskListWidget } from "./AuditeeTaskListWidget";

// Keep widget_id values in sync with backend/internal/domain/dashboard/dashboard_layout.go.
//
// Every Component is wrapped in withAuditeeDashboardContext so it can also
// render on a non-Auditee dashboard (Admin can mix these into any user's
// layout via Edit User → Tata Letak Dashboard — see widgets/registry.ts's
// allMainDashboardWidgets) without crashing for lack of AuditeeDashboardContext.
export const auditeeWidgetRegistry: WidgetDefinition[] = [
  { id: "auditee-stats", title: "Kartu Statistik", Component: withAuditeeDashboardContext(AuditeeStatsWidget), defaultLayout: { x: 0, y: 0, w: 12, h: 3 } },
  { id: "auditee-task-list", title: "Daftar Temuan Prioritas", Component: withAuditeeDashboardContext(AuditeeTaskListWidget), defaultLayout: { x: 0, y: 3, w: 12, h: 8 } },
];
