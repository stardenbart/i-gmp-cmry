import type { WidgetDefinition } from "@/components/dashboard/WidgetGrid";
import { AuditeeStatsWidget } from "./AuditeeStatsWidget";
import { AuditeeTaskListWidget } from "./AuditeeTaskListWidget";

// Keep widget_id values in sync with backend/internal/domain/dashboard/dashboard_layout.go.
export const auditeeWidgetRegistry: WidgetDefinition[] = [
  { id: "auditee-stats", title: "Kartu Statistik", Component: AuditeeStatsWidget },
  { id: "auditee-task-list", title: "Daftar Temuan Prioritas", Component: AuditeeTaskListWidget },
];
