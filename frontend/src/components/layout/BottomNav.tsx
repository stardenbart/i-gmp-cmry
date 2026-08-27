"use client";

import Link from "next/link";
import { usePathname, useParams } from "next/navigation";
import { Home, ClipboardCheck, AlertTriangle, User, FileBox } from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { usePermissions } from "@/lib/usePermissions";

export function BottomNav() {
  const pathname = usePathname();
  const params = useParams();
  const user = useAuthStore((state) => state.user);
  const mounted = useMounted();
  const { hasPermission } = usePermissions();

  const plantCode = (params?.plantCode as string) || user?.plant_id || "global";
  const basePath = `/cimory/${plantCode}/dashboard/${mounted ? user?.id : 'overview'}`;

  const NAV_ITEMS = [
    { href: `${basePath}`, label: "Home", icon: Home },
    ...(mounted && hasPermission("PERM-INSP-R")
      ? [{ href: `${basePath}/inspections`, label: "Inspeksi", icon: ClipboardCheck }]
      : []),
    { href: `${basePath}/issues`, label: "Temuan", icon: AlertTriangle },
    { href: `${basePath}/wowr`, label: "WO/WR", icon: FileBox },
    { href: `${basePath}/profile`, label: "Profil", icon: User },
  ];

  return (
    <nav
      className="pwa-bottom-nav-safe fixed bottom-0 left-0 right-0 z-40 flex items-center justify-around border-t border-border bg-card/90 px-2 backdrop-blur-xl md:hidden"
      aria-label="Navigasi utama"
    >
      {NAV_ITEMS.map((item) => {
        const isActive =
          (item.label === "Home")
            ? (pathname === basePath || pathname === basePath + "/")
            : pathname.startsWith(item.href);
        const Icon = item.icon;

        return (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              "flex flex-col items-center justify-center w-full h-full space-y-1 transition-colors",
              isActive ? "text-primary" : "text-muted-foreground hover:text-foreground"
            )}
          >
            <div
              className={cn(
                "flex h-8 w-14 items-center justify-center rounded-full transition-all",
                isActive ? "bg-primary/10" : ""
              )}
            >
              <Icon className="h-5 w-5" />
            </div>
            <span className="text-[10px] font-medium">{item.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}
