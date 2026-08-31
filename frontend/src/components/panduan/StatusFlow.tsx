import { ArrowRight, CornerDownLeft } from "lucide-react";
import { Pill } from "./Pill";
import type { PillStyle } from "./theme";

interface FlowStep {
  style: PillStyle;
  caption?: string;
}

export function StatusFlow({ steps, branches }: { steps: FlowStep[]; branches?: string[] }) {
  return (
    <div className="rounded-3xl border border-border bg-card p-5">
      <div className="flex flex-wrap items-start gap-x-2 gap-y-4">
        {steps.map((step, i) => (
          <div key={i} className="flex items-center gap-2">
            <div className="flex w-24 flex-col items-center gap-1.5 text-center">
              <Pill style={step.style} />
              {step.caption && <span className="text-label-sm leading-tight text-muted-foreground">{step.caption}</span>}
            </div>
            {i < steps.length - 1 && (
              <ArrowRight aria-hidden="true" className="mt-2 h-4 w-4 shrink-0 text-muted-foreground/50" />
            )}
          </div>
        ))}
      </div>
      {branches && branches.length > 0 && (
        <div className="mt-4 space-y-1.5 border-t border-border pt-3">
          {branches.map((b, i) => (
            <div key={i} className="flex items-start gap-2 text-xs text-muted-foreground">
              <CornerDownLeft aria-hidden="true" className="mt-0.5 h-3.5 w-3.5 shrink-0" />
              <span>{b}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
