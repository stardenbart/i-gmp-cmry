"use client";

import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/axios";
import { Check, X, Save } from "lucide-react";
import { Button } from "@/components/ui/button";

export function UserPermissionsTab({ userId, roleId }: { userId: string, roleId: string }) {
  const queryClient = useQueryClient();

  const [modifiedPerms, setModifiedPerms] = useState<Record<string, boolean | null>>({});
  
  // 1. Fetch Modules (to build matrix structure)
  const { data: modulesData, isLoading: modulesLoading } = useQuery({
    queryKey: ["modules"],
    queryFn: async () => {
      const res = await api.get("/master/modules");
      return res.data?.data || [];
    }
  });

  // 2. Fetch Role Permissions (baseline)
  const { data: rolePermsData, isLoading: rolePermsLoading } = useQuery({
    queryKey: ["role-permissions", roleId],
    queryFn: async () => {
      const res = await api.get(`/master/roles/${roleId}/permissions`);
      return res.data?.data || [];
    },
    enabled: !!roleId
  });

  // 3. Fetch User Permissions (overrides)
  const { data: userPermsData, isLoading: userPermsLoading } = useQuery({
    queryKey: ["user-permissions", userId],
    queryFn: async () => {
      const res = await api.get(`/users/${userId}/permissions`);
      return res.data?.data || [];
    },
    enabled: !!userId
  });

  useEffect(() => {
    // Reset modified perms when modal opens/data loads
    setModifiedPerms({});
  }, [userId, userPermsData]);

  const saveMutation = useMutation({
    mutationFn: async () => {
      // Build the final array of permissions to send.
      // We only send the ones that have been overridden (not null in our state)
      // Actually, wait, if we cleared an override, we might need to send something?
      // Our API expects the FULL list of overrides for this user.
      
      const currentOverrides = userPermsData || [];
      const overridesMap: Record<string, boolean> = {};
      
      currentOverrides.forEach((up: any) => {
        overridesMap[up.permission_id] = up.is_allowed;
      });

      // Apply modifications
      Object.keys(modifiedPerms).forEach((permId) => {
        const val = modifiedPerms[permId];
        if (val === null) {
          delete overridesMap[permId];
        } else {
          overridesMap[permId] = val;
        }
      });

      const payload = {
        user_id: userId,
        permissions: Object.keys(overridesMap).map(permId => ({
          permission_id: permId,
          is_allowed: overridesMap[permId]
        }))
      };

      return api.put(`/users/${userId}/permissions`, payload);
    },
    onSuccess: () => {
      toast.success("Pengecualian hak akses berhasil disimpan");
      queryClient.invalidateQueries({ queryKey: ["user-permissions", userId] });
    },
    onError: (err: any) => {
      toast.error(err.response?.data?.message || "Gagal menyimpan perubahan");
    }
  });

  if (modulesLoading || rolePermsLoading || userPermsLoading) {
    return <div className="h-40 flex items-center justify-center"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div></div>;
  }

  // Create fast lookup maps
  const roleMap: Record<string, boolean> = {};
  rolePermsData?.forEach((rp: any) => {
    roleMap[rp.permission_id] = rp.is_allowed;
  });

  const userMap: Record<string, boolean> = {};
  userPermsData?.forEach((up: any) => {
    userMap[up.permission_id] = up.is_allowed;
  });

  const handleToggle = (permId: string, currentRoleValue: boolean, hasUserOverride: boolean, currentUserValue: boolean) => {
    // State cycle: Follow Role (null) -> Force Allow (true) -> Force Deny (false) -> Follow Role (null)
    
    // Determine current effective state in the UI
    let currentUiState: boolean | null = null;
    if (modifiedPerms[permId] !== undefined) {
      currentUiState = modifiedPerms[permId];
    } else if (hasUserOverride) {
      currentUiState = currentUserValue;
    }

    let nextState: boolean | null;
    if (currentUiState === null) {
      nextState = true;
    } else if (currentUiState === true) {
      nextState = false;
    } else {
      nextState = null;
    }

    setModifiedPerms(prev => ({ ...prev, [permId]: nextState }));
  };

  const hasChanges = Object.keys(modifiedPerms).length > 0;

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center bg-muted/30 p-3 rounded-lg border border-border">
        <div className="text-sm">
          <p className="font-semibold">Keterangan Status:</p>
          <div className="flex gap-4 mt-2">
            <div className="flex items-center gap-2"><div className="w-4 h-4 rounded bg-primary/20 border border-primary/30 text-primary flex items-center justify-center"><Check className="w-3 h-3"/></div> <span className="text-xs text-muted-foreground">Default Role (Diizinkan)</span></div>
            <div className="flex items-center gap-2"><div className="w-4 h-4 rounded bg-destructive/20 border border-destructive/30 text-destructive flex items-center justify-center"><X className="w-3 h-3"/></div> <span className="text-xs text-muted-foreground">Default Role (Dilarang)</span></div>
            <div className="flex items-center gap-2"><div className="w-4 h-4 rounded bg-primary text-primary-foreground flex items-center justify-center"><Check className="w-3 h-3"/></div> <span className="text-xs">Override (Diizinkan)</span></div>
            <div className="flex items-center gap-2"><div className="w-4 h-4 rounded bg-destructive text-destructive-foreground flex items-center justify-center"><X className="w-3 h-3"/></div> <span className="text-xs">Override (Dilarang)</span></div>
          </div>
        </div>
        <Button size="sm" onClick={() => saveMutation.mutate()} disabled={!hasChanges || saveMutation.isPending} isLoading={saveMutation.isPending}>
          <Save className="w-4 h-4 mr-2" /> Simpan Akses
        </Button>
      </div>

      <div className="max-h-[50vh] overflow-y-auto space-y-6 pr-2">
        {modulesData?.map((module: any) => (
          <div key={module.module_id} className="rounded-xl border border-border bg-card overflow-hidden">
            <div className="bg-muted/50 px-4 py-2 border-b border-border font-semibold text-sm">
              {module.module_name}
            </div>
            <div className="p-4 grid grid-cols-2 sm:grid-cols-3 gap-4">
              {module.permissions?.map((perm: any) => {
                const roleAllowed = roleMap[perm.permission_id] || false;
                const hasUserOverride = userMap[perm.permission_id] !== undefined;
                const userAllowed = userMap[perm.permission_id];
                
                let uiState = modifiedPerms[perm.permission_id];
                if (uiState === undefined) {
                  uiState = hasUserOverride ? userAllowed : null;
                }

                // Determine styling based on uiState
                let boxClass = "bg-muted border-border text-transparent";
                let icon = null;
                let textClass = "text-muted-foreground";

                if (uiState === true) {
                  boxClass = "bg-primary border-primary text-primary-foreground";
                  icon = <Check className="w-4 h-4" />;
                  textClass = "text-foreground font-medium";
                } else if (uiState === false) {
                  boxClass = "bg-destructive border-destructive text-destructive-foreground";
                  icon = <X className="w-4 h-4" />;
                  textClass = "text-foreground font-medium";
                } else {
                  // Fallback to role
                  if (roleAllowed) {
                    boxClass = "bg-primary/20 border-primary/30 text-primary opacity-60";
                    icon = <Check className="w-4 h-4" />;
                  } else {
                    boxClass = "bg-destructive/20 border-destructive/30 text-destructive opacity-60";
                    icon = <X className="w-4 h-4" />;
                  }
                }

                return (
                  <div key={perm.permission_id} 
                    className="flex items-center gap-3 cursor-pointer group hover:bg-muted/30 p-2 rounded-lg transition-colors"
                    onClick={() => handleToggle(perm.permission_id, roleAllowed, hasUserOverride, userAllowed)}
                  >
                    <div className={`w-6 h-6 rounded border flex items-center justify-center transition-all ${boxClass}`}>
                      {icon}
                    </div>
                    <span className={`text-sm ${textClass} select-none`}>
                      {perm.permission_name}
                    </span>
                  </div>
                );
              })}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
