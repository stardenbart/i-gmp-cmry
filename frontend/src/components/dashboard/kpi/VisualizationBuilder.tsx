"use client";

import { useMemo, useState, type DragEvent } from "react";
import { createPortal } from "react-dom";
import { useQuery } from "@tanstack/react-query";
import { X, GripVertical, Plus, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { useMounted } from "@/lib/useMounted";
import { analyticsApi, type CatalogField, type AnalyticsChartType } from "@/lib/api/analytics.api";
import { getApiErrorMessage } from "@/lib/api/error";
import type { VizType } from "@/components/dashboard/types";
import { CategoryDropZone } from "./CategoryDropZone";
import { KPIVisualizationPreview } from "./KPIVisualizationPreview";
import { VisualizationDisplayOptions } from "./VisualizationDisplayOptions";
import { VisualizationTypePicker } from "./VisualizationTypePicker";
import {
  categoriesFromSavedConfig,
  compatibleCharts,
  isMatrixChart,
  MAX_CATEGORIES,
  maxMeasuresFor,
  resolveCategories,
} from "./visualizationCompatibility";
import {
  KPI_FIELD_DRAG_MIME,
  readFieldDragPayload,
  type VisualizationFieldDragPayload,
  type VisualizationFieldKind,
} from "./visualizationDnd";

export interface BuiltVisualization {
  title: string;
  measures: string[];
  dimension: string;
  /** Derived from the second ordered category for Heatmap/Sankey. */
  dimension2?: string;
  /** Derived from every ordered category after `dimension` for regular charts. */
  drillDimensions?: string[];
  /** Pure rendering options — see buildEChartsOption.ts's ChartDisplayOptions. */
  showLabels?: boolean;
  showLabelValues?: boolean;
  showTrendLine?: boolean;
  vizType: VizType;
}

interface VisualizationBuilderProps {
  onClose: () => void;
  onAdd: (config: BuiltVisualization) => void;
  isSubmitting?: boolean;
  /** "edit" pre-fills every field from `initial` and changes the modal's
   * title/submit label — used by the pencil icon on an existing custom
   * widget (see DynamicKPIWidget/DashboardKPI) so a widget's config can be
   * revised without deleting and rebuilding it. Defaults to "add". */
  mode?: "add" | "edit";
  /** Only meaningful with mode="edit" — this component always mounts fresh
   * per open (`{builderOpen && <VisualizationBuilder .../>}`), so every
   * piece of state below can safely lazy-init from this once, no effect
   * needed to keep it in sync. */
  initial?: BuiltVisualization;
}

// Every draggable "field" the user sees is a plain catalog entry fetched
// from the backend (GET /analytics/catalog) — there is deliberately no
// hardcoded list of measures/dimensions anywhere in this file. Dragging
// never carries anything beyond a field id + which list it came from.
function FieldCard({ field, kind, onSelect, selected }: { field: CatalogField; kind: VisualizationFieldKind; onSelect: () => void; selected: boolean }) {
  const handleDragStart = (event: DragEvent<HTMLDivElement>) => {
    const payload: VisualizationFieldDragPayload = { kind, id: field.id };
    event.dataTransfer.setData(KPI_FIELD_DRAG_MIME, JSON.stringify(payload));
    event.dataTransfer.setData("text/plain", field.label);
    event.dataTransfer.effectAllowed = "copy";
  };
  return (
    <div
      draggable
      unselectable="on"
      onDragStart={handleDragStart}
      onClick={onSelect}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          onSelect();
        }
      }}
      role="button"
      tabIndex={0}
      aria-pressed={selected}
      title={`${field.description} — klik atau seret untuk menambahkan`}
      className={`flex cursor-grab select-none items-center gap-1.5 rounded-lg border px-2.5 py-1.5 text-xs font-medium shadow-sm transition-colors active:cursor-grabbing ${
        selected ? "border-primary bg-primary/10 text-primary" : "border-border bg-card text-foreground hover:border-primary/50 hover:bg-primary/5"
      }`}
    >
      <GripVertical className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
      <span className="truncate">{field.label}</span>
      {field.fans_out && <span className="shrink-0 text-[10px] text-amber-500" title="Bisa dihitung lebih dari sekali per temuan">⚠</span>}
    </div>
  );
}

