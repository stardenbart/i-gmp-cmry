"use client";

import Link from "next/link";
import Image from "next/image";
import { usePathname, useParams } from "next/navigation";
import { LogOut, BookOpen } from "lucide-react";
import { useAuthStore } from "@/stores/authStore";
import { authApi } from "@/lib/api/auth.api";
import { useMounted } from "@/lib/useMounted";
import { NotificationBell } from "@/components/notifications/NotificationBell";
import { InstallAppButton } from "./InstallAppButton";
import { ThemeToggle } from "@/components/ui/ThemeToggle";

export function Header() {
  const pathname = usePathname();
  const params = useParams();
  const { logout, user } = useAuthStore();
  const mounted = useMounted();

  const plantCode = (params?.plantCode as string) || user?.plant_id || "global";

  const getTitle = () => {
    if (pathname.includes("/panduan")) return "Panduan";
    if (pathname.includes("/profile")) return "Profil Pengguna";
    if (pathname.includes("/inspections")) return "Inspeksi";
    if (pathname.includes("/issues")) return "Temuan Inspeksi";
    if (pathname.includes("/master")) return "Data Induk";
    if (pathname.includes("/users")) return "Manajemen Pengguna";
    if (pathname.includes("/logs")) return "Riwayat Aktivitas";
    if (pathname.includes("/settings")) return "Pengaturan";
    if (pathname.includes("/dashboard")) return "Dasbor Utama";
    return "";
  };

  const handleLogout = async () => {
    // access_token/refresh_token are httpOnly — only the backend can clear
    // them, so the revoke call must happen before (or at least alongside)
    // the local state clear, not be skipped.
    try {
      await authApi.logout();
    } catch {
      // Best-effort: still clear local state and leave even if the network
      // call fails — the access token cookie will simply expire on its own
      // shortly (15 min TTL) if the revoke call above didn't land.
    }
    logout();
    if (typeof window !== "undefined") {
      window.location.href = "/login";
    }
  };

  return (
    <header className="pwa-header-safe sticky top-0 z-30 flex items-center justify-between border-b border-border bg-background/80 px-3 backdrop-blur-xl sm:px-6">
      <div className="flex min-w-0 flex-1 items-center gap-4 overflow-hidden">
        {/* Mobile Logo */}
        <div className="md:hidden flex items-center gap-3">
          <div className="flex items-center justify-center">
            <Image 
              src="/Logo_Cimory.png" 
              alt="Cimory Logo" 
              width={88}
              height={32} 
              priority
              className="h-7 w-auto max-w-[88px] object-contain"
            />
          </div>
        </div>
        
        {/* Desktop Title */}
        <h1 className="hidden truncate pr-3 text-lg font-semibold tracking-tight lg:block">{getTitle()}</h1>
      </div>

      <div className="flex shrink-0 items-center gap-1 lg:gap-2">
        <Link
          href={`/cimory/${plantCode}/dashboard/${mounted ? user?.id : "overview"}/panduan`}
          aria-label="Buka Panduan Pengguna"
          title="Panduan"
          className="flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <BookOpen aria-hidden="true" className="h-4 w-4" />
        </Link>
        <InstallAppButton />
        <ThemeToggle />
        <NotificationBell />
        <div className="mx-1 hidden h-6 w-px bg-border lg:block"></div>
        <Link className="hidden lg:block" href={`/cimory/${plantCode}/dashboard/${mounted ? user?.id : 'overview'}/profile`}>
          <div className="flex items-center gap-2 rounded-full px-2 py-1.5 hover:bg-muted cursor-pointer transition-colors">
            <div className="h-7 w-7 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-xs">
              {mounted && user?.name ? user.name.charAt(0).toUpperCase() : "U"}
            </div>
            <span className="hidden max-w-30 truncate text-sm font-medium 2xl:inline-block">
              {mounted && user?.name ? user.name : "Profil"}
            </span>
          </div>
        </Link>
        <button
          onClick={handleLogout}
          type="button"
          aria-label="Keluar dari aplikasi"
          title="Keluar"
          className="hidden h-9 w-9 items-center justify-center rounded-lg text-sm font-medium text-muted-foreground hover:bg-muted hover:text-foreground lg:flex 2xl:w-auto 2xl:gap-2 2xl:rounded-full 2xl:px-3"
        >
          <LogOut aria-hidden="true" className="h-4 w-4" />
          <span className="hidden 2xl:inline-block">Keluar</span>
        </button>
      </div>
    </header>
  );
}
