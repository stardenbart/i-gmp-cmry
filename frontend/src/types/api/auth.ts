/**
 * Auth API Service
 * Type-safe API calls for authentication
 */

import { api } from "@/lib/api/axios";
import type {
  LoginRequest,
  LoginResponse,
  User,
  ChangePasswordRequest,
  UpdateProfileRequest,
  SingleItemResponse,
  PaginatedResponse,
} from "./index";

// Alias for backward compatibility
export type { User };

/**
 * Auth API endpoints
 */
export const authApi = {
  /**
   * Login user
   */
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const res = await api.post<{ data: LoginResponse }>("/auth/login", data);
    return res.data.data;
  },

  /**
   * Get current user info
   */
  me: async (): Promise<User> => {
    const res = await api.get<SingleItemResponse<User>>("/auth/me");
    return res.data.data;
  },

  /**
   * Change password
   */
  changePassword: async (data: ChangePasswordRequest): Promise<void> => {
    await api.put("/auth/change-password", data);
  },

  /**
   * Logout user
   */
  logout: async (): Promise<void> => {
    await api.post("/auth/logout");
  },
};

/**
 * User Management API endpoints
 */
export const userApi = {
  /**
   * Get all users with pagination
   */
  getAll: async (
    page = 1,
    limit = 10,
    search = ""
  ): Promise<PaginatedResponse<User>> => {
    const res = await api.get("/users", {
      params: { page, limit, search },
    });
    return res.data;
  },

  /**
   * Get user by ID
   */
  getById: async (id: string): Promise<User> => {
    const res = await api.get<SingleItemResponse<User>>(`/users/${id}`);
    return res.data.data;
  },

  /**
   * Create new user
   */
  create: async (
    data: Omit<User, "user_id" | "created_at" | "updated_at">
  ): Promise<User> => {
    const res = await api.post<SingleItemResponse<User>>("/users", data);
    return res.data.data;
  },

  /**
   * Update user
   */
  update: async (id: string, data: Partial<User>): Promise<User> => {
    const res = await api.put<SingleItemResponse<User>>(`/users/${id}`, data);
    return res.data.data;
  },

  /**
   * Delete user
   */
  delete: async (id: string): Promise<void> => {
    await api.delete(`/users/${id}`);
  },

  /**
   * Update profile (self)
   */
  updateProfile: async (
    id: string,
    data: UpdateProfileRequest
  ): Promise<User> => {
    const res = await api.put<SingleItemResponse<User>>(
      `/users/${id}`,
      data
    );
    return res.data.data;
  },

  /**
   * Change password
   */
  changePassword: async (
    id: string,
    data: ChangePasswordRequest
  ): Promise<void> => {
    await api.put(`/users/${id}/change-password`, data);
  },
};
