import { cn } from "@/lib/utils";
import type { TrendMode, TrendGranularity } from "@/lib/api/dashboard.api";
import type { TrendDateRange } from "@/components/dashboard/useTrendPeriodState";

const MODE_OPTIONS: ReadonlyArray<{ value: TrendMode; label: string }> = [
  { value: "range", label: "Rentang Tanggal" },
  { value: "quarter", label: "Kuartal" },
];

const GRANULARITY_OPTIONS: ReadonlyArray<{ value: TrendGranularity; label: string }> = [
  { value: "day", label: "Harian" },
  { value: "week", label: "Mingguan" },
  { value: "month", label: "Bulanan" },
  { value: "year", label: "Tahunan" },
];

interface TrendPeriodFilterProps {
  mode: TrendMode;
  onModeChange: (mode: TrendMode) => void;
  range: TrendDateRange;
  onRangeChange: (range: TrendDateRange) => void;
  granularity: TrendGranularity;
  onGranularityChange: (granularity: TrendGranularity) => void;
  disabled?: boolean;
  className?: string;
}

export function TrendPeriodFilter({
  mode,
  onModeChange,
  range,
  onRangeChange,
  granularity,
  onGranularityChange,
  disabled,
  className,
}: TrendPeriodFilterProps) {
  const today = new Date().toISOString().slice(0, 10);

  return (
    <div className={cn("flex flex-col items-end gap-2 min-w-0", className)}>
      {/* Kuartal tetap terpisah dari filter kalender — bukan bagian dari rentang tanggal. */}
      <div className="hidden flex-wrap items-center gap-1 rounded-xl border border-border/60 bg-muted p-1 sm:flex" role="group" aria-label="Mode filter grafik">
        {MODE_OPTIONS.map((option) => (
          <button
            key={option.value}
            type="button"
            disabled={disabled}
            aria-pressed={mode === option.value}
            onClick={() => onModeChange(option.value)}
            className={cn(
              "rounded-lg px-2.5 py-1.5 text-[11px] font-semibold transition-all disabled:opacity-60 lg:px-3 lg:text-xs",
              mode === option.value
                ? "border border-border/80 bg-background text-foreground shadow-2xs"
                : "text-muted-foreground hover:bg-background/60 hover:text-foreground"
            )}
          >
            {option.label}
          </button>
        ))}
      </div>
      <select
        aria-label="Mode filter grafik"
        value={mode}
        disabled={disabled}
        onChange={(event) => onModeChange(event.target.value as TrendMode)}
        className="h-10 w-full rounded-xl border border-border bg-card px-3 text-xs font-semibold text-foreground outline-none focus:ring-2 focus:ring-primary/40 disabled:opacity-60 sm:hidden"
      >
        {MODE_OPTIONS.map((option) => (
          <option key={option.value} value={option.value}>{option.label}</option>
        ))}
      </select>

      {mode === "range" && (
        <div className="flex flex-wrap items-center justify-end gap-2">
          <label className="flex items-center gap-1.5 text-[11px] font-medium text-muted-foreground">
            Dari
            <input
              type="date"
              value={range.start}
              max={range.end}
              disabled={disabled}
              onChange={(event) => onRangeChange({ ...range, start: event.target.value })}
              className="h-9 rounded-lg border border-border bg-card px-2 text-xs text-foreground outline-none focus:ring-2 focus:ring-primary/40 disabled:opacity-60"
            />
          </label>
          <label className="flex items-center gap-1.5 text-[11px] font-medium text-muted-foreground">
            Sampai
            <input
              type="date"
              value={range.end}
              min={range.start}
              max={today}
              disabled={disabled}
              onChange={(event) => onRangeChange({ ...range, end: event.target.value })}
              className="h-9 rounded-lg border border-border bg-card px-2 text-xs text-foreground outline-none focus:ring-2 focus:ring-primary/40 disabled:opacity-60"
            />
          </label>
          <label className="flex items-center gap-1.5 text-[11px] font-medium text-muted-foreground">
            Kelompokkan per
            <select
              value={granularity}
              disabled={disabled}
              onChange={(event) => onGranularityChange(event.target.value as TrendGranularity)}
              className="h-9 rounded-lg border border-border bg-card px-2 text-xs font-semibold text-foreground outline-none focus:ring-2 focus:ring-primary/40 disabled:opacity-60"
            >
              {GRANULARITY_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>{option.label}</option>
              ))}
            </select>
          </label>
        </div>
      )}
    </div>
  );
}
