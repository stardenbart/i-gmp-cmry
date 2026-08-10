"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createContext, useContext, useEffect, useState } from "react";
import { Toaster } from "sonner";

import { StoreProvider } from "@/store/provider";

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

  useEffect(() => {
    const stored = localStorage.getItem("theme") || "dark";
    setThemeState(stored);
    applyTheme(stored);
  }, []);

  const applyTheme = (t: string) => {
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
  };

  const setTheme = (newTheme: string) => {
    setThemeState(newTheme);
    try {
      localStorage.setItem("theme", newTheme);
    } catch (e) {}
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
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) return null;
  return <Toaster position="top-center" theme={(resolvedTheme as "light" | "dark" | "system") || "dark"} />;
}

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 2 * 60 * 1000, // 2 minutes staleTime to prevent duplicate requests on navigation/re-renders
            gcTime: 10 * 60 * 1000,    // 10 minutes in-memory cache time
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
        </QueryClientProvider>
      </ThemeProvider>
    </StoreProvider>
  );
}
