"use client";

import { X } from "lucide-react";

export interface DateRange {
  start: string;
  end: string;
}

const DATE_INPUT =
  "h-11 rounded-md border border-border bg-card px-2.5 text-sm text-foreground focus-visible:border-ring focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/15";

/** Overview date range (YYYY-MM-DD, inclusive). Empty = all time. */
export function DateRangeFilter({ value, onChange }: { value: DateRange; onChange: (range: DateRange) => void }) {
  const hasRange = Boolean(value.start || value.end);

  return (
    <div className="flex flex-wrap items-center gap-1.5">
      <input
        type="date"
        aria-label="Dari tanggal"
        title="Dari tanggal inspeksi"
        value={value.start}
        max={value.end || undefined}
        onChange={(e) => onChange({ ...value, start: e.target.value })}
        className={DATE_INPUT}
      />
      <span aria-hidden="true" className="text-xs text-muted-foreground">
        s.d.
      </span>
      <input
        type="date"
        aria-label="Sampai tanggal"
        title="Sampai tanggal inspeksi"
        value={value.end}
        min={value.start || undefined}
        onChange={(e) => onChange({ ...value, end: e.target.value })}
        className={DATE_INPUT}
      />
      {hasRange && (
        <button
          type="button"
          onClick={() => onChange({ start: "", end: "" })}
          className="inline-flex h-11 items-center gap-1 rounded-md px-2 text-xs font-medium text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <X aria-hidden="true" className="h-3.5 w-3.5" />
          Semua waktu
        </button>
      )}
    </div>
  );
}
