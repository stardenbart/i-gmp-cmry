"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import {
  User, Mail, Shield, Building2, KeyRound,
  CheckCircle2, LogOut, Edit3, Save, X
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { authApi } from "@/lib/api/auth.api";
import { api } from "@/lib/api/axios";

import { useAuthStore } from "@/stores/authStore";
import { cn } from "@/lib/utils";

const profileSchema = z.object({
  full_name: z.string().min(3, "Nama minimal 3 karakter"),
  email: z.string().email("Format email tidak valid"),
});

const passwordSchema = z.object({
  old_password: z.string().min(1, "Password lama wajib diisi"),
  new_password: z.string().min(8, "Password baru minimal 8 karakter"),
  confirm_password: z.string().min(8, "Konfirmasi password wajib diisi"),
}).refine((d) => d.new_password === d.confirm_password, {
  message: "Konfirmasi password tidak cocok",
  path: ["confirm_password"],
});

type ProfileFormValues = z.infer<typeof profileSchema>;
type PasswordFormValues = z.infer<typeof passwordSchema>;

// Known role ID to Human Name dictionary
const ROLE_NAME_MAP: Record<string, string> = {
  "ROLE-000": "Super Admin",
  "SUPERADMIN": "Super Admin",
  "ROLE-001": "Administrator",
  "ROLE-002": "Administrator",
  "ADMIN": "Administrator",
  "ADM": "Administrator",
  "ROLE-003": "Auditor",
  "AUDITOR": "Auditor",
  "ROLE-004": "Auditee",
  "AUDITEE": "Auditee",
  "ROLE-005": "User",
  "USER": "User",
};

