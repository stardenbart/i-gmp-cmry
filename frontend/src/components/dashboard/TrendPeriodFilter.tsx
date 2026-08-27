import { cn } from "@/lib/utils";
import type { TrendPeriod } from "@/lib/api/dashboard.api";
import type { TrendDateRange } from "@/components/dashboard/useTrendPeriodState";

export const TREND_PERIODS: ReadonlyArray<{ value: TrendPeriod; label: string }> = [
  { value: "daily", label: "Harian" },
  { value: "weekly", label: "Mingguan" },
  { value: "monthly", label: "Bulanan" },
  { value: "quarter", label: "Kuartal" },
  { value: "custom", label: "Rentang Tanggal" },
];

interface TrendPeriodFilterProps {
  value: TrendPeriod;
  onChange: (value: TrendPeriod) => void;
  customRange: TrendDateRange;
  onCustomRangeChange: (range: TrendDateRange) => void;
  disabled?: boolean;
  className?: string;
}

export function TrendPeriodFilter({ value, onChange, customRange, onCustomRangeChange, disabled, className }: TrendPeriodFilterProps) {
  const isCustom = value === "custom";

  return (
    <div className={cn("flex flex-col items-end gap-2 min-w-0", className)}>
      <div className="min-w-0 w-full sm:w-auto">
        <label htmlFor="trend-period" className="sr-only">Pilih periode grafik</label>
        <select
          id="trend-period"
          value={value}
          disabled={disabled}
          onChange={(event) => onChange(event.target.value as TrendPeriod)}
          className="h-10 w-full rounded-xl border border-border bg-card px-3 text-xs font-semibold text-foreground outline-none focus:ring-2 focus:ring-primary/40 disabled:opacity-60 sm:hidden"
        >
          {TREND_PERIODS.map((period) => (
            <option key={period.value} value={period.value}>{period.label}</option>
          ))}
        </select>

        <div className="hidden flex-wrap items-center gap-1 rounded-xl border border-border/60 bg-muted p-1 sm:flex" role="group" aria-label="Filter periode grafik">
          {TREND_PERIODS.map((period) => (
            <button
              key={period.value}
              type="button"
              disabled={disabled}
              aria-pressed={value === period.value}
              onClick={() => onChange(period.value)}
              className={cn(
                "rounded-lg px-2.5 py-1.5 text-[11px] font-semibold transition-all disabled:opacity-60 lg:px-3 lg:text-xs",
                value === period.value
                  ? "border border-border/80 bg-background text-foreground shadow-2xs"
                  : "text-muted-foreground hover:bg-background/60 hover:text-foreground"
              )}
            >
              {period.label}
            </button>
          ))}
        </div>
      </div>

      {isCustom && (
        <div className="flex flex-wrap items-center gap-2">
          <label className="flex items-center gap-1.5 text-[11px] font-medium text-muted-foreground">
            Dari
            <input
              type="date"
              value={customRange.start}
              max={customRange.end}
              disabled={disabled}
              onChange={(event) => onCustomRangeChange({ ...customRange, start: event.target.value })}
              className="h-9 rounded-lg border border-border bg-card px-2 text-xs text-foreground outline-none focus:ring-2 focus:ring-primary/40 disabled:opacity-60"
            />
          </label>
          <label className="flex items-center gap-1.5 text-[11px] font-medium text-muted-foreground">
            Sampai
            <input
              type="date"
              value={customRange.end}
              min={customRange.start}
              max={new Date().toISOString().slice(0, 10)}
              disabled={disabled}
              onChange={(event) => onCustomRangeChange({ ...customRange, end: event.target.value })}
              className="h-9 rounded-lg border border-border bg-card px-2 text-xs text-foreground outline-none focus:ring-2 focus:ring-primary/40 disabled:opacity-60"
            />
          </label>
        </div>
      )}
    </div>
  );
}
