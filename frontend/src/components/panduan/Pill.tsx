import { cn } from "@/lib/utils";
import type { PillStyle } from "./theme";

export function Pill({ style, className }: { style: PillStyle; className?: string }) {
  const Icon = style.icon;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 whitespace-nowrap rounded-full border px-2.5 py-1 text-xs font-semibold",
        style.bg,
        style.text,
        style.border,
        className
      )}
    >
      {Icon && <Icon className="h-3.5 w-3.5" aria-hidden="true" />}
      {style.label}
    </span>
  );
}
