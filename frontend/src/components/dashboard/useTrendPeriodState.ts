"use client";

import { useCallback, useEffect, useState } from "react";
import type { TrendMode, TrendGranularity } from "@/lib/api/dashboard.api";

const validModes = new Set<TrendMode>(["range", "quarter"]);
const validGranularities = new Set<TrendGranularity>(["day", "week", "month", "year"]);

/** Which top-level mode is active: the date-range picker, or the fixed
 * "8 Kuartal Terakhir" preset. Persists to the URL so a shared/reloaded
 * dashboard link keeps the same chart view. */
export function useTrendModeState(defaultMode: TrendMode = "range") {
  const [mode, setModeState] = useState<TrendMode>(defaultMode);

  useEffect(() => {
    const queryMode = new URLSearchParams(window.location.search).get("trend_mode") as TrendMode | null;
    if (queryMode && validModes.has(queryMode)) {
      const timer = window.setTimeout(() => setModeState(queryMode), 0);
      return () => window.clearTimeout(timer);
    }
  }, []);

  const setMode = useCallback((nextMode: TrendMode) => {
    setModeState(nextMode);
    const url = new URL(window.location.href);
    url.searchParams.set("trend_mode", nextMode);
    window.history.replaceState(window.history.state, "", `${url.pathname}${url.search}${url.hash}`);
  }, []);

  return [mode, setMode] as const;
}

export interface TrendDateRange {
  start: string; // YYYY-MM-DD
  end: string; // YYYY-MM-DD
}

function formatDate(date: Date): string {
  return date.toISOString().slice(0, 10);
}

function defaultTrendRange(): TrendDateRange {
  const end = new Date();
  const start = new Date();
  start.setDate(start.getDate() - 29); // last 30 days, matches the "day" default granularity
  return { start: formatDate(start), end: formatDate(end) };
}

/** Backs the date-range picker (Dari/Sampai) for "range" mode.
 * Empty values are normalized in the setter so consumers never observe an
 * invalid range and no corrective render/effect cycle is needed. */
export function useTrendRangeState() {
  const [range, setRangeState] = useState<TrendDateRange>(defaultTrendRange);

  const setRange = useCallback((next: TrendDateRange) => {
    const fallback = defaultTrendRange();
    setRangeState({
      start: next.start || fallback.start,
      end: next.end || fallback.end,
    });
  }, []);

  return [range, setRange] as const;
}

/** Backs the granularity dropdown (Harian/Mingguan/Bulanan/Tahunan) that
 * decides how the picked date range is bucketed. */
export function useTrendGranularityState(defaultGranularity: TrendGranularity = "day") {
  const [granularity, setGranularityState] = useState<TrendGranularity>(defaultGranularity);

  useEffect(() => {
    const queryGranularity = new URLSearchParams(window.location.search).get("trend_granularity") as TrendGranularity | null;
    if (queryGranularity && validGranularities.has(queryGranularity)) {
      const timer = window.setTimeout(() => setGranularityState(queryGranularity), 0);
      return () => window.clearTimeout(timer);
    }
  }, []);

  const setGranularity = useCallback((next: TrendGranularity) => {
    setGranularityState(next);
    const url = new URL(window.location.href);
    url.searchParams.set("trend_granularity", next);
    window.history.replaceState(window.history.state, "", `${url.pathname}${url.search}${url.hash}`);
  }, []);

  return [granularity, setGranularity] as const;
}
