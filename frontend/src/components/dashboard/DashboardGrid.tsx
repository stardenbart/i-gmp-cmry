"use client";

import "react-grid-layout/css/styles.css";
import "react-resizable/css/styles.css";

import { useCallback, useEffect, useMemo, useState, type DragEvent, type ReactElement } from "react";
import { Responsive, type Layout, type LayoutItem, type ResponsiveLayouts } from "react-grid-layout";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { LayoutGrid, Loader2, Save, X, Plus, EyeOff, LayoutDashboard } from "lucide-react";
import { Button } from "@/components/ui/button";
import { dashboardLayoutApi, type WidgetConfig, type LayoutTarget } from "@/lib/api/dashboard-layout.api";
import type { WidgetDefinition, VizType } from "@/components/dashboard/types";

// Labels only — purely additive to a Record<VizType,string> literal, no
// logic/JSX change. Safe for the 3 existing role dashboards: each widget's
// own `supportedVizTypes` (registry.ts) still lists only its original
// subset, so none of the new entries below ever appear in their picker —
// only the new Custom KPI Builder widgets (DashboardKPI.tsx) opt into them.
const VIZ_TYPE_LABELS: Record<VizType, string> = {
  list: "Tampilan Bawaan",
  table: "Tabel",
  bar: "Bar Chart",
  horizontal_bar: "Bar Horizontal",
  stacked_bar: "Bar Bertumpuk",
  line: "Line Chart",
  area: "Area Chart",
  radar: "Radar Chart",
  scatter: "Scatter Plot",
  pie: "Pie Chart",
  donut: "Donut Chart",
  treemap: "Treemap",
  funnel: "Funnel",
  number_card: "Kartu Angka",
  gauge: "Gauge",
  heatmap: "Heatmap",
  sankey: "Sankey",
};

const BREAKPOINTS = { lg: 768, xs: 0 };
const COLS = { lg: 12, xs: 1 };
const ROW_HEIGHT = 32;
const GRID_MARGIN_Y = 16;

function gridItemPixelHeight(rows: number): number {
  const normalizedRows = Math.max(1, Math.round(rows));
  return ROW_HEIGHT * normalizedRows + GRID_MARGIN_Y * (normalizedRows - 1);
}

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
  /** Read-only layout supplied by a public dashboard snapshot. When set,
   * DashboardGrid never calls the authenticated layout API. */
  providedLayout?: WidgetConfig[];
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
  /**
   * Which of the caller's several independently-saved dashboards this is.
   * Defaults to "main" (every dashboard before this existed) — passing a
   * different key (e.g. "kpi") persists/loads a completely separate layout
   * row so it never collides with the caller's main dashboard.
   */
  dashboardKey?: string;
  /**
   * Replace the "Disembunyikan: [+ pill]" click-to-restore bar with a real
   * drag-and-drop Toolbox panel (react-grid-layout's official Toolbox
   * pattern) — drag a card from the panel onto the grid to add it back,
   * at the exact position dropped. Off by default so the 3 existing role
   * dashboards and the Edit User dashboard tab keep their current
   * click-only behavior untouched; opt in per dashboard as needed.
   */
  toolboxMode?: boolean;
}

interface ResolvedWidget {
  def: WidgetDefinition;
  visible: boolean;
  layout: LayoutItem;
  vizType: VizType;
  // Opaque to this component — DashboardGrid never reads/interprets it
  // (only the Custom KPI Visualization Builder does, see DashboardKPI.tsx)
  // but it MUST round-trip through save/load unchanged, or a plain
  // "Simpan Tata Letak" from normal edit mode (move/resize/hide/viz-type)
  // would silently wipe a custom widget's query config and make it
  // disappear on the next reload.
  customQuery?: WidgetConfig["custom_query"];
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
      // A widget present in `registry` but with NO saved config for this
      // user defaults to HIDDEN, not shown. `registry` is no longer
      // necessarily "just this user's own role's widgets" — it's the
      // merged cross-role catalog (allMainDashboardWidgets) — so a
      // missing entry now genuinely means "never configured for this
      // user" (e.g. a widget that belongs to a different role, or one an
      // Admin hasn't turned on yet via Edit User), not "a brand-new
      // widget everyone should see immediately". Every role's actual
      // default widgets still show correctly because the backend's
      // DefaultLayoutForRole always returns an explicit visible:true
      // entry for each of them before a user has ever customized anything.
      visible: cfg ? cfg.visible : false,
      layout: { ...normalized, i: def.id },
      vizType: (cfg?.viz_type as VizType | undefined) ?? def.defaultVizType ?? "list",
      customQuery: cfg?.custom_query
        ? { ...cfg.custom_query, title: cfg.custom_query.title?.trim() || def.title }
        : undefined,
    };
  });
}

