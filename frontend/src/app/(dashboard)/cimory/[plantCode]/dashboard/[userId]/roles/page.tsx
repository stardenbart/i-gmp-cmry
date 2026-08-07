"use client";

import { useState, useEffect } from "react";
import { useAuthStore } from "@/stores/authStore";
import { usePermissions } from "@/lib/usePermissions";
import { useMounted } from "@/lib/useMounted";
import { api } from "@/lib/api/axios";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Save, ShieldAlert, Check, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";

// --- API Functions ---
const fetchRoles = async () => {
  const res = await api.get("/master/roles", { params: { limit: 100 } });
  return res.data?.data?.items || [];
};

const fetchModules = async () => {
  const res = await api.get("/master/modules");
  return res.data?.data || [];
};

const fetchRolePermissions = async (roleId: string) => {
  const res = await api.get(`/master/roles/${roleId}/permissions`);
  return res.data?.data || [];
};

export default function RolesPermissionsPage() {
  const user = useAuthStore((state) => state.user);
  const { hasPermission, isLoading: isGuardLoading } = usePermissions();
  const isAdmin = hasPermission("PERM-MSTR-R");
  const mounted = useMounted();
  const queryClient = useQueryClient();

  // State to track modified permissions: { [roleId]: { [permissionId]: isAllowed } }
  const [modifiedPerms, setModifiedPerms] = useState<Record<string, Record<string, boolean>>>({});
  // State to hold the current effective permissions (merged backend + local edits)
  const [effectivePerms, setEffectivePerms] = useState<Record<string, Record<string, boolean>>>({});

  // Data Fetching
  const { data: roles, isLoading: rolesLoading } = useQuery({
    queryKey: ["roles"],
    queryFn: fetchRoles,
    enabled: mounted && !!user && isAdmin,
  });

  const { data: modules, isLoading: modulesLoading } = useQuery({
    queryKey: ["modules"],
    queryFn: fetchModules,
    enabled: mounted && !!user && isAdmin,
  });

  // Fetch all permissions for all roles
  const [allRolePerms, setAllRolePerms] = useState<Record<string, any[]>>({});
  const [permsLoading, setPermsLoading] = useState(true);

  useEffect(() => {
    async function loadAllPermissions() {
      if (!roles || roles.length === 0) return;
      setPermsLoading(true);
      try {
        const permsMap: Record<string, any[]> = {};
        const effMap: Record<string, Record<string, boolean>> = {};
        
        const promises = roles.map((r: any) => fetchRolePermissions(r.role_id));
        const results = await Promise.all(promises);
        
        roles.forEach((r: any, idx: number) => {
          permsMap[r.role_id] = results[idx];
          effMap[r.role_id] = {};
          results[idx].forEach((rp: any) => {
            effMap[r.role_id][rp.permission_id] = rp.is_allowed;
          });
        });
        
        setAllRolePerms(permsMap);
        setEffectivePerms(effMap);
      } catch (err) {
        toast.error("Gagal memuat data permissions");
      } finally {
        setPermsLoading(false);
      }
    }
    loadAllPermissions();
  }, [roles]);

  // Mutations
  const savePermissionsMutation = useMutation({
    mutationFn: async () => {
      const promises = Object.keys(modifiedPerms).map(async (roleId) => {
        const perms = modifiedPerms[roleId];
        const payload = {
          role_id: roleId,
          permissions: Object.keys(perms).map((permId) => ({
            permission_id: permId,
            is_allowed: perms[permId],
          })),
        };
        return api.put(`/master/roles/${roleId}/permissions`, payload);
      });
      await Promise.all(promises);
    },
    onSuccess: () => {
      toast.success("Perubahan hak akses berhasil disimpan");
      setModifiedPerms({});
      // No need to refetch immediately, local state is accurate
    },
    onError: (err: any) => {
      toast.error(err.response?.data?.message || "Gagal menyimpan perubahan");
    }
  });

  const handleToggle = (roleId: string, permId: string, currentStatus: boolean) => {
    const newStatus = !currentStatus;
    
    setEffectivePerms(prev => ({
      ...prev,
      [roleId]: {
        ...prev[roleId],
        [permId]: newStatus
      }
    }));

    setModifiedPerms(prev => ({
      ...prev,
      [roleId]: {
        ...prev[roleId],
        [permId]: newStatus
      }
    }));
  };

  const handleReset = () => {
    // Rebuild effective perms from backend state
    const effMap: Record<string, Record<string, boolean>> = {};
    roles?.forEach((r: any) => {
      effMap[r.role_id] = {};
      const rps = allRolePerms[r.role_id] || [];
      rps.forEach((rp: any) => {
        effMap[r.role_id][rp.permission_id] = rp.is_allowed;
      });
    });
    setEffectivePerms(effMap);
    setModifiedPerms({});
  };

  const hasChanges = Object.keys(modifiedPerms).length > 0;
  const isLoadingAll = rolesLoading || modulesLoading || permsLoading;

  if (!isGuardLoading && !isAdmin) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] py-12 px-4 space-y-4 text-center w-full max-w-lg mx-auto">
        <div className="w-16 h-16 rounded-full bg-destructive/10 flex items-center justify-center shrink-0">
          <ShieldAlert className="h-8 w-8 text-destructive" />
        </div>
        <h2 className="text-xl font-semibold text-foreground">Akses Ditolak</h2>
        <p className="text-sm text-muted-foreground text-center leading-relaxed w-full">
          Anda tidak memiliki izin untuk mengelola Hak Akses &amp; Peran. Halaman ini khusus untuk Administrator.
        </p>
        <Button variant="outline" onClick={() => window.history.back()} className="rounded-xl px-6">Kembali</Button>
      </div>
    );
  }

  if (!mounted || isGuardLoading) return <div className="h-64 flex justify-center items-center"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div></div>;

  return (
    <div className="space-y-6 pb-20">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Team & Permissions</h1>
          <p className="text-muted-foreground mt-1">Control access levels and assign roles to your team.</p>
        </div>
        <div className="flex items-center gap-3">
          {hasChanges && (
            <Button variant="outline" onClick={handleReset} disabled={savePermissionsMutation.isPending}>
              <RefreshCw className="h-4 w-4 mr-2" /> Batal
            </Button>
          )}
          <Button 
            onClick={() => savePermissionsMutation.mutate()} 
            disabled={!hasChanges || savePermissionsMutation.isPending}
            isLoading={savePermissionsMutation.isPending}
          >
            <Save className="h-4 w-4 mr-2" /> Simpan Perubahan
          </Button>
        </div>
      </div>

      {isLoadingAll ? (
        <div className="h-64 flex justify-center items-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : (
        <div className="space-y-8">
          {modules?.map((module: any) => (
            <div key={module.module_id} className="rounded-xl border border-border bg-card overflow-hidden shadow-sm">
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-border bg-muted/30">
                      <th className="px-6 py-4 text-left font-semibold text-foreground min-w-[200px]">
                        {module.module_name}
                      </th>
                      {module.permissions?.map((perm: any) => (
                        <th key={perm.permission_id} className="px-4 py-4 text-center font-medium text-muted-foreground whitespace-nowrap">
                          {perm.permission_name}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    {roles?.map((role: any) => (
                      <tr key={role.role_id} className="hover:bg-muted/10 transition-colors">
                        <td className="px-6 py-4 text-muted-foreground font-medium">
                          {role.role_name}
                        </td>
                        {module.permissions?.map((perm: any) => {
                          const isAllowed = effectivePerms[role.role_id]?.[perm.permission_id] || false;
                          
                          return (
                            <td key={perm.permission_id} className="px-4 py-4 text-center">
                              <label className="relative inline-flex items-center cursor-pointer justify-center group">
                                <input
                                  type="checkbox"
                                  className="sr-only peer"
                                  checked={isAllowed}
                                  onChange={() => handleToggle(role.role_id, perm.permission_id, isAllowed)}
                                />
                                <div className={`
                                  w-6 h-6 rounded border transition-all duration-200 flex items-center justify-center
                                  ${isAllowed 
                                    ? 'bg-primary border-primary text-primary-foreground' 
                                    : 'border-input bg-transparent hover:border-primary/50 text-transparent'}
                                `}>
                                  <Check className={`w-4 h-4 ${isAllowed ? 'opacity-100 scale-100' : 'opacity-0 scale-50 transition-all'}`} />
                                </div>
                              </label>
                            </td>
                          );
                        })}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
