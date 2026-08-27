"use client";

import "react-grid-layout/css/styles.css";
import "react-resizable/css/styles.css";

import { useEffect, useMemo, useState } from "react";
import { Responsive, useContainerWidth, type Layout, type LayoutItem, type ResponsiveLayouts } from "react-grid-layout";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { LayoutGrid, Loader2, Save, X, Plus, EyeOff } from "lucide-react";
import { Button } from "@/components/ui/button";
import { dashboardLayoutApi, type WidgetConfig } from "@/lib/api/dashboard-layout.api";
import type { WidgetDefinition } from "@/components/dashboard/types";

const BREAKPOINTS = { lg: 768, xs: 0 };
const COLS = { lg: 12, xs: 1 };
const ROW_HEIGHT = 32;

interface DashboardGridProps {
  registry: WidgetDefinition[];
  enabled: boolean;
}

interface ResolvedWidget {
  def: WidgetDefinition;
  visible: boolean;
  layout: LayoutItem;
}

function resolveWidgets(registry: WidgetDefinition[], saved: WidgetConfig[] | undefined): ResolvedWidget[] {
  const byId = new Map((saved || []).map((w) => [w.widget_id, w]));
  return registry.map((def) => {
    const cfg = byId.get(def.id);
    const pos = cfg && cfg.x != null && cfg.y != null && cfg.w != null && cfg.h != null ? { x: cfg.x, y: cfg.y, w: cfg.w, h: cfg.h } : def.defaultLayout;
    return {
      def,
      visible: cfg ? cfg.visible : true,
      layout: { i: def.id, ...pos },
    };
  });
}

function toSingleColumnLayout(items: ResolvedWidget[]): Layout {
  let y = 0;
  return items
    .filter((w) => w.visible)
    .map((w) => {
      const item: LayoutItem = { i: w.def.id, x: 0, y, w: 1, h: w.layout.h };
      y += w.layout.h;
      return item;
    });
}

/**
 * Fase 3b: the customizable dashboard grid — drag to reorder, drag the
 * corner to resize. Replaces Fase 3a's checklist+order modal with direct
 * manipulation, in an explicit "Edit Layout" mode (view mode stays static
 * so normal dashboard browsing isn't accidentally draggable).
 */
