import { api } from "@/lib/api/axios";

export interface APIKey {
  key_id: string;
  name: string;
  prefix: string;
  created_by_id: string;
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
  is_single_use: boolean;
  created_at: string;
}

export const apiKeyApi = {
  list: async (): Promise<APIKey[]> => {
    const res = await api.get("/api-keys");
    return res.data.data;
  },

  create: async (name: string, isSingleUse: boolean = true): Promise<CreateAPIKeyResponse> => {
    const res = await api.post("/api-keys", { name, is_single_use: isSingleUse });
    return res.data.data;
  },

  revoke: async (id: string): Promise<void> => {
    await api.delete(`/api-keys/${id}`);
  },
};
