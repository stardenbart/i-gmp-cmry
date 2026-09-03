import type { Metadata } from "next";
import { PublicKPIThemeBoundary } from "@/components/dashboard/kpi/PublicKPIThemeBoundary";

export const metadata: Metadata = {
  title: "Dashboard KPI Publik | Cimory",
  robots: { index: false, follow: false, nocache: true },
  referrer: "no-referrer",
};

export default function SharedKPILayout({ children }: { children: React.ReactNode }) {
  return <PublicKPIThemeBoundary>{children}</PublicKPIThemeBoundary>;
}
