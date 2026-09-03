"use client";

import { useMemo } from "react";
import { BarChart3, Info, Loader2 } from "lucide-react";
import { useDebounce } from "@/hooks/useDebounce";
import { DynamicKPIWidget } from "./DynamicKPIWidget";
import type { AnalyticsChartType } from "@/lib/api/analytics.api";

interface KPIVisualizationPreviewProps {
  title: string;
  measureIds: string[];
  dimensionId: string | null;
  dimension2Id: string | null;
  drillIds: string[];
  vizType: AnalyticsChartType | null;
  showLabels: boolean;
  showLabelValues: boolean;
  showTrendLine: boolean;
}

interface PreviewQueryConfig {
  measureIds: string[];
  dimensionId: string;
  dimension2Id?: string;
  drillIds?: string[];
  vizType: AnalyticsChartType;
}

function previewEmptyMessage(
  dimensionId: string | null,
  measureIds: string[],
  vizType: AnalyticsChartType | null,
) {
  if (!dimensionId && measureIds.length === 0) {
    return "Seret satu Kategori dan minimal satu Nilai untuk melihat hasil visualisasi.";
  }
  if (!dimensionId) return "Tambahkan field Kategori agar data dapat dikelompokkan.";
  if (measureIds.length === 0) return "Tambahkan minimal satu field Nilai untuk membentuk visualisasi.";
  if (!vizType) return "Kombinasi field ini belum memiliki jenis visualisasi yang kompatibel.";
  return "Lengkapi konfigurasi untuk menampilkan preview.";
}

/**
 * Live, non-persisted preview for VisualizationBuilder. Data-related config
 * is debounced so a short series of drag/drop changes produces one analytics
 * request. Display-only options and the title still update immediately.
 */
export function KPIVisualizationPreview({
  title,
  measureIds,
  dimensionId,
  dimension2Id,
  drillIds,
  vizType,
  showLabels,
  showLabelValues,
  showTrendLine,
}: KPIVisualizationPreviewProps) {
  const isReady = !!dimensionId && measureIds.length > 0 && !!vizType;

  const queryConfig = useMemo<PreviewQueryConfig | null>(
    () =>
      isReady
        ? {
            measureIds,
            dimensionId,
            dimension2Id: dimension2Id || undefined,
            drillIds: drillIds.length ? drillIds : undefined,
            vizType,
          }
        : null,
    [dimension2Id, dimensionId, drillIds, isReady, measureIds, vizType],
  );
  const debouncedConfig = useDebounce(queryConfig, 300);
  const currentSignature = queryConfig
    ? `${queryConfig.measureIds.join(",")}|${queryConfig.dimensionId}|${queryConfig.dimension2Id ?? ""}|${queryConfig.drillIds?.join(",") ?? ""}|${queryConfig.vizType}`
    : "";
  const debouncedSignature = debouncedConfig
    ? `${debouncedConfig.measureIds.join(",")}|${debouncedConfig.dimensionId}|${debouncedConfig.dimension2Id ?? ""}|${debouncedConfig.drillIds?.join(",") ?? ""}|${debouncedConfig.vizType}`
    : "";
  const isSynchronizing = isReady && currentSignature !== debouncedSignature;

  return (
    <aside className="lg:sticky lg:top-0 lg:self-start" aria-label="Preview visualisasi">
      <div className="mb-2 flex items-center justify-between gap-2">
        <div>
          <p className="text-xs font-semibold text-foreground">Preview Visualisasi</p>
          <p className="text-[11px] text-muted-foreground">Menggunakan data aktual sesuai plant aktif.</p>
        </div>
        {isSynchronizing && (
          <span className="inline-flex items-center gap-1 rounded-full bg-primary/10 px-2 py-1 text-[10px] font-semibold text-primary">
            <Loader2 className="h-3 w-3 animate-spin" /> Menyesuaikan
          </span>
        )}
      </div>

      <div className="relative h-[380px] overflow-hidden rounded-2xl border border-border bg-muted/10 p-2 shadow-inner">
        {!isReady ? (
          <div className="flex h-full flex-col items-center justify-center gap-3 px-8 text-center">
            <span className="rounded-2xl bg-primary/10 p-3 text-primary">
              <BarChart3 className="h-7 w-7" />
            </span>
            <div>
              <p className="text-sm font-semibold text-foreground">Preview belum tersedia</p>
              <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
                {previewEmptyMessage(dimensionId, measureIds, vizType)}
              </p>
            </div>
          </div>
        ) : !debouncedConfig ? (
          <div className="flex h-full items-center justify-center gap-2 text-xs text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" /> Menyiapkan preview...
          </div>
        ) : (
          <div className="h-full">
            <DynamicKPIWidget
              title={title.trim() || "Preview Visualisasi"}
              measures={debouncedConfig.measureIds}
              dimension={debouncedConfig.dimensionId}
              dimension2={debouncedConfig.dimension2Id}
              drillDimensions={debouncedConfig.drillIds}
              vizType={debouncedConfig.vizType}
              showLabels={showLabels}
              showLabelValues={showLabelValues}
              showTrendLine={showTrendLine}
              badgeLabel="Preview"
            />
            {isSynchronizing && <div className="pointer-events-none absolute inset-2 rounded-xl bg-background/25 backdrop-blur-[1px]" />}
          </div>
        )}
      </div>

      <p className="mt-2 flex items-start gap-1.5 text-[11px] leading-relaxed text-muted-foreground">
        <Info className="mt-0.5 h-3 w-3 shrink-0" />
        Preview tidak disimpan. Dashboard baru berubah setelah Anda menekan tombol tambah atau simpan.
      </p>
    </aside>
  );
}
