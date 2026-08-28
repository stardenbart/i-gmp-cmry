"use client";

import "react-grid-layout/css/styles.css";
import "react-resizable/css/styles.css";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Responsive, type Layout, type LayoutItem, type ResponsiveLayouts } from "react-grid-layout";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { LayoutGrid, Loader2, Save, X, Plus, EyeOff, LayoutDashboard } from "lucide-react";
import { Button } from "@/components/ui/button";
import { dashboardLayoutApi, type WidgetConfig, type LayoutTarget } from "@/lib/api/dashboard-layout.api";
import type { WidgetDefinition } from "@/components/dashboard/types";

const BREAKPOINTS = { lg: 768, xs: 0 };
const COLS = { lg: 12, xs: 1 };
const ROW_HEIGHT = 32;

/**
 * Measures the width of the element it's attached to.
 *
 * NOTE: intentionally NOT react-grid-layout's own `useContainerWidth`. That
 * hook's ResizeObserver-attaching effect only re-runs when its internal
 * `mounted` flag flips — but `mounted` starts out already `true` (its
 * `measureBeforeMount` option defaults to `false`), so the effect fires
 * exactly once, on the very first render. DashboardGrid renders a loading
 * spinner (no grid container in the tree yet) while the layout query is in
 * flight, so that one-shot effect finds `containerRef.current === null` and
 * bails — permanently. The reported width then stays stuck at the hook's
 * `initialWidth` default (1280px) forever, so every screen — including a
 * 390px phone — gets treated as "desktop" and laid out with the saved
 * 12-column x/y/w/h coordinates, producing the cramped/overlapping mobile
 * layout. A callback ref sidesteps this: its identity only changes when the
 * DOM node itself changes, so the effect re-attaches correctly no matter how
 * many loading-gated renders happened first.
 */
function useElementWidth() {
  const [node, setNode] = useState<HTMLDivElement | null>(null);
  const [width, setWidth] = useState(0);
  const containerRef = useCallback((el: HTMLDivElement | null) => {
    setNode(el);
    setWidth(el ? Math.round(el.getBoundingClientRect().width) : 0);
  }, []);

  useEffect(() => {
    if (!node) return;
    if (typeof ResizeObserver === "undefined") return;
    const observer = new ResizeObserver((entries) => {
      const entry = entries[0];
      if (entry) setWidth(Math.round(entry.contentRect.width));
    });
    observer.observe(node);
    return () => observer.disconnect();
  }, [node]);

  return { width, containerRef, mounted: width > 0 };
}

interface DashboardGridProps {
  registry: WidgetDefinition[];
  enabled: boolean;
  /** Whose layout to load/save. Omit to act on the caller themself. */
  target?: LayoutTarget;
  /**
   * Whether the "Sesuaikan Dashboard" edit mode is available at all.
   * Dashboard customization is Admin/Super-Admin-only, configured from the
   * Edit User screen — so a user's own live dashboard renders `editable=false`
   * (pure display of whatever layout was configured for them), while the
   * Edit User "Dashboard" tab renders `editable=true` targeting that user.
   */
  editable?: boolean;
  /**
   * Render lightweight placeholder cards (title only) instead of mounting
   * the real widget components. Required outside a role's own live
   * dashboard page — e.g. the Edit User "Dashboard" tab arranges another
   * user's widgets without that user's role Context/data providers mounted
   * (and showing the *viewing admin's own* data there would be wrong
   * anyway, since those widgets read from useAuthStore's logged-in user).
   */
  previewOnly?: boolean;
}

interface ResolvedWidget {
  def: WidgetDefinition;
  visible: boolean;
  layout: LayoutItem;
}

function normalizeLayout(layout: { x: number; y: number; w: number; h: number }, fallback: WidgetDefinition["defaultLayout"]): LayoutItem {
  const width = Number.isFinite(layout.w) ? Math.min(12, Math.max(1, Math.round(layout.w))) : fallback.w;
  return {
    i: "",
    x: Number.isFinite(layout.x) ? Math.min(12 - width, Math.max(0, Math.round(layout.x))) : fallback.x,
    y: Number.isFinite(layout.y) ? Math.max(0, Math.round(layout.y)) : fallback.y,
    w: width,
    h: Number.isFinite(layout.h) ? Math.min(30, Math.max(2, Math.round(layout.h))) : fallback.h,
  };
}

