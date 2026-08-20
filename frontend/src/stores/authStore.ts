import { create } from "zustand";
import { persist } from "zustand/middleware";

import { setAuthCookie } from "@/lib/utils";

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
  token: string | null;
  user: User | null;
  setAuth: (token: string, user: User) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      setAuth: (token, user) => {
        setAuthCookie(token, user.id);
        set({ token, user });
      },
      logout: () => {
        setAuthCookie(null, null);
        set({ token: null, user: null });
      },
    }),
    {
      name: "auth-storage",
    }
  )
);
