import type { ReactNode } from "react";

export function StepList({ steps }: { steps: ReactNode[] }) {
  return (
    <ol className="space-y-4">
      {steps.map((step, i) => (
        <li key={i} className="flex gap-3">
          <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-semibold text-primary">
            {i + 1}
          </span>
          <div className="pt-0.5 text-sm text-muted-foreground [&_strong]:font-semibold [&_strong]:text-foreground">{step}</div>
        </li>
      ))}
    </ol>
  );
}
