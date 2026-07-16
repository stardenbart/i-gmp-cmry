import { api } from "./axios";

export type IssueStatus = "Open" | "InProgress" | "Closed" | "Verified";
export type PhotoType = "Initial" | "FollowUp";

export interface IssuePhoto {
  issue_photo_id: string;
  issue_id: string;
  pic_user_id: string;
  photo_type: PhotoType;
  image_url: string;
  file_name: string;
  follow_up_date?: string;
  jumlah_follow_up?: number;
  created_at: string;
  updated_at: string;
}

export interface Issue {
  issue_id: string;
  result_id: string;
  issue_pic_user_id: string;
  due_date?: string;
  issue_status: IssueStatus;
  keterangan: string;
  created_at: string;
  updated_at: string;
  photos?: IssuePhoto[];
}

export const issueApi = {
  getAll: async (params?: { page?: number; limit?: number; status?: string; pic_user_id?: string }) => {
    const res = await api.get("/issues", { params });
    return res.data;
  },

  getById: async (id: string) => {
    const res = await api.get(`/issues/${id}`);
    return res.data;
  },

  create: async (data: { result_id: string; issue_pic_user_id: string; due_date?: string; keterangan?: string }) => {
    const res = await api.post("/issues", data);
    return res.data;
  },

  update: async (id: string, data: { issue_pic_user_id?: string; due_date?: string; issue_status?: IssueStatus; keterangan?: string }) => {
    const res = await api.put(`/issues/${id}`, data);
    return res.data;
  },

  delete: async (id: string) => {
    const res = await api.delete(`/issues/${id}`);
    return res.data;
  },

  // Photos
  getPhotos: async (issueId: string) => {
    const res = await api.get(`/issues/${issueId}/photos`);
    return res.data;
  },

  uploadPhoto: async (issueId: string, formData: FormData) => {
    const res = await api.post(`/issues/${issueId}/photos`, formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return res.data;
  },

  deletePhoto: async (issueId: string, photoId: string) => {
    const res = await api.delete(`/issues/${issueId}/photos/${photoId}`);
    return res.data;
  },
};
