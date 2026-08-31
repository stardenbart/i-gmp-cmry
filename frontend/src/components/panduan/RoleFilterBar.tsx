"use client";

import { cn } from "@/lib/utils";
import { usePanduanFilter } from "./PanduanFilterContext";
import type { PanduanRole } from "./theme";

const OPTIONS: { value: PanduanRole; label: string }[] = [
  { value: "all", label: "Semua" },
  { value: "admin", label: "Saya Admin" },
  { value: "auditor", label: "Saya Auditor" },
  { value: "auditee", label: "Saya Auditee/PIC" },
];

export function RoleFilterBar() {
  const { filter, setFilter } = usePanduanFilter();
  return (
    <div className="flex flex-wrap gap-2" role="group" aria-label="Saring panduan berdasarkan peran">
      {OPTIONS.map((opt) => (
        <button
          key={opt.value}
          type="button"
          onClick={() => setFilter(opt.value)}
          aria-pressed={filter === opt.value}
          className={cn(
            "rounded-full border px-3.5 py-1.5 text-xs font-semibold transition-colors",
            filter === opt.value
              ? "border-primary bg-primary text-primary-foreground"
              : "border-border bg-card text-muted-foreground hover:text-foreground"
          )}
        >
          {opt.label}
        </button>
      ))}
    </div>
  );
}
