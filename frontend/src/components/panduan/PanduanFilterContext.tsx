"use client";

import { createContext, useContext, useState, type ReactNode } from "react";
import type { PanduanRole } from "./theme";

interface PanduanFilterValue {
  filter: PanduanRole;
  setFilter: (role: PanduanRole) => void;
}

const PanduanFilterContext = createContext<PanduanFilterValue | null>(null);

export function PanduanFilterProvider({ children }: { children: ReactNode }) {
  const [filter, setFilter] = useState<PanduanRole>("all");
  return (
    <PanduanFilterContext.Provider value={{ filter, setFilter }}>
      {children}
    </PanduanFilterContext.Provider>
  );
}

export function usePanduanFilter() {
  const ctx = useContext(PanduanFilterContext);
  if (!ctx) throw new Error("usePanduanFilter harus dipakai di dalam PanduanFilterProvider");
  return ctx;
}

/** Sembunyikan children bila peran yang sedang dipilih tidak termasuk `roles`. */
export function RoleScope({ roles, children }: { roles: Exclude<PanduanRole, "all">[]; children: ReactNode }) {
  const { filter } = usePanduanFilter();
  if (filter !== "all" && !roles.includes(filter)) return null;
  return <>{children}</>;
}
