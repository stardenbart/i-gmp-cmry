import { api } from "./axios";

/**
 * "quarter" is a fixed rolling-window preset (8 most recent calendar
 * quarters), kept standalone and deliberately NOT part of the date-range
 * picker. "range" is the everyday mode: always a start_date/end_date, with
 * TrendGranularity controlling how that range is bucketed for the chart.
 */
export type TrendMode = "range" | "quarter";
export type TrendGranularity = "day" | "week" | "month" | "year";

export interface DashboardTrendPoint {
  bucket_start: string;
  bucket_end: string;
  label: string;
  total_inspections: number;
  total_issues: number;
  compliance_rate: number;
}

export interface DashboardTrendData {
  period: TrendMode;
  period_label: string;
  granularity: TrendGranularity | "quarter";
  timezone: string;
  range_start: string;
  range_end: string;
  inspector_id?: string;
  summary: {
    total_inspections: number;
    total_issues: number;
    average_compliance: number;
  };
  points: DashboardTrendPoint[];
}

export interface DashboardTrendParams {
  period: TrendMode;
  plant_id?: string;
  area_id?: string;
  inspector_id?: string;
  /** Required when period is "range" (format YYYY-MM-DD). */
  start_date?: string;
  end_date?: string;
  /** Required when period is "range". */
  granularity?: TrendGranularity;
}

export const dashboardApi = {
  getTrend: async (params: DashboardTrendParams): Promise<DashboardTrendData> => {
    const response = await api.get("/dashboard/trend", { params });
    return response.data.data;
  },
};
