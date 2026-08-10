"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState } from "react";
import { ThemeProvider, useTheme } from "next-themes";
import { Toaster } from "sonner";

import { StoreProvider } from "@/store/provider";

function ToasterWithTheme() {
  const { resolvedTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) return null;
  return <Toaster position="top-center" theme={(resolvedTheme as "light" | "dark" | "system") || "system"} />;
}

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            refetchOnWindowFocus: false,
            retry: 1,
          },
        },
      })
  );

  return (
    <StoreProvider>
      <ThemeProvider attribute="class" defaultTheme="system" enableSystem={true}>
        <QueryClientProvider client={queryClient}>
          {children}
          <ToasterWithTheme />
        </QueryClientProvider>
      </ThemeProvider>
    </StoreProvider>
  );
}
