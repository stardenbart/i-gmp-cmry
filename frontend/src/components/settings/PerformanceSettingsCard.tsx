"use client";

import { toast } from "sonner";

import { isAdminUser } from "@/lib/useAdminGuard";
import { useAuthStore } from "@/stores/authStore";
import { useSettingsStore } from "@/stores/settingsStore";

interface SettingSwitchProps {
  checked: boolean;
  title: string;
  description: React.ReactNode;
  onToggle: () => void;
}

function SettingSwitch({ checked, title, description, onToggle }: SettingSwitchProps) {
  return (
    <div className="flex items-center justify-between gap-4 rounded-xl border border-border/80 bg-card p-4 shadow-sm">
      <div className="space-y-0.5">
        <p className="text-sm font-semibold text-foreground">{title}</p>
        <div className="text-xs text-muted-foreground">{description}</div>
      </div>
      <button
        type="button"
        role="switch"
        aria-label={title}
        aria-checked={checked}
        onClick={onToggle}
        className={`relative inline-flex h-6 w-11 shrink-0 rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary ${checked ? "bg-primary" : "bg-muted-foreground/30"}`}
      >
        <span className={`pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transition-transform ${checked ? "translate-x-5" : "translate-x-0"}`} />
      </button>
    </div>
  );
}

export function PerformanceSettingsCard() {
  const user = useAuthStore((state) => state.user);
  const {
    showSearchLatencyButton,
    toggleSearchLatencyButton,
    showCoreWebVitalsMonitor,
    toggleCoreWebVitalsMonitor,
  } = useSettingsStore();

  if (!user || !isAdminUser(user.role_id, user.role?.role_name)) return null;

  return (
    <section className="space-y-4 rounded-xl border border-border bg-muted/30 p-5">
      <h4 className="font-semibold text-primary">Pengujian & Performa</h4>
      <SettingSwitch
        checked={showSearchLatencyButton}
        title="Tombol Uji Latensi Search"
        description={<>Tampilkan alat uji latensi pada pencarian Inspeksi, Temuan, GMP Data, dan Logs.</>}
        onToggle={() => {
          toggleSearchLatencyButton();
          toast.success(showSearchLatencyButton ? "Tombol Uji Latensi dinonaktifkan" : "Tombol Uji Latensi diaktifkan");
        }}
      />
      <SettingSwitch
        checked={showCoreWebVitalsMonitor}
        title="Widget Core Web Vitals"
        description={<>Tampilkan metrik LCP, CLS, INP, dan TTFB untuk diagnosis performa browser.</>}
        onToggle={() => {
          toggleCoreWebVitalsMonitor();
          toast.success(showCoreWebVitalsMonitor ? "Widget Core Web Vitals dinonaktifkan" : "Widget Core Web Vitals diaktifkan");
        }}
      />
    </section>
  );
}
