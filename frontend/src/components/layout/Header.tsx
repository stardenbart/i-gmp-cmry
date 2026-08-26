"use client";

import Link from "next/link";
import Image from "next/image";
import { usePathname, useParams } from "next/navigation";
import { LogOut } from "lucide-react";
import { useAuthStore } from "@/stores/authStore";
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

  const handleLogout = () => {
    logout();
    if (typeof window !== "undefined") {
      window.location.href = "/login";
    }
  };

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-border bg-background/80 px-4 backdrop-blur-xl sm:px-6">
      <div className="flex items-center gap-4">
        {/* Mobile Logo */}
        <div className="md:hidden flex items-center gap-3">
          <div className="flex items-center justify-center">
            <Image 
              src="/Logo_Cimory.png" 
              alt="Cimory Logo" 
              width={100} 
              height={32} 
              priority
              className="h-7 w-auto object-contain"
            />
          </div>
        </div>
        
        {/* Desktop Title */}
        <h1 className="hidden md:block text-lg font-semibold tracking-tight">{getTitle()}</h1>
      </div>

      <div className="flex items-center gap-2">
        <InstallAppButton />
        <ThemeToggle />
        <NotificationBell />
        <div className="h-6 w-px bg-border mx-1"></div>
        <Link href={`/cimory/${plantCode}/dashboard/${mounted ? user?.id : 'overview'}/profile`}>
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
