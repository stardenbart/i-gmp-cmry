"use client";

import { useEffect, useState } from "react";
import { cn } from "@/lib/utils";

export interface ChapterItem {
  id: string;
  label: string;
}

export function ChapterTabs({ chapters }: { chapters: ChapterItem[] }) {
  const [active, setActive] = useState(chapters[0]?.id);

  useEffect(() => {
    const sections = chapters
      .map((c) => document.getElementById(c.id))
      .filter((el): el is HTMLElement => !!el);
    if (sections.length === 0) return;

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) setActive(entry.target.id);
        });
      },
      { rootMargin: "-20% 0px -70% 0px" }
    );
    sections.forEach((el) => observer.observe(el));
    return () => observer.disconnect();
  }, [chapters]);

  return (
    <nav className="-mx-1 flex gap-1.5 overflow-x-auto px-1 pb-1" aria-label="Navigasi bab panduan">
      {chapters.map((c, i) => (
        <a
          key={c.id}
          href={`#${c.id}`}
          className={cn(
            "shrink-0 rounded-full border px-3 py-1.5 text-xs font-medium transition-colors",
            active === c.id
              ? "border-primary bg-primary/10 text-primary"
              : "border-border bg-card text-muted-foreground hover:text-foreground"
          )}
        >
          <span className="mr-1 font-mono text-[10px] text-muted-foreground/70">{String(i).padStart(2, "0")}</span>
          {c.label}
        </a>
      ))}
    </nav>
  );
}
