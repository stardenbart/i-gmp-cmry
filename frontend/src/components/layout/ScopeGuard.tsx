"use client";

import { useEffect, useRef } from "react";
import { useParams, useRouter, usePathname } from "next/navigation";
import { useAuthHydrated, useAuthStore } from "@/stores/authStore";
import { authApi } from "@/lib/api/auth.api";
import { Loader2 } from "lucide-react";

export function ScopeGuard({ children }: { children: React.ReactNode }) {
  const { plantCode: urlPlantCode, userId: urlUserId } = useParams() as { plantCode: string; userId: string };
  const router = useRouter();
  const pathname = usePathname();
  const user = useAuthStore((state) => state.user);
  const token = useAuthStore((state) => state.token);
  const setAuth = useAuthStore((state) => state.setAuth);
  const hydrated = useAuthHydrated();
  const hasRefreshedRole = useRef(false);

  const userPlant = user?.plant_id || "global";
  const isSuperAdmin = user?.role_id === "ROLE-000" || user?.role_id === "SUPERADMIN";
  const isInvalidPlant = Boolean(user && !isSuperAdmin && userPlant !== "global" && urlPlantCode !== userPlant);
  const isInvalidUser = Boolean(user && user.id !== urlUserId);
  const isAuthorized = Boolean(user && !isInvalidPlant && !isInvalidUser);

  useEffect(() => {
    if (!hydrated) {
      return;
    }

    if (!user) {
      // A token cookie can outlive the persisted browser state (for example
      // after site data is partially cleared). Remove that stale cookie first,
      // otherwise the proxy redirects /login back here indefinitely.
      useAuthStore.getState().logout();
      router.replace("/login");
      return;
    }

    if (isInvalidPlant || isInvalidUser) {
      const targetPlant = isSuperAdmin ? urlPlantCode : userPlant;
      const newPath = pathname.replace(`/cimory/${urlPlantCode}/dashboard/${urlUserId}`, `/cimory/${targetPlant}/dashboard/${user.id}`);
      router.replace(newPath);
    }
  }, [hydrated, user, urlPlantCode, urlUserId, pathname, router, userPlant, isSuperAdmin, isInvalidPlant, isInvalidUser]);

  // An Admin changing this user's role/plant elsewhere never reaches an
  // already-open session otherwise — the app only fetches fresh role info
  // at actual login time, so a stale localStorage session keeps routing to
  // the old role's dashboard until a full logout+login. Refresh once per
  // page load (guarded by the ref, since this layout persists across
  // client-side navigation) so a plain browser refresh is enough to pick
  // up a role change instead of requiring the affected user to re-login.
  useEffect(() => {
    if (!hydrated || !user || !token || hasRefreshedRole.current) return;
    hasRefreshedRole.current = true;
    authApi
      .me()
      .then((res) => {
        const fresh = res.data;
        setAuth(token, {
          ...user,
          role_id: fresh.role_id,
          plant_id: fresh.plant_id,
          department_id: fresh.department_id,
          user_status: fresh.user_status,
          name: fresh.full_name || user.name,
          full_name: fresh.full_name,
          username: fresh.username,
        });
      })
      .catch(() => {
        // Best-effort: keep the existing session data on failure (e.g. a
        // transient network hiccup) rather than blocking navigation.
      });
  }, [hydrated, user, token, setAuth]);

  if (!hydrated) {
    return (
      <div className="flex min-h-dvh w-full items-center justify-center bg-background" role="status" aria-label="Memuat aplikasi">
        <Loader2 aria-hidden="true" className="h-8 w-8 animate-spin text-primary motion-reduce:animate-none" />
        <span className="sr-only">Memuat aplikasi</span>
      </div>
    );
  }

  // A hydrated session without a user is redirected above. Rendering nothing
  // avoids Lighthouse and users getting stuck on a permanent verification UI.
  if (!isAuthorized) return null;

  return <>{children}</>;
}