function DropZone({
  label,
  hint,
  items,
  onDrop,
  onRemove,
  emptyLabel,
  acceptKind,
}: {
  label: string;
  hint: string;
  items: CatalogField[];
  onDrop: (payload: VisualizationFieldDragPayload) => void;
  onRemove: (id: string) => void;
  emptyLabel: string;
  acceptKind: VisualizationFieldKind;
}) {
  const [isOver, setIsOver] = useState(false);
  return (
    <div>
      <p className="mb-1.5 text-xs font-semibold text-foreground">{label}</p>
      <div
        onDragOver={(e) => {
          e.preventDefault();
          setIsOver(true);
        }}
        onDragLeave={() => setIsOver(false)}
        onDrop={(e) => {
          e.preventDefault();
          setIsOver(false);
          const payload = readFieldDragPayload(e.dataTransfer);
          if (payload?.kind === acceptKind) onDrop(payload);
        }}
        className={`flex min-h-[64px] flex-wrap items-center gap-2 rounded-xl border-2 border-dashed p-3 transition-colors ${
          isOver ? "border-primary bg-primary/5" : "border-border bg-muted/20"
        }`}
      >
        {items.length === 0 && <span className="text-xs italic text-muted-foreground">{emptyLabel}</span>}
        {items.map((item) => (
          <span
            key={item.id}
            className="inline-flex items-center gap-1.5 rounded-lg bg-primary/10 px-2.5 py-1 text-xs font-semibold text-primary"
          >
            {item.label}
            <button type="button" onClick={() => onRemove(item.id)} aria-label={`Hapus ${item.label}`} className="hover:text-red-500">
              <X className="h-3 w-3" />
            </button>
          </span>
        ))}
      </div>
      <p className="mt-1 text-[11px] text-muted-foreground">{hint}</p>
    </div>
  );
}

/**
 * Power BI-style drag-and-drop panel for building a custom KPI widget:
 * drag a Dimension into "Kategori" and one or more Measures into "Nilai",
 * pick a compatible chart type, then add it to the dashboard. The user
 * never sees a field id, SQL, or formula — only the catalog's label/
 * description text (see analyticsApi.getCatalog).
 */
