"use client";

import { useState, useEffect } from "react";
import {
  Users,
  Search,
  Plus,
  Edit2,
  Trash2,
  KeyRound,
  ShieldAlert,
  ChevronLeft,
  ChevronRight,
  X,
  Filter,
  ChevronDown
} from "lucide-react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/axios";
import { filterApi, UserFilterParams } from "@/lib/api/filter.api";
import { useMounted } from "@/lib/useMounted";
import { useAdminGuard } from "@/lib/useAdminGuard";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuthStore } from "@/stores/authStore";
import { UserPermissionsTab } from "./UserPermissionsTab";
import { masterApi } from "@/lib/api/master.api";

// --- API Functions (CRUD only, filter uses filterApi) ---
const fetchRoles = async () => {
  const res = await api.get("/master/roles", { params: { limit: 100 } });
  return res.data;
};

const fetchDepartments = async () => {
  const res = await api.get("/master/departments", { params: { limit: 100 } });
  return res.data;
};


const createUser = async (data: Record<string, unknown>) => {
  const res = await api.post("/users", data);
  return res.data;
};

const updateUser = async (id: string, data: Record<string, unknown>) => {
  const res = await api.put(`/users/${id}`, data);
  return res.data;
};

const deleteUser = async (id: string) => {
  const res = await api.delete(`/users/${id}`);
  return res.data;
};

const resetPassword = async (id: string, newPassword: string) => {
  const res = await api.put(`/users/${id}/reset-password`, { new_password: newPassword });
  return res.data;
};

// --- Modals ---
function FormModal({
  isOpen, onClose, title, onSubmit, isLoading, children,
}: any) {
  if (!isOpen) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />
      <div className={`relative z-10 w-[90vw] ${title === "Edit User" ? "sm:w-[700px]" : "sm:w-[500px]"} rounded-2xl bg-card border border-border p-6 shadow-2xl max-h-[90vh] overflow-hidden flex flex-col`}>
        <div className="mb-4 flex items-center justify-between shrink-0">
          <h2 className="text-lg font-semibold">{title}</h2>
          <button type="button" onClick={onClose} className="rounded-lg p-1 hover:bg-muted">
            <X className="h-5 w-5" />
          </button>
        </div>
        <div className="overflow-y-auto flex-1">
          {children}
        </div>
      </div>
    </div>
  );
}

function ActionModal({
  isOpen, onClose, onConfirm, isLoading, title, description, confirmText, isDestructive = false
}: any) {
  if (!isOpen) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />
      <div className="relative z-10 w-[90vw] sm:w-[400px] rounded-2xl bg-card border border-border p-6 shadow-2xl">
        <h2 className="text-lg font-semibold mb-2">{title}</h2>
        <p className="text-muted-foreground mb-6">{description}</p>
        <div className="flex gap-3">
          <Button type="button" variant="outline" onClick={onClose} className="flex-1">Batal</Button>
          <Button variant={isDestructive ? "destructive" : "default"} onClick={onConfirm} isLoading={isLoading} className="flex-1">
            {confirmText}
          </Button>
        </div>
      </div>
    </div>
  );
}

