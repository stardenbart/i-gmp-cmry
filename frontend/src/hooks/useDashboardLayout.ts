"use client";

import { useQuery } from "@tanstack/react-query";
import { dashboardLayoutApi, type WidgetConfig } from "@/lib/api/dashboard-layout.api";

/**
 * Resolves which widgets to render and in what order for the current user.
 *
 * Reconciles the saved/default layout (server) against the widget registry
 * (code) so neither side can break the other:
 *  - a widget the registry knows about but the saved layout doesn't mention
 *    yet (newly added widget) is appended, visible by default;
 *  - a widget the saved layout mentions but the registry no longer has
 *    (removed/renamed widget) is silently dropped.
 */
export function useDashboardLayout(knownWidgetIds: readonly string[], enabled: boolean) {
  const query = useQuery({
    queryKey: ["dashboard-layout"],
    queryFn: () => dashboardLayoutApi.get(),
    enabled,
    staleTime: 60_000,
  });

  const resolvedOrder = resolveOrder(query.data, knownWidgetIds);

  return {
    order: resolvedOrder,
    isLoading: query.isLoading,
    refetch: query.refetch,
  };
}

function resolveOrder(saved: WidgetConfig[] | undefined, knownWidgetIds: readonly string[]): string[] {
  const knownSet = new Set(knownWidgetIds);
  const fromSaved = (saved || [])
    .filter((w) => w.visible && knownSet.has(w.widget_id))
    .sort((a, b) => a.order - b.order)
    .map((w) => w.widget_id);

  const seen = new Set(fromSaved);
  const missing = knownWidgetIds.filter((id) => !seen.has(id));

  return [...fromSaved, ...missing];
}
