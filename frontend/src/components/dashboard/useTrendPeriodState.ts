"use client";

import { useCallback, useEffect, useState } from "react";

import { TREND_PERIODS } from "@/components/dashboard/TrendPeriodFilter";
import type { TrendPeriod } from "@/lib/api/dashboard.api";

const validPeriods = new Set<TrendPeriod>(TREND_PERIODS.map((period) => period.value));

export function useTrendPeriodState(defaultPeriod: TrendPeriod = "monthly") {
  const [period, setPeriodState] = useState<TrendPeriod>(defaultPeriod);

  useEffect(() => {
    const queryPeriod = new URLSearchParams(window.location.search).get("trend_period") as TrendPeriod | null;
    if (queryPeriod && validPeriods.has(queryPeriod)) {
      const timer = window.setTimeout(() => setPeriodState(queryPeriod), 0);
      return () => window.clearTimeout(timer);
    }
  }, []);

  const setPeriod = useCallback((nextPeriod: TrendPeriod) => {
    setPeriodState(nextPeriod);
    const url = new URL(window.location.href);
    url.searchParams.set("trend_period", nextPeriod);
    window.history.replaceState(window.history.state, "", `${url.pathname}${url.search}${url.hash}`);
  }, []);

  return [period, setPeriod] as const;
}
