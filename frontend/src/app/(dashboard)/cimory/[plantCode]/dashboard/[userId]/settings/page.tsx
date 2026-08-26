"use client";

import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { getApiErrorMessage } from "@/lib/api/error";
import { api } from "@/lib/api/axios";
import { useAuthStore } from "@/stores/authStore";
import { usePermissions } from "@/lib/usePermissions";
import { useMounted } from "@/lib/useMounted";
import { ShieldAlert, Mail, Settings as SettingsIcon, PenSquare, Save, Key, Server, Eye, EyeOff } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { EmailEditorModal } from "./EmailEditorModal";
import { useSettingsStore } from "@/stores/settingsStore";
import { ApiKeyManager } from "@/components/settings/ApiKeyManager";

import { masterApi, type SystemSetting } from "@/lib/api/master.api";
import { ChevronDown, Filter } from "lucide-react";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { setActiveTab, setPlantFilter } from "@/store/slices/settingsSlice";

// The keys we care about for email templates
const EMAIL_TEMPLATES = [
  { key: "EMAIL_TEMPLATE_FORGOT_PASSWORD", title: "Lupa Password", description: "Email yang dikirim saat user meminta reset password.", variables: ["FullName", "Username", "TempPassword"] },
  { key: "EMAIL_TEMPLATE_ISSUE_ASSIGNMENT", title: "Penugasan Temuan (Issue)", description: "Email notifikasi saat user ditugaskan memperbaiki suatu temuan.", variables: ["PICName", "IssueID", "Keterangan", "Status", "DueDate"] },
  { key: "EMAIL_TEMPLATE_INSPECTION_CONFIRMED", title: "Konfirmasi Inspeksi", description: "Email saat inspeksi area telah selesai dan dikonfirmasi.", variables: ["AreaID", "TotalDetailKawasan", "CompletedDetailKawasan", "Status"] },
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
    { key: "ISSUE_DEADLINE_DAYS", label: "Tenggat Waktu Penyelesaian (Hari)", type: "number", desc: "Waktu default yang diberikan kepada PIC untuk menyelesaikan sebuah temuan sejak issue dibuat, sebelum jadi 'Overdue'." },
    { key: "ISSUE_AUTO_APPROVE_DAYS", label: "Waktu Auto-Approve Follow-Up (Hari)", type: "number", desc: "Waktu tunggu auditor untuk meninjau follow-up perbaikan temuan (foto bukti perbaikan) sebelum disetujui otomatis oleh sistem." },
    { key: "ISSUE_WOWR_AUTO_APPROVE_DAYS", label: "Waktu Auto-Approve WO/WR (Hari)", type: "number", desc: "Waktu tunggu auditor untuk memvalidasi bukti WO/WR yang diunggah PIC sebelum disetujui otomatis oleh sistem. Terpisah dari waktu auto-approve follow-up di atas." },
  ]}
];

type EmailTemplateConfig = (typeof EMAIL_TEMPLATES)[number];

const SMTP_DEFAULTS: Record<string, string> = {
  SMTP_ENABLED: "true",
  SMTP_HOST: "",
  SMTP_PORT: "",
  SMTP_USER: "",
  SMTP_PASSWORD: "",
  SMTP_SENDER_EMAIL: "",
};

