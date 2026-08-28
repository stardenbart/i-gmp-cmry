"use client";

import { useState, useEffect } from "react";
import dynamic from "next/dynamic";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { getApiErrorMessage } from "@/lib/api/error";
import { api } from "@/lib/api/axios";
import { useAuthStore } from "@/stores/authStore";
import { usePermissions } from "@/lib/usePermissions";
import { useMounted } from "@/lib/useMounted";
import { ShieldAlert, Save, Eye, EyeOff } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { SettingsNavigation } from "@/components/settings/SettingsNavigation";
import { GeneralSettingsPanel } from "@/components/settings/GeneralSettingsPanel";
import { EmailTemplatesPanel } from "@/components/settings/EmailTemplatesPanel";
import { EMAIL_TEMPLATES, GENERAL_SETTINGS, SMTP_DEFAULTS, type EmailTemplateConfig } from "@/components/settings/settingsConfig";

import { masterApi, type SystemSetting } from "@/lib/api/master.api";
import { ChevronDown, Filter } from "lucide-react";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { setActiveTab, setPlantFilter } from "@/store/slices/settingsSlice";

const EmailEditorModal = dynamic(
  () => import("./EmailEditorModal").then((module) => module.EmailEditorModal),
  { ssr: false }
);
const ApiKeyManager = dynamic(
  () => import("@/components/settings/ApiKeyManager").then((module) => module.ApiKeyManager),
  { ssr: false }
);

export default function SettingsPage() {
  const user = useAuthStore((state) => state.user);
  const { hasPermission, isLoading: isGuardLoading } = usePermissions();
  const isAdmin = hasPermission("PERM-MSTR-R");
  const mounted = useMounted();
  const queryClient = useQueryClient();
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
        <SettingsNavigation activeTab={activeTab} onChange={(tab) => dispatch(setActiveTab(tab))} />

        {/* Content Area */}
        <div className="flex-1">
          <div className="bg-card border border-border rounded-2xl shadow-sm overflow-hidden">
            
            {activeTab === "general" && (
              <GeneralSettingsPanel
                isLoading={settingsLoading}
                values={generalValues}
                onChange={(key, value) => setGeneralValues((current) => ({ ...current, [key]: value }))}
                onSubmit={handleSaveGeneral}
              />
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
              <EmailTemplatesPanel
                isLoading={settingsLoading}
                plantId={settingsPlantID}
                plantName={plantsList.find((plant) => plant.plant_id === settingsPlantID)?.plant_name}
                settings={settingsData}
                onEdit={handleOpenEditor}
              />
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
          requiredVariables={EMAIL_TEMPLATES.find((template) => template.key === editorKey)?.requiredVariables || []}
          onSave={handleSaveEmailTemplate}
          isSaving={saveSettingMutation.isPending}
        />
      )}
    </div>
  );
}
