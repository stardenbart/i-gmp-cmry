"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, ClipboardCheck, AlertTriangle, User, FileBox } from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { isAuditorUser } from "@/lib/useAdminGuard";

export function BottomNav() {
  const pathname = usePathname();
  const user = useAuthStore((state) => state.user);
  const mounted = useMounted();

  const basePath = `/cimory/dashboard/${mounted ? user?.id : 'overview'}`;

  const NAV_ITEMS = [
    { href: `${basePath}`, label: "Home", icon: Home },
    ...(mounted && isAuditorUser(user?.role_id)
      ? [{ href: `${basePath}/inspections`, label: "Inspeksi", icon: ClipboardCheck }]
      : []),
    { href: `${basePath}/issues`, label: "Temuan", icon: AlertTriangle },
    { href: `${basePath}/wowr`, label: "WO/WR", icon: FileBox },
    { href: `${basePath}/profile`, label: "Profil", icon: User },
  ];

  return (
    <div className="fixed bottom-0 left-0 right-0 z-50 flex h-16 items-center justify-around border-t border-border bg-card/80 px-2 pb-safe backdrop-blur-xl md:hidden">
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
    </div>
  );
}
