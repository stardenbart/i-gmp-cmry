import { api } from "./axios";

export type TrendPeriod =
  | "daily"
  | "weekly"
  | "monthly"
  | "quarter"
  | "previous_week"
  | "previous_month"
  | "previous_year";

export interface DashboardTrendPoint {
  bucket_start: string;
  bucket_end: string;
  label: string;
  total_inspections: number;
  total_issues: number;
  compliance_rate: number;
}

export interface DashboardTrendData {
  period: TrendPeriod;
  period_label: string;
  granularity: "day" | "week" | "month" | "quarter";
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
  period: TrendPeriod;
  plant_id?: string;
  area_id?: string;
  inspector_id?: string;
}

export const dashboardApi = {
  getTrend: async (params: DashboardTrendParams): Promise<DashboardTrendData> => {
    const response = await api.get("/dashboard/trend", { params });
    return response.data.data;
  },
};
