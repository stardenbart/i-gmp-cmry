"use client";

import { Key, Mail, Server, Settings } from "lucide-react";

import type { SettingsState } from "@/store/slices/settingsSlice";

interface SettingsNavigationProps {
  activeTab: SettingsState["activeTab"];
  onChange: (tab: SettingsState["activeTab"]) => void;
}

const ITEMS = [
  { id: "general", label: "Umum", icon: Settings },
  { id: "smtp", label: "Konfigurasi SMTP", icon: Server },
  { id: "email", label: "Template Email", icon: Mail },
  { id: "apikey", label: "API Key Power BI", icon: Key },
] as const;

export function SettingsNavigation({ activeTab, onChange }: SettingsNavigationProps) {
  return (
    <nav className="w-full shrink-0 space-y-2 md:w-64" aria-label="Kategori pengaturan">
      {ITEMS.map(({ id, label, icon: Icon }) => (
        <button
          key={id}
          type="button"
          onClick={() => onChange(id)}
          aria-current={activeTab === id ? "page" : undefined}
          className={`flex w-full items-center gap-3 rounded-xl px-4 py-3 text-left font-medium transition-colors ${activeTab === id ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-muted hover:text-foreground"}`}
        >
          <Icon className="h-5 w-5" />
          {label}
        </button>
      ))}
    </nav>
  );
}
