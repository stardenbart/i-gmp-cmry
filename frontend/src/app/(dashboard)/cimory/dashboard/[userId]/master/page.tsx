"use client";

import { useState } from "react";
import {
  Building2,
  MapPin,
  Layers,
  ListChecks,
  ClipboardList,
  FileText,
  CheckSquare,
  Settings,
  Plus,
  Search,
  Edit2,
  Trash2,
  ChevronLeft,
  ChevronRight,
  X,
  ShieldAlert,
} from "lucide-react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/axios";
import { useMounted } from "@/lib/useMounted";
import { useAdminGuard } from "@/lib/useAdminGuard";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuthStore } from "@/stores/authStore";

// Types
interface Department {
  department_id: string;
  department_name: string;
  department_code: string;
  is_active: boolean;
}

interface Area {
  area_id: string;
  area_name: string;
  area_code: string;
  department_id?: string;
}

interface Kawasan {
  kawasan_id: string;
  kawasan_name: string;
  kawasan_code: string;
  area_id: string;
}

interface DetailKawasan {
  detail_kawasan_id: string;
  detail_kawasan_name: string;
  detail_kawasan_code: string;
  kawasan_id: string;
}

interface Aspek {
  aspek_id: string;
  aspek_name: string;
  aspek_code: string;
  aspek_weight: number;
  area_id?: string;
}

interface Detail {
  detail_id: string;
  detail_name: string;
  detail_code: string;
  aspek_id: string;
}

interface Uraian {
  uraian_id: string;
  uraian_name: string;
  uraian_code: string;
  detail_id: string;
}

// API Functions
const fetchItems = async (endpoint: string, page = 1, search = "") => {
  const res = await api.get(endpoint, { params: { page, limit: 10, search } });
  return res.data;
};

const createItem = async (endpoint: string, data: Record<string, unknown>) => {
  const res = await api.post(endpoint, data);
  return res.data;
};

const updateItem = async (endpoint: string, id: string, data: Record<string, unknown>) => {
  const res = await api.put(`${endpoint}/${id}`, data);
  return res.data;
};

const deleteItem = async (endpoint: string, id: string) => {
  const res = await api.delete(`${endpoint}/${id}`);
  return res.data;
};

// Master Data Config
const MASTER_TABS = [
  { id: "departments", label: "Department", icon: Building2, endpoint: "/master/departments" },
  { id: "areas", label: "Area", icon: MapPin, endpoint: "/master/areas" },
  { id: "kawasans", label: "Kawasan", icon: Layers, endpoint: "/master/kawasans" },
  { id: "detail-kawasans", label: "Detail Kawasan", icon: Layers, endpoint: "/master/detail-kawasans" },
  { id: "aspeks", label: "Aspek", icon: ClipboardList, endpoint: "/master/aspeks" },
  { id: "details", label: "Detail Audit", icon: FileText, endpoint: "/master/details" },
  { id: "urains", label: "Uraian", icon: CheckSquare, endpoint: "/master/urains" },
] as const;

// Modal Component
function FormModal({
  isOpen,
  onClose,
  title,
  onSubmit,
  isLoading,
  children,
}: {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  onSubmit: (e: React.FormEvent<HTMLFormElement>) => void;
  isLoading: boolean;
  children: React.ReactNode;
}) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />
      <div className="relative z-10 w-full max-w-md rounded-2xl bg-card border border-border p-6 shadow-2xl">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold">{title}</h2>
          <button onClick={onClose} className="rounded-lg p-1 hover:bg-muted">
            <X className="h-5 w-5" />
          </button>
        </div>
        <form onSubmit={onSubmit} className="space-y-4">
          {children}
          <div className="flex gap-3 pt-2">
            <Button type="button" variant="outline" onClick={onClose} className="flex-1">
              Batal
            </Button>
            <Button type="submit" isLoading={isLoading} className="flex-1">
              Simpan
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

// Delete Confirmation Modal
function DeleteModal({
  isOpen,
  onClose,
  onConfirm,
  isLoading,
  itemName,
}: {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  isLoading: boolean;
  itemName: string;
}) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />
      <div className="relative z-10 w-full max-w-sm rounded-2xl bg-card border border-border p-6 shadow-2xl">
        <h2 className="text-lg font-semibold mb-2">Hapus Data</h2>
        <p className="text-muted-foreground mb-6">
          Apakah Anda yakin ingin menghapus <strong>{itemName}</strong>? Tindakan ini tidak dapat dibatalkan.
        </p>
        <div className="flex gap-3">
          <Button type="button" variant="outline" onClick={onClose} className="flex-1">
            Batal
          </Button>
          <Button variant="destructive" onClick={onConfirm} isLoading={isLoading} className="flex-1">
            Hapus
          </Button>
        </div>
      </div>
    </div>
  );
}

