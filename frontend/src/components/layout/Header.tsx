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
import { cn } from "@/lib/utils";

// Translucent white controls that sit on the brand gradient (style guide
// `.btn-header`): white 15% fill, white 30% border, 4px radius.
const HEADER_ACTION =
  "h-8 w-8 min-h-8 min-w-8 rounded-[4px] border border-white/30 bg-white/15 text-white shadow-none transition-colors duration-150 hover:bg-white/25 hover:text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/70";

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
    if (pathname.includes("/wowr")) return "Perintah Kerja";
    if (pathname.includes("/kpi")) return "Dashboard KPI";
    if (pathname.includes("/gmp-data")) return "Data Inspeksi (GMP)";
    if (pathname.includes("/notifications")) return "Notifikasi";
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
    <header className="app-header pwa-header-safe fixed inset-x-0 top-0 z-50 flex items-center justify-between gap-3 px-3 text-white sm:px-5">
      <div className="flex min-w-0 flex-1 items-center gap-3 overflow-hidden">
        <Image
          src="/Logo_Cimory.png"
          alt="Cimory"
          width={68}
          height={36}
          priority
          className="h-6 w-auto shrink-0 object-contain md:h-9"
        />
        <div className="flex min-w-0 flex-col leading-tight">
          <span className="truncate text-[13px] font-semibold md:text-base">I-GMP</span>
          <span className="truncate text-[8px] tracking-wide text-white/60 md:text-[11px]">PT CISARUA MOUNTAIN DAIRY TBK</span>
        </div>
        {getTitle() && (
          <>
            <span aria-hidden="true" className="mx-1 hidden h-6 w-px bg-white/25 lg:block" />
            <h1 className="hidden truncate text-sm font-medium text-white/85 lg:block">{getTitle()}</h1>
          </>
        )}
      </div>

      <div className="flex shrink-0 items-center gap-1.5">
        <span className="mr-1 hidden max-w-48 truncate text-xs text-white/80 xl:inline">
          Halo, {mounted && user?.name ? user.name : "Pengguna"}
        </span>
        <Link
          href={`/cimory/${plantCode}/dashboard/${mounted ? user?.id : "overview"}/panduan`}
          aria-label="Buka Panduan Pengguna"
          title="Panduan"
          className={cn(HEADER_ACTION, "flex items-center justify-center")}
        >
          <BookOpen aria-hidden="true" className="h-4 w-4" />
        </Link>
        <InstallAppButton className={cn(HEADER_ACTION, "xl:rounded-[4px]")} />
        <ThemeToggle className={HEADER_ACTION} />
        <NotificationBell triggerClassName={HEADER_ACTION} />
        <Link
          className={cn(HEADER_ACTION, "hidden items-center justify-center text-xs font-bold lg:flex")}
          href={`/cimory/${plantCode}/dashboard/${mounted ? user?.id : "overview"}/profile`}
          aria-label="Profil pengguna"
          title="Profil"
        >
          {mounted && user?.name ? user.name.charAt(0).toUpperCase() : "U"}
        </Link>
        <button
          onClick={handleLogout}
          type="button"
          aria-label="Keluar dari aplikasi"
          title="Keluar"
          className={cn(HEADER_ACTION, "hidden items-center justify-center gap-2 text-[11px] font-medium lg:flex 2xl:w-auto 2xl:px-3")}
        >
          <LogOut aria-hidden="true" className="h-4 w-4" />
          <span className="hidden 2xl:inline-block">Keluar</span>
        </button>
      </div>
    </header>
  );
}
