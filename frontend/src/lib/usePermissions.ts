"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/axios";
import { useAuthStore } from "@/stores/authStore";
import { isAdminUser } from "@/lib/useAdminGuard";
import { useMounted } from "@/lib/useMounted";

export function usePermissions() {
  const user = useAuthStore((state) => state.user);
  const mounted = useMounted();

  const roleId = user?.role_id;
  const userId = user?.id;
  const isAdmin = isAdminUser(roleId, user?.role?.role_name);

  // Fetch Role Permissions dynamically from backend DB
  const { data: rolePermsData, isLoading: isRoleLoading } = useQuery({
    queryKey: ["role-permissions", roleId],
    queryFn: async () => {
      const res = await api.get(`/master/roles/${roleId}/permissions`);
      return res.data?.data || [];
    },
    enabled: mounted && !!user && !isAdmin && !!roleId,
    staleTime: 5 * 60 * 1000,
  });

  // Fetch User Permissions Overrides dynamically from backend DB
  const { data: userPermsData, isLoading: isUserLoading } = useQuery({
    queryKey: ["user-permissions", userId],
    queryFn: async () => {
      const res = await api.get(`/users/${userId}/permissions`);
      return res.data?.data || [];
    },
    enabled: mounted && !!user && !isAdmin && !!userId,
    staleTime: 5 * 60 * 1000,
  });

  const isLoading = !mounted || (!isAdmin && (isRoleLoading || isUserLoading));

  /**
   * Pure dynamic permission checker against Database Role_Permission & User_Permission
   */
  const hasPermission = (permissionId: string): boolean => {
    if (!user) return false;
    if (isAdmin) return true;

    // 1. User override check (Highest priority in dynamic RBAC)
    if (userPermsData && Array.isArray(userPermsData)) {
      const userOverride = userPermsData.find((up) => up.permission_id === permissionId);
      if (userOverride !== undefined) {
        return !!userOverride.is_allowed;
      }
    }

    // 2. Role permission check from DB
    if (rolePermsData && Array.isArray(rolePermsData)) {
      const rolePerm = rolePermsData.find((rp) => rp.permission_id === permissionId);
      if (rolePerm !== undefined) {
        return !!rolePerm.is_allowed;
      }
    }

    // 3. Fallback during initial load before queries resolve
    if (isLoading) {
      // Do NOT grant admin/management permissions speculatively during initial load to prevent UI flashing
      const isRestrictedAdminKey = ["PERM-MSTR", "PERM-USR", "PERM-LOG", "PERM-STNG", "PERM-GMP", "PERM-KPI"].some(
        (prefix) => permissionId.startsWith(prefix)
      );
      if (isRestrictedAdminKey) return false;

      // Optimistic read only for standard non-restricted feature pages (e.g. PERM-INSP-R, PERM-ISS-R)
      return permissionId.endsWith("-R");
    }

    return false;
  };

  return {
    hasPermission,
    isLoading,
  };
}
