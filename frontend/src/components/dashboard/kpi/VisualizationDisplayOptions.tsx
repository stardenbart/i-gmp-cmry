"use client";

import type { AnalyticsChartType } from "@/lib/api/analytics.api";
import { labelToggleDisabledReason, PIE_FAMILY, TREND_LINE_CHARTS } from "./visualizationCompatibility";

interface VisualizationDisplayOptionsProps {
  vizType: AnalyticsChartType | null;
  showLabels: boolean;
  showLabelValues: boolean;
  showTrendLine: boolean;
  onShowLabelsChange: (checked: boolean) => void;
  onShowLabelValuesChange: (checked: boolean) => void;
  onShowTrendLineChange: (checked: boolean) => void;
}

export function VisualizationDisplayOptions({
  vizType,
  showLabels,
  showLabelValues,
  showTrendLine,
  onShowLabelsChange,
  onShowLabelValuesChange,
  onShowTrendLineChange,
}: VisualizationDisplayOptionsProps) {
  const labelsDisabledReason = labelToggleDisabledReason(vizType);
  const labelValuesDisabled = !showLabels || !vizType || !PIE_FAMILY.includes(vizType);
  const trendDisabled = !vizType || !TREND_LINE_CHARTS.includes(vizType);

  return (
    <div>
      <p className="mb-1.5 text-xs font-semibold text-foreground">Opsi Tampilan</p>
      <div className="flex flex-wrap gap-4">
        <label
          title={labelsDisabledReason ?? "Tampilkan nilai/nama langsung di atas chart, bukan cuma melalui hover"}
          className={`flex items-center gap-1.5 text-xs font-medium ${labelsDisabledReason ? "cursor-not-allowed text-muted-foreground/50" : "cursor-pointer text-foreground"}`}
        >
          <input
            type="checkbox"
            checked={showLabels}
            disabled={!!labelsDisabledReason}
            onChange={(event) => onShowLabelsChange(event.target.checked)}
            className="h-3.5 w-3.5 rounded border-border accent-primary"
          />
          Aktifkan Label
        </label>
        <label
          title={labelValuesDisabled ? "Hanya berlaku untuk Pie/Donut/Treemap/Funnel dengan Label aktif" : "Tampilkan angka nilai di label"}
          className={`flex items-center gap-1.5 text-xs font-medium ${labelValuesDisabled ? "cursor-not-allowed text-muted-foreground/50" : "cursor-pointer text-foreground"}`}
        >
          <input
            type="checkbox"
            checked={showLabelValues}
            disabled={labelValuesDisabled}
            onChange={(event) => onShowLabelValuesChange(event.target.checked)}
            className="h-3.5 w-3.5 rounded border-border accent-primary"
          />
          Tampilkan Nilai di Label
        </label>
        <label
          title={trendDisabled ? "Hanya berlaku untuk Bar Chart/Bar Horizontal/Bar Bertumpuk" : "Tambah garis yang mengikuti nilai batang"}
          className={`flex items-center gap-1.5 text-xs font-medium ${trendDisabled ? "cursor-not-allowed text-muted-foreground/50" : "cursor-pointer text-foreground"}`}
        >
          <input
            type="checkbox"
            checked={showTrendLine}
            disabled={trendDisabled}
            onChange={(event) => onShowTrendLineChange(event.target.checked)}
            className="h-3.5 w-3.5 rounded border-border accent-primary"
          />
          Tambahkan Garis Tren (Line)
        </label>
      </div>
    </div>
  );
}
