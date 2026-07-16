import { api } from "./axios";

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

export const inspectionApi = {
  getAll: async (params?: { page?: number; limit?: number; status?: string }) => {
    const res = await api.get("/inspections", { params });
    return res.data;
  },
  
  getById: async (id: string) => {
    const res = await api.get(`/inspections/${id}`);
    return res.data;
  },
  
  create: async (data: { area_id: string; kawasan_id: string; detail_kawasan_id: string }) => {
    const res = await api.post("/inspections", data);
    return res.data;
  },

  updateStatus: async (id: string, status: InspectionStatus) => {
    const res = await api.patch(`/inspections/${id}/status`, { status });
    return res.data;
  }
};
