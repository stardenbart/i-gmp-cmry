import axios from "axios";
import { useAuthStore } from "@/stores/authStore";
import { decryptApiResponseData } from "@/lib/crypto";

const baseURL = process.env.NEXT_PUBLIC_API_URL || "/api/v1";

export const api = axios.create({
  baseURL,
  headers: {
    "Content-Type": "application/json",
  },
});

// Interceptor to attach token
api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token;
    if (config.headers) {
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      config.headers["Cache-Control"] = "no-cache, no-store, must-revalidate";
      config.headers["Pragma"] = "no-cache";
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Interceptor to handle 401 Unauthorized globally & auto-decrypt encrypted API payloads
api.interceptors.response.use(
  async (response) => {
    if (response.data) {
      response.data = await decryptApiResponseData(response.data);
    }
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      // Clear token and redirect to login
      useAuthStore.getState().logout();
      if (typeof window !== "undefined") {
        window.location.href = "/login";
      }
    }
    return Promise.reject(error);
  }
);
