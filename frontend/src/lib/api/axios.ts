import axios from "axios";
import { useAuthStore } from "@/stores/authStore";

const baseURL = process.env.NEXT_PUBLIC_API_URL || "/api/v1";

// A separate, un-intercepted client for the refresh call itself — reusing
// `api` would recurse into the response interceptor below on a failed
// refresh.
const refreshClient = axios.create({ baseURL, withCredentials: true });

export const api = axios.create({
  baseURL,
  // Session auth now rides on httpOnly cookies (access_token/refresh_token)
  // instead of a Bearer header pulled from localStorage — withCredentials
  // makes the browser attach and accept those cookies on cross-port requests
  // (the backend's CORS config allows exactly this origin with credentials).
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

function readCookie(name: string): string | null {
  if (typeof document === "undefined") return null;
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
  return match ? decodeURIComponent(match[1]) : null;
}

const CSRF_PROTECTED_METHODS = new Set(["post", "put", "patch", "delete"]);

// Double-submit CSRF: csrf_token is a non-httpOnly cookie set alongside the
// session cookies specifically so this same-origin code can read it and echo
// it back as a header. A cross-site page riding on the victim's cookies
// can't read this cookie value, so it can't forge a matching header — see
// backend/internal/middleware/csrf_middleware.go.
api.interceptors.request.use(
  (config) => {
    if (config.headers) {
      const method = (config.method || "get").toLowerCase();
      if (CSRF_PROTECTED_METHODS.has(method)) {
        const csrfToken = readCookie("csrf_token");
        if (csrfToken) {
          config.headers["X-CSRF-Token"] = csrfToken;
        }
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

let refreshInFlight: Promise<boolean> | null = null;

function refreshSession(): Promise<boolean> {
  if (!refreshInFlight) {
    refreshInFlight = refreshClient
      .post("/auth/refresh")
      .then(() => true)
      .catch(() => false)
      .finally(() => {
        refreshInFlight = null;
      });
  }
  return refreshInFlight;
}

// On a 401, try exactly one silent refresh (the access token cookie simply
// expired — a normal, expected event with a 15-minute access TTL) and retry
// the original request once. Only if the refresh itself fails (refresh
// token missing/expired/reused) do we treat this as a real logout.
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    const status = error.response?.status;
    const isAuthEndpoint =
      typeof originalRequest?.url === "string" &&
      (originalRequest.url.includes("/auth/login") || originalRequest.url.includes("/auth/refresh"));

    if (status === 401 && originalRequest && !originalRequest._retry && !isAuthEndpoint) {
      originalRequest._retry = true;
      const refreshed = await refreshSession();
      if (refreshed) {
        return api(originalRequest);
      }
    }

    if (status === 401) {
      useAuthStore.getState().logout();
      if (typeof window !== "undefined") {
        window.location.href = "/login";
      }
    }
    return Promise.reject(error);
  }
);