function resolveWidgets(registry: WidgetDefinition[], saved: WidgetConfig[] | undefined): ResolvedWidget[] {
  const byId = new Map((saved || []).map((w) => [w.widget_id, w]));
  return registry.map((def) => {
    const cfg = byId.get(def.id);
    const pos = cfg && cfg.x != null && cfg.y != null && cfg.w != null && cfg.h != null ? { x: cfg.x, y: cfg.y, w: cfg.w, h: cfg.h } : def.defaultLayout;
    const normalized = normalizeLayout(pos, def.defaultLayout);
    return {
      def,
      visible: cfg ? cfg.visible : true,
      layout: { ...normalized, i: def.id },
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

function groupWidgetsByRow(items: ResolvedWidget[]): ResolvedWidget[][] {
  const rows = new Map<number, ResolvedWidget[]>();
  [...items]
    .sort((a, b) => a.layout.y - b.layout.y || a.layout.x - b.layout.x)
    .forEach((widget) => {
      const row = rows.get(widget.layout.y) || [];
      row.push(widget);
      rows.set(widget.layout.y, row);
    });
  return [...rows.values()];
}

interface DashboardWidgetFrameProps {
  widget: ResolvedWidget;
  isEditing: boolean;
  previewOnly: boolean;
  onHide: (id: string) => void;
}

function DashboardWidgetFrame({ widget, isEditing, previewOnly, onHide }: DashboardWidgetFrameProps) {
  const { Component } = widget.def;
  return (
    <div className="relative h-full min-w-0 group">
      {isEditing && (
        <div className="widget-drag-handle absolute inset-x-0 top-0 z-10 flex items-center justify-between rounded-t-xl bg-primary/90 px-2 py-1 text-xs font-semibold text-primary-foreground cursor-move">
          <span className="truncate">{widget.def.title}</span>
          <button
            type="button"
            onClick={(event) => {
              event.stopPropagation();
              onHide(widget.def.id);
            }}
            className="shrink-0 rounded p-0.5 transition-colors hover:bg-white/20"
            aria-label={`Sembunyikan ${widget.def.title}`}
            title="Sembunyikan widget ini"
          >
            <X className="h-3.5 w-3.5" />
          </button>
        </div>
      )}
      <div className={isEditing ? "h-full min-w-0 pt-7 pointer-events-none select-none" : "h-full min-w-0"}>
        {previewOnly ? (
          <div className="flex h-full min-h-28 w-full flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-border bg-muted/30 p-4 text-muted-foreground">
            <LayoutDashboard className="h-6 w-6 opacity-50" />
            <span className="text-center text-xs font-medium">{widget.def.title}</span>
          </div>
        ) : (
          <Component />
        )}
      </div>
    </div>
  );
}

function NaturalDashboardLayout({
  widgets,
  wide,
  isEditing,
  previewOnly,
  onHide,
}: {
  widgets: ResolvedWidget[];
  wide: boolean;
  isEditing: boolean;
  previewOnly: boolean;
  onHide: (id: string) => void;
}) {
  return (
    <div className="space-y-4">
      {groupWidgetsByRow(widgets).map((row) => (
        <div
          key={`${row[0].layout.y}-${row.map((widget) => widget.def.id).join("-")}`}
          className="grid min-w-0 gap-4"
          style={{ gridTemplateColumns: wide ? "repeat(12, minmax(0, 1fr))" : "minmax(0, 1fr)" }}
        >
          {row.map((widget) => {
            const span = Math.min(12, Math.max(1, widget.layout.w));
            const start = Math.min(13 - span, Math.max(1, widget.layout.x + 1));
            return (
              <div
                key={widget.def.id}
                className="min-w-0"
                style={wide ? { gridColumn: `${start} / span ${span}` } : undefined}
              >
                <DashboardWidgetFrame widget={widget} isEditing={isEditing} previewOnly={previewOnly} onHide={onHide} />
              </div>
            );
          })}
        </div>
      ))}
    </div>
  );
}

/**
 * Fase 3b: the customizable dashboard grid — drag to reorder, drag the
 * corner to resize. Replaces Fase 3a's checklist+order modal with direct
 * manipulation, in an explicit "Edit Layout" mode (view mode stays static
 * so normal dashboard browsing isn't accidentally draggable).
 */
export function DashboardGrid({ registry, enabled, target, editable = false, previewOnly = false }: DashboardGridProps) {
  const queryClient = useQueryClient();
  const { width, containerRef, mounted } = useElementWidth();
  const queryKey = ["dashboard-layout", target?.userId ?? "self"];

  const { data: saved, isLoading } = useQuery({
    queryKey,
    queryFn: () => dashboardLayoutApi.get(target),
    enabled,
    staleTime: 60_000,
  });

  const resolved = useMemo(() => resolveWidgets(registry, saved), [registry, saved]);

  const [isEditing, setIsEditing] = useState(false);
  const [draft, setDraft] = useState<ResolvedWidget[]>(resolved);
  const activeWidgets = isEditing ? draft : resolved;
  const visibleWidgets = activeWidgets.filter((w) => w.visible);
  const hiddenWidgets = activeWidgets.filter((w) => !w.visible);
  const hasDesktopGrid = width >= BREAKPOINTS.lg;

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
      return dashboardLayoutApi.save(widgets, target);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey });
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

  const startEditing = () => {
    setDraft(resolved);
    setIsEditing(true);
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
      {editable && (
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
            <Button variant="outline" size="sm" onClick={startEditing} className="gap-2">
              <LayoutGrid className="h-4 w-4" />
              <span className="hidden sm:inline">Sesuaikan Dashboard</span>
            </Button>
          )}
        </div>
      )}

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

      {isEditing && mounted && !hasDesktopGrid && (
        <p className="mb-4 rounded-xl border border-border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
          Pada layar ini widget mengikuti tinggi kontennya. Gunakan layar yang lebih lebar untuk mengubah posisi dan ukuran; pengaturan tampil/sembunyi tetap dapat disimpan.
        </p>
      )}

      <div ref={containerRef}>
        {mounted && width > 0 && (
          isEditing && hasDesktopGrid ? (
            <Responsive
              layouts={layouts}
              breakpoints={BREAKPOINTS}
              cols={COLS}
              width={width}
              rowHeight={ROW_HEIGHT}
              margin={[16, 16]}
              dragConfig={{ enabled: true, handle: ".widget-drag-handle" }}
              resizeConfig={{ enabled: true }}
              onLayoutChange={handleLayoutChange}
            >
              {visibleWidgets.map((widget) => (
                <div key={widget.def.id} className="min-w-0">
                  <DashboardWidgetFrame widget={widget} isEditing previewOnly={previewOnly} onHide={hideWidget} />
                </div>
              ))}
            </Responsive>
          ) : (
            <NaturalDashboardLayout
              widgets={visibleWidgets}
              wide={hasDesktopGrid}
              isEditing={isEditing}
              previewOnly={previewOnly}
              onHide={hideWidget}
            />
          )
        )}
      </div>
    </div>
  );
}
