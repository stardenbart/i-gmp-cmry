"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { LogOut, Menu } from "lucide-react";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { NotificationBell } from "@/components/notifications/NotificationBell";

export function Header() {
  const pathname = usePathname();
  const { logout, user } = useAuthStore();
  const mounted = useMounted();

  const getTitle = () => {
    if (pathname.includes("/profile")) return "Profil";
    if (pathname.includes("/inspections")) return "Inspeksi";
    if (pathname.includes("/issues")) return "Temuan (Issue)";
    if (pathname.includes("/master")) return "Master Data";
    if (pathname.includes("/users")) return "Manajemen User";
    if (pathname.includes("/logs")) return "Audit Trail";
    if (pathname.includes("/settings")) return "Pengaturan";
    if (pathname.includes("/cimory/dashboard")) return "Dashboard";
    return "";
  };

  const handleLogout = () => {
    logout();
    if (typeof window !== "undefined") {
      window.location.href = "/login";
    }
  };

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-border bg-background/80 px-4 backdrop-blur-xl sm:px-6">
      <div className="flex items-center gap-4">
        <button className="md:hidden flex items-center justify-center rounded-lg p-2 hover:bg-muted">
          <Menu className="h-5 w-5" />
        </button>
        <h1 className="text-lg font-semibold tracking-tight">{getTitle()}</h1>
      </div>

      <div className="flex items-center gap-2">
        <NotificationBell />
        <div className="h-6 w-px bg-border mx-1"></div>
        <Link href={`/cimory/dashboard/${mounted ? user?.id : 'overview'}/profile`}>
          <div className="flex items-center gap-2 rounded-full px-2 py-1.5 hover:bg-muted cursor-pointer transition-colors">
            <div className="h-7 w-7 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-xs">
              {mounted && user?.name ? user.name.charAt(0).toUpperCase() : "U"}
            </div>
            <span className="hidden sm:inline-block text-sm font-medium max-w-30 truncate">
              {mounted && user?.name ? user.name : "Profil"}
            </span>
          </div>
        </Link>
        <button
          onClick={handleLogout}
          className="flex items-center gap-2 rounded-full px-3 py-1.5 text-sm font-medium text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <LogOut className="h-4 w-4" />
          <span className="hidden sm:inline-block">Keluar</span>
        </button>
      </div>
    </header>
  );
}
