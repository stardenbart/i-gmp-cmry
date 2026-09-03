"use client";

import { useLayoutEffect, type ReactNode } from "react";

function shouldRestoreDarkTheme() {
  let savedTheme = "dark";
  try {
    savedTheme = window.localStorage.getItem("theme") || "dark";
  } catch {
    // Keep the application default when storage is unavailable.
  }
  return savedTheme === "dark"
    || (savedTheme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches);
}

/**
 * Public KPI links always use a presentation-friendly light theme. The global
 * ThemeProvider may re-apply the viewer's saved dark theme after hydration, so
 * this boundary observes the root class while mounted and restores the saved
 * preference when the visitor leaves the public page.
 */
export function PublicKPIThemeBoundary({ children }: { children: ReactNode }) {
  useLayoutEffect(() => {
    const root = document.documentElement;
    const enforceLightTheme = () => {
      if (root.classList.contains("dark")) root.classList.remove("dark");
      root.dataset.publicKpiTheme = "light";
    };

    enforceLightTheme();
    const observer = new MutationObserver(enforceLightTheme);
    observer.observe(root, { attributes: true, attributeFilter: ["class"] });

    return () => {
      observer.disconnect();
      delete root.dataset.publicKpiTheme;
      root.classList.toggle("dark", shouldRestoreDarkTheme());
    };
  }, []);

  return <div className="public-kpi-light min-h-screen bg-white text-slate-950">{children}</div>;
}
