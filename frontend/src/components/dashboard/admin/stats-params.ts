import type { DateRange } from "@/components/dashboard/DateRangeFilter";

/** Query params for GET /dashboard/stats; empty filters are left out. */
export function buildStatsParams({ areaId, plantId, range }: { areaId: string; plantId: string; range: DateRange }) {
  return {
    ...(areaId ? { area_id: areaId } : {}),
    ...(plantId ? { plant_id: plantId } : {}),
    include_trend: false,
    ...(range.start ? { start_date: range.start } : {}),
    ...(range.end ? { end_date: range.end } : {}),
  };
}