export function VisualizationBuilder({ onClose, onAdd, isSubmitting = false, mode = "add", initial }: VisualizationBuilderProps) {
  const mounted = useMounted();
  const { data: catalog, isLoading, isError, error } = useQuery({
    queryKey: ["analytics-catalog"],
    queryFn: () => analyticsApi.getCatalog(),
    staleTime: 10 * 60_000,
  });

  const [categoryIds, setCategoryIds] = useState<string[]>(() =>
    categoriesFromSavedConfig(initial?.dimension, initial?.dimension2, initial?.drillDimensions),
  );
  const [measureIds, setMeasureIds] = useState<string[]>(() => initial?.measures ?? []);
  const [vizType, setVizType] = useState<AnalyticsChartType | null>(() => (initial?.vizType && initial.vizType !== "list" ? (initial.vizType as AnalyticsChartType) : null));
  const [title, setTitle] = useState(() => initial?.title ?? "");
  const [showLabels, setShowLabels] = useState(() => initial?.showLabels ?? false);
  const [showLabelValues, setShowLabelValues] = useState(() => initial?.showLabelValues ?? false);
  const [showTrendLine, setShowTrendLine] = useState(() => initial?.showTrendLine ?? false);
  const [submitAttempted, setSubmitAttempted] = useState(false);

  const dimensionById = useMemo(() => new Map((catalog?.dimensions ?? []).map((d) => [d.id, d])), [catalog]);
  const measureById = useMemo(() => new Map((catalog?.measures ?? []).map((m) => [m.id, m])), [catalog]);

  const selectedCategories = categoryIds.map((id) => dimensionById.get(id)).filter((field): field is CatalogField => !!field);
  const selectedDimension = selectedCategories[0] ?? null;
  const selectedMeasures = measureIds.map((id) => measureById.get(id)).filter((m): m is CatalogField => !!m);
  const availableCharts = useMemo(
    () => compatibleCharts(selectedMeasures, selectedDimension, selectedCategories.length),
    [selectedMeasures, selectedDimension, selectedCategories.length],
  );
  const effectiveVizType: AnalyticsChartType | null = vizType && availableCharts.includes(vizType) ? vizType : availableCharts[0] ?? null;
  const resolvedCategories = resolveCategories(categoryIds, effectiveVizType);
  const matrixMode = isMatrixChart(effectiveVizType);
  // Only "table" can hold more than MAX_MEASURES columns — see
  // compatibleCharts()'s matching filter, which is what forces
  // effectiveVizType back to "table" if it's the only option left once
  // this cap is exceeded.
  const measuresLimit = maxMeasuresFor(effectiveVizType);

  const normalizedTitle = title.trim();
  const validationIssues = [
    normalizedTitle.length === 0 ? "isi judul visualisasi" : null,
    !selectedDimension ? "tambahkan minimal 1 kategori" : null,
    selectedMeasures.length === 0 ? "tambahkan minimal 1 nilai" : null,
    selectedDimension && selectedMeasures.length > 0 && !effectiveVizType ? "pilih kombinasi visualisasi yang kompatibel" : null,
  ].filter((issue): issue is string => !!issue);
  const canAdd = validationIssues.length === 0;

  const handleDropMeasure = (payload: VisualizationFieldDragPayload) => {
    setMeasureIds((prev) => {
      if (prev.includes(payload.id)) return prev;
      if (prev.length >= measuresLimit) return prev;
      return [...prev, payload.id];
    });
  };

  const handleAddCategory = (fieldId: string) => {
    setCategoryIds((previous) => {
      if (previous.includes(fieldId) || previous.length >= MAX_CATEGORIES) return previous;
      return [...previous, fieldId];
    });
  };

  if (!mounted || typeof window === "undefined") return null;

  return createPortal(
    <div
      className="fixed inset-0 z-[110] flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      aria-labelledby="visualization-builder-title"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget && !isSubmitting) onClose();
      }}
    >
      <div className="flex max-h-[90vh] w-full max-w-6xl flex-col rounded-2xl border border-border/60 bg-background shadow-2xl">
        <div className="flex items-center justify-between border-b border-border px-5 py-4">
          <div>
            <h2 id="visualization-builder-title" className="text-base font-semibold text-foreground">
              {mode === "edit" ? "Edit Visualisasi" : "Tambah Visualisasi"}
            </h2>
            <p className="text-xs text-muted-foreground">
              {mode === "edit"
                ? "Ubah konfigurasi dan periksa preview sebelum menyimpan perubahan."
                : "Seret Kategori dan Nilai, lalu periksa preview sebelum menambahkannya ke dashboard."}
            </p>
          </div>
          <button type="button" onClick={onClose} disabled={isSubmitting} className="rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground">
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="flex-1 space-y-5 overflow-y-auto px-5 py-4">
          {isLoading ? (
            <div className="flex items-center justify-center py-10">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : isError ? (
            <p className="py-10 text-center text-sm text-red-500">{getApiErrorMessage(error, "Gagal memuat daftar field")}</p>
          ) : (
            <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1.15fr)_minmax(340px,0.85fr)]">
              <div className="min-w-0 space-y-5">
              <div>
                <div className="mb-1.5 flex items-center justify-between gap-3">
                  <label htmlFor="custom-visualization-title" className="text-xs font-semibold text-foreground">
                    Judul visualisasi
                  </label>
                  <span className="text-[10px] tabular-nums text-muted-foreground">{title.length}/100</span>
                </div>
                <input
                  id="custom-visualization-title"
                  type="text"
                  value={title}
                  maxLength={100}
                  onChange={(event) => setTitle(event.target.value)}
                  placeholder="Contoh: Tren Temuan per Bulan"
                  className={`h-10 w-full rounded-xl border bg-card px-3 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-primary focus:ring-2 focus:ring-primary/20 ${submitAttempted && !normalizedTitle ? "border-red-500" : "border-border"}`}
                />
                <p className="mt-1 text-[11px] text-muted-foreground">Gunakan nama yang menjelaskan isi dan tujuan grafik.</p>
                {submitAttempted && !normalizedTitle && <p className="mt-1 text-[11px] font-medium text-red-500">Judul visualisasi wajib diisi.</p>}
              </div>

              <div>
                <p className="mb-1.5 text-xs font-semibold text-foreground">Field tersedia</p>
                <div className="space-y-2 rounded-xl border border-border bg-muted/10 p-3">
                  <div>
                    <p className="mb-1 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Kategori <span className="normal-case font-normal">— klik atau seret</span></p>
                    <div className="flex flex-wrap gap-1.5">
                      {(catalog?.dimensions ?? []).map((d) => (
                        <FieldCard key={d.id} field={d} kind="dimension" selected={categoryIds.includes(d.id)} onSelect={() => handleAddCategory(d.id)} />
                      ))}
                    </div>
                  </div>
                  <div>
                    <p className="mb-1 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Nilai <span className="normal-case font-normal">— klik atau seret</span></p>
                    <div className="flex flex-wrap gap-1.5">
                      {(catalog?.measures ?? []).map((m) => (
                        <FieldCard key={m.id} field={m} kind="measure" selected={measureIds.includes(m.id)} onSelect={() => handleDropMeasure({ kind: "measure", id: m.id })} />
                      ))}
                    </div>
                  </div>
                </div>
              </div>

              <div className="space-y-4">
                <CategoryDropZone
                  items={selectedCategories}
                  onChange={setCategoryIds}
                  matrixMode={matrixMode}
                  maxItems={matrixMode ? 2 : MAX_CATEGORIES}
                />
                <DropZone
                  label={`Nilai (maks. ${measuresLimit})`}
                  hint="Satu atau lebih angka yang ingin ditampilkan."
                  items={selectedMeasures}
                  onDrop={handleDropMeasure}
                  onRemove={(id) => setMeasureIds((prev) => prev.filter((m) => m !== id))}
                  emptyLabel="Seret 1+ field Nilai ke sini"
                  acceptKind="measure"
                />
              </div>

              <VisualizationTypePicker
                availableCharts={availableCharts}
                selectedType={effectiveVizType}
                selectedMeasures={selectedMeasures}
                primaryCategory={selectedDimension}
                categoryCount={selectedCategories.length}
                onChange={setVizType}
              />

              <VisualizationDisplayOptions
                vizType={effectiveVizType}
                showLabels={showLabels}
                showLabelValues={showLabelValues}
                showTrendLine={showTrendLine}
                onShowLabelsChange={setShowLabels}
                onShowLabelValuesChange={setShowLabelValues}
                onShowTrendLineChange={setShowTrendLine}
              />

              {selectedMeasures.some((m) => m.fans_out) && (
                <p className="rounded-lg border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-[11px] text-amber-700 dark:text-amber-400">
                  ⚠ Salah satu field yang dipilih bisa terhitung lebih dari sekali per temuan (lihat deskripsi field). Totalnya bisa lebih besar dari jumlah temuan unik.
                </p>
              )}
              </div>

              <KPIVisualizationPreview
                title={title}
                measureIds={measureIds}
                dimensionId={resolvedCategories.dimensionId}
                dimension2Id={resolvedCategories.dimension2Id}
                drillIds={resolvedCategories.drillIds}
                vizType={effectiveVizType}
                showLabels={showLabels}
                showLabelValues={showLabelValues}
                showTrendLine={showTrendLine}
              />
            </div>
          )}
        </div>

        <div className="flex flex-col gap-3 border-t border-border px-5 py-4 sm:flex-row sm:items-center">
          {validationIssues.length > 0 && (
            <p className={`mr-auto text-xs ${submitAttempted ? "font-medium text-red-500" : "text-muted-foreground"}`} role={submitAttempted ? "alert" : undefined}>
              Belum bisa ditambahkan: {validationIssues.join("; ")}.
            </p>
          )}
          <div className="flex items-center justify-end gap-2">
          <Button variant="outline" size="sm" onClick={onClose} disabled={isSubmitting}>
            Batal
          </Button>
          <Button
            size="sm"
            className="gap-2"
            disabled={isSubmitting || isLoading || isError}
            isLoading={isSubmitting}
            onClick={() => {
              if (!canAdd || !selectedDimension || !effectiveVizType || selectedMeasures.length === 0) {
                setSubmitAttempted(true);
                toast.error(`Lengkapi konfigurasi: ${validationIssues.join(", ")}`);
                return;
              }
              onAdd({
                title: normalizedTitle,
                measures: selectedMeasures.map((m) => m.id),
                dimension: selectedDimension.id,
                dimension2: resolvedCategories.dimension2Id ?? undefined,
                drillDimensions: resolvedCategories.drillIds.length ? resolvedCategories.drillIds : undefined,
                showLabels,
                showLabelValues,
                showTrendLine,
                vizType: effectiveVizType,
              });
            }}
          >
            <Plus className="h-4 w-4" />
            {mode === "edit" ? "Simpan Perubahan" : "Tambah ke Dashboard"}
          </Button>
          </div>
        </div>
      </div>
    </div>,
    document.body
  );
}
