"use client";

import { Save } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { GENERAL_SETTINGS } from "@/components/settings/settingsConfig";
import { PerformanceSettingsCard } from "@/components/settings/PerformanceSettingsCard";

interface GeneralSettingsPanelProps {
  isLoading: boolean;
  values: Record<string, string>;
  onChange: (key: string, value: string) => void;
  onSubmit: (event: React.FormEvent) => void;
}

export function GeneralSettingsPanel({ isLoading, values, onChange, onSubmit }: GeneralSettingsPanelProps) {
  return (
    <section className="p-6">
      <header className="mb-6 border-b border-border pb-4">
        <h2 className="text-lg font-semibold">Pengaturan Umum</h2>
        <p className="text-sm text-muted-foreground">Batas login, sesi, penyimpanan, dan tenggat waktu.</p>
      </header>
      {isLoading ? (
        <div className="flex justify-center p-12"><div className="h-8 w-8 animate-spin rounded-full border-b-2 border-primary" /></div>
      ) : (
        <form onSubmit={onSubmit} className="space-y-8">
          {GENERAL_SETTINGS.map((group) => (
            <div key={group.group} className="rounded-xl border border-border bg-muted/30 p-5">
              <h3 className="mb-4 font-semibold text-primary">{group.group}</h3>
              <div className="grid grid-cols-1 gap-6 xl:grid-cols-2">
                {group.keys.map((item) => (
                  <div key={item.key}>
                    <label className="mb-1 block text-sm font-medium" htmlFor={`setting-${item.key}`}>{item.label}</label>
                    <Input id={`setting-${item.key}`} type={item.type} value={values[item.key] || ""} onChange={(event) => onChange(item.key, event.target.value)} required />
                    <p className="mt-1.5 text-[11px] text-muted-foreground">{item.desc}</p>
                  </div>
                ))}
              </div>
            </div>
          ))}
          <PerformanceSettingsCard />
          <div className="flex justify-end border-t border-border pt-4">
            <Button type="submit"><Save className="mr-2 h-4 w-4" />Simpan Pengaturan Umum</Button>
          </div>
        </form>
      )}
    </section>
  );
}
