import { DynamicKPIWidget } from "@/components/dashboard/kpi/DynamicKPIWidget";
import { DRILLABLE_CHARTS, MATRIX_FAMILY } from "@/components/dashboard/kpi/visualizationCompatibility";
import type { WidgetDefinition, VizType } from "@/components/dashboard/types";
import type { WidgetConfig } from "@/lib/api/dashboard-layout.api";

const DYNAMIC_WIDGET_VIZ_TYPES: VizType[] = [
  "bar", "horizontal_bar", "stacked_bar", "line", "area", "radar", "scatter",
  "pie", "donut", "treemap", "funnel", "number_card", "gauge", "heatmap", "sankey", "table",
];

function supportedTypes(customQuery: NonNullable<WidgetConfig["custom_query"]>): VizType[] {
  if (customQuery.dimension2) return [...MATRIX_FAMILY];
  if (customQuery.drillDimensions?.length) return [...DRILLABLE_CHARTS];
  return DYNAMIC_WIDGET_VIZ_TYPES;
}

export function buildDynamicWidgetDefinition(cfg: WidgetConfig, onEditWidget?: (cfg: WidgetConfig) => void): WidgetDefinition | null {
  const query = cfg.custom_query;
  if (!query || query.version !== 1 || !query.dimension || query.measures.length === 0) return null;
  const title = query.title?.trim() || "Visualisasi Kustom";
  return {
    id: cfg.widget_id,
    title,
    Component: ({ vizType }: { vizType?: VizType }) => (
      <DynamicKPIWidget
        widgetId={cfg.widget_id}
        title={title}
        measures={query.measures}
        dimension={query.dimension}
        dimension2={query.dimension2}
        drillDimensions={query.drillDimensions}
        showLabels={query.showLabels}
        showLabelValues={query.showLabelValues}
        showTrendLine={query.showTrendLine}
        vizType={vizType}
        onEdit={onEditWidget ? () => onEditWidget(cfg) : undefined}
      />
    ),
    defaultLayout: { x: 0, y: 9999, w: 6, h: 10 },
    supportedVizTypes: supportedTypes(query),
    defaultVizType: query.dimension2 ? "heatmap" : "bar",
  };
}