// --- Main Page ---
export default function UsersPage() {
  const user = useAuthStore((state) => state.user);
  const { isAdmin, isLoading: isGuardLoading } = useAdminGuard();
  const mounted = useMounted();
  const queryClient = useQueryClient();

  const [page, setPage] = useState(1);
  const [searchQuery, setSearchQuery] = useState("");
  const [roleFilter, setRoleFilter] = useState("ALL");
  const [statusFilter, setStatusFilter] = useState("ALL");

  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<any>(null);
  const [activeTab, setActiveTab] = useState<"profile" | "permissions">("profile");
  const [deleteItemId, setDeleteItemId] = useState<string | null>(null);
  const [resetPassId, setResetPassId] = useState<string | null>(null);
  const [newPassword, setNewPassword] = useState("");

  // Controlled state for PIC fields
  const [picKawasanIds, setPicKawasanIds] = useState<string[]>([]);
  const [picKategori, setPicKategori] = useState<string>("");

  // Sync PIC state when editingItem changes
  useEffect(() => {
    if (editingItem) {
      const kawasanIds = editingItem.pic_mappings?.map((m: any) => m.kawasan_id) || [];
      const kategori = editingItem.pic_mappings?.[0]?.kategori_pic || "";
      setPicKawasanIds(kawasanIds);
      setPicKategori(kategori);
    } else {
      setPicKawasanIds([]);
      setPicKategori("");
    }
  }, [editingItem, isFormOpen]);

  // Data Fetching via /users/filter endpoint
  const filterParams: UserFilterParams = {
    page,
    limit: 10,
    ...(searchQuery && { q: searchQuery }),
    ...(roleFilter !== "ALL" && { role_id: roleFilter }),
    ...(statusFilter !== "ALL" && { user_status: statusFilter }),
    sort_by: "created_at",
    sort_order: "desc",
  };

  const { data: usersRes, isLoading, isFetching } = useQuery({
    queryKey: ["users-filter", filterParams],
    queryFn: () => filterApi.users(filterParams),
    enabled: mounted && !!user && isAdmin,
  });

  const { data: rolesRes } = useQuery({
    queryKey: ["roles-list"],
    queryFn: fetchRoles,
    enabled: mounted && !!user && isAdmin,
  });

  const { data: deptsRes } = useQuery({
    queryKey: ["master-departments"],
    queryFn: fetchDepartments,
    enabled: mounted && !!user && isAdmin,
  });

  const { data: kawasansRes } = useQuery({
    queryKey: ["master-kawasans"],
    queryFn: () => masterApi.getKawasans({ limit: 1000 }),
    enabled: mounted && !!user && isAdmin,
  });

  // Extract data from filter response
  const usersList = usersRes?.items || [];
  const pagination = {
    total: usersRes?.total ?? 0,
    page: usersRes?.page ?? 1,
    limit: usersRes?.limit ?? 10,
    total_pages: usersRes?.total_pages ?? 1,
  };
  const facets = usersRes?.facets;
  const rolesList = rolesRes?.data?.items || [];
  const deptsList = deptsRes?.data?.items || [];
  const kawasansList = kawasansRes || [];

  // Mutations
  const createMutation = useMutation({
    mutationFn: createUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users-filter"] });
      setIsFormOpen(false);
      toast.success("User berhasil ditambahkan");
    },
    onError: (err: any) => toast.error(err.response?.data?.message || "Gagal menambah user"),
  });

  const updateMutation = useMutation({
    mutationFn: (data: any) => updateUser(editingItem.user_id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users-filter"] });
      setIsFormOpen(false);
      setEditingItem(null);
      toast.success("User berhasil diperbarui");
    },
    onError: (err: any) => toast.error(err.response?.data?.message || "Gagal memperbarui user"),
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteUser(deleteItemId!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users-filter"] });
      setDeleteItemId(null);
      toast.success("User berhasil dihapus");
    },
    onError: (err: any) => toast.error(err.response?.data?.message || "Gagal menghapus user"),
  });

  const resetPassMutation = useMutation({
    mutationFn: () => resetPassword(resetPassId!, newPassword),
    onSuccess: () => {
      setResetPassId(null);
      setNewPassword("");
      toast.success("Password berhasil direset");
    },
    onError: (err: any) => toast.error(err.response?.data?.message || "Gagal reset password"),
  });

  // Handlers
  const handleFormSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const data: Record<string, unknown> = {};
    
    // Process non-PIC fields from FormData
    const keys = Array.from(new Set(formData.keys()));
    for (const key of keys) {
      // PIC fields are handled separately via controlled state
      if (key === 'pic_kawasan_ids' || key === 'pic_kategori') continue;
      const value = formData.get(key);
      if (value !== "") {
        data[key] = value;
      }
    }

    // Always include PIC fields from controlled state (never from FormData)
    // This ensures empty array is sent correctly when all kawasans are deselected
    data['pic_kawasan_ids'] = picKawasanIds;
    data['pic_kategori'] = picKategori;

    if (editingItem) updateMutation.mutate(data);
    else createMutation.mutate(data);
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "Active": return <span className="bg-green-500/10 text-green-500 px-2 py-1 rounded text-xs font-semibold">Active</span>;
      case "Inactive": return <span className="bg-orange-500/10 text-orange-500 px-2 py-1 rounded text-xs font-semibold">Inactive</span>;
      case "Suspended": return <span className="bg-red-500/10 text-red-500 px-2 py-1 rounded text-xs font-semibold">Suspended</span>;
      default: return <span className="bg-muted text-muted-foreground px-2 py-1 rounded text-xs font-semibold">{status}</span>;
    }
  };

  if (!isGuardLoading && !isAdmin) {
    return (
      <div className="flex flex-col items-center justify-center min-h-100 space-y-4">
        <ShieldAlert className="h-12 w-12 text-destructive" />
        <h2 className="text-xl font-semibold">Akses Ditolak</h2>
        <Button onClick={() => window.history.back()}>Kembali</Button>
      </div>
    );
  }

  if (!mounted || isGuardLoading) return <div className="h-64 flex justify-center items-center"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div></div>;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">User Management</h1>
        <p className="text-muted-foreground">Kelola pengguna, hak akses, dan departemen</p>
      </div>

      <div className="flex flex-col md:flex-row gap-4 justify-between items-center bg-card p-4 rounded-xl border border-border shadow-sm">
        <div className="flex flex-col sm:flex-row gap-3 w-full md:w-auto flex-1">
          <div className="relative w-full sm:max-w-xs min-w-[250px]">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
            <Input
              placeholder="Cari nama, username..."
              value={searchQuery}
              onChange={(e) => { setSearchQuery(e.target.value); setPage(1); }}
              className="pl-9 w-full"
            />
          </div>
          <div className="flex gap-2 items-center">
            <Filter className="h-4 w-4 text-muted-foreground" />
            <div className="relative">
              <select
                className="text-sm bg-background appearance-none border border-border rounded-md pl-3 pr-8 py-2 outline-none"
                value={roleFilter}
                onChange={(e) => { setRoleFilter(e.target.value); setPage(1); }}
              >
                <option value="ALL">Semua Role</option>
                {rolesList.map((r: any) => (
                  <option key={r.role_id} value={r.role_id}>{r.role_name}</option>
                ))}
              </select>
              <ChevronDown className="absolute right-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
            </div>
            <div className="relative">
              <select
                className="text-sm bg-background appearance-none border border-border rounded-md pl-3 pr-8 py-2 outline-none"
                value={statusFilter}
                onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
              >
                <option value="ALL">Semua Status</option>
                <option value="Active">Active</option>
                <option value="Inactive">Inactive</option>
                <option value="Suspended">Suspended</option>
              </select>
              <ChevronDown className="absolute right-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
            </div>
          </div>
        </div>
        <Button onClick={() => { setEditingItem(null); setActiveTab("profile"); setIsFormOpen(true); }} className="w-full md:w-auto">
          <Plus className="h-4 w-4 mr-2" /> Tambah User
        </Button>
      </div>

      <div className="rounded-xl border border-border bg-card overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border bg-muted/50 text-left text-muted-foreground uppercase tracking-wider text-xs font-semibold">
                <th className="px-4 py-3">Pengguna</th>
                <th className="px-4 py-3">Role</th>
                <th className="px-4 py-3">Departemen</th>
                <th className="px-4 py-3 text-center">Status</th>
                <th className="px-4 py-3 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {isLoading || isFetching ? (
                <tr><td colSpan={5} className="p-8 text-center"><div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary mx-auto"></div></td></tr>
              ) : usersList.length === 0 ? (
                <tr><td colSpan={5} className="p-8 text-center text-muted-foreground">Tidak ada user ditemukan</td></tr>
              ) : (
                usersList.map((u: any) => (
                  <tr key={u.user_id} className="hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary font-bold text-xs uppercase">
                          {u.full_name?.substring(0, 2) || "U"}
                        </div>
                        <div className="flex flex-col">
                          <span className="font-semibold">{u.full_name}</span>
                          <span className="text-xs text-muted-foreground">{u.username} • {u.email}</span>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3 font-medium">{rolesList.find((r:any) => r.role_id === u.role_id)?.role_name || u.role_id}</td>
                    <td className="px-4 py-3 text-muted-foreground">{deptsList.find((d:any) => d.department_id === u.department_id)?.department_name || u.department_id}</td>
                    <td className="px-4 py-3 text-center">{getStatusBadge(u.user_status)}</td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex justify-end gap-1">
                        <Button variant="ghost" size="sm" onClick={() => { setEditingItem(u); setActiveTab("profile"); setIsFormOpen(true); }} className="h-8 w-8 p-0" title="Edit">
                          <Edit2 className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => setResetPassId(u.user_id)} className="h-8 w-8 p-0 text-orange-500 hover:text-orange-600" title="Reset Password">
                          <KeyRound className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => setDeleteItemId(u.user_id)} className="h-8 w-8 p-0 text-destructive hover:text-destructive" title="Hapus">
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        {pagination.total_pages > 1 && (
          <div className="flex items-center justify-between border-t border-border px-4 py-3">
            <p className="text-sm text-muted-foreground">
              Hal {page} dari {pagination.total_pages} ({pagination.total} total)
            </p>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1}><ChevronLeft className="h-4 w-4" /></Button>
              <Button variant="outline" size="sm" onClick={() => setPage(p => Math.min(pagination.total_pages, p + 1))} disabled={page === pagination.total_pages}><ChevronRight className="h-4 w-4" /></Button>
            </div>
          </div>
        )}
      </div>

      {/* CREATE/EDIT MODAL */}
      <FormModal
        isOpen={isFormOpen}
        onClose={() => { setIsFormOpen(false); setEditingItem(null); }}
        title={editingItem ? "Edit User" : "Tambah User Baru"}
      >
        {editingItem && (
          <div className="flex border-b border-border mb-4">
            <button
              className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === "profile" ? "border-primary text-primary" : "border-transparent text-muted-foreground hover:text-foreground"}`}
              onClick={() => setActiveTab("profile")}
            >
              Profil User
            </button>
            <button
              className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === "permissions" ? "border-primary text-primary" : "border-transparent text-muted-foreground hover:text-foreground"}`}
              onClick={() => setActiveTab("permissions")}
            >
              Hak Akses Khusus
            </button>
          </div>
        )}

        {activeTab === "profile" || !editingItem ? (
          <form onSubmit={handleFormSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="col-span-2">
                <label className="text-sm font-medium mb-1 block">Nama Lengkap</label>
                <Input name="full_name" defaultValue={editingItem?.full_name} required />
              </div>
              <div className="col-span-2 md:col-span-1">
                <label className="text-sm font-medium mb-1 block">Username</label>
                <Input name="username" defaultValue={editingItem?.username} required disabled={!!editingItem} />
                {editingItem && <p className="text-[10px] text-muted-foreground mt-1">Username tidak dapat diubah</p>}
              </div>
              <div className="col-span-2 md:col-span-1">
                <label className="text-sm font-medium mb-1 block">Email</label>
                <Input name="email" type="email" defaultValue={editingItem?.email} required />
              </div>
              
              <div className="col-span-2 md:col-span-1">
                <label className="text-sm font-medium mb-1 block">Role</label>
                <div className="relative">
                  <select name="role_id" defaultValue={editingItem?.role_id || ""} required className="flex h-10 w-full appearance-none rounded-md border border-input bg-background px-3 py-2 pr-10 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                    <option value="" disabled>Pilih Role</option>
                    {rolesList.map((r: any) => <option key={r.role_id} value={r.role_id}>{r.role_name}</option>)}
                  </select>
                  <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
                </div>
              </div>
              <div className="col-span-2 md:col-span-1">
                <label className="text-sm font-medium mb-1 block">Departemen</label>
                <div className="relative">
                  <select name="department_id" defaultValue={editingItem?.department_id || ""} required className="flex h-10 w-full appearance-none rounded-md border border-input bg-background px-3 py-2 pr-10 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                    <option value="" disabled>Pilih Departemen</option>
                    {deptsList.map((d: any) => <option key={d.department_id} value={d.department_id}>{d.department_name}</option>)}
                  </select>
                  <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
                </div>
              </div>

              {editingItem && (
                <div className="col-span-2">
                  <label className="text-sm font-medium mb-1 block">Status</label>
                  <div className="relative">
                    <select name="user_status" defaultValue={editingItem?.user_status} className="flex h-10 w-full appearance-none rounded-md border border-input bg-background px-3 py-2 pr-10 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                      <option value="Active">Active</option>
                      <option value="Inactive">Inactive</option>
                      <option value="Suspended">Suspended</option>
                    </select>
                    <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
                  </div>
                </div>
              )}

              {/* PIC Mapping Fields - fully controlled via state */}
              <div className="col-span-2 md:col-span-1">
                <label className="text-sm font-medium mb-1 block">Kategori PIC <span className="text-muted-foreground text-xs font-normal">(Opsional)</span></label>
                <div className="relative">
                  <select
                    value={picKategori}
                    onChange={(e) => setPicKategori(e.target.value)}
                    className="flex h-10 w-full appearance-none rounded-md border border-input bg-background px-3 py-2 pr-10 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                  >
                    <option value="">Tidak Ada Kategori</option>
                    <option value="Manager">Manager</option>
                    <option value="Supervisor">Supervisor</option>
                    <option value="Staff">Staff</option>
                  </select>
                  <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
                </div>
              </div>
              <div className="col-span-2 md:col-span-1">
                <label className="text-sm font-medium mb-1 block">Kawasan PIC <span className="text-muted-foreground text-xs font-normal">(Opsional)</span></label>
                <select
                  multiple
                  value={picKawasanIds}
                  onChange={(e) => {
                    const selected = Array.from(e.target.selectedOptions).map(o => o.value);
                    setPicKawasanIds(selected);
                  }}
                  className="flex min-h-20 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                >
                  {kawasansList.map((k: any) => <option key={k.kawasan_id} value={k.kawasan_id}>{k.kawasan_name}</option>)}
                </select>
                <p className="text-[10px] text-muted-foreground mt-1">Tahan Ctrl/Cmd untuk memilih lebih dari satu kawasan</p>
              </div>

              {!editingItem && (
                <div className="col-span-2">
                  <label className="text-sm font-medium mb-1 block">Password</label>
                  <Input name="password" type="password" required minLength={6} placeholder="Minimal 6 karakter" />
                </div>
              )}
            </div>
            
            <div className="flex gap-3 pt-4 border-t border-border mt-4">
              <Button type="button" variant="outline" onClick={() => setIsFormOpen(false)} className="flex-1">Batal</Button>
              <Button type="submit" isLoading={createMutation.isPending || updateMutation.isPending} className="flex-1">Simpan Profil</Button>
            </div>
          </form>
        ) : (
          <UserPermissionsTab userId={editingItem?.user_id} roleId={editingItem?.role_id} />
        )}
      </FormModal>

      {/* RESET PASSWORD MODAL */}
      <FormModal
        isOpen={!!resetPassId}
        onClose={() => { setResetPassId(null); setNewPassword(""); }}
        title="Reset Password"
      >
        <form onSubmit={(e: any) => { e.preventDefault(); resetPassMutation.mutate(); }}>
          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium mb-1 block">Password Baru</label>
              <Input 
                type="password" 
                value={newPassword} 
                onChange={(e) => setNewPassword(e.target.value)} 
                required minLength={6} 
                placeholder="Masukkan password baru..." 
              />
            </div>
          </div>
          <div className="flex gap-3 pt-4 mt-4 border-t border-border">
            <Button type="button" variant="outline" onClick={() => setResetPassId(null)} className="flex-1">Batal</Button>
            <Button type="submit" isLoading={resetPassMutation.isPending} className="flex-1">Simpan Password</Button>
          </div>
        </form>
      </FormModal>

      {/* DELETE MODAL */}
      <ActionModal
        isOpen={!!deleteItemId}
        onClose={() => setDeleteItemId(null)}
        onConfirm={() => deleteMutation.mutate()}
        isLoading={deleteMutation.isPending}
        title="Hapus User"
        description="Apakah Anda yakin ingin menghapus user ini? Tindakan ini tidak dapat dibatalkan."
        confirmText="Hapus Permanen"
        isDestructive={true}
      />
    </div>
  );
}
