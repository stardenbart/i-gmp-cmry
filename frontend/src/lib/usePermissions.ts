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
  const isAdmin = isAdminUser(roleId);

  // Fetch Role Permissions
  const { data: rolePermsData, isLoading: isRoleLoading } = useQuery({
    queryKey: ["role-permissions", roleId],
    queryFn: async () => {
      const res = await api.get(`/master/roles/${roleId}/permissions`);
      return res.data?.data || [];
    },
    enabled: mounted && !!user && !isAdmin && !!roleId,
  });

  // Fetch User Permissions Overrides
  const { data: userPermsData, isLoading: isUserLoading } = useQuery({
    queryKey: ["user-permissions", userId],
    queryFn: async () => {
      const res = await api.get(`/users/${userId}/permissions`);
      return res.data?.data || [];
    },
    enabled: mounted && !!user && !isAdmin && !!userId,
  });

  const isLoading = !mounted || (!isAdmin && (isRoleLoading || isUserLoading));

  const hasPermission = (permissionId: string): boolean => {
    if (!user) return false;
    if (isAdmin) return true;
    if (isLoading) return false; // pessimistic default while loading to prevent flashes and unauthorized requests

    // User override check first
    const userOverride = userPermsData?.find((up: any) => up.permission_id === permissionId);
    if (userOverride !== undefined) {
      return !!userOverride.is_allowed;
    }

    // Role permission check
    const rolePerm = rolePermsData?.find((rp: any) => rp.permission_id === permissionId);
    if (rolePerm !== undefined) {
      return !!rolePerm.is_allowed;
    }

    // Default to false if permission explicitly checked but not granted
    return false;
  };

  return {
    hasPermission,
    isLoading,
  };
}
