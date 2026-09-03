"use client";

import {
  AreaChart,
  BarChart3,
  BarChartHorizontal,
  Disc3,
  Filter,
  Gauge as GaugeIcon,
  Grid3x3,
  Hash,
  Layers,
  LayoutGrid,
  LineChart,
  PieChart,
  Radar,
  ScatterChart,
  Table2,
  Workflow,
} from "lucide-react";
import type { AnalyticsChartType, CatalogField } from "@/lib/api/analytics.api";
import { ALL_CHART_TYPES, incompatibleReason } from "./visualizationCompatibility";

type ChartGroup = "Perbandingan" | "Proporsi" | "Nilai Tunggal" | "Matriks (2 Kategori)" | "Lainnya";

const GROUP_ORDER: ChartGroup[] = ["Perbandingan", "Proporsi", "Nilai Tunggal", "Matriks (2 Kategori)", "Lainnya"];
const CHART_META: Record<AnalyticsChartType, { label: string; icon: typeof BarChart3; group: ChartGroup }> = {
  bar: { label: "Bar Chart", icon: BarChart3, group: "Perbandingan" },
  horizontal_bar: { label: "Bar Horizontal", icon: BarChartHorizontal, group: "Perbandingan" },
  stacked_bar: { label: "Bar Bertumpuk", icon: Layers, group: "Perbandingan" },
  line: { label: "Line Chart", icon: LineChart, group: "Perbandingan" },
  area: { label: "Area Chart", icon: AreaChart, group: "Perbandingan" },
  radar: { label: "Radar Chart", icon: Radar, group: "Perbandingan" },
  pie: { label: "Pie Chart", icon: PieChart, group: "Proporsi" },
  donut: { label: "Donut Chart", icon: Disc3, group: "Proporsi" },
  treemap: { label: "Treemap", icon: LayoutGrid, group: "Proporsi" },
  funnel: { label: "Funnel", icon: Filter, group: "Proporsi" },
  number_card: { label: "Kartu Angka", icon: Hash, group: "Nilai Tunggal" },
  gauge: { label: "Gauge", icon: GaugeIcon, group: "Nilai Tunggal" },
  heatmap: { label: "Heatmap", icon: Grid3x3, group: "Matriks (2 Kategori)" },
  sankey: { label: "Sankey", icon: Workflow, group: "Matriks (2 Kategori)" },
  scatter: { label: "Scatter Plot", icon: ScatterChart, group: "Lainnya" },
  table: { label: "Tabel", icon: Table2, group: "Lainnya" },
};

interface VisualizationTypePickerProps {
  availableCharts: AnalyticsChartType[];
  selectedType: AnalyticsChartType | null;
  selectedMeasures: CatalogField[];
  primaryCategory: CatalogField | null;
  categoryCount: number;
  onChange: (type: AnalyticsChartType) => void;
}

export function VisualizationTypePicker({
  availableCharts,
  selectedType,
  selectedMeasures,
  primaryCategory,
  categoryCount,
  onChange,
}: VisualizationTypePickerProps) {
  const chartsByGroup = new Map<ChartGroup, AnalyticsChartType[]>();
  for (const type of ALL_CHART_TYPES) {
    const group = CHART_META[type].group;
    chartsByGroup.set(group, [...(chartsByGroup.get(group) ?? []), type]);
  }

  return (
    <div>
      <p className="mb-1.5 text-xs font-semibold text-foreground">Jenis Visualisasi</p>
      {availableCharts.length === 0 ? (
        <p className="text-xs italic text-muted-foreground">
          {categoryCount === 0 || selectedMeasures.length === 0
            ? "Tambahkan Kategori dan Nilai terlebih dahulu untuk melihat pilihan visualisasi."
            : "Kombinasi Kategori dan Nilai ini belum didukung."}
        </p>
      ) : (
        <div className="space-y-3">
          {GROUP_ORDER.map((group) => (
            <div key={group}>
              <p className="mb-1 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">{group}</p>
              <div className="flex flex-wrap gap-2">
                {(chartsByGroup.get(group) ?? []).map((type) => {
                  const meta = CHART_META[type];
                  const Icon = meta.icon;
                  const isCompatible = availableCharts.includes(type);
                  const isActive = selectedType === type;
                  return (
                    <button
                      key={type}
                      type="button"
                      disabled={!isCompatible}
                      onClick={() => onChange(type)}
                      title={
                        !isCompatible
                          ? incompatibleReason(type, selectedMeasures.length, primaryCategory, categoryCount)
                          : type === "gauge"
                            ? "Skala otomatis, bukan target bisnis"
                            : meta.label
                      }
                      className={`flex items-center gap-1.5 rounded-lg border px-3 py-1.5 text-xs font-medium transition-colors ${
                        !isCompatible
                          ? "cursor-not-allowed border-border/50 text-muted-foreground/50"
                          : isActive
                            ? "border-primary bg-primary/10 text-primary"
                            : "border-border text-foreground hover:border-primary/50"
                      }`}
                    >
                      <Icon className="h-3.5 w-3.5" />
                      {meta.label}
                    </button>
                  );
                })}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
