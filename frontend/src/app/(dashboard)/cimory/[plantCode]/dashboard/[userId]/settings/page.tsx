"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/axios";
import { useAuthStore } from "@/stores/authStore";
import { useAdminGuard } from "@/lib/useAdminGuard";
import { useMounted } from "@/lib/useMounted";
import { ShieldAlert, Mail, Settings as SettingsIcon, PenSquare, Save, Key } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { EmailEditorModal } from "./EmailEditorModal";
import { useSettingsStore } from "@/stores/settingsStore";
import { ApiKeyManager } from "@/components/settings/ApiKeyManager";

// The keys we care about for email templates
const EMAIL_TEMPLATES = [
  { key: "EMAIL_TEMPLATE_FORGOT_PASSWORD", title: "Lupa Password", description: "Email yang dikirim saat user meminta reset password." },
  { key: "EMAIL_TEMPLATE_ISSUE_ASSIGNMENT", title: "Penugasan Temuan (Issue)", description: "Email notifikasi saat user ditugaskan memperbaiki suatu temuan." },
  { key: "EMAIL_TEMPLATE_INSPECTION_CONFIRMED", title: "Konfirmasi Inspeksi", description: "Email saat jadwal inspeksi telah disetujui/dikonfirmasi." },
];

const GENERAL_SETTINGS = [
  { group: "Keamanan & Akses", keys: [
    { key: "MAX_LOGIN_ATTEMPTS", label: "Maksimal Percobaan Login", type: "number", desc: "Jumlah maksimal percobaan sebelum akun terkunci sementara." },
    { key: "SESSION_IDLE_TIMEOUT_MINUTES", label: "Batas Waktu Sesi (Menit)", type: "number", desc: "Waktu tidak aktif sebelum pengguna dikeluarkan (logout) otomatis." },
  ]},
  { group: "Sistem & Penyimpanan", keys: [
    { key: "MAX_UPLOAD_SIZE_MB", label: "Maksimal Ukuran Unggahan (MB)", type: "number", desc: "Batas ukuran maksimal untuk setiap file yang diunggah ke sistem." },
    { key: "MINIO_ALLOWED_IPS", label: "IP yang Diizinkan untuk Penyimpanan", type: "text", desc: "Daftar IP yang diizinkan mengakses storage Minio (pisahkan dengan koma)." },
  ]},
  { group: "Tenggat Waktu Temuan (Issue)", keys: [
    { key: "ISSUE_DEADLINE_DAYS", label: "Tenggat Waktu Penyelesaian (Hari)", type: "number", desc: "Waktu default yang diberikan untuk menyelesaikan sebuah temuan." },
    { key: "ISSUE_AUTO_APPROVE_DAYS", label: "Waktu Auto-Approve (Hari)", type: "number", desc: "Waktu sebelum perbaikan temuan disetujui otomatis jika tidak diulas." },
  ]}
];

