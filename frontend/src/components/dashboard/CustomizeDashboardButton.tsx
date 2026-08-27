"use client";

import { useState } from "react";
import { createPortal } from "react-dom";
import { LayoutGrid, X, ChevronUp, ChevronDown, Loader2, Save } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useEditableDashboardLayout } from "@/hooks/useEditableDashboardLayout";
import type { WidgetDefinition } from "@/components/dashboard/WidgetGrid";

interface CustomizeDashboardButtonProps {
  registry: WidgetDefinition[];
}

/**
 * Fase 3a: a simple "Sesuaikan Dashboard" modal — checklist to show/hide each
 * widget, up/down buttons to reorder. No drag-and-drop yet (that's Fase 3b,
 * once this ships and a react-grid-layout upgrade lands separately).
 */
export function CustomizeDashboardButton({ registry }: CustomizeDashboardButtonProps) {
  const [isOpen, setIsOpen] = useState(false);
  const { rows, isLoading, toggleVisible, move, save, isSaving } = useEditableDashboardLayout(registry, isOpen);

  const handleSave = async () => {
    try {
      await save();
      toast.success("Dashboard berhasil disesuaikan");
      setIsOpen(false);
    } catch {
      toast.error("Gagal menyimpan pengaturan dashboard");
    }
  };

  return (
    <>
      <Button variant="outline" size="sm" onClick={() => setIsOpen(true)} className="gap-2">
        <LayoutGrid className="h-4 w-4" />
        <span className="hidden sm:inline">Sesuaikan Dashboard</span>
      </Button>

      {isOpen &&
        typeof document !== "undefined" &&
        createPortal(
          <div
            className="fixed inset-0 z-[9999] flex items-center justify-center bg-black/70 p-4 backdrop-blur-sm animate-in fade-in duration-200"
            role="dialog"
            aria-modal="true"
            aria-labelledby="customize-dashboard-title"
            onClick={() => !isSaving && setIsOpen(false)}
          >
            <div
              className="relative w-full max-w-md max-h-[85dvh] overflow-y-auto rounded-2xl border border-border bg-card p-5 shadow-2xl"
              onClick={(e) => e.stopPropagation()}
            >
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h2 id="customize-dashboard-title" className="text-lg font-bold text-foreground">
                    Sesuaikan Dashboard
                  </h2>
                  <p className="text-xs text-muted-foreground">Pilih widget yang ingin ditampilkan dan atur urutannya.</p>
                </div>
                <button
                  type="button"
                  onClick={() => setIsOpen(false)}
                  disabled={isSaving}
                  className="p-1.5 rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground transition-colors disabled:opacity-50"
                  aria-label="Tutup"
                >
                  <X className="h-4 w-4" />
                </button>
              </div>

              {isLoading ? (
                <div className="flex items-center justify-center py-10">
                  <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                </div>
              ) : (
                <div className="space-y-2 mb-5">
                  {rows.map((row, index) => (
                    <div
                      key={row.id}
                      className={cn(
                        "flex items-center gap-3 p-3 rounded-xl border border-border/70 bg-muted/30 transition-opacity",
                        !row.visible && "opacity-50"
                      )}
                    >
                      <span className="text-xs font-mono font-semibold text-muted-foreground w-5 shrink-0">{index + 1}.</span>

                      <label className="flex items-center gap-2 flex-1 min-w-0 cursor-pointer">
                        <input
                          type="checkbox"
                          checked={row.visible}
                          onChange={() => toggleVisible(row.id)}
                          className="h-4 w-4 rounded border-border accent-primary shrink-0"
                        />
                        <span className="text-sm font-medium text-foreground truncate">{row.title}</span>
                      </label>

                      <div className="flex items-center gap-0.5 shrink-0">
                        <button
                          type="button"
                          onClick={() => move(row.id, "up")}
                          disabled={index === 0}
                          className="p-1.5 rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground transition-colors disabled:opacity-30 disabled:pointer-events-none"
                          aria-label={`Pindahkan ${row.title} ke atas`}
                          title="Naikkan urutan"
                        >
                          <ChevronUp className="h-4 w-4" />
                        </button>
                        <button
                          type="button"
                          onClick={() => move(row.id, "down")}
                          disabled={index === rows.length - 1}
                          className="p-1.5 rounded-lg text-muted-foreground hover:bg-muted hover:text-foreground transition-colors disabled:opacity-30 disabled:pointer-events-none"
                          aria-label={`Pindahkan ${row.title} ke bawah`}
                          title="Turunkan urutan"
                        >
                          <ChevronDown className="h-4 w-4" />
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}

              <div className="flex items-center gap-2">
                <Button variant="outline" onClick={() => setIsOpen(false)} disabled={isSaving} className="flex-1">
                  Batal
                </Button>
                <Button onClick={handleSave} isLoading={isSaving} disabled={isLoading} className="flex-1 gap-2">
                  <Save className="h-4 w-4" />
                  Simpan
                </Button>
              </div>
            </div>
          </div>,
          document.body
        )}
    </>
  );
}
