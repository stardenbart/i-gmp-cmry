import { api } from "@/lib/api/axios";

export interface APIKey {
  key_id: string;
  name: string;
  prefix: string;
  created_by_id: string;
  plant_id?: string;
  is_active: boolean;
  is_single_use: boolean;
  used_at?: string;
  expires_at?: string;
  created_at: string;
}

export interface CreateAPIKeyResponse {
  key_id: string;
  name: string;
  raw_token: string;
  prefix: string;
  plant_id?: string;
  is_single_use: boolean;
  created_at: string;
}

export const apiKeyApi = {
  list: async (plant_id?: string): Promise<APIKey[]> => {
    const params: Record<string, any> = {};
    if (plant_id && plant_id !== "ALL") params.plant_id = plant_id;
    const res = await api.get("/api-keys", { params });
    return res.data.data;
  },

  create: async (name: string, isSingleUse: boolean = true, plant_id?: string): Promise<CreateAPIKeyResponse> => {
    const payload: Record<string, any> = { name, is_single_use: isSingleUse };
    if (plant_id && plant_id !== "ALL") payload.plant_id = plant_id;
    const res = await api.post("/api-keys", payload);
    return res.data.data;
  },

  revoke: async (id: string): Promise<void> => {
    await api.delete(`/api-keys/${id}`);
  },
};
