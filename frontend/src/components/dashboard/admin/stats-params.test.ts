import { describe, expect, it } from "vitest";

import { buildStatsParams } from "./stats-params";

describe("buildStatsParams", () => {
  it("sends the overview date range with area and plant", () => {
    expect(buildStatsParams({ areaId: "AREA-1", plantId: "PLT-SENTUL", range: { start: "2026-10-01", end: "2026-10-08" } })).toEqual({
      area_id: "AREA-1",
      plant_id: "PLT-SENTUL",
      include_trend: false,
      start_date: "2026-10-01",
      end_date: "2026-10-08",
    });
  });

  it("omits empty filters so the backend counts all time", () => {
    expect(buildStatsParams({ areaId: "", plantId: "", range: { start: "", end: "" } })).toEqual({ include_trend: false });
  });

  it("allows an open-ended range", () => {
    expect(buildStatsParams({ areaId: "", plantId: "", range: { start: "2026-10-01", end: "" } })).toEqual({
      include_trend: false,
      start_date: "2026-10-01",
    });
  });
});
