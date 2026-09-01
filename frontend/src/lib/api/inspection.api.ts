import { api } from "./axios";
import { getApiErrorStatus } from "./error";

export type InspectionStatus = "Draft" | "Ongoing" | "Completed" | "Approved";

export interface InspectionHeader {
  inspection_id: string;
  area_id: string;
  kawasan_id: string;
  detail_kawasan_id: string;
  inspector_id: string;
  status: InspectionStatus;
  created_at: string;
  updated_at: string;
  area_name?: string;
  kawasan_name?: string;
  detail_kawasan_name?: string;
  inspector_name?: string;
  score?: number;
}

export interface InspectionResult {
  result_id: string;
  inspection_id: string;
  uraian_id: string;
  checking: "OK" | "NG" | "NA";
  nilai: number;
  keterangan: string;
  created_at: string;
  updated_at: string;
}

export interface ChecklistPhoto {
  issue_photo_id: string;
  image_url: string;
  keterangan?: string;
  hei_id?: string;
  hei_category?: string;
}

export interface ChecklistUraian {
  uraian_id: string;
  uraian_name?: string;
  uraian_text?: string;
  standard_score?: number;
  checking?: "OK" | "NG" | "NA";
  nilai?: string | number;
  result?: {
    result_id?: string;
    checking?: "OK" | "NG" | "NA";
    photos?: ChecklistPhoto[];
  };
}

export interface ChecklistDetail {
  detail_id?: string;
  detail_name?: string;
  detail_aspek_name?: string;
  uraians?: ChecklistUraian[];
}

export interface ChecklistAspek {
  aspek_id: string;
  aspek_name: string;
  aspek_weight?: number;
  details?: ChecklistDetail[];
}

export interface InspectionChecklist {
  aspeks: ChecklistAspek[];
}

export interface BulkInspectionResult {
  uraian_id: string;
  checking: "OK" | "NG";
  nilai: number;
  keterangan: string;
}

// Boundaries of the currently-running "inspection period" (the recurring
// monthly cycle a DetailKawasan must be re-inspected within) for the plant
// that owns a given Area — see backend ResolveInspectionPeriod. cutoff_day=1
// is a plain calendar month; period_end is exclusive.
export interface InspectionPeriodInfo {
  cutoff_day: number;
  period_start: string;
  period_end: string;
}

export const inspectionApi = {
  getAll: async (params?: { page?: number; limit?: number; status?: string; inspector_id?: string }) => {
    const res = await api.get("/inspections", { params });
    return res.data.data;
  },
  
  getById: async (id: string): Promise<{ data: InspectionHeader & { plant_id?: string } }> => {
    const res = await api.get(`/inspections/${id}`);
    return res.data;
  },
  
  create: async (data: { area_id: string; kawasan_id: string; detail_kawasan_id: string }) => {
    const res = await api.post("/inspections", data);
    return res.data;
  },

  getChecklist: async (id: string): Promise<{ data: InspectionChecklist }> => {
    const res = await api.get(`/inspections/${id}/checklist`);
    return res.data;
  },

  bulkSaveResults: async (inspectionId: string, results: BulkInspectionResult[]) => {
    const res = await api.post(`/inspections/${inspectionId}/results/bulk`, {
      inspection_id: inspectionId,
      results
    });
    return res.data;
  },

  updateStatus: async (id: string, status: InspectionStatus) => {
    const res = await api.put(`/inspections/${id}/status`, { status });
    return res.data;
  },

  delete: async (id: string) => {
    const res = await api.delete(`/inspections/${id}`);
    return res.data;
  },

  getAnalyticsTrend: async (context_id: string, year?: number) => {
    const res = await api.get("/analytics/inspections-trend", { params: { context_id, year } });
    return res.data;
  },

  // Lets the client check "is this DetailKawasan done for the current
  // cycle" without re-deriving the admin-configurable cutoff-day math
  // itself — that duplication is exactly what caused the old client-side
  // calendar-month check to drift from the backend’s actual gate.
  getCurrentPeriodInfo: async (areaId: string): Promise<InspectionPeriodInfo> => {
    const res = await api.get("/inspections/period-info", { params: { area_id: areaId } });
    return res.data.data;
  },

  // Distributed Inspection Redis APIs
  acquireLock: async (scopeId: string, aspekId: string) => {
    const res = await api.post(`/inspeksi/${scopeId}/${aspekId}/lock`);
    return res.data;
  },

  releaseLock: async (scopeId: string, aspekId: string, lockToken: string) => {
    try {
      const res = await api.delete(`/inspeksi/${scopeId}/${aspekId}/lock`, {
        headers: { "X-Lock-Token": lockToken },
      });
      return res.data;
    } catch (err) {
      const status = getApiErrorStatus(err);
      if (status === 403 || status === 404 || status === 409) {
        return { success: true };
      }
      throw err;
    }
  },

  saveAspekDraft: async (
    scopeId: string,
    aspekId: string,
    lockToken: string,
    payload: {
      data: Record<string, unknown>;
      skor?: number;
      is_final?: boolean;
      session_id?: string;
    }
  ) => {
    const res = await api.put(`/inspeksi/${scopeId}/${aspekId}`, payload, {
      headers: { "X-Lock-Token": lockToken },
    });
    return res.data;
  },

  getDraftState: async (scopeId: string, aspekId: string) => {
    const res = await api.get(`/inspeksi/${scopeId}/${aspekId}/state`);
    return res.data;
  },

  getAllDrafts: async (scopeId: string) => {
    const res = await api.get(`/inspeksi/${scopeId}/drafts`);
    return res.data;
  },

  getKawasanStatus: async (scopeId: string) => {
    const res = await api.get(`/inspeksi/${scopeId}/status`);
    return res.data;
  }
};
