"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Home,
  ClipboardCheck,
  AlertTriangle,
  Database,
  Users,
  History,
  Settings,
  ShieldCheck
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { isAdminUser } from "@/lib/useAdminGuard";


export function Sidebar() {
  const pathname = usePathname();
  const user = useAuthStore((state) => state.user);
  const mounted = useMounted();

  const basePath = `/cimory/dashboard/${mounted ? user?.id : 'overview'}`;

  const MAIN_MENU = [
    { href: `${basePath}`, label: "Dashboard", icon: Home },
    { href: `${basePath}/inspections`, label: "Inspeksi", icon: ClipboardCheck },
    { href: `${basePath}/issues`, label: "Temuan (Issue)", icon: AlertTriangle },
  ];

  // Admin menu - only visible for admin users
  const ADMIN_MENU = mounted && isAdminUser(user?.role_id) ? [
    { href: `${basePath}/master`, label: "Master Data", icon: Database },
    { href: `${basePath}/users`, label: "Manajemen User", icon: Users },
    { href: `${basePath}/logs`, label: "Audit Trail", icon: History },
    { href: `${basePath}/settings`, label: "Pengaturan", icon: Settings },
  ] : [];

  const renderLinks = (links: typeof MAIN_MENU) => {
    return links.map((item) => {
      // For dashboard home, we need exact match logic or checking if it's the base path without sub-routes
      const isActive =
        (item.label === "Dashboard")
          ? (pathname === basePath || pathname === basePath + "/")
          : pathname.startsWith(item.href);
      const Icon = item.icon;

      return (
        <Link
          key={item.href}
          href={item.href}
          className={cn(
            "flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-all",
            isActive
              ? "bg-primary/10 text-primary"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          )}
        >
          <Icon className="h-4 w-4" />
          {item.label}
        </Link>
      );
    });
  };

  return (
    <aside className="fixed inset-y-0 left-0 z-40 hidden w-64 flex-col border-r border-border bg-card md:flex">
      <div className="flex h-16 items-center gap-2 border-b border-border px-6">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
          <ShieldCheck className="h-5 w-5 text-primary-foreground" />
        </div>
        <span className="font-bold tracking-tight">AuditSys</span>
      </div>

      <div className="flex-1 overflow-y-auto py-4">
        <nav className="space-y-1 px-3">
          <div className="mb-2 px-4 text-xs font-semibold uppercase tracking-wider text-muted-foreground/70">
            Utama
          </div>
          {renderLinks(MAIN_MENU)}
        </nav>

        <nav className="mt-8 space-y-1 px-3">
          <div className="mb-2 px-4 text-xs font-semibold uppercase tracking-wider text-muted-foreground/70">
            Administrator
          </div>
          {renderLinks(ADMIN_MENU)}
        </nav>
      </div>

      <div className="border-t border-border p-4">
        <div className="flex items-center gap-3 rounded-xl bg-muted/50 p-3">
          <div className="h-9 w-9 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-sm">
            {mounted && user?.name ? user.name.charAt(0).toUpperCase() : "U"}
          </div>
          <div className="flex flex-col min-w-0">
            <span className="text-sm font-medium truncate">{mounted && user?.name ? user.name : "Loading..."}</span>
            <span className="text-xs text-muted-foreground truncate">{mounted ? (user?.role_id || "-") : "–"}</span>
          </div>
        </div>
      </div>
    </aside>
  );
}
