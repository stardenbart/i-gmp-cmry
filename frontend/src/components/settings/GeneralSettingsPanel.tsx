"use client";

import { Save } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { GENERAL_SETTINGS } from "@/components/settings/settingsConfig";
import { PerformanceSettingsCard } from "@/components/settings/PerformanceSettingsCard";

const dateFmt = new Intl.DateTimeFormat("id-ID", { day: "numeric", month: "long", year: "numeric" });
const monthFmt = new Intl.DateTimeFormat("id-ID", { month: "long", year: "numeric" });

// Mirrors the backend's ResolveInspectionPeriod (see
// backend/internal/usecase/inspectionusecase/period.go) purely for this
// settings-page preview — never used for actual gating, so the two
// implementations drifting slightly would only ever mislead an admin
// reading the preview, not break real behavior.
function resolveInspectionPeriodPreview(cutoffDay: number, now: Date) {
  const day = Number.isInteger(cutoffDay) && cutoffDay >= 1 && cutoffDay <= 28 ? cutoffDay : 1;
  let month = now.getMonth();
  if (now.getDate() < day) month -= 1;
  const start = new Date(now.getFullYear(), month, day);
  const end = new Date(now.getFullYear(), month + 1, day);
  const lastDay = new Date(end.getTime() - 24 * 60 * 60 * 1000);
  return { start, lastDay };
}

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
                  item.type === "cutoff-day" ? (
                    (() => {
                      const cutoffDay = parseInt(values[item.key] || "1", 10);
                      const endDay = Number.isInteger(cutoffDay) && cutoffDay > 1 ? String(cutoffDay - 1) : "Akhir bulan";
                      const { start, lastDay } = resolveInspectionPeriodPreview(cutoffDay, new Date());
                      return (
                        <div key={item.key} className="xl:col-span-2">
                          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                            <div>
                              <label className="mb-1 block text-sm font-medium" htmlFor={`setting-${item.key}`}>Tanggal Mulai Periode</label>
                              <Input
                                id={`setting-${item.key}`}
                                type="number"
                                min={1}
                                max={28}
                                value={values[item.key] || ""}
                                onChange={(event) => onChange(item.key, event.target.value)}
                                required
                              />
                            </div>
                            <div>
                              <label className="mb-1 block text-sm font-medium text-muted-foreground">Tanggal Akhir Periode</label>
                              <Input type="text" value={endDay} disabled readOnly className="text-muted-foreground" />
                              <p className="mt-1 text-[10px] text-muted-foreground">Otomatis: sehari sebelum tanggal mulai di bulan berikutnya.</p>
                            </div>
                          </div>
                          <p className="mt-2 rounded-lg bg-primary/5 px-3 py-2 text-[12px] text-foreground">
                            Periode saat ini: <strong>{dateFmt.format(start)} – {dateFmt.format(lastDay)}</strong> → dihitung sebagai <strong>bulan {monthFmt.format(start)}</strong>.
                          </p>
                          <p className="mt-1.5 text-[11px] text-muted-foreground">{item.desc}</p>
                        </div>
                      );
                    })()
                  ) : item.type === "boolean" ? (
                    <div key={item.key} className="flex items-start justify-between gap-4">
                      <div>
                        <label className="mb-1 block text-sm font-medium" htmlFor={`setting-${item.key}`}>{item.label}</label>
                        <p className="text-[11px] text-muted-foreground">{item.desc}</p>
                      </div>
                      <label className="relative inline-flex shrink-0 cursor-pointer items-center">
                        <input
                          id={`setting-${item.key}`}
                          type="checkbox"
                          className="peer sr-only"
                          checked={values[item.key] === "true"}
                          onChange={(event) => onChange(item.key, event.target.checked ? "true" : "false")}
                        />
                        <div className="h-6 w-11 rounded-full bg-muted transition-colors peer-checked:bg-primary" />
                        <div className="absolute left-1 h-4 w-4 rounded-full bg-background transition-transform peer-checked:translate-x-5" />
                      </label>
                    </div>
                  ) : (
                    <div key={item.key}>
                      <label className="mb-1 block text-sm font-medium" htmlFor={`setting-${item.key}`}>{item.label}</label>
                      <Input id={`setting-${item.key}`} type={item.type} value={values[item.key] || ""} onChange={(event) => onChange(item.key, event.target.value)} required />
                      <p className="mt-1.5 text-[11px] text-muted-foreground">{item.desc}</p>
                    </div>
                  )
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
