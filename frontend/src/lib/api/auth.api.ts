import { api } from "./axios";

export interface UserInfo {
  user_id: string;
  username: string;
  full_name: string;
  email: string;
  department_id: string;
  role_id: string;
  plant_id?: string;
  user_status: "Active" | "Inactive" | "Suspended";
}

export const authApi = {
  me: async (): Promise<{ data: UserInfo }> => {
    const res = await api.get("/auth/me");
    return res.data;
  },

  changePassword: async (data: {
    old_password: string;
    new_password: string;
  }) => {
    const res = await api.put("/auth/change-password", data);
    return res.data;
  },

  updateProfile: async (id: string, data: { full_name?: string; email?: string }) => {
    const res = await api.put(`/users/${id}`, data);
    return res.data;
  },

  login: async (data: { username: string; password: string }) => {
    const res = await api.post("/auth/login", data);
    return res.data;
  },

  logout: async () => {
    const res = await api.post("/auth/logout");
    return res.data;
  },
};
