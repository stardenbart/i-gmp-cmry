import { api } from "./axios";

export type IssueStatus = "Open" | "InProgress" | "PendingValidation" | "Closed" | "Verified" | "OpenOverdue" | "ClosedOverdue" | "Overdue";
export type WOWRStatus = "None" | "PendingValidation" | "Verified" | "Rejected";
export type PhotoType = "Initial" | "FollowUp" | "WOWR";

export interface IssuePhoto {
  issue_photo_id: string;
  issue_id: string;
  pic_user_id: string;
  photo_type: PhotoType;
  image_url: string;
  file_name: string;
  keterangan?: string;
  follow_up_date?: string;
  jumlah_follow_up?: number;
  created_at: string;
  updated_at: string;
}

export interface Issue {
  issue_id: string;
  result_id: string;
  issue_pic_user_id: string;
  pic_name?: string;
  due_date?: string;
  issue_status: IssueStatus;
  computed_status?: IssueStatus;
  follow_up_delay?: number;
  keterangan: string;
  needs_wo_wr?: boolean;
  wo_id?: string;
  wr_id?: string;
  wowr_status?: WOWRStatus;
  created_at: string;
  updated_at: string;
  photos?: IssuePhoto[];
  area_name?: string;
  kawasan_name?: string;
  detail_kawasan_name?: string;
}

export const issueApi = {
  getAll: async (params?: { page?: number; limit?: number; status?: string; pic_user_id?: string; needs_wo_wr?: boolean }) => {
    const res = await api.get("/issues", { params });
    return res.data.data;
  },

  getById: async (id: string) => {
    const res = await api.get(`/issues/${id}`);
    return res.data;
  },

  create: async (data: { result_id: string; issue_pic_user_id: string; due_date?: string; keterangan?: string }) => {
    const res = await api.post("/issues", data);
    return res.data;
  },

  update: async (id: string, data: { issue_pic_user_id?: string; due_date?: string; issue_status?: IssueStatus; wowr_status?: WOWRStatus; keterangan?: string; needs_wo_wr?: boolean; wo_id?: string; wr_id?: string }) => {
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
    const res = await api.post(`/issues/${issueId}/photos/upload`, formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return res.data;
  },

  uploadPhotoChunked: async (
    issueId: string,
    file: File,
    onProgress?: (progress: number) => void
  ) => {
    const CHUNK_SIZE = 2 * 1024 * 1024; // 2MB chunks
    const totalChunks = Math.ceil(file.size / CHUNK_SIZE);
    
    // Generate a unique file ID for this upload session
    const fileId = `${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
    
    let lastResponse = null;

    for (let i = 0; i < totalChunks; i++) {
      const start = i * CHUNK_SIZE;
      const end = Math.min(file.size, start + CHUNK_SIZE);
      const chunk = file.slice(start, end);

      const formData = new FormData();
      formData.append("photo", chunk, file.name);
      formData.append("chunk_index", i.toString());
      formData.append("total_chunks", totalChunks.toString());
      formData.append("file_id", fileId);
      // Optional: Add photo_type if needed for the backend logic (can default to "Issue")
      formData.append("photo_type", "Issue");

      const res = await api.post(`/issues/${issueId}/photos/upload`, formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      
      lastResponse = res.data;
      
      if (onProgress) {
        onProgress(Math.round(((i + 1) / totalChunks) * 100));
      }
    }
    
    return lastResponse;
  },

  updatePhoto: async (photoId: string, keterangan: string) => {
    const res = await api.put(`/issues/photos/${photoId}`, { keterangan });
    return res.data;
  },

  deletePhoto: async (issueId: string, photoId: string) => {
    const res = await api.delete(`/issues/photos/${photoId}`);
    return res.data;
  },
};
