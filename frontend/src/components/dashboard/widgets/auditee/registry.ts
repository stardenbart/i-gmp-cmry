import type { WidgetDefinition } from "@/components/dashboard/types";
import { AuditeeStatsWidget } from "./AuditeeStatsWidget";
import { AuditeeTaskListWidget } from "./AuditeeTaskListWidget";

// Keep widget_id values in sync with backend/internal/domain/dashboard/dashboard_layout.go.
export const auditeeWidgetRegistry: WidgetDefinition[] = [
  { id: "auditee-stats", title: "Kartu Statistik", Component: AuditeeStatsWidget, defaultLayout: { x: 0, y: 0, w: 12, h: 3 } },
  { id: "auditee-task-list", title: "Daftar Temuan Prioritas", Component: AuditeeTaskListWidget, defaultLayout: { x: 0, y: 3, w: 12, h: 8 } },
];
