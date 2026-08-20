/**
 * Issue API Service
 * Type-safe API calls for issue management
 */

import { api } from "@/lib/api/axios";
import type {
  Issue,
  CreateIssueRequest,
  UpdateIssueRequest,
  IssuePhoto,
  UploadIssuePhotoRequest,
  PaginatedResponse,
  SingleItemResponse,
} from "./index";

/**
 * Issue API
 */
export const issueApi = {
  /**
   * Get all issues with pagination
   */
  getAll: async (
    page = 1,
    limit = 10,
    status?: string,
    picUserId?: string,
    search = ""
  ): Promise<PaginatedResponse<Issue>> => {
    const res = await api.get("/issues", {
      params: { page, limit, status, pic_user_id: picUserId, search },
    });
    return res.data;
  },

  /**
   * Get issue by ID with photos
   */
  getById: async (id: string): Promise<Issue> => {
    const res = await api.get<SingleItemResponse<Issue>>(`/issues/${id}`);
    return res.data.data;
  },

  /**
   * Create new issue
   */
  create: async (data: CreateIssueRequest): Promise<Issue> => {
    const res = await api.post<SingleItemResponse<Issue>>("/issues", data);
    return res.data.data;
  },

  /**
   * Update issue
   */
  update: async (id: string, data: UpdateIssueRequest): Promise<Issue> => {
    const res = await api.put<SingleItemResponse<Issue>>(`/issues/${id}`, data);
    return res.data.data;
  },

  /**
   * Delete issue
   */
  delete: async (id: string): Promise<void> => {
    await api.delete(`/issues/${id}`);
  },

  /**
   * Close issue
   */
  close: async (id: string): Promise<Issue> => {
    const res = await api.put<SingleItemResponse<Issue>>(`/issues/${id}/close`, {
      issue_status: "Closed",
    });
    return res.data.data;
  },

  /**
   * Verify issue
   */
  verify: async (id: string): Promise<Issue> => {
    const res = await api.put<SingleItemResponse<Issue>>(`/issues/${id}/verify`, {
      issue_status: "Verified",
    });
    return res.data.data;
  },

  /**
   * Reopen issue
   */
  reopen: async (id: string): Promise<Issue> => {
    const res = await api.put<SingleItemResponse<Issue>>(`/issues/${id}/reopen`, {
      issue_status: "Open",
    });
    return res.data.data;
  },

  /**
   * Assign PIC to issue
   */
  assignPic: async (id: string, picUserId: string): Promise<Issue> => {
    const res = await api.put<SingleItemResponse<Issue>>(`/issues/${id}/assign`, {
      issue_pic_user_id: picUserId,
    });
    return res.data.data;
  },

  /**
   * Update due date
   */
  updateDueDate: async (id: string, dueDate: string): Promise<Issue> => {
    const res = await api.put<SingleItemResponse<Issue>>(`/issues/${id}/due-date`, {
      due_date: dueDate,
    });
    return res.data.data;
  },
};

/**
 * Issue Photo API
 */
export const issuePhotoApi = {
  /**
   * Get photos for an issue
   */
  getByIssueId: async (issueId: string): Promise<IssuePhoto[]> => {
    const res = await api.get<PaginatedResponse<IssuePhoto>>(
      `/issues/${issueId}/photos`
    );
    return res.data.data.items;
  },

  /**
   * Upload photo for issue
   */
  upload: async (
    issueId: string,
    file: File,
    data: Omit<UploadIssuePhotoRequest, "issue_id">
  ): Promise<IssuePhoto> => {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("photo_type", data.photo_type);
    if (data.follow_up_date) {
      formData.append("follow_up_date", data.follow_up_date);
    }
    if (data.jumlah_follow_up !== undefined) {
      formData.append("jumlah_follow_up", String(data.jumlah_follow_up));
    }

    const res = await api.post<SingleItemResponse<IssuePhoto>>(
      `/issues/${issueId}/photos`,
      formData,
      {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      }
    );
    return res.data.data;
  },

  /**
   * Delete photo
   */
  delete: async (issueId: string, photoId: string): Promise<void> => {
    await api.delete(`/issues/${issueId}/photos/${photoId}`);
  },
};
