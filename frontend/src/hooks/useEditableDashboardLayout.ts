"use client";

import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { dashboardLayoutApi, type WidgetConfig } from "@/lib/api/dashboard-layout.api";
import type { WidgetDefinition } from "@/components/dashboard/WidgetGrid";

export interface EditableWidgetRow {
  id: string;
  title: string;
  visible: boolean;
}

/**
 * Local, editable copy of a dashboard layout for the "Sesuaikan Dashboard"
 * modal (Fase 3a: checklist + up/down order, no drag). Initializes from the
 * saved/default layout, lets the caller toggle visibility and move rows, and
 * exposes a save() that persists the result via PUT /dashboard/layout.
 */
export function useEditableDashboardLayout(registry: WidgetDefinition[], isOpen: boolean) {
  const queryClient = useQueryClient();
  const registryIds = registry.map((w) => w.id);

  const { data: saved, isLoading } = useQuery({
    queryKey: ["dashboard-layout"],
    queryFn: () => dashboardLayoutApi.get(),
    enabled: isOpen,
    staleTime: 60_000,
  });

  const [rows, setRows] = useState<EditableWidgetRow[]>([]);

  // Re-seed local editable state whenever the modal (re-)opens with fresh data.
  useEffect(() => {
    if (!isOpen || isLoading) return;

    const byId = new Map(registry.map((w) => [w.id, w.title]));
    const knownSet = new Set(registryIds);

    const fromSaved = (saved || [])
      .filter((w) => knownSet.has(w.widget_id))
      .sort((a, b) => a.order - b.order)
      .map((w) => ({ id: w.widget_id, title: byId.get(w.widget_id) || w.widget_id, visible: w.visible }));

    const seen = new Set(fromSaved.map((r) => r.id));
    const missing = registry.filter((w) => !seen.has(w.id)).map((w) => ({ id: w.id, title: w.title, visible: true }));

    setRows([...fromSaved, ...missing]);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, isLoading, saved]);

  const toggleVisible = (id: string) => {
    setRows((prev) => prev.map((r) => (r.id === id ? { ...r, visible: !r.visible } : r)));
  };

  const move = (id: string, direction: "up" | "down") => {
    setRows((prev) => {
      const index = prev.findIndex((r) => r.id === id);
      const targetIndex = direction === "up" ? index - 1 : index + 1;
      if (index === -1 || targetIndex < 0 || targetIndex >= prev.length) return prev;
      const next = [...prev];
      [next[index], next[targetIndex]] = [next[targetIndex], next[index]];
      return next;
    });
  };

  const saveMutation = useMutation({
    mutationFn: () => {
      const widgets: WidgetConfig[] = rows.map((r, index) => ({
        widget_id: r.id,
        visible: r.visible,
        order: index,
      }));
      return dashboardLayoutApi.save(widgets);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dashboard-layout"] });
    },
  });

  return {
    rows,
    isLoading,
    toggleVisible,
    move,
    save: saveMutation.mutateAsync,
    isSaving: saveMutation.isPending,
  };
}
