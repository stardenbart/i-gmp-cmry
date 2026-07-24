"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter, usePathname } from "next/navigation";
import { useAuthStore } from "@/stores/authStore";
import { isAdminUser } from "@/lib/useAdminGuard";
import { Loader2 } from "lucide-react";

export function ScopeGuard({ children }: { children: React.ReactNode }) {
  const { userId: urlUserId } = useParams() as { userId: string };
  const router = useRouter();
  const pathname = usePathname();
  const user = useAuthStore((state) => state.user);
  const [isAuthorized, setIsAuthorized] = useState(false);

  useEffect(() => {
    if (!user) {
      // User state not initialized yet
      return;
    }

    // If the userId in the URL does not match the logged-in user's ID
    if (user.id !== urlUserId) {
      // We block access and redirect them to their own URL scope
      // (Unless they are an Admin and we explicitly want to allow impersonation, 
      // but for strict security, we'll force redirect everyone to their own scope for now)
      
      const newPath = pathname.replace(`/cimory/dashboard/${urlUserId}`, `/cimory/dashboard/${user.id}`);
      router.replace(newPath);
    } else {
      setIsAuthorized(true);
    }
  }, [user, urlUserId, pathname, router]);

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
