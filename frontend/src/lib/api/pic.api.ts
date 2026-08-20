import { api } from "./axios";

export interface PICMapping {
  pic_map_id: string;
  area_id: string;
  kawasan_id: string;
  user_id: string;
  kategori_pic: string;
}

export const picApi = {
  getAll: async (params?: { page?: number; limit?: number; area_id?: string; kawasan_id?: string }) => {
    const res = await api.get("/pic-mappings", { params });
    return res.data.data;
  },
};