export function DashboardGrid({ registry, enabled }: DashboardGridProps) {
  const queryClient = useQueryClient();
  const { width, containerRef, mounted } = useContainerWidth();

  const { data: saved, isLoading } = useQuery({
    queryKey: ["dashboard-layout"],
    queryFn: () => dashboardLayoutApi.get(),
    enabled,
    staleTime: 60_000,
  });

  const resolved = useMemo(() => resolveWidgets(registry, saved), [registry, saved]);

  const [isEditing, setIsEditing] = useState(false);
  const [draft, setDraft] = useState<ResolvedWidget[]>(resolved);

  // Re-seed the draft whenever fresh data arrives and we're not mid-edit, so
  // editing always starts from the latest saved state.
  useEffect(() => {
    if (!isEditing) setDraft(resolved);
  }, [resolved, isEditing]);

  const visibleWidgets = draft.filter((w) => w.visible);
  const hiddenWidgets = draft.filter((w) => !w.visible);

  const lgLayout: Layout = visibleWidgets.map((w) => w.layout);
  const layouts: ResponsiveLayouts = {
    lg: lgLayout,
    xs: toSingleColumnLayout(draft),
  };

  const handleLayoutChange = (_layout: Layout, allLayouts: ResponsiveLayouts) => {
    if (!isEditing) return;
    const lg = allLayouts.lg;
    if (!lg) return;
    setDraft((prev) =>
      prev.map((w) => {
        const match = lg.find((item) => item.i === w.def.id);
        return match ? { ...w, layout: { i: w.def.id, x: match.x, y: match.y, w: match.w, h: match.h } } : w;
      })
    );
  };

  const hideWidget = (id: string) => {
    setDraft((prev) => prev.map((w) => (w.def.id === id ? { ...w, visible: false } : w)));
  };

  const showWidget = (id: string) => {
    setDraft((prev) => prev.map((w) => (w.def.id === id ? { ...w, visible: true } : w)));
  };

  const saveMutation = useMutation({
    mutationFn: () => {
      const widgets: WidgetConfig[] = draft.map((w, index) => ({
        widget_id: w.def.id,
        visible: w.visible,
        order: index,
        x: w.layout.x,
        y: w.layout.y,
        w: w.layout.w,
        h: w.layout.h,
      }));
      return dashboardLayoutApi.save(widgets);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dashboard-layout"] });
      setIsEditing(false);
      toast.success("Tata letak dashboard berhasil disimpan");
    },
    onError: () => {
      toast.error("Gagal menyimpan tata letak dashboard");
    },
  });

  const handleCancel = () => {
    setDraft(resolved);
    setIsEditing(false);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-end gap-2 mb-3">
        {isEditing ? (
          <>
            <span className="text-xs text-muted-foreground mr-auto hidden sm:inline">
              Seret untuk pindahkan, tarik sudut kanan-bawah untuk ubah ukuran.
            </span>
            <Button variant="outline" size="sm" onClick={handleCancel} disabled={saveMutation.isPending}>
              Batal
            </Button>
            <Button size="sm" onClick={() => saveMutation.mutate()} isLoading={saveMutation.isPending} className="gap-2">
              <Save className="h-4 w-4" />
              Simpan Tata Letak
            </Button>
          </>
        ) : (
          <Button variant="outline" size="sm" onClick={() => setIsEditing(true)} className="gap-2">
            <LayoutGrid className="h-4 w-4" />
            <span className="hidden sm:inline">Sesuaikan Dashboard</span>
          </Button>
        )}
      </div>

      {isEditing && hiddenWidgets.length > 0 && (
        <div className="flex flex-wrap items-center gap-2 mb-4 p-3 rounded-xl border border-dashed border-border bg-muted/30">
          <span className="text-xs font-medium text-muted-foreground flex items-center gap-1.5 mr-1">
            <EyeOff className="h-3.5 w-3.5" /> Disembunyikan:
          </span>
          {hiddenWidgets.map((w) => (
            <button
              key={w.def.id}
              type="button"
              onClick={() => showWidget(w.def.id)}
              className="inline-flex items-center gap-1 px-2.5 py-1 text-xs font-medium rounded-full bg-card border border-border hover:border-primary/50 hover:text-primary transition-colors"
            >
              <Plus className="h-3 w-3" />
              {w.def.title}
            </button>
          ))}
        </div>
      )}

      <div ref={containerRef}>
        {mounted && width > 0 && (
          <Responsive
            layouts={layouts}
            breakpoints={BREAKPOINTS}
            cols={COLS}
            width={width}
            rowHeight={ROW_HEIGHT}
            margin={[16, 16]}
            dragConfig={{ enabled: isEditing, handle: ".widget-drag-handle" }}
            resizeConfig={{ enabled: isEditing }}
            onLayoutChange={handleLayoutChange}
          >
            {visibleWidgets.map((w) => {
              const { Component } = w.def;
              return (
                <div key={w.def.id} className="relative group">
                  {isEditing && (
                    <div className="widget-drag-handle absolute inset-x-0 top-0 z-10 flex items-center justify-between px-2 py-1 bg-primary/90 text-primary-foreground rounded-t-xl cursor-move text-xs font-semibold">
                      <span className="truncate">{w.def.title}</span>
                      <button
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation();
                          hideWidget(w.def.id);
                        }}
                        className="p-0.5 rounded hover:bg-white/20 transition-colors shrink-0"
                        aria-label={`Sembunyikan ${w.def.title}`}
                        title="Sembunyikan widget ini"
                      >
                        <X className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  )}
                  <div className={isEditing ? "h-full pt-7 pointer-events-none select-none" : "h-full"}>
                    <Component />
                  </div>
                </div>
              );
            })}
          </Responsive>
        )}
      </div>
    </div>
  );
}
