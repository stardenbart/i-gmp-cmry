"use client";

import Link from "next/link";
import { usePathname, useParams } from "next/navigation";
import {
  Home,
  ClipboardCheck,
  ClipboardList,
  AlertTriangle,
  Database,
  Users,
  History,
  Settings,
  FileBox,
  BarChart3
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { usePermissions } from "@/lib/usePermissions";


export function Sidebar() {
  const pathname = usePathname();
  const params = useParams();
  const user = useAuthStore((state) => state.user);
  const mounted = useMounted();
  const { hasPermission } = usePermissions();

  const plantCode = (params?.plantCode as string) || user?.plant_id || "global";
  const basePath = `/cimory/${plantCode}/dashboard/${mounted ? user?.id : 'overview'}`;

  const MAIN_MENU = [
    { href: `${basePath}`, label: "Dasbor", icon: Home },
    // Dashboard KPI/analitik (PERM-KPI-R). Endpoint /dashboard/* yang
    // dipakai widget bawaan tetap tidak digate (shared dengan fitur lain),
    // tapi /analytics/* (Custom Visualization Builder) sudah digate di
    // backend — gate menu ini juga supaya konsisten dengan halamannya.
    ...(hasPermission("PERM-KPI-R")
      ? [{ href: `${basePath}/kpi`, label: "Dashboard KPI", icon: BarChart3 }]
      : []),
    // Inspeksi (PERM-INSP-R)
    ...(hasPermission("PERM-INSP-R")
      ? [{ href: `${basePath}/inspections`, label: "Inspections", icon: ClipboardCheck }]
      : []),
    // Temuan Inspeksi / Issues (PERM-ISS-R)
    ...(hasPermission("PERM-ISS-R")
      ? [{ href: `${basePath}/issues`, label: "Temuan Inspeksi", icon: AlertTriangle }]
      : []),
    // Perintah Kerja / WO/WR (PERM-WOWR-R)
    ...(hasPermission("PERM-WOWR-R")
      ? [{ href: `${basePath}/wowr`, label: "Perintah Kerja", icon: FileBox }]
      : []),
  ];

  // Management menu - dynamically controlled by Role_Permission & User_Permission in DB
  const ADMIN_MENU = [
    ...(hasPermission("PERM-GMP-R")
      ? [{ href: `${basePath}/gmp-data`, label: "Data Inspeksi (GMP)", icon: ClipboardList }]
      : []),
    ...(hasPermission("PERM-MSTR-R")
      ? [{ href: `${basePath}/master`, label: "Data Induk", icon: Database }]
      : []),
    ...(hasPermission("PERM-USR-R")
      ? [{ href: `${basePath}/users`, label: "Manajemen Pengguna", icon: Users }]
      : []),
    ...(hasPermission("PERM-LOG-R")
      ? [{ href: `${basePath}/logs`, label: "Riwayat Aktivitas", icon: History }]
      : []),
    ...(hasPermission("PERM-STNG-R")
      ? [{ href: `${basePath}/settings`, label: "Pengaturan", icon: Settings }]
      : []),
  ];

  const renderLinks = (links: typeof MAIN_MENU) => {
    return links.map((item) => {
      // For dashboard home, we need exact match logic or checking if it's the base path without sub-routes
      const isDashboardHome = item.href === basePath;
      const isActive = isDashboardHome
        ? pathname === basePath || pathname === `${basePath}/`
        : pathname.startsWith(item.href);
      const Icon = item.icon;

      return (
        <Link
          key={item.href}
          href={item.href}
          aria-current={isActive ? "page" : undefined}
          className={cn(
            "flex items-center gap-3 border-l-[3px] px-4 py-[9px] text-[13px] transition-colors duration-150",
            isActive
              ? "border-accent bg-sidebar-active font-semibold text-sidebar-foreground"
              : "border-transparent text-sidebar-muted hover:bg-white/[0.08] hover:text-sidebar-foreground"
          )}
        >
          <Icon aria-hidden="true" className="h-4 w-4 shrink-0" />
          {item.label}
        </Link>
      );
    });
  };

  return (
    <aside className="fixed bottom-0 left-0 top-14 z-40 hidden w-[220px] flex-col bg-sidebar text-sidebar-foreground md:flex">
      <div className="flex-1 overflow-y-auto py-3 [scrollbar-color:rgba(255,255,255,0.2)_transparent] [scrollbar-width:thin]">
        <nav aria-label="Navigasi utama">
          <div className="mb-1 px-4 text-[11px] font-semibold uppercase tracking-wider text-sidebar-muted">
            Utama
          </div>
          {renderLinks(MAIN_MENU)}
        </nav>

        {ADMIN_MENU.length > 0 && (
          <nav className="mt-6" aria-label="Navigasi administrator">
            <div className="mb-1 px-4 text-[11px] font-semibold uppercase tracking-wider text-sidebar-muted">
              Administrator
            </div>
            {renderLinks(ADMIN_MENU)}
          </nav>
        )}
      </div>

      <div className="border-t border-white/10 bg-black/[0.12] px-4 py-3">
        <div className="flex items-center gap-3">
          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-white/15 text-xs font-bold text-sidebar-foreground">
            {mounted && user?.name ? user.name.charAt(0).toUpperCase() : "U"}
          </div>
          <div className="flex min-w-0 flex-col">
            <span className="truncate text-[13px] font-medium">{mounted && user?.name ? user.name : "Loading..."}</span>
            <span className="truncate text-[11px] text-sidebar-muted">{mounted ? (user?.role_id || "-") : "–"}</span>
          </div>
        </div>
      </div>
    </aside>
  );
}
