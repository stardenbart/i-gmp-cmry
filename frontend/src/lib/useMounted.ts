"use client";

import { useEffect, useState } from "react";

/**
 * Hook to track when a component has been mounted on the client.
 * Useful for preventing hydration mismatches with client-side only state (e.g., localStorage, cookies).
 *
 * @returns boolean - true after the component has mounted on the client
 *
 * @example
 * // Before: Redundant in every component
 * const [mounted, setMounted] = useState(false);
 * useEffect(() => setMounted(true), []);
 *
 * // After: Reusable hook
 * const mounted = useMounted();
 * const basePath = `/dashboard/${mounted ? user?.id : 'default'}`;
 */
export function useMounted(): boolean {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  return mounted;
}
