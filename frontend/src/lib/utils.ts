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

export function isEncryptedBase64(str: string | null | undefined): boolean {
  if (!str) return false;
  const trimmed = str.trim();
  return (
    trimmed.length >= 30 &&
    !trimmed.includes(" ") &&
    !trimmed.includes("/") &&
    !trimmed.match(/\.(jpg|jpeg|png|webp|gif|svg|bmp)$/i) &&
    (trimmed.endsWith("==") || trimmed.endsWith("=") || trimmed.includes("+"))
  );
}

export function formatImageUrl(url: string | null | undefined): string {
  if (!url) return "";
  let formatted = url.trim();
  if (!formatted) return "";

  // If string is an undecrypted AES base64 ciphertext, do not send as raw MinIO URL
  if (isEncryptedBase64(formatted)) {
    return "";
  }

  // Reject malformed paths that point to a directory/issue ID prefix instead of an actual image file
  const cleanPath = formatted.split('?')[0].split('#')[0];
  const lastSegment = cleanPath.substring(cleanPath.lastIndexOf('/') + 1);
  const isLikelyIssueFolder = /^ISS-[A-Za-z0-9_-]+$/i.test(lastSegment) || /^ISSUE-[A-Za-z0-9_-]+$/i.test(lastSegment);
  
  if (isLikelyIssueFolder && !cleanPath.match(/\.(jpg|jpeg|png|webp|gif|svg|bmp)$/i)) {
    return "";
  }

  // Detect current hostname and protocol when running in browser
  let currentHost = "localhost";
  let protocol = "http:";
  if (typeof window !== "undefined") {
    currentHost = window.location.hostname;
    protocol = window.location.protocol;
  }

  const backendPort = "8080";
  const apiHost = process.env.NEXT_PUBLIC_API_URL || `${protocol}//${currentHost}:${backendPort}`;
  const cleanHost = apiHost.replace(/\/api\/v1\/?$/, "");

  // If MinIO URL with port 9000 or /monitoring-audit-bucket/uploads/, convert to backend static /uploads route
  if (formatted.includes(":9000/") || formatted.includes("/monitoring-audit-bucket/uploads/")) {
    const uploadPathIdx = formatted.indexOf("/uploads/");
    if (uploadPathIdx !== -1) {
      return `${cleanHost}${formatted.substring(uploadPathIdx)}`;
    }
  }

  // Replace localhost or 127.0.0.1 with current connected server hostname
  if (formatted.includes("localhost") || formatted.includes("127.0.0.1")) {
    formatted = formatted.replace(/localhost|127\.0\.0\.1/g, currentHost);
  }

  if (
    formatted.startsWith("http://") ||
    formatted.startsWith("https://") ||
    formatted.startsWith("data:") ||
    formatted.startsWith("blob:")
  ) {
    return formatted;
  }

  // Local uploads folder
  if (formatted.startsWith("/uploads/") || formatted.startsWith("uploads/")) {
    const cleanP = formatted.startsWith("/") ? formatted : `/${formatted}`;
    return `${cleanHost}${cleanP}`;
  }

  // Default object key -> prefix with /uploads/
  const key = formatted.replace(/^\/+/, '').replace(/^monitoring-audit-bucket\//, '');
  return `${cleanHost}/uploads/${key}`;
}