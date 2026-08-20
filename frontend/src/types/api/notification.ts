/**
 * Notification API Service
 * Type-safe API calls for notification management
 */

import { api } from "@/lib/api/axios";
import type {
  Notification,
  PaginatedNotifications,
  PaginatedResponse,
} from "./index";

/**
 * Notification API
 */
export const notificationApi = {
  /**
   * Get all notifications with pagination
   */
  getAll: async (
    page = 1,
    limit = 20
  ): Promise<PaginatedNotifications> => {
    const res = await api.get<PaginatedResponse<Notification>>("/notifications", {
      params: { page, limit },
    });
    return {
      items: res.data.data.items,
      pagination: res.data.data.pagination,
    };
  },

  /**
   * Get unread notification count
   */
  getUnreadCount: async (): Promise<{ count: number }> => {
    const res = await api.get("/notifications/unread-count");
    return res.data.data;
  },

  /**
   * Mark notification as read
   */
  markAsRead: async (id: string): Promise<void> => {
    await api.put(`/notifications/${id}/read`);
  },

  /**
   * Mark all notifications as read
   */
  markAllAsRead: async (): Promise<void> => {
    await api.put("/notifications/read-all");
  },

  /**
   * Delete notification
   */
  delete: async (id: string): Promise<void> => {
    await api.delete(`/notifications/${id}`);
  },

  /**
   * Clear all notifications
   */
  clearAll: async (): Promise<void> => {
    await api.delete("/notifications");
  },
};
