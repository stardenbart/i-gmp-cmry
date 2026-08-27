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

export interface LayoutTarget {
  /** Defaults to the caller themself. Targeting another user is only
   * permitted for Admin/Super Admin (enforced server-side, 403 otherwise). */
  userId?: string;
  /** Needed alongside userId so the backend can resolve that user's default
   * layout if they've never customized one — the caller's own role isn't
   * necessarily the target's role. */
  roleId?: string;
}

export const dashboardLayoutApi = {
  get: async (target?: LayoutTarget): Promise<WidgetConfig[]> => {
    const res = await api.get<SingleItemResponse<WidgetConfig[]>>("/dashboard/layout", {
      params: { user_id: target?.userId, role_id: target?.roleId },
    });
    return res.data.data;
  },

  save: async (widgets: WidgetConfig[], target?: LayoutTarget): Promise<WidgetConfig[]> => {
    const res = await api.put<SingleItemResponse<WidgetConfig[]>>("/dashboard/layout", widgets, {
      params: { user_id: target?.userId, role_id: target?.roleId },
    });
    return res.data.data;
  },
};
