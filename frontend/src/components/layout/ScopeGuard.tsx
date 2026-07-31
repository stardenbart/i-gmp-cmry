"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter, usePathname } from "next/navigation";
import { useAuthStore } from "@/stores/authStore";
import { isAdminUser } from "@/lib/useAdminGuard";
import { Loader2 } from "lucide-react";

export function ScopeGuard({ children }: { children: React.ReactNode }) {
  const { plantCode: urlPlantCode, userId: urlUserId } = useParams() as { plantCode: string; userId: string };
  const router = useRouter();
  const pathname = usePathname();
  const user = useAuthStore((state) => state.user);
  const [isAuthorized, setIsAuthorized] = useState(false);

  useEffect(() => {
    if (!user) {
      // User state not initialized yet
      return;
    }

    const userPlant = user.plant_id || "global";
    const isSuperAdmin = user.role_id === "ROLE-000" || user.role_id === "SUPERADMIN";

    // Validate plant scope if user is non-superadmin
    const isInvalidPlant = !isSuperAdmin && userPlant !== "global" && urlPlantCode !== userPlant;
    const isInvalidUser = user.id !== urlUserId;

    if (isInvalidPlant || isInvalidUser) {
      const targetPlant = isSuperAdmin ? urlPlantCode : userPlant;
      const newPath = pathname.replace(`/cimory/${urlPlantCode}/dashboard/${urlUserId}`, `/cimory/${targetPlant}/dashboard/${user.id}`);
      router.replace(newPath);
    } else {
      setIsAuthorized(true);
    }
  }, [user, urlPlantCode, urlUserId, pathname, router]);

  if (!isAuthorized) {
    return (
      <div className="flex h-screen w-full flex-col items-center justify-center bg-background">
        <Loader2 className="h-10 w-10 animate-spin text-primary mb-4" />
        <p className="text-sm text-muted-foreground animate-pulse">Memverifikasi akses...</p>
      </div>
    );
  }

  return <>{children}</>;
}
