"use client";

import { useEffect, useRef, useState } from "react";
import { useParams, useRouter, usePathname } from "next/navigation";
import { useAuthHydrated, useAuthStore, type User } from "@/stores/authStore";
import { authApi, type UserInfo } from "@/lib/api/auth.api";
import { Loader2 } from "lucide-react";

function freshUserInfoToUser(fresh: UserInfo, previous?: User | null): User {
  return {
    ...previous,
    id: fresh.user_id,
    name: fresh.full_name || previous?.name || "",
    email: fresh.email,
    role_id: fresh.role_id,
    plant_id: fresh.plant_id,
    department_id: fresh.department_id,
    user_status: fresh.user_status,
    full_name: fresh.full_name,
    username: fresh.username,
  };
}

export function ScopeGuard({ children }: { children: React.ReactNode }) {
  const { plantCode: urlPlantCode, userId: urlUserId } = useParams() as { plantCode: string; userId: string };
  const router = useRouter();
  const pathname = usePathname();
  const user = useAuthStore((state) => state.user);
  const setAuth = useAuthStore((state) => state.setAuth);
  const hydrated = useAuthHydrated();
  const hasCheckedSession = useRef(false);
  const [sessionChecked, setSessionChecked] = useState(false);

  const userPlant = user?.plant_id || "global";
  const isSuperAdmin = user?.role_id === "ROLE-000" || user?.role_id === "SUPERADMIN";
  const isInvalidPlant = Boolean(user && !isSuperAdmin && userPlant !== "global" && urlPlantCode !== userPlant);
  const isInvalidUser = Boolean(user && user.id !== urlUserId);
  const isAuthorized = Boolean(user && !isInvalidPlant && !isInvalidUser);

  // Validate the session against the backend exactly once per mount.
  // access_token/refresh_token are httpOnly, so the persisted `user` profile
  // is only ever a best-effort local cache — the cookie the browser is
  // actually holding is the real source of truth, and /auth/me is what
  // reconciles the two. This also picks up an Admin changing this user's
  // role/plant elsewhere (same as the previous per-mount role refresh), and
  // fixes the case where a valid session cookie outlives a wiped/cleared
  // local profile: instead of assuming "no persisted user" means logged out,
  // this lets a still-valid cookie sign the user back in.
  useEffect(() => {
    if (!hydrated || hasCheckedSession.current) return;
    hasCheckedSession.current = true;

    authApi
      .me()
      .then((res) => {
        setAuth(freshUserInfoToUser(res.data, user));
      })
      .catch(() => {
        // No valid session cookie (expired, revoked, or never existed) —
        // any stale local profile is meaningless without it.
        useAuthStore.getState().logout();
      })
      .finally(() => setSessionChecked(true));
    // Deliberately runs once per mount (guarded by the ref above), not on
    // every `user`/setAuth change — setAuth() inside this effect would
    // otherwise retrigger it.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [hydrated]);

  useEffect(() => {
    if (!hydrated) return;
    // Nothing persisted to render optimistically yet — wait for the
    // /auth/me check above before deciding this is really a logged-out
    // browser rather than one with a valid cookie but a cleared local cache.
    if (!user && !sessionChecked) return;

    if (!user) {
      router.replace("/login");
      return;
    }

    if (isInvalidPlant || isInvalidUser) {
      const targetPlant = isSuperAdmin ? urlPlantCode : userPlant;
      const newPath = pathname.replace(`/cimory/${urlPlantCode}/dashboard/${urlUserId}`, `/cimory/${targetPlant}/dashboard/${user.id}`);
      router.replace(newPath);
    }
  }, [hydrated, sessionChecked, user, urlPlantCode, urlUserId, pathname, router, userPlant, isSuperAdmin, isInvalidPlant, isInvalidUser]);

  // A session already persisted from a previous visit renders optimistically
  // while /auth/me reconciles silently in the background (see effect
  // above) — only a completely empty local cache waits on that call before
  // deciding whether to show the loader or redirect to /login.
  const isLoading = !hydrated || (!user && !sessionChecked);
  if (isLoading) {
    return (
      <div className="flex min-h-dvh w-full items-center justify-center bg-background" role="status" aria-label="Memuat aplikasi">
        <Loader2 aria-hidden="true" className="h-8 w-8 animate-spin text-primary motion-reduce:animate-none" />
        <span className="sr-only">Memuat aplikasi</span>
      </div>
    );
  }

  // An unauthorized state is redirected above. Rendering nothing avoids
  // Lighthouse and users getting stuck on a permanent verification UI.
  if (!isAuthorized) return null;

  return <>{children}</>;
}
