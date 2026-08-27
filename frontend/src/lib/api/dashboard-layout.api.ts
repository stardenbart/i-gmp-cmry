import { api } from "./axios";
import type { SingleItemResponse } from "@/types/api/types";

export interface WidgetConfig {
  widget_id: string;
  visible: boolean;
  order: number;
  // Reserved for Fase 3b (react-grid-layout drag/resize) — absent today.
  x?: number;
  y?: number;
  w?: number;
  h?: number;
}

export const dashboardLayoutApi = {
  get: async (): Promise<WidgetConfig[]> => {
    const res = await api.get<SingleItemResponse<WidgetConfig[]>>("/dashboard/layout");
    return res.data.data;
  },

  save: async (widgets: WidgetConfig[]): Promise<WidgetConfig[]> => {
    const res = await api.put<SingleItemResponse<WidgetConfig[]>>("/dashboard/layout", widgets);
    return res.data.data;
  },
};
