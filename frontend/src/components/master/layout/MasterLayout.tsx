"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { ShieldAlert } from "lucide-react";
import { Button } from "@/components/ui/button";

import { useMounted } from "@/lib/useMounted";
import { useAdminGuard } from "@/lib/useAdminGuard";
import { useAuthStore } from "@/stores/authStore";

import { MASTER_TABS } from "../master.types";
import { fetchItems, createItem, updateItem, deleteItem } from "../master.api";

import { MasterTabs } from "../fragments/MasterTabs";
import { MasterToolbar } from "../fragments/MasterToolbar";
import { MasterTable } from "../fragments/MasterTable";
import { MasterFormFields } from "../fragments/MasterFormFields";
import { MasterFormModal, MasterDeleteModal } from "../fragments/MasterModals";

export function MasterLayout() {
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

  const currentTab = MASTER_TABS.find((tab) => tab.id === activeTab)!;

  // Main table query
  const { data, isLoading } = useQuery({
    queryKey: ["master", activeTab, page, searchQuery],
    queryFn: () => fetchItems(currentTab.endpoint, page, searchQuery),
    enabled: mounted && !!user,
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

  const getColumns = () => {
    const columns: Record<string, { key: string; label: string }[]> = {
      departments: [
        { key: "department_id", label: "ID" },
        { key: "department_name", label: "Nama" },
      ],
      areas: [
        { key: "area_id", label: "ID" },
        { key: "area_name", label: "Nama" },
      ],
      kawasans: [
        { key: "kawasan_id", label: "ID" },
        { key: "kawasan_name", label: "Nama" },
      ],
      "detail-kawasans": [
        { key: "detail_kawasan_id", label: "ID" },
        { key: "detail_kawasan_name", label: "Nama" },
      ],
      aspeks: [
        { key: "aspek_id", label: "ID" },
        { key: "aspek_name", label: "Nama" },
      ],
      details: [
        { key: "detail_id", label: "ID" },
        { key: "detail_name", label: "Nama" },
      ],
      urains: [
        { key: "uraian_id", label: "ID" },
        { key: "uraian_text", label: "Uraian" },
      ],
    };
    return columns[activeTab] || [];
  };

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

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold">Master Data</h1>
        <p className="text-muted-foreground">Kelola data master sistem audit</p>
      </div>

      <MasterTabs
        activeTab={activeTab}
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
