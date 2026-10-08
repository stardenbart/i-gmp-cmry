export const APP_TIME_ZONE = "Asia/Jakarta";

const MINUTE = 60 * 1000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

/**
 * Relative time in Indonesian ("5 menit lalu"). Future timestamps (clock skew
 * between server and device) read as "Baru saja"; anything older than a week
 * shows the WIB calendar date.
 */
export function formatTimeAgo(value: string | Date, now: Date = new Date()): string {
  const date = value instanceof Date ? value : new Date(value);
  const diffMs = now.getTime() - date.getTime();

  if (diffMs < MINUTE) return "Baru saja";
  if (diffMs < HOUR) return `${Math.floor(diffMs / MINUTE)} menit lalu`;
  if (diffMs < DAY) return `${Math.floor(diffMs / HOUR)} jam lalu`;
  if (diffMs < 7 * DAY) return `${Math.floor(diffMs / DAY)} hari lalu`;
  return date.toLocaleDateString("id-ID", {
    day: "numeric",
    month: "short",
    timeZone: APP_TIME_ZONE,
  });
}