export default function MasterPage() {
  const user = useAuthStore((state) => state.user);
  const { isAdmin, isLoading: isGuardLoading } = useAdminGuard();
  const mounted = useMounted();
  const queryClient = useQueryClient();

  const [activeTab, setActiveTab] = useState<string>("departments");
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<Record<string, unknown> | null>(null);
  const [deleteItemId, setDeleteItemId] = useState<string | null>(null);
  const [deleteItemName, setDeleteItemName] = useState("");

  // Get current tab config
  const currentTab = MASTER_TABS.find((tab) => tab.id === activeTab)!;

  // Fetch data
  const { data, isLoading, isFetching } = useQuery({
    queryKey: ["master", activeTab, page, searchQuery],
    queryFn: () => fetchItems(currentTab.endpoint, page, searchQuery),
    enabled: mounted && !!user,
  });

  const items = data?.data?.items || [];
  const pagination = data?.data?.pagination || { total: 0, page: 1, limit: 10, total_pages: 1 };

  // Mutations
  const createMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) => createItem(currentTab.endpoint, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["master", activeTab] });
      setIsModalOpen(false);
      toast.success("Data berhasil ditambahkan");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || "Gagal menambahkan data");
    },
  });

  const updateMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) =>
      updateItem(currentTab.endpoint, editingItem?.id as string, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["master", activeTab] });
      setIsModalOpen(false);
      setEditingItem(null);
      toast.success("Data berhasil diperbarui");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || "Gagal memperbarui data");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteItem(currentTab.endpoint, deleteItemId!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["master", activeTab] });
      setDeleteItemId(null);
      toast.success("Data berhasil dihapus");
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || "Gagal menghapus data");
    },
  });

  // Access denied component
  if (!isGuardLoading && !isAdmin) {
    return (
      <div className="flex flex-col items-center justify-center min-h-100 space-y-4">
        <div className="w-16 h-16 rounded-full bg-destructive/10 flex items-center justify-center">
          <ShieldAlert className="h-8 w-8 text-destructive" />
        </div>
        <h2 className="text-xl font-semibold">Akses Ditolak</h2>
        <p className="text-muted-foreground text-center max-w-md">
          Anda tidak memiliki izin untuk mengakses halaman Master Data.
          Hanya administrator yang dapat mengakses halaman ini.
        </p>
        {mounted && user && (
          <Button variant="outline" asChild>
            <a href={`/cimory/dashboard/${user.id}`}>Kembali ke Dashboard</a>
          </Button>
        )}
      </div>
    );
  }

  if (!mounted || isGuardLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  // Form Fields based on active tab
  const renderFormFields = () => {
    const fields: Record<string, React.ReactNode> = {
      departments: (
        <>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Kode Department</label>
            <Input
              name="department_code"
              defaultValue={editingItem?.department_code as string}
              placeholder="Contoh: DEPT001"
              required
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Nama Department</label>
            <Input
              name="department_name"
              defaultValue={editingItem?.department_name as string}
              placeholder="Contoh: Quality Assurance"
              required
            />
          </div>
        </>
      ),
      areas: (
        <>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Kode Area</label>
            <Input
              name="area_code"
              defaultValue={editingItem?.area_code as string}
              placeholder="Contoh: AREA001"
              required
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Nama Area</label>
            <Input
              name="area_name"
              defaultValue={editingItem?.area_name as string}
              placeholder="Contoh: Pabrik Utama"
              required
            />
          </div>
        </>
      ),
      kawasans: (
        <>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Kode Kawasan</label>
            <Input
              name="kawasan_code"
              defaultValue={editingItem?.kawasan_code as string}
              placeholder="Contoh: KWS001"
              required
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Nama Kawasan</label>
            <Input
              name="kawasan_name"
              defaultValue={editingItem?.kawasan_name as string}
              placeholder="Contoh: Gedung A"
              required
            />
          </div>
        </>
      ),
      "detail-kawasans": (
        <>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Kode Detail Kawasan</label>
            <Input
              name="detail_kawasan_code"
              defaultValue={editingItem?.detail_kawasan_code as string}
              placeholder="Contoh: DK001"
              required
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Nama Detail Kawasan</label>
            <Input
              name="detail_kawasan_name"
              defaultValue={editingItem?.detail_kawasan_name as string}
              placeholder="Contoh: Lantai 1"
              required
            />
          </div>
        </>
      ),
      aspeks: (
        <>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Kode Aspek</label>
            <Input
              name="aspek_code"
              defaultValue={editingItem?.aspek_code as string}
              placeholder="Contoh: ASP001"
              required
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Nama Aspek</label>
            <Input
              name="aspek_name"
              defaultValue={editingItem?.aspek_name as string}
              placeholder="Contoh: Kebersihan"
              required
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Bobot (%)</label>
            <Input
              name="aspek_weight"
              type="number"
              min="0"
              max="100"
              defaultValue={editingItem?.aspek_weight as number || 0}
              required
            />
          </div>
        </>
      ),
      details: (
        <>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Kode Detail</label>
            <Input
              name="detail_code"
              defaultValue={editingItem?.detail_code as string}
              placeholder="Contoh: DTL001"
              required
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Nama Detail</label>
            <Input
              name="detail_name"
              defaultValue={editingItem?.detail_name as string}
              placeholder="Contoh: WC Umum"
              required
            />
          </div>
        </>
      ),
      urains: (
        <>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Kode Uraian</label>
            <Input
              name="uraian_code"
              defaultValue={editingItem?.uraian_code as string}
              placeholder="Contoh: URI001"
              required
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium">Nama Uraian</label>
            <Input
              name="uraian_name"
              defaultValue={editingItem?.uraian_name as string}
              placeholder="Contoh: Lantai bersih"
              required
            />
          </div>
        </>
      ),
    };

    return fields[activeTab] || null;
  };

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const data: Record<string, unknown> = {};

    formData.forEach((value, key) => {
      if (value !== "") {
        data[key] = value;
      }
    });

    if (editingItem?.id) {
      updateMutation.mutate(data);
    } else {
      createMutation.mutate(data);
    }
  };

  const handleEdit = (item: Record<string, unknown>) => {
    setEditingItem(item);
    setIsModalOpen(true);
  };

  const handleDelete = (id: string, name: string) => {
    setDeleteItemId(id);
    setDeleteItemName(name);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setEditingItem(null);
  };

  // Get display columns based on tab
  const getColumns = () => {
    const columns: Record<string, { key: string; label: string }[]> = {
      departments: [
        { key: "department_code", label: "Kode" },
        { key: "department_name", label: "Nama" },
      ],
      areas: [
        { key: "area_code", label: "Kode" },
        { key: "area_name", label: "Nama" },
      ],
      kawasans: [
        { key: "kawasan_code", label: "Kode" },
        { key: "kawasan_name", label: "Nama" },
      ],
      "detail-kawasans": [
        { key: "detail_kawasan_code", label: "Kode" },
        { key: "detail_kawasan_name", label: "Nama" },
      ],
      aspeks: [
        { key: "aspek_code", label: "Kode" },
        { key: "aspek_name", label: "Nama" },
        { key: "aspek_weight", label: "Bobot" },
      ],
      details: [
        { key: "detail_code", label: "Kode" },
        { key: "detail_name", label: "Nama" },
      ],
      urains: [
        { key: "uraian_code", label: "Kode" },
        { key: "uraian_name", label: "Nama" },
      ],
    };
    return columns[activeTab] || [];
  };

  // Get ID field based on tab
  const getIdField = () => {
    const idFields: Record<string, string> = {
      departments: "department_id",
      areas: "area_id",
      kawasans: "kawasan_id",
      "detail-kawasans": "detail_kawasan_id",
      aspeks: "aspek_id",
      details: "detail_id",
      urains: "uraian_id",
    };
    return idFields[activeTab] || "id";
  };

  // Get name field based on tab
  const getNameField = () => {
    const nameFields: Record<string, string> = {
      departments: "department_name",
      areas: "area_name",
      kawasans: "kawasan_name",
      "detail-kawasans": "detail_kawasan_name",
      aspeks: "aspek_name",
      details: "detail_name",
      urains: "uraian_name",
    };
    return nameFields[activeTab] || "name";
  };

  if (!mounted) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold">Master Data</h1>
        <p className="text-muted-foreground">Kelola data master sistem audit</p>
      </div>

      {/* Tabs */}
      <div className="border-b border-border">
        <div className="flex gap-1 overflow-x-auto pb-px -mb-px scrollbar-hide">
          {MASTER_TABS.map((tab) => {
            const Icon = tab.icon;
            return (
              <button
                key={tab.id}
                onClick={() => {
                  setActiveTab(tab.id);
                  setPage(1);
                  setSearchQuery("");
                }}
                className={`flex items-center gap-2 px-4 py-2.5 text-sm font-medium whitespace-nowrap border-b-2 transition-colors ${
                  activeTab === tab.id
                    ? "border-primary text-primary"
                    : "border-transparent text-muted-foreground hover:text-foreground hover:border-border"
                }`}
              >
                <Icon className="h-4 w-4" />
                {tab.label}
              </button>
            );
          })}
        </div>
      </div>

      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row gap-4 justify-between">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={`Cari ${currentTab.label.toLowerCase()}...`}
            value={searchQuery}
            onChange={(e) => {
              setSearchQuery(e.target.value);
              setPage(1);
            }}
            className="pl-9"
          />
        </div>
        <Button onClick={() => setIsModalOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Tambah {currentTab.label}
        </Button>
      </div>

      {/* Table */}
      <div className="rounded-xl border border-border bg-card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border bg-muted/50">
                {getColumns().map((col) => (
                  <th
                    key={col.key}
                    className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted-foreground"
                  >
                    {col.label}
                  </th>
                ))}
                <th className="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  Aksi
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {isLoading ? (
                <tr>
                  <td colSpan={getColumns().length + 1} className="px-4 py-12 text-center">
                    <div className="flex justify-center">
                      <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
                    </div>
                  </td>
                </tr>
              ) : items.length === 0 ? (
                <tr>
                  <td
                    colSpan={getColumns().length + 1}
                    className="px-4 py-12 text-center text-muted-foreground"
                  >
                    <Settings className="h-12 w-12 mx-auto mb-3 opacity-20" />
                    <p>Belum ada data {currentTab.label.toLowerCase()}</p>
                    <p className="text-sm">Klik tombol &quot;Tambah&quot; untuk menambahkan data</p>
                  </td>
                </tr>
              ) : (
                items.map((item: Record<string, unknown>) => (
                  <tr key={item[getIdField()] as string} className="hover:bg-muted/30 transition-colors">
                    {getColumns().map((col) => (
                      <td key={col.key} className="px-4 py-3 text-sm">
                        {col.key === "aspek_weight"
                          ? `${item[col.key]}%`
                          : (item[col.key] as string) || "-"}
                      </td>
                    ))}
                    <td className="px-4 py-3 text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleEdit(item)}
                          className="h-8 w-8 p-0"
                        >
                          <Edit2 className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() =>
                            handleDelete(
                              item[getIdField()] as string,
                              item[getNameField()] as string
                            )
                          }
                          className="h-8 w-8 p-0 text-destructive hover:text-destructive"
                        >
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

        {/* Pagination */}
        {pagination.total_pages > 1 && (
          <div className="flex items-center justify-between border-t border-border px-4 py-3">
            <p className="text-sm text-muted-foreground">
              Menampilkan {(page - 1) * pagination.limit + 1} -{" "}
              {Math.min(page * pagination.limit, pagination.total)} dari {pagination.total} data
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
              >
                <ChevronLeft className="h-4 w-4 mr-1" />
                Sebelumnya
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage((p) => Math.min(pagination.total_pages, p + 1))}
                disabled={page === pagination.total_pages}
              >
                Selanjutnya
                <ChevronRight className="h-4 w-4 ml-1" />
              </Button>
            </div>
          </div>
        )}
      </div>

      {/* Create/Edit Modal */}
      <FormModal
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        title={editingItem ? `Edit ${currentTab.label}` : `Tambah ${currentTab.label}`}
        onSubmit={handleSubmit}
        isLoading={createMutation.isPending || updateMutation.isPending}
      >
        {renderFormFields()}
      </FormModal>

      {/* Delete Confirmation Modal */}
      <DeleteModal
        isOpen={!!deleteItemId}
        onClose={() => setDeleteItemId(null)}
        onConfirm={() => deleteMutation.mutate()}
        isLoading={deleteMutation.isPending}
        itemName={deleteItemName}
      />
    </div>
  );
}