export default function ProfilePage() {
  const queryClient = useQueryClient();
  const { user, setAuth, token, logout } = useAuthStore();
  const [isEditingProfile, setIsEditingProfile] = useState(false);

  // Fetch latest user data from server
  const { data: meData } = useQuery({
    queryKey: ["me"],
    queryFn: authApi.me,
    enabled: !!token,
  });

  // Optional: Fetch master roles for exact dynamic name matching
  const { data: masterRoles } = useQuery({
    queryKey: ["profile-master-roles"],
    queryFn: async () => {
      try {
        const res = await api.get("/master/roles", { params: { limit: 100 } });
        return res.data?.data?.items || [];
      } catch (e) {
        return [];
      }
    },
    staleTime: 10 * 60 * 1000,
  });

  // Optional: Fetch master departments for exact dynamic name matching
  const { data: masterDepts } = useQuery({
    queryKey: ["profile-master-departments"],
    queryFn: async () => {
      try {
        const res = await api.get("/master/departments", { params: { limit: 100 } });
        return res.data?.data?.items || [];
      } catch (e) {
        return [];
      }
    },
    staleTime: 10 * 60 * 1000,
  });

  const currentUser = meData?.data || user;

  // Format Role Name (never display raw ROLE-00x ID)
  const getRoleDisplayName = () => {
    if (currentUser?.role?.role_name) return currentUser.role.role_name;
    if ((currentUser as any)?.role_name) return (currentUser as any).role_name;
    
    const id = currentUser?.role_id;
    if (!id) return "–";

    // Try finding in master roles list
    const foundMaster = masterRoles?.find((r: any) => r.role_id === id);
    if (foundMaster?.role_name) return foundMaster.role_name;

    // Try finding in static dictionary
    const upperId = id.toUpperCase();
    if (ROLE_NAME_MAP[upperId]) return ROLE_NAME_MAP[upperId];

    // Fallback: strip ROLE- prefix if any
    return id.replace(/^ROLE-?/i, "").replace(/_/g, " ");
  };

  // Format Department Name (never display raw DEPT-00x ID)
  const getDeptDisplayName = () => {
    if (currentUser?.department?.department_name) return currentUser.department.department_name;
    if ((currentUser as any)?.department_name) return (currentUser as any).department_name;
    
    const id = currentUser?.department_id;
    if (!id) return "–";

    // Try finding in master departments list
    const foundMaster = masterDepts?.find((d: any) => d.department_id === id);
    if (foundMaster?.department_name) return foundMaster.department_name;

    // Fallback: humanize DEPT- ID
    if (id.startsWith("DEPT-")) {
      return `Departemen ${id.replace("DEPT-", "")}`;
    }

    return id;
  };

  const roleName = getRoleDisplayName();
  const deptName = getDeptDisplayName();

  // Profile form
  const profileForm = useForm<ProfileFormValues>({
    resolver: zodResolver(profileSchema),
    values: {
      full_name: currentUser?.full_name || "",
      email: currentUser?.email || "",
    },
  });

  // Password form
  const passwordForm = useForm<PasswordFormValues>({
    resolver: zodResolver(passwordSchema),
  });

  // Update profile mutation
  const updateProfileMutation = useMutation({
    mutationFn: (data: ProfileFormValues) =>
      authApi.updateProfile(user!.id, data),
    onSuccess: (res) => {
      queryClient.invalidateQueries({ queryKey: ["me"] });
      // Update zustand store
      if (token && user) {
        setAuth(token, { ...user, name: res.data.full_name, email: res.data.email });
      }
      toast.success("Profil berhasil diperbarui");
      setIsEditingProfile(false);
    },
    onError: () => toast.error("Gagal memperbarui profil"),
  });

  // Change password mutation
  const changePasswordMutation = useMutation({
    mutationFn: (data: PasswordFormValues) =>
      authApi.changePassword({ old_password: data.old_password, new_password: data.new_password }),
    onSuccess: () => {
      toast.success("Password berhasil diubah");
      passwordForm.reset();
    },
    onError: (err: any) =>
      toast.error(err?.response?.data?.message || "Gagal mengubah password"),
  });

  const handleLogout = () => {
    logout();
    window.location.href = "/login";
  };

  const statusColor = {
    Active: "text-green-500 bg-green-500/10",
    Inactive: "text-zinc-500 bg-zinc-500/10",
    Suspended: "text-red-500 bg-red-500/10",
  };

  return (
    <div className="space-y-6 max-w-2xl mx-auto">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Profil Saya</h2>
        <p className="text-muted-foreground">
          Kelola informasi akun dan keamanan Anda.
        </p>
      </div>

      {/* Avatar & Identity Card */}
      <Card className="p-6 bg-card/60 backdrop-blur-md">
        <div className="flex flex-col sm:flex-row items-center sm:items-start gap-6">
          {/* Avatar */}
          <div className="relative shrink-0">
            <div className="h-24 w-24 rounded-3xl bg-primary/20 flex items-center justify-center ring-4 ring-primary/10">
              <span className="text-4xl font-bold text-primary">
                {currentUser?.full_name?.charAt(0)?.toUpperCase() || "U"}
              </span>
            </div>
            {currentUser?.user_status && (
              <span className={cn(
                "absolute -bottom-2 left-1/2 -translate-x-1/2 text-[10px] font-semibold px-2 py-0.5 rounded-full whitespace-nowrap",
                statusColor[currentUser.user_status as keyof typeof statusColor]
              )}>
                {currentUser.user_status}
              </span>
            )}
          </div>

          {/* Info */}
          <div className="flex-1 text-center sm:text-left">
            <h3 className="text-xl font-bold">{currentUser?.full_name || "–"}</h3>
            <p className="text-muted-foreground text-sm mt-0.5">@{currentUser?.username || "–"}</p>
            <div className="mt-3 flex flex-wrap justify-center sm:justify-start gap-2">
              <span className="inline-flex items-center gap-1.5 text-xs font-medium px-3 py-1 rounded-full bg-primary/10 text-primary">
                <Shield className="h-3 w-3" />
                {roleName}
              </span>
              <span className="inline-flex items-center gap-1.5 text-xs font-medium px-3 py-1 rounded-full bg-muted text-muted-foreground">
                <Building2 className="h-3 w-3" />
                {deptName}
              </span>
            </div>
          </div>

          {/* Logout Button */}
          <div className="sm:self-start">
            <Button variant="ghost" size="sm" onClick={handleLogout} className="text-red-500 hover:text-red-400 hover:bg-red-500/10">
              <LogOut className="mr-2 h-4 w-4" /> Keluar
            </Button>
          </div>
        </div>
      </Card>

      {/* Edit Profile Card */}
      <Card className="p-6 bg-card/60 backdrop-blur-md">
        <div className="flex items-center justify-between mb-5 border-b border-border pb-3">
          <h3 className="font-semibold flex items-center gap-2">
            <User className="h-4 w-4 text-muted-foreground" />
            Informasi Profil
          </h3>
          {!isEditingProfile ? (
            <Button variant="ghost" size="sm" onClick={() => setIsEditingProfile(true)}>
              <Edit3 className="mr-2 h-3.5 w-3.5" /> Edit
            </Button>
          ) : (
            <Button variant="ghost" size="sm" onClick={() => { setIsEditingProfile(false); profileForm.reset(); }}>
              <X className="mr-2 h-3.5 w-3.5" /> Batal
            </Button>
          )}
        </div>

        {isEditingProfile ? (
          <form onSubmit={profileForm.handleSubmit((d) => updateProfileMutation.mutate(d))} className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-muted-foreground ml-1 mb-2">Nama Lengkap</label>
              <Input {...profileForm.register("full_name")} placeholder="Nama lengkap" />
              {profileForm.formState.errors.full_name && (
                <p className="text-red-500 text-xs mt-1 ml-1">{profileForm.formState.errors.full_name.message}</p>
              )}
            </div>
            <div>
              <label className="block text-xs font-medium text-muted-foreground ml-1 mb-2">Email</label>
              <Input type="email" {...profileForm.register("email")} placeholder="email@perusahaan.com" />
              {profileForm.formState.errors.email && (
                <p className="text-red-500 text-xs mt-1 ml-1">{profileForm.formState.errors.email.message}</p>
              )}
            </div>
            <div className="flex justify-end pt-2">
              <Button type="submit" isLoading={updateProfileMutation.isPending}>
                <Save className="mr-2 h-4 w-4" /> Simpan Perubahan
              </Button>
            </div>
          </form>
        ) : (
          <div className="space-y-4">
            <InfoRow icon={User} label="Nama Lengkap" value={currentUser?.full_name} />
            <InfoRow icon={Mail} label="Email" value={currentUser?.email} />
            <InfoRow icon={Shield} label="Role / Jabatan" value={roleName} />
            <InfoRow icon={Building2} label="Departemen" value={deptName} />
          </div>
        )}
      </Card>

      {/* Change Password Card */}
      <Card className="p-6 bg-card/60 backdrop-blur-md">
        <div className="flex items-center gap-2 mb-5 border-b border-border pb-3">
          <KeyRound className="h-4 w-4 text-muted-foreground" />
          <h3 className="font-semibold">Ubah Password</h3>
        </div>

        <form onSubmit={passwordForm.handleSubmit((d) => changePasswordMutation.mutate(d))} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-muted-foreground ml-1 mb-2">Password Lama</label>
            <Input type="password" {...passwordForm.register("old_password")} placeholder="••••••••" />
            {passwordForm.formState.errors.old_password && (
              <p className="text-red-500 text-xs mt-1 ml-1">{passwordForm.formState.errors.old_password.message}</p>
            )}
          </div>
          <div>
            <label className="block text-xs font-medium text-muted-foreground ml-1 mb-2">Password Baru</label>
            <Input type="password" {...passwordForm.register("new_password")} placeholder="Min. 8 karakter" />
            {passwordForm.formState.errors.new_password && (
              <p className="text-red-500 text-xs mt-1 ml-1">{passwordForm.formState.errors.new_password.message}</p>
            )}
          </div>
          <div>
            <label className="block text-xs font-medium text-muted-foreground ml-1 mb-2">Konfirmasi Password Baru</label>
            <Input type="password" {...passwordForm.register("confirm_password")} placeholder="Ulangi password baru" />
            {passwordForm.formState.errors.confirm_password && (
              <p className="text-red-500 text-xs mt-1 ml-1">{passwordForm.formState.errors.confirm_password.message}</p>
            )}
          </div>

          {/* Password strength hint */}
          <div className="rounded-2xl bg-muted/50 p-3 text-xs text-muted-foreground space-y-1">
            <p className="font-medium text-foreground/70">Syarat password:</p>
            <ul className="space-y-0.5 ml-2">
              <li className={cn("flex items-center gap-1.5", passwordForm.watch("new_password")?.length >= 8 ? "text-green-500" : "")}>
                <CheckCircle2 className="h-3 w-3" /> Minimal 8 karakter
              </li>
            </ul>
          </div>

          <div className="flex justify-end pt-2">
            <Button type="submit" isLoading={changePasswordMutation.isPending}>
              <KeyRound className="mr-2 h-4 w-4" /> Ubah Password
            </Button>
          </div>
        </form>
      </Card>
    </div>
  );
}

function InfoRow({
  icon: Icon,
  label,
  value,
}: {
  icon: React.ElementType;
  label: string;
  value?: string | null;
}) {
  return (
    <div className="flex items-center gap-4 py-2 border-b border-border/50 last:border-0">
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-muted">
        <Icon className="h-4 w-4 text-muted-foreground" />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-xs text-muted-foreground">{label}</p>
        <p className="text-sm font-medium truncate">{value || "–"}</p>
      </div>
    </div>
  );
}
