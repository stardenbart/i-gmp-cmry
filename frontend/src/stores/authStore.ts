import { create } from "zustand";
import { persist } from "zustand/middleware";
import { useSyncExternalStore } from "react";

import { setSessionHintCookie } from "@/lib/utils";

export interface User {
  id: string;
  name: string;
  email: string;
  role_id: string;
  plant_id?: string;
  role?: {
    role_name: string;
  };
  username?: string;
  full_name?: string;
  department_id?: string;
  user_status?: "Active" | "Inactive" | "Suspended";
}

interface AuthState {
  user: User | null;
  // Sets the locally-cached user profile. The actual session lives entirely
  // in httpOnly cookies set by the backend (see lib/api/axios.ts) — this
  // store never holds a token, only non-sensitive profile fields used to
  // paint the UI instantly before ScopeGuard's /auth/me call confirms (or
  // corrects) it against the real session.
  setAuth: (user: User) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      setAuth: (user) => {
        setSessionHintCookie(user.id);
        set({ user });
      },
      logout: () => {
        setSessionHintCookie(null);
        set({ user: null });
      },
    }),
    {
      name: "auth-storage",
    }
  )
);

const subscribeToAuthHydration = (onStoreChange: () => void) =>
  useAuthStore.persist.onFinishHydration(onStoreChange);

/**
 * Distinguishes an unauthenticated browser from a browser whose persisted
 * session is still being restored. This prevents protected routes from
 * showing an indefinite verification screen during initial hydration.
 */
export function useAuthHydrated() {
  return useSyncExternalStore(
    subscribeToAuthHydration,
    () => useAuthStore.persist.hasHydrated(),
    () => false
  );
}