export default function SettingsPage() {
  const user = useAuthStore((state) => state.user);
  const { hasPermission, isLoading: isGuardLoading } = usePermissions();
  const isAdmin = hasPermission("PERM-MSTR-R");
  const mounted = useMounted();
  const queryClient = useQueryClient();
  const {
    showSearchLatencyButton,
    toggleSearchLatencyButton,
    showCoreWebVitalsMonitor,
    toggleCoreWebVitalsMonitor,
  } = useSettingsStore();

  const isSuperAdmin = user?.role_id === "ROLE-000" || user?.role_id === "SUPERADMIN" || user?.role?.role_name === "Super Admin";

  const dispatch = useAppDispatch();
  const { activeTab, plantFilter } = useAppSelector((state) => state.settings);
  const smtpPlantID = isSuperAdmin
    ? (plantFilter && plantFilter !== "ALL" ? plantFilter : "GLOBAL")
    : (user?.plant_id || "GLOBAL");
  const settingsPlantID = smtpPlantID;
  
  useEffect(() => {
    if (!isSuperAdmin && user?.plant_id) {
      dispatch(setPlantFilter(user.plant_id));
    }
  }, [isSuperAdmin, user?.plant_id, dispatch]);

  // General Settings Form State
  const [generalValues, setGeneralValues] = useState<Record<string, string>>({});
  const [smtpValues, setSmtpValues] = useState<Record<string, string>>(SMTP_DEFAULTS);
  const [smtpPasswordConfigured, setSmtpPasswordConfigured] = useState(false);
  const [smtpHasPlantOverride, setSmtpHasPlantOverride] = useState(false);
  const [showSmtpPassword, setShowSmtpPassword] = useState(false);
  const [isSavingSmtp, setIsSavingSmtp] = useState(false);
  
  // Email Editor State
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorKey, setEditorKey] = useState("");
  const [editorTitle, setEditorTitle] = useState("");

  // Fetch plants
  const { data: plantsList = [] } = useQuery({
    queryKey: ["master-plants"],
    queryFn: () => masterApi.getPlants({ limit: 1000 }),
    enabled: mounted && !!user && isAdmin,
  });

  // Fetch all settings
  const { data: settingsData, isLoading: settingsLoading } = useQuery({
    queryKey: ["settings", settingsPlantID],
    queryFn: async () => {
      const params = { plant_id: settingsPlantID };
      const res = await api.get<{ data: SystemSetting[] }>("/master/settings", { params });
      
      // Initialize general values
      const fetched = res.data?.data || [];
      const initialGen: Record<string, string> = {};
      fetched.forEach((s) => {
        initialGen[s.setting_key] = s.setting_value;
      });
      setGeneralValues(initialGen);
      
      return fetched;
    },
    enabled: mounted && !!user && isAdmin,
  });

  const { isLoading: smtpLoading } = useQuery({
    queryKey: ["settings-smtp", smtpPlantID],
    queryFn: async () => {
      const res = await api.get<{ data: SystemSetting[] }>("/master/settings", { params: { plant_id: smtpPlantID } });
      const fetched = res.data?.data || [];
      const nextValues = { ...SMTP_DEFAULTS };
      let passwordConfigured = false;
      fetched.forEach((setting) => {
        if (!(setting.setting_key in nextValues)) return;
        if (setting.setting_key === "SMTP_PASSWORD") {
          passwordConfigured = setting.setting_value === "***";
          nextValues.SMTP_PASSWORD = "";
          return;
        }
        nextValues[setting.setting_key] = setting.setting_value;
      });
      setSmtpValues(nextValues);
      setSmtpPasswordConfigured(passwordConfigured);
      setSmtpHasPlantOverride(
        smtpPlantID !== "GLOBAL" && fetched.some((setting) =>
          setting.setting_key in SMTP_DEFAULTS && setting.plant_id === smtpPlantID
        )
      );
      return fetched;
    },
    enabled: mounted && !!user && isAdmin,
  });

  // Save Mutation
  const saveSettingMutation = useMutation({
    mutationFn: async ({ key, value }: { key: string, value: string }) => {
      const params = { plant_id: settingsPlantID };
      return api.put(`/master/settings/${key}`, { setting_value: value, plant_id: settingsPlantID }, { params });
    },
    onSuccess: () => {
      toast.success("Pengaturan berhasil disimpan");
      queryClient.invalidateQueries({ queryKey: ["settings"] });
      setEditorOpen(false);
    },
    onError: (err) => {
      toast.error(getApiErrorMessage(err, "Gagal menyimpan pengaturan"));
    }
  });

  const handleOpenEditor = (template: EmailTemplateConfig) => {
    setEditorKey(template.key);
    setEditorTitle(`Edit Template: ${template.title}`);
    setEditorOpen(true);
  };

  // Derive initial data so it's always up to date even if settingsData updates in the background
  const activeSetting = settingsData?.find((s) => s.setting_key === editorKey);
  const derivedInitialData = activeSetting?.setting_value || "";

  const handleSaveEmailTemplate = (html: string, design: unknown) => {
    // Keep the editor design for future edits. The backend extracts only HTML
    // from this versioned envelope before rendering and sending the email.
    const payload = JSON.stringify({ version: 1, html, design });
    saveSettingMutation.mutate({ key: editorKey, value: payload });
  };

  const handleSaveGeneral = (e: React.FormEvent) => {
    e.preventDefault();
    const params = { plant_id: settingsPlantID };
    // Save all general settings sequentially
    const generalKeys = GENERAL_SETTINGS.flatMap(group => group.keys.map(item => item.key));
    const promises = generalKeys.map(key =>
      api.put(`/master/settings/${key}`, { setting_value: generalValues[key], plant_id: settingsPlantID }, { params })
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

  const handleSaveSMTP = async (e: React.FormEvent) => {
    e.preventDefault();
    const smtpEnabled = smtpValues.SMTP_ENABLED === "true";
    const port = Number(smtpValues.SMTP_PORT);
    if (smtpEnabled && (!smtpValues.SMTP_HOST.trim() || !Number.isInteger(port) || port < 1 || port > 65535)) {
      toast.error("Host SMTP dan port 1–65535 wajib diisi");
      return;
    }
    if (smtpEnabled && !/^\S+@\S+\.\S+$/.test(smtpValues.SMTP_SENDER_EMAIL.trim())) {
      toast.error("Email pengirim SMTP tidak valid");
      return;
    }

    setIsSavingSmtp(true);
    try {
      const keys = smtpEnabled
        ? ["SMTP_ENABLED", "SMTP_HOST", "SMTP_PORT", "SMTP_USER", "SMTP_SENDER_EMAIL"]
        : ["SMTP_ENABLED"];
      if (smtpEnabled && smtpValues.SMTP_PASSWORD.trim()) keys.push("SMTP_PASSWORD");
      await Promise.all(keys.map(key => api.put(
        `/master/settings/${key}`,
        { setting_value: smtpValues[key], plant_id: smtpPlantID },
        { params: { plant_id: smtpPlantID } }
      )));
      toast.success(`Konfigurasi SMTP ${smtpPlantID === "GLOBAL" ? "global" : "plant"} berhasil disimpan dan langsung aktif`);
      setSmtpValues(prev => ({ ...prev, SMTP_PASSWORD: "" }));
      if (keys.includes("SMTP_PASSWORD")) setSmtpPasswordConfigured(true);
      if (smtpPlantID !== "GLOBAL") setSmtpHasPlantOverride(true);
      queryClient.invalidateQueries({ queryKey: ["settings-smtp"] });
      queryClient.invalidateQueries({ queryKey: ["settings"] });
    } catch (error) {
      toast.error(getApiErrorMessage(error, "Gagal menyimpan konfigurasi SMTP"));
    } finally {
      setIsSavingSmtp(false);
    }
  };

  if (!isGuardLoading && !isAdmin) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] py-12 px-4 space-y-4 text-center w-full max-w-lg mx-auto">
        <div className="w-16 h-16 rounded-full bg-destructive/10 flex items-center justify-center shrink-0">
          <ShieldAlert className="h-8 w-8 text-destructive" />
        </div>
        <h2 className="text-xl font-semibold text-foreground">Akses Ditolak</h2>
        <p className="text-sm text-muted-foreground text-center leading-relaxed w-full">
          Anda tidak memiliki izin untuk mengelola Pengaturan Sistem. Halaman ini khusus untuk Administrator.
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
          <h1 className="text-2xl font-bold">Pengaturan Sistem</h1>
          <p className="text-muted-foreground mt-1">Kelola konfigurasi umum dan template email sistem per plant.</p>
        </div>

        {isSuperAdmin && (
          <div className="flex items-center gap-2">
            <Filter className="h-4 w-4 text-muted-foreground" />
            <div className="relative">
              <select
                className="text-sm bg-background appearance-none border border-border rounded-xl pl-3 pr-8 py-2 outline-none h-10 font-medium"
                value={plantFilter}
                onChange={(e) => dispatch(setPlantFilter(e.target.value))}
              >
                <option value="ALL">Global (Default semua plant)</option>
                {plantsList.map((p) => (
                  <option key={p.plant_id} value={p.plant_id}>{p.plant_name}</option>
                ))}
              </select>
              <ChevronDown className="absolute right-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
            </div>
          </div>
        )}
      </div>

      <div className="flex flex-col md:flex-row gap-6">
        {/* Sidebar Tabs */}
        <div className="w-full md:w-64 shrink-0 space-y-2">
          <button
            onClick={() => dispatch(setActiveTab("general"))}
            className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-colors font-medium text-left ${activeTab === "general" ? "bg-primary text-primary-foreground" : "hover:bg-muted text-muted-foreground hover:text-foreground"}`}
          >
            <SettingsIcon className="w-5 h-5" />
            Umum
          </button>
          <button
            onClick={() => dispatch(setActiveTab("smtp"))}
            className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-colors font-medium text-left ${activeTab === "smtp" ? "bg-primary text-primary-foreground" : "hover:bg-muted text-muted-foreground hover:text-foreground"}`}
          >
            <Server className="w-5 h-5" />
            Konfigurasi SMTP
          </button>
          <button
            onClick={() => dispatch(setActiveTab("email"))}
            className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-colors font-medium text-left ${activeTab === "email" ? "bg-primary text-primary-foreground" : "hover:bg-muted text-muted-foreground hover:text-foreground"}`}
          >
            <Mail className="w-5 h-5" />
            Template Email
          </button>
          <button
            onClick={() => dispatch(setActiveTab("apikey"))}
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
                    <div className="bg-muted/30 p-5 rounded-xl border border-border space-y-4">
                      <h4 className="font-semibold text-primary flex items-center gap-2">
                        Pengaturan Pengujian & Performa (Khusus Admin & SuperAdmin)
                      </h4>

                      {/* Switch 1: Search Latency */}
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

                      {/* Switch 2: Core Web Vitals Monitor */}
                      <div className="flex items-center justify-between p-4 bg-card rounded-xl border border-border/80 shadow-sm gap-4">
                        <div className="space-y-0.5">
                          <label className="text-sm font-semibold text-foreground block">
                            Widget Live Monitor Core Web Vitals
                          </label>
                          <p className="text-xs text-muted-foreground">
                            Tampilkan widget pengujian performa real-time <strong>Core Web Vitals</strong> (LCP, CLS, INP, TTFB) di pojok kiri bawah.
                          </p>
                        </div>
                        <button
                          type="button"
                          role="switch"
                          aria-checked={showCoreWebVitalsMonitor}
                          onClick={() => {
                            toggleCoreWebVitalsMonitor();
                            toast.success(
                              !showCoreWebVitalsMonitor
                                ? "Widget Core Web Vitals Diaktifkan"
                                : "Widget Core Web Vitals Dinonaktifkan"
                            );
                          }}
                          className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                            showCoreWebVitalsMonitor ? "bg-primary" : "bg-muted-foreground/30"
                          }`}
                        >
                          <span
                            className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow-lg ring-0 transition duration-200 ease-in-out ${
                              showCoreWebVitalsMonitor ? "translate-x-5" : "translate-x-0"
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

            {activeTab === "smtp" && (
              <div className="p-6">
                <div className="border-b border-border pb-4 mb-6">
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="text-lg font-semibold">Konfigurasi SMTP</h3>
                    <span className="rounded-full bg-primary/10 px-2.5 py-1 text-[10px] font-bold uppercase text-primary">
                      {smtpPlantID === "GLOBAL"
                        ? "Default Global"
                        : `Plant: ${plantsList.find((plant) => plant.plant_id === smtpPlantID)?.plant_name || smtpPlantID}${smtpHasPlantOverride ? "" : " · Fallback Global"}`}
                    </span>
                  </div>
                  <p className="text-sm text-muted-foreground mt-1">
                    Pilih plant pada filter di atas untuk membuat override. Plant yang belum memiliki konfigurasi menggunakan SMTP global, kemudian environment sebagai fallback terakhir.
                  </p>
                </div>

                {smtpLoading ? (
                  <div className="p-12 flex justify-center items-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                  </div>
                ) : (
                  <form onSubmit={handleSaveSMTP} className="space-y-6">
                    <div className="flex items-center justify-between gap-4 rounded-xl border border-border bg-muted/30 p-4">
                      <div>
                        <label className="text-sm font-semibold">Aktifkan Pengiriman Email</label>
                        <p className="text-xs text-muted-foreground mt-1">Jika dimatikan, seluruh pengiriman email otomatis akan ditolak.</p>
                      </div>
                      <button
                        type="button"
                        role="switch"
                        aria-checked={smtpValues.SMTP_ENABLED === "true"}
                        onClick={() => setSmtpValues(prev => ({
                          ...prev,
                          SMTP_ENABLED: prev.SMTP_ENABLED === "true" ? "false" : "true",
                        }))}
                        className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors ${smtpValues.SMTP_ENABLED === "true" ? "bg-primary" : "bg-muted-foreground/30"}`}
                      >
                        <span className={`pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transition-transform ${smtpValues.SMTP_ENABLED === "true" ? "translate-x-5" : "translate-x-0"}`} />
                      </button>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-5 rounded-xl border border-border bg-muted/20 p-5">
                      <div>
                        <label className="text-sm font-medium mb-1.5 block">SMTP Host</label>
                        <Input
                          value={smtpValues.SMTP_HOST}
                          onChange={e => setSmtpValues(prev => ({ ...prev, SMTP_HOST: e.target.value }))}
                          placeholder="smtp.gmail.com"
                          autoComplete="off"
                          required={smtpValues.SMTP_ENABLED === "true"}
                        />
                        <p className="text-[11px] text-muted-foreground mt-1.5">Hostname server penyedia email.</p>
                      </div>

                      <div>
                        <label className="text-sm font-medium mb-1.5 block">SMTP Port</label>
                        <Input
                          type="number"
                          min={1}
                          max={65535}
                          value={smtpValues.SMTP_PORT}
                          onChange={e => setSmtpValues(prev => ({ ...prev, SMTP_PORT: e.target.value }))}
                          placeholder="587"
                          required={smtpValues.SMTP_ENABLED === "true"}
                        />
                        <p className="text-[11px] text-muted-foreground mt-1.5">Umumnya 587 untuk STARTTLS atau 25 untuk relay internal.</p>
                      </div>

                      <div>
                        <label className="text-sm font-medium mb-1.5 block">SMTP Username</label>
                        <Input
                          value={smtpValues.SMTP_USER}
                          onChange={e => setSmtpValues(prev => ({ ...prev, SMTP_USER: e.target.value }))}
                          placeholder="email@perusahaan.com"
                          autoComplete="username"
                          required={smtpValues.SMTP_ENABLED === "true"}
                        />
                        <p className="text-[11px] text-muted-foreground mt-1.5">Akun untuk autentikasi ke server SMTP.</p>
                      </div>

                      <div>
                        <label className="text-sm font-medium mb-1.5 flex items-center gap-2">
                          SMTP Password / App Password
                          {smtpPasswordConfigured && (
                            <span className="rounded bg-green-500/10 px-1.5 py-0.5 text-[9px] font-bold uppercase text-green-600">Tersimpan</span>
                          )}
                        </label>
                        <div className="relative">
                          <Input
                            type={showSmtpPassword ? "text" : "password"}
                            value={smtpValues.SMTP_PASSWORD}
                            onChange={e => setSmtpValues(prev => ({ ...prev, SMTP_PASSWORD: e.target.value }))}
                            placeholder={smtpPasswordConfigured ? "Kosongkan untuk mempertahankan password" : "Masukkan password SMTP"}
                            autoComplete="new-password"
                            className="pr-10"
                          />
                          <button
                            type="button"
                            onClick={() => setShowSmtpPassword(value => !value)}
                            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                            aria-label={showSmtpPassword ? "Sembunyikan password" : "Tampilkan password"}
                          >
                            {showSmtpPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                          </button>
                        </div>
                        <p className="text-[11px] text-muted-foreground mt-1.5">Password dienkripsi di database dan tidak pernah ditampilkan kembali.</p>
                      </div>

                      <div className="md:col-span-2">
                        <label className="text-sm font-medium mb-1.5 block">Email Pengirim</label>
                        <Input
                          type="email"
                          value={smtpValues.SMTP_SENDER_EMAIL}
                          onChange={e => setSmtpValues(prev => ({ ...prev, SMTP_SENDER_EMAIL: e.target.value }))}
                          placeholder="noreply@perusahaan.com"
                          required={smtpValues.SMTP_ENABLED === "true"}
                        />
                        <p className="text-[11px] text-muted-foreground mt-1.5">Alamat yang tampil pada bagian From di email penerima.</p>
                      </div>
                    </div>

                    <div className="flex justify-end pt-4 border-t border-border">
                      <Button type="submit" disabled={isSavingSmtp}>
                        {isSavingSmtp ? (
                          <span className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                        ) : (
                          <Save className="w-4 h-4 mr-2" />
                        )}
                        {isSavingSmtp ? "Menyimpan..." : "Simpan Konfigurasi SMTP"}
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
                  <p className="text-sm text-muted-foreground">
                    Scope aktif: <span className="font-semibold text-foreground">{settingsPlantID === "GLOBAL" ? "Global (fallback semua plant)" : plantsList.find((plant) => plant.plant_id === settingsPlantID)?.plant_name || settingsPlantID}</span>.
                    {settingsPlantID !== "GLOBAL" && " Saat disimpan, template ini menjadi override khusus plant tersebut."}
                  </p>
                </div>
                
                {settingsLoading ? (
                  <div className="p-12 flex justify-center items-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                  </div>
                ) : (
                  <div className="divide-y divide-border">
                    {EMAIL_TEMPLATES.map((template) => {
                      const existing = settingsData?.find((s) => s.setting_key === template.key);
                      const isConfigured = !!existing?.setting_value;
                      const isPlantOverride = settingsPlantID !== "GLOBAL" && existing?.plant_id === settingsPlantID;
                      const usesGlobalFallback = settingsPlantID !== "GLOBAL" && existing?.plant_id !== settingsPlantID;

                      return (
                        <div key={template.key} className="p-6 flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between hover:bg-muted/10 transition-colors">
                          <div>
                            <div className="flex items-center gap-2 mb-1">
                              <h4 className="font-semibold">{template.title}</h4>
                              {isPlantOverride ? (
                                <span className="bg-green-500/10 text-green-500 text-[10px] uppercase font-bold px-2 py-0.5 rounded">Khusus Plant</span>
                              ) : usesGlobalFallback ? (
                                <span className="bg-blue-500/10 text-blue-500 text-[10px] uppercase font-bold px-2 py-0.5 rounded">Fallback Global</span>
                              ) : isConfigured ? (
                                <span className="bg-green-500/10 text-green-500 text-[10px] uppercase font-bold px-2 py-0.5 rounded">Template Global</span>
                              ) : (
                                <span className="bg-orange-500/10 text-orange-500 text-[10px] uppercase font-bold px-2 py-0.5 rounded">Bawaan Sistem</span>
                              )}
                            </div>
                            <p className="text-sm text-muted-foreground">{template.description}</p>
                            <p className="text-xs text-muted-foreground mt-2">Variabel: {template.variables.map((variable) => `{{.${variable}}}`).join(", ")}</p>
                            <p className="text-xs text-muted-foreground mt-2 font-mono bg-muted inline-block px-2 py-1 rounded">Key: {template.key}</p>
                          </div>
                          
                          <Button variant={isConfigured ? "outline" : "default"} onClick={() => handleOpenEditor(template)} className="shrink-0">
                            <PenSquare className="w-4 h-4 mr-2" />
                            {usesGlobalFallback ? "Buat Override Plant" : isConfigured ? "Edit Template" : "Buat Template"}
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
                <ApiKeyManager plantFilter={plantFilter} />
              </div>
            )}
          </div>
        </div>
      </div>

      {editorOpen && (
        <EmailEditorModal
          isOpen
          onClose={() => setEditorOpen(false)}
          title={editorTitle}
          initialData={derivedInitialData}
          variables={EMAIL_TEMPLATES.find((template) => template.key === editorKey)?.variables || []}
          onSave={handleSaveEmailTemplate}
          isSaving={saveSettingMutation.isPending}
        />
      )}
    </div>
  );
}
