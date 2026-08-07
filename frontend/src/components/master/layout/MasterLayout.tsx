"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { ShieldAlert } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";

import { useMounted } from "@/lib/useMounted";
import { usePermissions } from "@/lib/usePermissions";
import { useAuthStore } from "@/stores/authStore";
import { useDebounce } from "@/hooks/useDebounce";

import { MASTER_TABS } from "../master.types";
import { fetchItems, createItem, updateItem, deleteItem } from "../master.api";

import { MasterTabs } from "../fragments/MasterTabs";
import { MasterToolbar } from "../fragments/MasterToolbar";
import { MasterTable } from "../fragments/MasterTable";
import { MasterFormFields } from "../fragments/MasterFormFields";
import { MasterFormModal, MasterDeleteModal } from "../fragments/MasterModals";

export function MasterLayout() {
  const user = useAuthStore((state) => state.user);
  const { hasPermission, isLoading: isGuardLoading } = usePermissions();
  const isAdmin = hasPermission("PERM-MSTR-R");
  const mounted = useMounted();
  const queryClient = useQueryClient();

  const [activeTab, setActiveTab] = useState<string>("departments");
  const [searchQuery, setSearchQuery] = useState("");
  const debouncedSearchQuery = useDebounce(searchQuery, 400);
  const [page, setPage] = useState(1);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<Record<string, unknown> | null>(null);
  const [deleteItemId, setDeleteItemId] = useState<string | null>(null);
  const [deleteItemName, setDeleteItemName] = useState("");

  const currentTab = MASTER_TABS.find((tab) => tab.id === activeTab)!;

  // Main table query — only run when user has PERM-MSTR-R access
  const { data, isLoading } = useQuery({
    queryKey: ["master", activeTab, page, debouncedSearchQuery],
    queryFn: () => fetchItems(currentTab.endpoint, page, debouncedSearchQuery),
    enabled: mounted && !!user && !isGuardLoading && isAdmin,
  });

  const items = data?.items || [];
  const pagination = data?.pagination || { total: 0, page: 1, limit: 10, total_pages: 1 };

  // Mutations
  const createMutation = useMutation({
    mutationFn: (newData: Record<string, unknown>) => createItem(currentTab.endpoint, newData),
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
    mutationFn: (newData: Record<string, unknown>) => {
      const idField = getIdField();
      const id = editingItem?.[idField] as string;
      return updateItem(currentTab.endpoint, id, newData);
    },
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

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const newData: Record<string, unknown> = {};

    formData.forEach((value, key) => {
      if (value !== "") {
        if (key === "standard_score") {
          newData[key] = parseInt(value as string, 10);
        } else {
          newData[key] = value;
        }
      }
    });

    if (activeTab === "urains" && (newData["standard_score"] === undefined || newData["standard_score"] === "")) {
      newData["standard_score"] = 2;
    }

    if (editingItem) {
      updateMutation.mutate(newData);
    } else {
      createMutation.mutate(newData);
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

  const getIdField = () => {
    const idFields: Record<string, string> = {
      plants: "plant_id",
      departments: "department_id",
      areas: "area_id",
      kawasans: "kawasan_id",
      "detail-kawasans": "detail_kawasan_id",
      aspeks: "aspek_id",
      details: "detail_id",
      urains: "uraian_id",
      habits: "habit_id",
      equipments: "equipment_id",
      infrastructures: "infrastructure_id",
    };
    return idFields[activeTab] || "id";
  };

  const getNameField = () => {
    const nameFields: Record<string, string> = {
      plants: "plant_name",
      departments: "department_name",
      areas: "area_name",
      kawasans: "kawasan_name",
      "detail-kawasans": "detail_kawasan_name",
      aspeks: "aspek_name",
      details: "detail_name",
      urains: "uraian_name",
      habits: "habit_name",
      equipments: "equipment_name",
      infrastructures: "infrastructure_name",
    };
    return nameFields[activeTab] || "name";
  };

  const getColumns = () => {
    const columns: Record<string, { key: string; label: string }[]> = {
      plants: [
        { key: "plant_code", label: "Kode Plant" },
        { key: "plant_name", label: "Nama Plant (Pabrik)" },
        { key: "address", label: "Alamat" },
      ],
      departments: [
        { key: "department_id", label: "ID" },
        { key: "department_name", label: "Nama Department" },
      ],
      areas: [
        { key: "area_id", label: "ID" },
        { key: "area_name", label: "Nama Area" },
        { key: "plant_name", label: "Plant Induk" },
      ],
      kawasans: [
        { key: "kawasan_id", label: "ID" },
        { key: "kawasan_name", label: "Nama Kawasan" },
        { key: "area_name", label: "Area Induk" },
      ],
      "detail-kawasans": [
        { key: "detail_kawasan_id", label: "ID" },
        { key: "detail_kawasan_name", label: "Nama Detail Kawasan" },
        { key: "kawasan_name", label: "Kawasan Induk" },
        { key: "kawasan_area_name", label: "Area Induk" },
      ],
      aspeks: [
        { key: "aspek_id", label: "ID" },
        { key: "aspek_name", label: "Nama Aspek" },
        { key: "area_name", label: "Area Induk" },
      ],
      details: [
        { key: "detail_id", label: "ID" },
        { key: "detail_name", label: "Nama Detail" },
        { key: "aspek_name", label: "Aspek Induk" },
        { key: "aspek_area_name", label: "Area Induk" },
      ],
      urains: [
        { key: "uraian_id", label: "ID" },
        { key: "uraian_text", label: "Teks Uraian" },
        { key: "standard_score", label: "Skor Standar" },
        { key: "detail_name", label: "Detail Induk" },
        { key: "detail_aspek_name", label: "Aspek Induk" },
      ],
      habits: [
        { key: "habit_code", label: "Kode Habit" },
        { key: "habit_name", label: "Nama Habit" },
        { key: "habit_category", label: "Kategori" },
        { key: "description", label: "Deskripsi" },
      ],
      equipments: [
        { key: "equipment_code", label: "Kode Equipment" },
        { key: "equipment_name", label: "Nama Equipment" },
        { key: "equipment_type", label: "Tipe" },
        { key: "kawasan_name", label: "Kawasan Induk" },
      ],
      infrastructures: [
        { key: "infrastructure_code", label: "Kode Infrastructure" },
        { key: "infrastructure_name", label: "Nama Infrastructure" },
        { key: "infrastructure_type", label: "Tipe" },
        { key: "kawasan_name", label: "Kawasan Induk" },
      ],
    };
    return columns[activeTab] || [];
  };

  if (!isGuardLoading && !isAdmin) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] py-12 px-4 space-y-4 text-center w-full max-w-lg mx-auto">
        <div className="w-16 h-16 rounded-full bg-destructive/10 flex items-center justify-center shrink-0">
          <ShieldAlert className="h-8 w-8 text-destructive" />
        </div>
        <h2 className="text-xl font-semibold text-foreground">Akses Ditolak</h2>
        <p className="text-sm text-muted-foreground text-center leading-relaxed w-full">
          Anda tidak memiliki izin untuk mengakses halaman Master Data. Hanya administrator yang dapat mengakses halaman ini.
        </p>
        {mounted && user && (
          <Link href={`/cimory/${user.plant_id || 'global'}/dashboard/${user.id || (user as any).user_id}`}>
            <Button variant="outline" className="rounded-xl px-6">Kembali ke Dashboard</Button>
          </Link>
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

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold">Master Data</h1>
        <p className="text-muted-foreground">Kelola data master sistem audit</p>
      </div>

      <MasterTabs
        activeTab={activeTab}
        userRole={user?.role_id}
        onTabChange={(tabId) => {
          setActiveTab(tabId);
          setPage(1);
          setSearchQuery("");
        }}
      />

      <MasterToolbar
        currentTabLabel={currentTab.label}
        searchQuery={searchQuery}
        onSearchChange={(query) => {
          setSearchQuery(query);
          setPage(1);
        }}
        onAddClick={() => setIsModalOpen(true)}
      />

      <MasterTable
        items={items}
        columns={getColumns()}
        isLoading={isLoading}
        currentTabLabel={currentTab.label}
        idField={getIdField()}
        nameField={getNameField()}
        pagination={pagination}
        onPageChange={setPage}
        onEdit={handleEdit}
        onDelete={handleDelete}
      />

      <MasterFormModal
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        title={editingItem ? `Edit ${currentTab.label}` : `Tambah ${currentTab.label}`}
        onSubmit={handleSubmit}
        isLoading={createMutation.isPending || updateMutation.isPending}
      >
        <MasterFormFields activeTab={activeTab} editingItem={editingItem} />
      </MasterFormModal>

      <MasterDeleteModal
        isOpen={!!deleteItemId}
        onClose={() => setDeleteItemId(null)}
        onConfirm={() => deleteMutation.mutate()}
        isLoading={deleteMutation.isPending}
        itemName={deleteItemName}
      />
    </div>
  );
}
