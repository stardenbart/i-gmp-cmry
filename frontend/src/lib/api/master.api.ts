import { api } from "./axios";

export interface Plant {
  plant_id: string;
  plant_code: string;
  plant_name: string;
  address?: string;
  created_at?: string;
  updated_at?: string;
}

export interface Area {
  area_id: string;
  area_name: string;
  area_code?: string;
  plant_id?: string;
  plant?: Plant;
}

export interface Kawasan {
  kawasan_id: string;
  kawasan_name: string;
  kawasan_code?: string;
  area_id: string;
  last_inspection?: string;
}

export interface DetailKawasan {
  detail_kawasan_id: string;
  detail_kawasan_name: string;
  detail_kawasan_code?: string;
  kawasan_id: string;
  last_inspection?: string;
  active_inspection_status?: string; // "Ongoing" | "Draft" | "" — from live join
}

interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

interface ApiResponse<T> {
  success: boolean;
  status_code: number;
  message: string;
  data: PaginatedResponse<T>;
}

export const masterApi = {
  // Plants
  getPlants: async (params?: { page?: number; limit?: number; search?: string }): Promise<Plant[]> => {
    const res = await api.get<ApiResponse<Plant>>("/master/plants", { params });
    return res.data.data.items;
  },

  // Areas
  getAreas: async (params?: { page?: number; limit?: number; search?: string }): Promise<Area[]> => {
    const res = await api.get<ApiResponse<Area>>("/master/area", { params });
    return res.data.data.items;
  },

  // Kawasans (filtered by area_id)
  getKawasans: async (params?: { area_id?: string; page?: number; limit?: number; search?: string }): Promise<Kawasan[]> => {
    const res = await api.get<ApiResponse<Kawasan>>("/master/kawasan", { params });
    return res.data.data.items;
  },

  // Detail Kawasans (filtered by kawasan_id)
  getDetailKawasans: async (params?: { kawasan_id?: string; page?: number; limit?: number; search?: string }): Promise<DetailKawasan[]> => {
    const res = await api.get<ApiResponse<DetailKawasan>>("/master/detail-kawasan", { params });
    return res.data.data.items;
  },
};
