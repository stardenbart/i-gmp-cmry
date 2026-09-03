import axios from "axios";
import { api } from "./axios";
import type { AnalyticsCatalog, AnalyticsQueryResult } from "./analytics.api";
import { deserializeWidgetConfigs, type WidgetConfig } from "./dashboard-layout.api";
import type { SingleItemResponse } from "@/types/api/types";

const publicApi = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "/api/v1",
  headers: { "Content-Type": "application/json" },
});

export interface KPIShareFilter {
  period: "range" | "quarter";
  start_date?: string;
  end_date?: string;
  granularity?: "day" | "week" | "month" | "year";
}

export interface KPIShareRecord {
  share_id: string;
  share_name: string;
  public_title: string;
  token_prefix: string;
  plant_id: string;
  plant_name: string;
  allow_period_change: boolean;
  expires_at?: string;
  revoked_at?: string;
  last_accessed_at?: string;
  access_count: number;
  created_at: string;
}

export interface CreatedKPIShare extends KPIShareRecord {
  raw_token: string;
}

export interface PublicKPIBootstrap {
  share_id: string;
  public_title: string;
  plant_id: string;
  plant_name: string;
  layout: WidgetConfig[];
  filter: KPIShareFilter;
  allow_period_change: boolean;
  expires_at?: string;
  updated_at: string;
  catalog: AnalyticsCatalog;
}

interface WireBootstrap extends Omit<PublicKPIBootstrap, "layout"> {
  layout: Parameters<typeof deserializeWidgetConfigs>[0];
}

export const kpiShareApi = {
  create: async (payload: {
    share_name: string;
    public_title: string;
    plant_id: string;
    filter: KPIShareFilter;
    allow_period_change: boolean;
    expires_at: string;
  }): Promise<CreatedKPIShare> => {
    const res = await api.post<SingleItemResponse<CreatedKPIShare>>("/dashboard/kpi/shares", payload);
    return res.data.data;
  },

  list: async (): Promise<KPIShareRecord[]> => {
    const res = await api.get<SingleItemResponse<KPIShareRecord[]>>("/dashboard/kpi/shares");
    return res.data.data;
  },

  revoke: async (shareId: string): Promise<void> => {
    await api.delete(`/dashboard/kpi/shares/${encodeURIComponent(shareId)}`);
  },

  rotate: async (shareId: string): Promise<CreatedKPIShare> => {
    const res = await api.post<SingleItemResponse<CreatedKPIShare>>(`/dashboard/kpi/shares/${encodeURIComponent(shareId)}/rotate`);
    return res.data.data;
  },

  bootstrap: async (token: string): Promise<PublicKPIBootstrap> => {
    const res = await publicApi.get<SingleItemResponse<WireBootstrap>>(`/public/kpi/${encodeURIComponent(token)}/bootstrap`);
    return { ...res.data.data, layout: deserializeWidgetConfigs(res.data.data.layout) };
  },

  runCustomQuery: async (
    token: string,
    widgetId: string,
    drillLevel: number,
    signal?: AbortSignal,
  ): Promise<AnalyticsQueryResult> => {
    const res = await publicApi.post<SingleItemResponse<AnalyticsQueryResult>>(
      `/public/kpi/${encodeURIComponent(token)}/query`,
      { widget_id: widgetId, drill_level: drillLevel },
      { signal },
    );
    return res.data.data;
  },
};