export default function SettingsPage() {
  const user = useAuthStore((state) => state.user);
  const { isAdmin, isLoading: isGuardLoading } = useAdminGuard();
  const mounted = useMounted();
  const queryClient = useQueryClient();
  const { showSearchLatencyButton, toggleSearchLatencyButton } = useSettingsStore();

  const [activeTab, setActiveTab] = useState<"general" | "email" | "apikey">("general");
  
  // General Settings Form State
  const [generalValues, setGeneralValues] = useState<Record<string, string>>({});
  
  // Email Editor State
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorKey, setEditorKey] = useState("");
  const [editorTitle, setEditorTitle] = useState("");

  // Fetch all settings
  const { data: settingsData, isLoading: settingsLoading } = useQuery({
    queryKey: ["settings"],
    queryFn: async () => {
      const res = await api.get("/master/settings");
      
      // Initialize general values
      const fetched = res.data?.data || [];
      const initialGen: Record<string, string> = {};
      fetched.forEach((s: any) => {
        initialGen[s.setting_key] = s.setting_value;
      });
      setGeneralValues(prev => ({ ...initialGen, ...prev }));
      
      return fetched;
    },
    enabled: mounted && !!user && isAdmin,
  });

  // Save Mutation
  const saveSettingMutation = useMutation({
    mutationFn: async ({ key, value }: { key: string, value: string }) => {
      return api.put(`/master/settings/${key}`, { setting_value: value });
    },
    onSuccess: () => {
      toast.success("Pengaturan berhasil disimpan");
      queryClient.invalidateQueries({ queryKey: ["settings"] });
      setEditorOpen(false);
    },
    onError: (err: any) => {
      toast.error(err.response?.data?.message || "Gagal menyimpan pengaturan");
    }
  });

  const handleOpenEditor = (template: any) => {
    setEditorKey(template.key);
    setEditorTitle(`Edit Template: ${template.title}`);
    setEditorOpen(true);
  };

  // Derive initial data so it's always up to date even if settingsData updates in the background
  const activeSetting = settingsData?.find((s: any) => s.setting_key === editorKey);
  const derivedInitialData = activeSetting?.setting_value || "";

  const handleSaveEmailTemplate = (html: string, design: any) => {
    // Combine HTML and JSON design into a single JSON string
    const payload = JSON.stringify({ html, design });
    saveSettingMutation.mutate({ key: editorKey, value: payload });
  };

  const handleSaveGeneral = (e: React.FormEvent) => {
    e.preventDefault();
    // Save all general settings sequentially
    const promises = Object.keys(generalValues).map(key => 
      api.put(`/master/settings/${key}`, { setting_value: generalValues[key] })
    );

    toast.promise(Promise.all(promises), {
      loading: 'Menyimpan pengaturan...',
      success: () => {
        queryClient.invalidateQueries({ queryKey: ["settings"] });
        return 'Semua pengaturan berhasil disimpan!';
      },
      error: 'Gagal menyimpan beberapa pengaturan',
    });
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
    <div className="space-y-6 pb-20">
      <div>
        <h1 className="text-2xl font-bold">Pengaturan Sistem</h1>
        <p className="text-muted-foreground mt-1">Kelola konfigurasi umum dan template email sistem.</p>
      </div>

      <div className="flex flex-col md:flex-row gap-6">
        {/* Sidebar Tabs */}
        <div className="w-full md:w-64 shrink-0 space-y-2">
          <button
            onClick={() => setActiveTab("general")}
            className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-colors font-medium text-left ${activeTab === "general" ? "bg-primary text-primary-foreground" : "hover:bg-muted text-muted-foreground hover:text-foreground"}`}
          >
            <SettingsIcon className="w-5 h-5" />
            Umum
          </button>
          <button
            onClick={() => setActiveTab("email")}
            className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-colors font-medium text-left ${activeTab === "email" ? "bg-primary text-primary-foreground" : "hover:bg-muted text-muted-foreground hover:text-foreground"}`}
          >
            <Mail className="w-5 h-5" />
            Template Email
          </button>
          <button
            onClick={() => setActiveTab("apikey")}
            className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-colors font-medium text-left ${activeTab === "apikey" ? "bg-primary text-primary-foreground" : "hover:bg-muted text-muted-foreground hover:text-foreground"}`}
          >
            <Key className="w-5 h-5" />
            API Key Power BI
          </button>
        </div>

        {/* Content Area */}
        <div className="flex-1">
          <div className="bg-card border border-border rounded-2xl shadow-sm overflow-hidden">
            
            {activeTab === "general" && (
              <div className="p-6">
                <div className="border-b border-border pb-4 mb-6">
                  <h3 className="text-lg font-semibold">Pengaturan Umum</h3>
                  <p className="text-sm text-muted-foreground">Konfigurasi umum seperti Batas Login, Timeout, dan Tenggat Waktu.</p>
                </div>
                
                {settingsLoading ? (
                  <div className="p-12 flex justify-center items-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                  </div>
                ) : (
                  <form onSubmit={handleSaveGeneral} className="space-y-8">
                    {GENERAL_SETTINGS.map((group, idx) => (
                      <div key={idx} className="bg-muted/30 p-5 rounded-xl border border-border">
                        <h4 className="font-semibold mb-4 text-primary">{group.group}</h4>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                          {group.keys.map(item => (
                            <div key={item.key}>
                              <label className="text-sm font-medium mb-1 block">{item.label}</label>
                              <Input 
                                type={item.type} 
                                value={generalValues[item.key] || ""} 
                                onChange={(e) => setGeneralValues(prev => ({...prev, [item.key]: e.target.value}))} 
                                required 
                              />
                              <p className="text-[11px] text-muted-foreground mt-1.5">{item.desc}</p>
                            </div>
                          ))}
                        </div>
                      </div>
                    ))}

                    {/* Fitur Pengujian & Performa Switch Toggle */}
                    <div className="bg-muted/30 p-5 rounded-xl border border-border">
                      <h4 className="font-semibold mb-4 text-primary flex items-center gap-2">
                        Pengaturan Pengujian & Performa
                      </h4>
                      <div className="flex items-center justify-between p-4 bg-card rounded-xl border border-border/80 shadow-sm gap-4">
                        <div className="space-y-0.5">
                          <label className="text-sm font-semibold text-foreground block">
                            Tombol Uji Latensi Search
                          </label>
                          <p className="text-xs text-muted-foreground">
                            Aktifkan atau nonaktifkan tombol <strong>Uji Latensi Search</strong> di bilah pencarian (Inspeksi, Temuan, GMP Data, & Logs).
                          </p>
                        </div>
                        <button
                          type="button"
                          role="switch"
                          aria-checked={showSearchLatencyButton}
                          onClick={() => {
                            toggleSearchLatencyButton();
                            toast.success(
                              !showSearchLatencyButton
                                ? "Tombol Uji Latensi Search Diaktifkan"
                                : "Tombol Uji Latensi Search Dinonaktifkan"
                            );
                          }}
                          className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                            showSearchLatencyButton ? "bg-primary" : "bg-muted-foreground/30"
                          }`}
                        >
                          <span
                            className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow-lg ring-0 transition duration-200 ease-in-out ${
                              showSearchLatencyButton ? "translate-x-5" : "translate-x-0"
                            }`}
                          />
                        </button>
                      </div>
                    </div>

                    <div className="flex justify-end pt-4 border-t border-border">
                      <Button type="submit">
                        <Save className="w-4 h-4 mr-2" /> Simpan Pengaturan Umum
                      </Button>
                    </div>
                  </form>
                )}
              </div>
            )}

            {activeTab === "email" && (
              <div>
                <div className="p-6 border-b border-border bg-muted/20">
                  <h3 className="text-lg font-semibold">Template Email</h3>
                  <p className="text-sm text-muted-foreground">Sesuaikan tampilan dan konten email yang dikirim otomatis oleh sistem.</p>
                </div>
                
                {settingsLoading ? (
                  <div className="p-12 flex justify-center items-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                  </div>
                ) : (
                  <div className="divide-y divide-border">
                    {EMAIL_TEMPLATES.map((template) => {
                      const existing = settingsData?.find((s: any) => s.setting_key === template.key);
                      const isConfigured = !!existing?.setting_value;

                      return (
                        <div key={template.key} className="p-6 flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between hover:bg-muted/10 transition-colors">
                          <div>
                            <div className="flex items-center gap-2 mb-1">
                              <h4 className="font-semibold">{template.title}</h4>
                              {isConfigured ? (
                                <span className="bg-green-500/10 text-green-500 text-[10px] uppercase font-bold px-2 py-0.5 rounded">Terkonfigurasi</span>
                              ) : (
                                <span className="bg-orange-500/10 text-orange-500 text-[10px] uppercase font-bold px-2 py-0.5 rounded">Bawaan Sistem</span>
                              )}
                            </div>
                            <p className="text-sm text-muted-foreground">{template.description}</p>
                            <p className="text-xs text-muted-foreground mt-2 font-mono bg-muted inline-block px-2 py-1 rounded">Key: {template.key}</p>
                          </div>
                          
                          <Button variant={isConfigured ? "outline" : "default"} onClick={() => handleOpenEditor(template)} className="shrink-0">
                            <PenSquare className="w-4 h-4 mr-2" />
                            {isConfigured ? "Edit Template" : "Buat Template"}
                          </Button>
                        </div>
                      )
                    })}
                  </div>
                )}
              </div>
            )}

            {activeTab === "apikey" && (
              <div className="p-6">
                <ApiKeyManager />
              </div>
            )}
          </div>
        </div>
      </div>

      <EmailEditorModal
        isOpen={editorOpen}
        onClose={() => setEditorOpen(false)}
        title={editorTitle}
        initialData={derivedInitialData}
        onSave={handleSaveEmailTemplate}
        isSaving={saveSettingMutation.isPending}
      />
    </div>
  );
}