function getWidgetTitle(widget: ResolvedWidget) {
  return widget.customQuery?.title?.trim() || widget.def.title;
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
  onVizTypeChange: (id: string, vizType: VizType) => void;
  onTitleChange: (id: string, title: string) => void;
}

function DashboardWidgetFrame({ widget, isEditing, previewOnly, onHide, onVizTypeChange, onTitleChange }: DashboardWidgetFrameProps) {
  // Not every widget declares vizType support/takes a vizType prop (the 15
  // widgets across the Admin/Auditor/Auditee registries don't) — cast once
  // here rather than constraining the shared WidgetDefinition.Component
  // type, so those are unaffected.
  const Component = widget.def.Component as unknown as (props: { vizType?: VizType }) => ReactElement;
  const hasVizPicker = isEditing && (widget.def.supportedVizTypes?.length ?? 0) > 0;
  const title = getWidgetTitle(widget);
  return (
    <div className="relative h-full min-w-0 group">
      {isEditing && (
        <div className="widget-drag-handle absolute inset-x-0 top-0 z-10 flex items-center justify-between gap-1 rounded-t-xl bg-primary/90 px-2 py-1 text-xs font-semibold text-primary-foreground cursor-move">
          {widget.customQuery ? (
            <input
              type="text"
              value={widget.customQuery.title ?? ""}
              maxLength={100}
              aria-label={`Ubah judul ${title}`}
              title="Ubah judul visualisasi"
              placeholder="Judul visualisasi"
              onPointerDown={(event) => event.stopPropagation()}
              onClick={(event) => event.stopPropagation()}
              onChange={(event) => onTitleChange(widget.def.id, event.target.value)}
              className={`min-w-0 flex-1 rounded border px-1.5 py-0.5 text-xs font-semibold text-primary-foreground outline-none placeholder:text-primary-foreground/60 ${
                widget.customQuery.title?.trim() ? "border-white/20 bg-white/10 focus:border-white/60" : "border-red-200 bg-red-500/30"
              }`}
            />
          ) : (
            <span className="truncate">{title}</span>
          )}
          <div className="flex shrink-0 items-center gap-1">
            {hasVizPicker && (
              <select
                aria-label={`Ubah visualisasi ${title}`}
                title="Ubah tampilan visualisasi"
                value={widget.vizType}
                onPointerDown={(event) => event.stopPropagation()}
                onClick={(event) => event.stopPropagation()}
                onChange={(event) => onVizTypeChange(widget.def.id, event.target.value as VizType)}
                className="cursor-pointer rounded bg-white/15 px-1 py-0.5 text-[11px] font-medium text-primary-foreground outline-none hover:bg-white/25"
              >
                {(widget.def.supportedVizTypes ?? []).map((type) => (
                  <option key={type} value={type} className="text-foreground">
                    {VIZ_TYPE_LABELS[type]}
                  </option>
                ))}
              </select>
            )}
            <button
              type="button"
              onClick={(event) => {
                event.stopPropagation();
                onHide(widget.def.id);
              }}
              className="shrink-0 rounded p-0.5 transition-colors hover:bg-white/20"
              aria-label={`Sembunyikan ${title}`}
              title="Sembunyikan widget ini"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>
      )}
      <div className={isEditing ? "h-full min-w-0 pt-7 pointer-events-none select-none" : "h-full min-w-0"}>
        {previewOnly ? (
          <div className="flex h-full min-h-28 w-full flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-border bg-muted/30 p-4 text-muted-foreground">
            <LayoutDashboard className="h-6 w-6 opacity-50" />
            <span className="text-center text-xs font-medium">{title}</span>
          </div>
        ) : (
          <Component vizType={widget.vizType} />
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
  onVizTypeChange,
  onTitleChange,
}: {
  widgets: ResolvedWidget[];
  wide: boolean;
  isEditing: boolean;
  previewOnly: boolean;
  onHide: (id: string) => void;
  onVizTypeChange: (id: string, vizType: VizType) => void;
  onTitleChange: (id: string, title: string) => void;
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
                style={
                  wide
                    ? {
                        gridColumn: `${start} / span ${span}`,
                        // Custom widgets contain optional drill controls. In
                        // natural/view mode their card must still honor the
                        // same saved `h` used by react-grid-layout; otherwise
                        // controls/breadcrumbs grow the DOM height and push
                        // every following row farther down after an edit.
                        ...(widget.customQuery ? { height: gridItemPixelHeight(widget.layout.h) } : {}),
                      }
                    : undefined
                }
              >
                <DashboardWidgetFrame
                  widget={widget}
                  isEditing={isEditing}
                  previewOnly={previewOnly}
                  onHide={onHide}
                  onVizTypeChange={onVizTypeChange}
                  onTitleChange={onTitleChange}
                />
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
export function DashboardGrid({ registry, enabled, providedLayout, target, editable = false, previewOnly = false, dashboardKey = "main", toolboxMode = false }: DashboardGridProps) {
  const queryClient = useQueryClient();
  const { width, containerRef, mounted } = useElementWidth();
  const queryKey = ["dashboard-layout", target?.userId ?? "self", dashboardKey];
  const layoutTarget = { ...target, dashboardKey };

  const { data: saved, isLoading } = useQuery({
    queryKey,
    queryFn: () => dashboardLayoutApi.get(layoutTarget),
    enabled: enabled && !providedLayout,
    staleTime: 60_000,
  });

  const effectiveSaved = providedLayout ?? saved;
  const resolved = useMemo(() => resolveWidgets(registry, effectiveSaved), [registry, effectiveSaved]);

  const [isEditing, setIsEditing] = useState(false);
  const [draft, setDraft] = useState<ResolvedWidget[]>(resolved);
  const activeWidgets = isEditing ? draft : resolved;

  // `draft` is a deliberate snapshot so free-form drag/resize/hide edits
  // don't touch the live view until "Simpan Tata Letak" — but that means it
  // never sees a save that happens through a DIFFERENT mutation while this
  // session is still open (the Custom KPI Builder's own "Tambah Visualisasi"
  // / edit-pencil flow, see DashboardKPI.tsx, which writes straight to the
  // same query cache). Without this, adding/editing a custom widget while
  // "Sesuaikan Dashboard" happens to be open silently doesn't show up —
  // the widget (or its new display options) only appears after Batal/Simpan
  // Tata Letak or a reload, which reads as "tersimpan tapi tidak muncul".
  // `resolved` can only change here from such an EXTERNAL save — this
  // component's OWN in-session edits (vizType/title via the header
  // controls) live in `draft` alone and never reach `saved` until "Simpan
  // Tata Letak" — so it's always safe to take the fresh def/vizType/
  // customQuery here, as long as this session's own structural edits
  // (position/size/visibility) are preserved for widgets that already
  // existed in the draft. Uses React's "adjust state during render" pattern
  // (comparing against the last-synced reference) rather than an effect +
  // setState, per https://react.dev/learn/you-might-not-need-an-effect —
  // no extra render/commit cycle, and avoids the react-hooks/set-state-in-effect rule.
  const [lastSyncedResolved, setLastSyncedResolved] = useState(resolved);
  if (isEditing && resolved !== lastSyncedResolved) {
    setLastSyncedResolved(resolved);
    setDraft((prev) => {
      const prevById = new Map(prev.map((w) => [w.def.id, w]));
      return resolved.map((fresh) => {
        const existing = prevById.get(fresh.def.id);
        return existing ? { ...fresh, layout: existing.layout, visible: existing.visible } : fresh;
      });
    });
  }
  const visibleWidgets = activeWidgets.filter((w) => w.visible);
  const hiddenWidgets = activeWidgets.filter((w) => !w.visible);
  const hasDesktopGrid = width >= BREAKPOINTS.lg;

  // Toolbox drag-and-drop (only used when toolboxMode is on): which hidden
  // widget is currently being dragged from the panel, so the grid knows
  // what size placeholder to show and which widget to reveal on drop.
  const [draggingWidget, setDraggingWidget] = useState<ResolvedWidget | null>(null);

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

  const changeVizType = (id: string, vizType: VizType) => {
    setDraft((prev) => prev.map((w) => (w.def.id === id ? { ...w, vizType } : w)));
  };

  const changeCustomTitle = (id: string, title: string) => {
    setDraft((prev) =>
      prev.map((w) =>
        w.def.id === id && w.customQuery
          ? { ...w, customQuery: { ...w.customQuery, title } }
          : w
      )
    );
  };

  const hasInvalidCustomTitle = draft.some((widget) => widget.customQuery && !widget.customQuery.title?.trim());

  // Toolbox: a card being dragged from the panel sets `draggingWidget` so
  // the grid can size its drop placeholder to that widget's default w/h.
  // `text/plain` must be set for the drag to register at all in Firefox.
  const handleToolboxDragStart = (widget: ResolvedWidget) => (event: DragEvent<HTMLDivElement>) => {
    setDraggingWidget(widget);
    event.dataTransfer.effectAllowed = "copy";
    event.dataTransfer.setData("text/plain", widget.def.id);
  };

  const handleToolboxDragEnd = () => setDraggingWidget(null);

  // react-grid-layout's onDrop hands back the full layout including the
  // placeholder item at its computed drop x/y — that placeholder's own `i`
  // is a synthetic id (see droppingItem below), so the real widget being
  // added is read from `draggingWidget` (set on drag-start), not from the
  // dropped item itself.
  const handleGridDrop = (_layout: Layout, item: LayoutItem | undefined) => {
    const widget = draggingWidget;
    setDraggingWidget(null);
    if (!widget || !item) return;
    setDraft((prev) =>
      prev.map((w) =>
        w.def.id === widget.def.id
          ? { ...w, visible: true, layout: { i: w.def.id, x: item.x, y: item.y, w: item.w, h: item.h } }
          : w
      )
    );
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
        viz_type: w.vizType,
        custom_query: w.customQuery
          ? { ...w.customQuery, title: w.customQuery.title?.trim() }
          : undefined,
      }));
      return dashboardLayoutApi.save(widgets, layoutTarget);
    },
    onSuccess: (savedWidgets) => {
      // Keep view mode and the next edit session in sync with exactly what
      // the backend persisted, including custom visualization titles.
      queryClient.setQueryData(queryKey, savedWidgets);
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

  if (!providedLayout && isLoading) {
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
              <Button
                size="sm"
                onClick={() => saveMutation.mutate()}
                isLoading={saveMutation.isPending}
                disabled={hasInvalidCustomTitle}
                title={hasInvalidCustomTitle ? "Judul visualisasi kustom wajib diisi" : undefined}
                className="gap-2"
              >
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
        toolboxMode && hasDesktopGrid ? (
          <div className="flex flex-wrap items-center gap-2 mb-4 p-3 rounded-xl border border-dashed border-border bg-muted/30">
            <span className="text-xs font-medium text-muted-foreground flex items-center gap-1.5 mr-1">
              <LayoutGrid className="h-3.5 w-3.5" /> Toolbox — seret ke grid untuk menampilkan:
            </span>
            {hiddenWidgets.map((w) => (
              <div
                key={w.def.id}
                draggable
                unselectable="on"
                onDragStart={handleToolboxDragStart(w)}
                onDragEnd={handleToolboxDragEnd}
                className="inline-flex cursor-grab select-none items-center gap-1.5 rounded-lg border-2 border-dashed border-primary/40 bg-card px-3 py-1.5 text-xs font-medium text-foreground shadow-sm transition-colors active:cursor-grabbing hover:border-primary hover:bg-primary/5"
              >
                <LayoutDashboard className="h-3.5 w-3.5 text-primary/70" />
                {getWidgetTitle(w)}
              </div>
            ))}
          </div>
        ) : (
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
                {getWidgetTitle(w)}
              </button>
            ))}
          </div>
        )
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
              margin={[16, GRID_MARGIN_Y]}
              dragConfig={{ enabled: true, handle: ".widget-drag-handle" }}
              resizeConfig={{ enabled: true }}
              onLayoutChange={handleLayoutChange}
              {...(toolboxMode
                ? {
                    dropConfig: { enabled: true },
                    droppingItem: draggingWidget
                      ? { i: "__dropping-elem__", x: 0, y: 0, w: draggingWidget.def.defaultLayout.w, h: draggingWidget.def.defaultLayout.h }
                      : undefined,
                    onDrop: handleGridDrop,
                  }
                : {})}
            >
              {visibleWidgets.map((widget) => (
                <div key={widget.def.id} className="min-w-0">
                  <DashboardWidgetFrame
                    widget={widget}
                    isEditing
                    previewOnly={previewOnly}
                    onHide={hideWidget}
                    onVizTypeChange={changeVizType}
                    onTitleChange={changeCustomTitle}
                  />
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
              onVizTypeChange={changeVizType}
              onTitleChange={changeCustomTitle}
            />
          )
        )}
      </div>
    </div>
  );
}
