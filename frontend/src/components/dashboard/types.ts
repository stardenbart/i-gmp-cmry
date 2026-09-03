import type { ComponentType } from "react";

export interface WidgetGridPosition {
  x: number;
  y: number;
  w: number;
  h: number;
}

// "list" = the widget's own bespoke markup (ranked cards, progress bars,
// etc. — whatever it already renders today). Custom KPI widgets render the
// chart types below through DynamicKPIWidget/ECharts.
export type VizType =
  | "list"
  | "table"
  | "bar"
  | "horizontal_bar"
  | "stacked_bar"
  | "line"
  | "area"
  | "radar"
  | "scatter"
  | "pie"
  | "donut"
  | "treemap"
  | "funnel"
  | "number_card"
  | "gauge"
  | "heatmap"
  | "sankey";

export interface WidgetDefinition {
  id: string;
  title: string;
  Component: ComponentType;
  /** Position/size used the first time a user opens this dashboard, before
   * they've ever dragged/resized anything (12-column grid). */
  defaultLayout: WidgetGridPosition;
  /** Which visualization types this widget can be switched between (Power
   * BI-style picker in the widget header, edit mode only). Omit/empty for
   * widgets that can't sensibly become a chart (e.g. stat cards) — no
   * picker is shown for those. */
  supportedVizTypes?: VizType[];
  /** Used before the user has ever picked one explicitly. */
  defaultVizType?: VizType;
}
