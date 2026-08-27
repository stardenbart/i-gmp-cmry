"use client";

import type { ComponentType } from "react";
import { useDashboardLayout } from "@/hooks/useDashboardLayout";

export interface WidgetDefinition {
  id: string;
  title: string;
  Component: ComponentType;
  /** Grid span classes applied to the wrapping cell (Tailwind), e.g. "lg:col-span-2". */
  className?: string;
}

interface WidgetGridProps {
  registry: WidgetDefinition[];
  enabled: boolean;
}

/**
 * Renders every visible widget from `registry`, in the order the user saved
 * (or the registry's own default order, before that endpoint resolves / for
 * a user who never customized anything).
 */
export function WidgetGrid({ registry, enabled }: WidgetGridProps) {
  const widgetIds = registry.map((w) => w.id);
  const { order } = useDashboardLayout(widgetIds, enabled);
  const byId = new Map(registry.map((w) => [w.id, w]));

  return (
    <>
      {order.map((id) => {
        const widget = byId.get(id);
        if (!widget) return null;
        const { Component } = widget;
        return (
          <div key={id} className={widget.className}>
            <Component />
          </div>
        );
      })}
    </>
  );
}
