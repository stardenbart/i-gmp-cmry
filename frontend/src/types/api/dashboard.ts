/**
 * Dashboard API Service
 * Type-safe API calls for dashboard and statistics
 */

import { api } from "@/lib/api/axios";
import type { DashboardStats, SingleItemResponse } from "./index";

export interface GmpFollowUpEvidence {
  photo_id: string;
  image_url: string;
  keterangan?: string;
  follow_up_date?: string;
}

export interface PreviewExportRow {
  inspection_id?: string;
  area?: string;
  kawasan_id?: string;
  kawasan?: string;
  detail_kawasan_id?: string;
  detail_kawasan?: string;
  pic?: string;
  aspek?: string;
  detail?: string;
  uraian_id?: string;
  keterangan?: string;
  nilai?: number;
  total_nilai?: number;
  /** Σ Nilai ÷ Σ StandardScore for this row's Detail Kawasan × 100. */
  persentase_kepatuhan_detail_kawasan?: number;
  total_temuan?: number;
  image_url?: string;
  image_urls?: string[];
  follow_up_image_url?: string;
  follow_up_image_urls?: string[];
  follow_up_evidence?: GmpFollowUpEvidence[];
  issue_id?: string;
  due_date?: string;
  follow_up_date?: string;
  /** Calendar-day difference: follow-up date minus due date. */
  follow_up_gap_days?: number;
  [key: string]: string | number | boolean | string[] | GmpFollowUpEvidence[] | null | undefined;
}

/**
 * Dashboard API
 */
export const dashboardApi = {
  /**
   * Get dashboard statistics
   */
  getStats: async (): Promise<DashboardStats> => {
    const res = await api.get<SingleItemResponse<DashboardStats>>("/dashboard/stats");
    return res.data.data;
  },

  /**
   * Get recent activity
   */
  getRecentActivity: async (limit = 10) => {
    const res = await api.get("/dashboard/recent-activity", {
      params: { limit },
    });
    return res.data;
  },

  /**
   * Get preview export data (table format)
   */
  getPreviewExport: async (filters?: {
    area_id?: string;
    kawasan_id?: string;
    detail_kawasan_id?: string;
    start_date?: string;
    end_date?: string;
    plant_id?: string;
    q?: string;
  }): Promise<{ data: PreviewExportRow[] }> => {
    const res = await api.get<{ data: PreviewExportRow[] }>("/dashboard/preview-export", {
      params: filters,
    });
    return res.data;
  },

  /**
   * Get issues summary by status
   */
  getIssuesSummary: async () => {
    const res = await api.get("/dashboard/issues-summary");
    return res.data;
  },

  /**
   * Get inspections summary
   */
  getInspectionsSummary: async () => {
    const res = await api.get("/dashboard/inspections-summary");
    return res.data;
  },
};

/**
 * Activity Log API
 */
export const activityLogApi = {
  /**
   * Get all activity logs with pagination
   */
  getAll: async (
    page = 1,
    limit = 20,
    userId?: string,
    module?: string,
    startDate?: string,
    endDate?: string
  ) => {
    const res = await api.get("/activity-logs", {
      params: { page, limit, user_id: userId, module, start_date: startDate, end_date: endDate },
    });
    return res.data;
  },

  /**
   * Get activity log by ID
   */
  getById: async (id: string) => {
    const res = await api.get(`/activity-logs/${id}`);
    return res.data;
  },
};
