"use client";

import { AlertTriangle, PenSquare } from "lucide-react";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { EMAIL_TEMPLATES, type EmailTemplateConfig } from "@/components/settings/settingsConfig";
import type { SystemSetting } from "@/lib/api/master.api";

interface EmailTemplatesPanelProps {
  isLoading: boolean;
  plantId: string;
  plantName?: string;
  settings?: SystemSetting[];
  onEdit: (template: EmailTemplateConfig) => void;
  /** Called with the template's `<KEY>_ENABLED` setting and its new value. */
  onToggle?: (settingKey: string, enabled: boolean) => void;
}

/** On/off switch for a template's email; a missing setting means enabled. */
export const emailEnabledKey = (templateKey: string) => `${templateKey}_ENABLED`;

export function EmailTemplatesPanel({ isLoading, plantId, plantName, settings, onEdit, onToggle }: EmailTemplatesPanelProps) {
  const isGlobal = plantId === "GLOBAL";
  return (
    <section>
      <header className="border-b border-border bg-muted/20 p-6">
        <h2 className="text-lg font-semibold">Template Email</h2>
        <p className="text-sm text-muted-foreground">
          Scope aktif: <strong className="text-foreground">{isGlobal ? "Global (fallback semua plant)" : plantName || plantId}</strong>.
          {!isGlobal && " Template yang disimpan menjadi override khusus plant."}
        </p>
      </header>
      {isLoading ? (
        <div className="flex justify-center p-12"><div className="h-8 w-8 animate-spin rounded-full border-b-2 border-primary" /></div>
      ) : (
        <div className="divide-y divide-border">
          {EMAIL_TEMPLATES.map((template) => {
            const existing = settings?.find((setting) => setting.setting_key === template.key);
            const alwaysOn = "alwaysOn" in template && template.alwaysOn;
            const enabledKey = emailEnabledKey(template.key);
            const enabled = settings?.find((setting) => setting.setting_key === enabledKey)?.setting_value?.toLowerCase() !== "false";
            const configured = !!existing?.setting_value;
            const missingRequiredVariables = template.requiredVariables.filter(
              (variable) => !existing?.setting_value?.includes(`{{.${variable}}}`)
            );
            const invalid = configured && missingRequiredVariables.length > 0;
            const plantOverride = !isGlobal && existing?.plant_id === plantId;
            const globalFallback = !isGlobal && existing?.plant_id !== plantId;
            const badge = invalid
              ? ["Perlu Diperbaiki", "bg-destructive/10 text-destructive"]
              : plantOverride
              ? ["Khusus Plant", "bg-green-500/10 text-green-500"]
              : globalFallback
                ? ["Fallback Global", "bg-blue-500/10 text-blue-500"]
                : configured
                  ? ["Template Global", "bg-green-500/10 text-green-500"]
                  : ["Bawaan Sistem", "bg-orange-500/10 text-orange-500"];

            return (
              <article key={template.key} className="flex flex-col items-start justify-between gap-4 p-6 transition-colors hover:bg-muted/10 sm:flex-row sm:items-center">
                <div>
                  <div className="mb-1 flex flex-wrap items-center gap-2">
                    <h3 className="font-semibold">{template.title}</h3>
                    <span className={`rounded px-2 py-0.5 text-[10px] font-bold uppercase ${badge[1]}`}>{badge[0]}</span>
                  </div>
                  <p className="text-sm text-muted-foreground">{template.description}</p>
                  <p className="mt-2 text-xs text-muted-foreground">Variabel: {template.variables.map((variable) => `{{.${variable}}}`).join(", ")}</p>
                  {invalid && (
                    <p className="mt-2 flex items-center gap-1.5 text-xs font-medium text-destructive">
                      <AlertTriangle className="h-3.5 w-3.5" />
                      Variabel wajib belum ada: {missingRequiredVariables.map((variable) => `{{.${variable}}}`).join(", ")}
                    </p>
                  )}
                  <code className="mt-2 inline-block rounded bg-muted px-2 py-1 text-xs text-muted-foreground">{template.key}</code>
                </div>
                <div className="flex shrink-0 flex-wrap items-center gap-3">
                  {alwaysOn ? (
                    <span className="text-xs font-medium text-muted-foreground" title="Dibutuhkan untuk reset password">Selalu aktif</span>
                  ) : (
                    <button
                      type="button"
                      role="switch"
                      aria-checked={enabled}
                      aria-label={`Kirim email ${template.title}`}
                      title={enabled ? "Email aktif — klik untuk menonaktifkan" : "Email nonaktif — klik untuk mengaktifkan"}
                      onClick={() => onToggle?.(enabledKey, !enabled)}
                      className="inline-flex items-center gap-2 text-xs font-medium text-muted-foreground"
                    >
                      <span className={cn("relative inline-flex h-5 w-9 items-center rounded-full transition-colors", enabled ? "bg-success" : "bg-muted-foreground/40")}>
                        <span className={cn("inline-block h-4 w-4 rounded-full bg-white shadow transition-transform", enabled ? "translate-x-4" : "translate-x-0.5")} />
                      </span>
                      {enabled ? "Email aktif" : "Email nonaktif"}
                    </button>
                  )}
                  <Button type="button" variant={configured ? "outline" : "default"} onClick={() => onEdit(template)} className="shrink-0">
                    <PenSquare className="mr-2 h-4 w-4" />
                    {invalid ? "Perbaiki Template" : globalFallback ? "Buat Override Plant" : configured ? "Edit Template" : "Buat Template"}
                  </Button>
                </div>
              </article>
            );
          })}
        </div>
      )}
    </section>
  );
}
