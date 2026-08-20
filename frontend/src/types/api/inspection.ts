/**
 * Inspection API Service
 * Type-safe API calls for inspection management
 */

import { api } from "@/lib/api/axios";
import type {
  InspectionHeader,
  InspectionResult,
  CreateInspectionRequest,
  UpdateInspectionRequest,
  SubmitInspectionResultRequest,
  PaginatedResponse,
  SingleItemResponse,
} from "./index";

/**
 * Inspection API
 */
export const inspectionApi = {
  /**
   * Get all inspections with pagination
   */
  getAll: async (
    page = 1,
    limit = 10,
    status?: string,
    areaId?: string,
    search = ""
  ): Promise<PaginatedResponse<InspectionHeader>> => {
    const res = await api.get("/inspections", {
      params: { page, limit, status, area_id: areaId, search },
    });
    return res.data;
  },

  /**
   * Get inspection by ID with results
   */
  getById: async (id: string): Promise<InspectionHeader & { results: InspectionResult[] }> => {
    const res = await api.get<SingleItemResponse<InspectionHeader & { results: InspectionResult[] }>>(
      `/inspections/${id}`
    );
    return res.data.data;
  },

  /**
   * Create new inspection
   */
  create: async (data: CreateInspectionRequest): Promise<InspectionHeader> => {
    const res = await api.post<SingleItemResponse<InspectionHeader>>("/inspections", data);
    return res.data.data;
  },

  /**
   * Update inspection
   */
  update: async (
    id: string,
    data: UpdateInspectionRequest
  ): Promise<InspectionHeader> => {
    const res = await api.put<SingleItemResponse<InspectionHeader>>(
      `/inspections/${id}`,
      data
    );
    return res.data.data;
  },

  /**
   * Delete inspection
   */
  delete: async (id: string): Promise<void> => {
    await api.delete(`/inspections/${id}`);
  },

  /**
   * Start inspection (change status to in_progress)
   */
  start: async (id: string): Promise<InspectionHeader> => {
    const res = await api.put<SingleItemResponse<InspectionHeader>>(
      `/inspections/${id}/start`
    );
    return res.data.data;
  },

  /**
   * Complete inspection (change status to completed)
   */
  complete: async (id: string): Promise<InspectionHeader> => {
    const res = await api.put<SingleItemResponse<InspectionHeader>>(
      `/inspections/${id}/complete`
    );
    return res.data.data;
  },

  /**
   * Cancel inspection
   */
  cancel: async (id: string): Promise<InspectionHeader> => {
    const res = await api.put<SingleItemResponse<InspectionHeader>>(
      `/inspections/${id}/cancel`
    );
    return res.data.data;
  },
};

/**
 * Inspection Result API
 */
export const inspectionResultApi = {
  /**
   * Get results for an inspection
   */
  getByInspectionId: async (
    inspectionId: string
  ): Promise<InspectionResult[]> => {
    const res = await api.get<PaginatedResponse<InspectionResult>>(
      `/inspections/${inspectionId}/results`
    );
    return res.data.data.items;
  },

  /**
   * Submit/update inspection result
   */
  submit: async (
    inspectionId: string,
    data: SubmitInspectionResultRequest
  ): Promise<InspectionResult> => {
    const res = await api.post<SingleItemResponse<InspectionResult>>(
      `/inspections/${inspectionId}/results`,
      data
    );
    return res.data.data;
  },

  /**
   * Bulk submit results
   */
  bulkSubmit: async (
    inspectionId: string,
    results: SubmitInspectionResultRequest[]
  ): Promise<InspectionResult[]> => {
    const res = await api.post<PaginatedResponse<InspectionResult>>(
      `/inspections/${inspectionId}/results/bulk`,
      { results }
    );
    return res.data.data.items;
  },
};
