import type { ComponentType } from "react";

export interface WidgetGridPosition {
  x: number;
  y: number;
  w: number;
  h: number;
}

export interface WidgetDefinition {
  id: string;
  title: string;
  Component: ComponentType;
  /** Position/size used the first time a user opens this dashboard, before
   * they've ever dragged/resized anything (12-column grid). */
  defaultLayout: WidgetGridPosition;
}
