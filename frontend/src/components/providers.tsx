"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createContext, useCallback, useContext, useEffect, useState } from "react";
import { Toaster } from "sonner";

import { StoreProvider } from "@/store/provider";
import { useMounted } from "@/lib/useMounted";

interface ThemeContextType {
  theme: string;
  setTheme: (theme: string) => void;
  resolvedTheme: string;
}

const ThemeContext = createContext<ThemeContextType>({
  theme: "dark",
  setTheme: () => {},
  resolvedTheme: "dark",
});

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<string>("dark");
  const [resolvedTheme, setResolvedTheme] = useState<string>("dark");

  const applyTheme = useCallback((t: string) => {
    const root = document.documentElement;
    let actualTheme = t;
    if (t === "system") {
      actualTheme = typeof window !== "undefined" && window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
    }
    setResolvedTheme(actualTheme);
    if (actualTheme === "dark") {
      root.classList.add("dark");
    } else {
      root.classList.remove("dark");
    }
  }, []);

  useEffect(() => {
    const frame = requestAnimationFrame(() => {
      const stored = localStorage.getItem("theme") || "dark";
      setThemeState(stored);
      applyTheme(stored);
    });
    return () => cancelAnimationFrame(frame);
  }, [applyTheme]);

  const setTheme = (newTheme: string) => {
    setThemeState(newTheme);
    try {
      localStorage.setItem("theme", newTheme);
    } catch {
      // Storage can be unavailable in restricted browser contexts.
    }
    applyTheme(newTheme);
  };

  return (
    <ThemeContext.Provider value={{ theme, setTheme, resolvedTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

export const useTheme = () => useContext(ThemeContext);

function ToasterWithTheme() {
  const { resolvedTheme } = useTheme();
  const mounted = useMounted();

  if (!mounted) return null;
  return <Toaster position="top-center" theme={(resolvedTheme as "light" | "dark" | "system") || "dark"} />;
}

import { CoreWebVitalsOverlay } from "@/components/ui/CoreWebVitalsOverlay";

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 5 * 60 * 1000, // 5 minutes staleTime for instant 0ms page transitions
            gcTime: 30 * 60 * 1000,   // 30 minutes in-memory cache time
            refetchOnWindowFocus: false,
            refetchIntervalInBackground: false,
            retry: 1,
          },
        },
      })
  );

  return (
    <StoreProvider>
      <ThemeProvider>
        <QueryClientProvider client={queryClient}>
          {children}
          <ToasterWithTheme />
          <CoreWebVitalsOverlay />
        </QueryClientProvider>
      </ThemeProvider>
    </StoreProvider>
  );
}
