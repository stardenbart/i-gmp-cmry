export type VisualizationFieldKind = "measure" | "dimension";

export const KPI_FIELD_DRAG_MIME = "application/x-kpi-field";
export const KPI_CATEGORY_DRAG_MIME = "application/x-kpi-category";

export interface VisualizationFieldDragPayload {
  kind: VisualizationFieldKind;
  id: string;
}

export function readFieldDragPayload(dataTransfer: DataTransfer): VisualizationFieldDragPayload | null {
  const raw = dataTransfer.getData(KPI_FIELD_DRAG_MIME);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as Partial<VisualizationFieldDragPayload>;
    if ((parsed.kind === "measure" || parsed.kind === "dimension") && typeof parsed.id === "string" && parsed.id) {
      return { kind: parsed.kind, id: parsed.id };
    }
  } catch {
    // Malformed drag payloads are ignored instead of breaking the builder.
  }
  return null;
}
