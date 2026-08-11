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

export function formatImageUrl(url: string | null | undefined): string {
  if (!url) return "";
  let formatted = url.trim();
  if (!formatted) return "";

  // Reject malformed paths that point to a directory/issue ID prefix instead of an actual image file
  const cleanPath = formatted.split('?')[0].split('#')[0];
  const lastSegment = cleanPath.substring(cleanPath.lastIndexOf('/') + 1);
  const isLikelyIssueFolder = /^ISS-[A-Za-z0-9_-]+$/i.test(lastSegment) || /^ISSUE-[A-Za-z0-9_-]+$/i.test(lastSegment);
  
  if (isLikelyIssueFolder && !cleanPath.match(/\.(jpg|jpeg|png|webp|gif|svg|bmp)$/i)) {
    return "";
  }

  // If it's a MinIO image URL containing /monitoring-audit-bucket/, convert to relative path for Next.js proxying
  const bucketIndex = formatted.indexOf("/monitoring-audit-bucket/");
  if (bucketIndex !== -1) {
    return formatted.substring(bucketIndex);
  }

  // Fallback if port 9000 is included in full URL
  const portIndex = formatted.indexOf(":9000/");
  if (portIndex !== -1) {
    const rawPath = formatted.substring(portIndex + 5).replace(/^\/+/, '');
    return rawPath.startsWith("monitoring-audit-bucket/") ? `/${rawPath}` : `/monitoring-audit-bucket/${rawPath}`;
  }

  // If minio internal hostname is used
  if (formatted.includes("minio:9000")) {
    const minioPath = formatted.substring(formatted.indexOf("minio:9000") + 10).replace(/^\/+/, '');
    return minioPath.startsWith("monitoring-audit-bucket/") ? `/${minioPath}` : `/monitoring-audit-bucket/${minioPath}`;
  }

  // Ensure full HTTP/HTTPS URLs are preserved for external browser image loading
  if (formatted.startsWith("http://") || formatted.startsWith("https://")) {
    return formatted;
  }

  // Local uploads folder
  if (formatted.startsWith("/uploads/") || formatted.startsWith("uploads/")) {
    return formatted.startsWith("/") ? formatted : `/${formatted}`;
  }

  // If already prefixed with /monitoring-audit-bucket/
  if (formatted.startsWith("/monitoring-audit-bucket/")) {
    return formatted;
  }

  // Default object key -> prefix with /monitoring-audit-bucket/
  const key = formatted.replace(/^\/+/, '');
  const encodedKey = key.includes('+') ? encodeURIComponent(key) : key;
  return `/monitoring-audit-bucket/${encodedKey}`;
}