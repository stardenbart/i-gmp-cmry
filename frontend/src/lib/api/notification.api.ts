import { api } from "./axios";

export interface Notification {
  id: string;
  user_id: string;
  type: "info" | "warning" | "error" | "success";
  title: string;
  message: string;
  is_read: boolean;
  link?: string;
  created_at: string;
}

export interface NotificationResponse {
  data: {
    items: Notification[];
    pagination: {
      total: number;
      page: number;
      limit: number;
      total_pages: number;
    };
  };
  message: string;
}

export const notificationApi = {
  getAll: async (page = 1, limit = 20): Promise<NotificationResponse> => {
    const res = await api.get("/notifications", { params: { page, limit } });
    return res.data;
  },

  markAsRead: async (id: string): Promise<void> => {
    await api.put(`/notifications/${id}/read`);
  },

  markAllAsRead: async (): Promise<void> => {
    await api.put("/notifications/read-all");
  },
};
