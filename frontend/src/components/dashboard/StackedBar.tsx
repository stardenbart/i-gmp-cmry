import { cn } from "@/lib/utils";

export interface StackedBarSegment {
  key: string;
  label: string;
  value: number;
  /** Background utility for the bar part and legend dot, e.g. "bg-success". */
  className: string;
}

function percent(value: number, total: number) {
  return total > 0 ? Math.round((value / total) * 1000) / 10 : 0;
}

/**
 * One 100% bar for parts of a single total (e.g. 3 completed + 1 running = 4
 * inspections), with a legend of counts and shares. Replaces separate
 * "3 / 4" + "1 / 4" progress bars, which read as two totals.
 */
export function StackedBar({ label, total, segments }: { label: string; total: number; segments: StackedBarSegment[] }) {
  const ariaLabel =
    total > 0
      ? `${label}: ${segments.map((s) => `${s.label} ${s.value} dari ${total}`).join(", ")}`
      : `${label}: belum ada data`;

  return (
    <div className="space-y-3">
      <div role="img" aria-label={ariaLabel} className="flex h-3 w-full gap-px overflow-hidden rounded-full bg-muted">
        {total > 0 &&
          segments
            .filter((s) => s.value > 0)
            .map((s) => (
              <div
                key={s.key}
                data-testid={`segment-${s.key}`}
                title={`${s.label}: ${s.value}`}
                className={cn("h-full transition-[width] duration-700", s.className)}
                style={{ width: `${percent(s.value, total)}%` }}
              />
            ))}
      </div>

      <ul className="grid grid-cols-2 gap-x-4 gap-y-2 text-xs">
        {segments.map((s) => (
          <li key={s.key} className="flex items-center justify-between gap-2">
            <span className="flex min-w-0 items-center gap-1.5 text-muted-foreground">
              <span aria-hidden="true" className={cn("h-2.5 w-2.5 shrink-0 rounded-sm", s.className)} />
              <span className="truncate">{s.label}</span>
            </span>
            <span className="whitespace-nowrap font-semibold text-foreground">
              {s.value} <span className="font-normal text-muted-foreground">({percent(s.value, total)}%)</span>
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}
