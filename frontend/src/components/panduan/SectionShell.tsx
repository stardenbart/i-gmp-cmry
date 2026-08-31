import type { ReactNode } from "react";

export function SectionShell({
  id,
  index,
  title,
  lede,
  role,
  children,
}: {
  id: string;
  index: number;
  title: string;
  lede: string;
  role?: ReactNode;
  children: ReactNode;
}) {
  return (
    <section id={id} className="scroll-mt-32 space-y-6 border-t border-border pt-10">
      <div className="space-y-2">
        <p className="font-mono text-xs uppercase tracking-wider text-muted-foreground/70">
          Bab {String(index).padStart(2, "0")}
        </p>
        <h2 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl">{title}</h2>
        <p className="max-w-2xl text-sm text-muted-foreground sm:text-base">{lede}</p>
        {role}
      </div>
      {children}
    </section>
  );
}
