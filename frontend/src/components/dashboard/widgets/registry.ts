import { adminWidgetRegistry } from "./admin/registry";
import { auditorWidgetRegistry } from "./auditor/registry";
import { auditeeWidgetRegistry } from "./auditee/registry";
import type { WidgetDefinition } from "@/components/dashboard/types";

// Every main-dashboard widget, across every role, in one list. IDs are
// confirmed unique across the three per-role registries below. Used by
// all 3 role dashboards (Admin/Auditor/Auditee) AND the Edit User "Tata
// Letak Dashboard" tab — so a widget built for one role can be shown on
// another role's dashboard, but ONLY once an Admin explicitly turns it on
// via Edit User (see DashboardGrid.tsx's resolveWidgets: a widget with no
// saved config for a given user defaults to hidden, not shown, so simply
// being present in this merged list never surfaces it on its own).
export const allMainDashboardWidgets: WidgetDefinition[] = [
  ...adminWidgetRegistry,
  ...auditorWidgetRegistry,
  ...auditeeWidgetRegistry,
];
