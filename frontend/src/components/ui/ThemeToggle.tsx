"use client";

import { useTheme } from "next-themes";
import { useMounted } from "@/lib/useMounted";
import { Sun, Moon, Laptop } from "lucide-react";
import { cn } from "@/lib/utils";

export function ThemeToggle({ className }: { className?: string }) {
  const mounted = useMounted();
  const { theme, setTheme, resolvedTheme } = useTheme();

  if (!mounted) {
    return (
      <div className={cn("w-9 h-9 rounded-lg bg-muted/50 border border-border/50 animate-pulse", className)} />
    );
  }

  const cycleTheme = () => {
    if (theme === "system") {
      setTheme(resolvedTheme === "dark" ? "light" : "dark");
    } else if (theme === "dark") {
      setTheme("light");
    } else {
      setTheme("dark");
    }
  };

  const isDark = resolvedTheme === "dark";

  return (
    <button
      type="button"
      onClick={cycleTheme}
      title={
        theme === "system" 
          ? `Tema Sistem (${isDark ? "Gelap" : "Terang"}) - Klik untuk ganti` 
          : isDark 
          ? "Ganti ke Mode Terang" 
          : "Ganti ke Mode Gelap"
      }
      className={cn(
        "relative flex items-center justify-center w-9 h-9 rounded-lg border border-border/80 bg-background/80 hover:bg-muted text-foreground transition-all duration-200 active:scale-95 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring shadow-sm",
        className
      )}
      aria-label="Toggle Theme"
    >
      {/* Sun Icon (shown in Dark mode) */}
      <Sun 
        className={cn(
          "h-4 w-4 transition-all duration-300 transform",
          isDark ? "rotate-0 scale-100 opacity-100 text-amber-400" : "-rotate-90 scale-0 opacity-0 absolute"
        )} 
      />
      
      {/* Moon Icon (shown in Light mode) */}
      <Moon 
        className={cn(
          "h-4 w-4 transition-all duration-300 transform",
          !isDark ? "rotate-0 scale-100 opacity-100 text-indigo-600 dark:text-indigo-400" : "rotate-90 scale-0 opacity-0 absolute"
        )} 
      />

      {/* Small badge dot if using system preference */}
      {theme === "system" && (
        <span className="absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full bg-primary ring-2 ring-background" />
      )}
    </button>
  );
}
