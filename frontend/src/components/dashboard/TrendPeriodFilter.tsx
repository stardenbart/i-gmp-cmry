import { cn } from "@/lib/utils";
import type { TrendPeriod } from "@/lib/api/dashboard.api";

export const TREND_PERIODS: ReadonlyArray<{ value: TrendPeriod; label: string }> = [
  { value: "daily", label: "Harian" },
  { value: "weekly", label: "Mingguan" },
  { value: "monthly", label: "Bulanan" },
  { value: "quarter", label: "Kuartal" },
  { value: "previous_week", label: "Minggu Lalu" },
  { value: "previous_month", label: "Bulan Lalu" },
  { value: "previous_year", label: "Tahun Lalu" },
];

interface TrendPeriodFilterProps {
  value: TrendPeriod;
  onChange: (value: TrendPeriod) => void;
  disabled?: boolean;
  className?: string;
}

export function TrendPeriodFilter({ value, onChange, disabled, className }: TrendPeriodFilterProps) {
  return (
    <div className={cn("min-w-0", className)}>
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
  );
}
