import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// Utility to set/remove auth cookie for Next.js middleware
export function setAuthCookie(token: string | null, userId?: string | null) {
  if (typeof document === "undefined") return;
  if (token) {
    document.cookie = `auth-token=${token}; path=/; max-age=86400; SameSite=Lax`;
    if (userId) {
      document.cookie = `user-id=${userId}; path=/; max-age=86400; SameSite=Lax`;
    }
  } else {
    document.cookie = "auth-token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";
    document.cookie = "user-id=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";
  }
}


export function formatTimeAgo(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / (1000 * 60));
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (diffMins < 1) return "Baru saja";
  if (diffMins < 60) return `${diffMins} menit lalu`;
  if (diffHours < 24) return `${diffHours} jam lalu`;
  if (diffDays < 7) return `${diffDays} hari lalu`;
  return date.toLocaleDateString("id-ID", { day: "numeric", month: "short" });
}