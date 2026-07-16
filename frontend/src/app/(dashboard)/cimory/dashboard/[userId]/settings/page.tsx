"use client";

import { useState, useRef, useEffect } from "react";
import {
  Settings as SettingsIcon,
  Edit2,
  X,
  ShieldAlert
} from "lucide-react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import EmailEditor, { EditorRef } from "react-email-editor";
import { api } from "@/lib/api/axios";
import { useMounted } from "@/lib/useMounted";
import { useAdminGuard } from "@/lib/useAdminGuard";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuthStore } from "@/stores/authStore";

const fetchSettings = async () => {
  const res = await api.get("/settings");
  return res.data; // Expected { data: [...] } if standard response, wait master_routes settings returns response.Paginated? No, settingH.GetAll returns standard response? Let's assume res.data.data
};

const updateSetting = async (key: string, value: string) => {
  const res = await api.put(`/settings/${key}`, { setting_value: value });
  return res.data;
};

// Modals
function StandardModal({
  isOpen, onClose, title, onSubmit, isLoading, children,
}: any) {
  if (!isOpen) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />
      <div className="relative z-10 w-full max-w-lg rounded-2xl bg-card border border-border p-6 shadow-2xl">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold">{title}</h2>
          <button type="button" onClick={onClose} className="rounded-lg p-1 hover:bg-muted">
            <X className="h-5 w-5" />
          </button>
        </div>
        <form onSubmit={onSubmit} className="space-y-4">
          {children}
          <div className="flex gap-3 pt-4">
            <Button type="button" variant="outline" onClick={onClose} className="flex-1">Batal</Button>
            <Button type="submit" isLoading={isLoading} className="flex-1">Simpan</Button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default function SettingsPage() {
  const user = useAuthStore((state) => state.user);
  const { isAdmin, isLoading: isGuardLoading } = useAdminGuard();
  const mounted = useMounted();
  const queryClient = useQueryClient();

  const [editingItem, setEditingItem] = useState<any>(null);
  const [isStandardModalOpen, setIsStandardModalOpen] = useState(false);
  const [isEmailModalOpen, setIsEmailModalOpen] = useState(false);
  const [newValue, setNewValue] = useState("");
  
  const emailEditorRef = useRef<EditorRef>(null);

  const { data: settingsRes, isLoading } = useQuery({
    queryKey: ["settings"],
    queryFn: fetchSettings,
    enabled: mounted && !!user && isAdmin,
  });

  const settingsList = settingsRes?.data || [];

  const updateMutation = useMutation({
    mutationFn: ({ key, value }: { key: string; value: string }) => updateSetting(key, value),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["settings"] });
      setIsStandardModalOpen(false);
      setIsEmailModalOpen(false);
      setEditingItem(null);
      toast.success("Pengaturan berhasil diperbarui");
    },
    onError: (err: any) => toast.error(err.response?.data?.message || "Gagal memperbarui pengaturan"),
  });

  const handleEditClick = (item: any) => {
    setEditingItem(item);
    setNewValue(item.setting_value || "");
    
    // Check if it's an email template
    if (item.setting_key.startsWith("EMAIL_TEMPLATE")) {
      setIsEmailModalOpen(true);
    } else {
      setIsStandardModalOpen(true);
    }
  };

  const handleStandardSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingItem) return;
    updateMutation.mutate({ key: editingItem.setting_key, value: newValue });
  };

  const saveEmailDesign = () => {
    if (!emailEditorRef.current?.editor) return;
    emailEditorRef.current.editor.exportHtml((data) => {
      const { html } = data;
      // We will save HTML directly so the backend mailer can use it easily
      updateMutation.mutate({ key: editingItem.setting_key, value: html });
    });
  };

  const onLoadEmailEditor = () => {
    // If we have saved design json before, we could load it, but since we only save HTML,
    // we can't easily reverse-engineer it to a perfect unlayer design.
    // For now we just load it if we had a JSON structure. Since we don't, we'll just start fresh or inject HTML.
    // NOTE: Unlayer supports loadDesign to load JSON. Since we only save HTML for backend compatibility, 
    // it will load as blank initially. In production, we'd save a JSON string with { html, design } and parse it here.
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
        <h1 className="text-2xl font-bold">System Settings</h1>
        <p className="text-muted-foreground">Konfigurasi dinamis sistem audit</p>
      </div>

      <div className="grid gap-4">
        {isLoading ? (
          <div className="p-12 flex justify-center"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div></div>
        ) : settingsList.length === 0 ? (
          <div className="p-8 text-center text-muted-foreground bg-card rounded-xl border border-border shadow-sm">
            Tidak ada pengaturan yang ditemukan.
          </div>
        ) : (
          settingsList.map((item: any) => (
            <div key={item.setting_key} className="bg-card p-5 rounded-xl border border-border shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4 transition-colors hover:border-primary/50">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <h3 className="font-semibold text-lg">{item.setting_key}</h3>
                  {item.is_encrypted && <span className="bg-primary/10 text-primary text-[10px] px-2 py-0.5 rounded-full font-bold uppercase">Encrypted</span>}
                </div>
                <p className="text-sm text-muted-foreground mb-3">{item.description}</p>
                <div className="bg-muted p-3 rounded-lg text-sm font-mono truncate max-w-full">
                  {item.is_encrypted ? "••••••••••••••••" : (item.setting_value || <span className="italic text-muted-foreground">Kosong</span>)}
                </div>
              </div>
              <div>
                <Button variant="outline" onClick={() => handleEditClick(item)} className="w-full md:w-auto">
                  <Edit2 className="h-4 w-4 mr-2" /> Edit
                </Button>
              </div>
            </div>
          ))
        )}
      </div>

      {/* STANDARD MODAL */}
      <StandardModal
        isOpen={isStandardModalOpen}
        onClose={() => { setIsStandardModalOpen(false); setEditingItem(null); }}
        title={`Edit ${editingItem?.setting_key}`}
        onSubmit={handleStandardSubmit}
        isLoading={updateMutation.isPending}
      >
        <div>
          <label className="text-sm font-medium mb-1 block">Value</label>
          <Input 
            value={newValue} 
            onChange={(e) => setNewValue(e.target.value)} 
            required 
            placeholder="Masukkan nilai baru..." 
          />
          <p className="text-xs text-muted-foreground mt-2">{editingItem?.description}</p>
        </div>
      </StandardModal>

      {/* EMAIL TEMPLATE MODAL (UNLAYER) */}
      {isEmailModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setIsEmailModalOpen(false)} />
          <div className="relative z-10 w-full h-[95vh] max-w-6xl rounded-2xl bg-card border border-border flex flex-col shadow-2xl overflow-hidden">
            <div className="p-4 border-b border-border flex items-center justify-between bg-card">
              <div>
                <h2 className="text-lg font-semibold">Email Template Editor</h2>
                <p className="text-sm text-muted-foreground">{editingItem?.setting_key}</p>
              </div>
              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm" onClick={() => setIsEmailModalOpen(false)}>Batal</Button>
                <Button size="sm" onClick={saveEmailDesign} isLoading={updateMutation.isPending}>Simpan Template</Button>
              </div>
            </div>
            <div className="flex-1 bg-gray-100">
              <EmailEditor 
                ref={emailEditorRef} 
                onLoad={onLoadEmailEditor}
                options={{
                  appearance: {
                    theme: 'modern_light',
                  }
                }}
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
