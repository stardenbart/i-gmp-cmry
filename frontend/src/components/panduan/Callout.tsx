import { Info, AlertTriangle } from "lucide-react";
import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

export function Callout({ variant = "info", children }: { variant?: "info" | "warn"; children: ReactNode }) {
  const Icon = variant === "warn" ? AlertTriangle : Info;
  return (
    <div
      className={cn(
        "flex gap-3 rounded-2xl border p-4 text-sm",
        variant === "warn" ? "border-amber-500/30 bg-amber-500/10" : "border-primary/30 bg-primary/10"
      )}
    >
      <Icon aria-hidden="true" className={cn("mt-0.5 h-4 w-4 shrink-0", variant === "warn" ? "text-amber-500" : "text-primary")} />
      <div className="space-y-1 text-foreground [&_strong]:font-semibold">{children}</div>
    </div>
  );
}
