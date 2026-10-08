import { describe, expect, it } from "vitest";

import { formatTimeAgo } from "./date";

// 2026-10-08 10:00 WIB
const NOW = new Date("2026-10-08T03:00:00Z");

describe("formatTimeAgo", () => {
  it("shows 'Baru saja' for a timestamp the backend just wrote (+07:00)", () => {
    expect(formatTimeAgo("2026-10-08T10:00:00+07:00", NOW)).toBe("Baru saja");
  });

  it("treats small future clock skew as just now", () => {
    expect(formatTimeAgo("2026-10-08T10:00:40+07:00", NOW)).toBe("Baru saja");
  });

  it("counts minutes, hours and days", () => {
    expect(formatTimeAgo("2026-10-08T09:55:00+07:00", NOW)).toBe("5 menit lalu");
    expect(formatTimeAgo("2026-10-08T07:00:00+07:00", NOW)).toBe("3 jam lalu");
    expect(formatTimeAgo("2026-10-06T10:00:00+07:00", NOW)).toBe("2 hari lalu");
  });

  it("falls back to a WIB calendar date after a week", () => {
    // 30 Sep 20:00 UTC is already 1 Oct in Jakarta
    expect(formatTimeAgo("2026-09-30T20:00:00Z", NOW)).toBe("1 Okt");
  });

  it("accepts Date objects", () => {
    expect(formatTimeAgo(new Date("2026-10-08T02:30:00Z"), NOW)).toBe("30 menit lalu");
  });
});
