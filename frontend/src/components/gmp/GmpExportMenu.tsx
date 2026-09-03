"use client";

import { useEffect, useRef, useState } from "react";
import { ChevronDown, Download, FileSpreadsheet, Loader2, Table2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
  downloadGmpExport,
  exportGmp,
  getGmpExportErrorMessage,
  type GmpExportFilters,
  type GmpExportFormat,
} from "./gmp-export.api";

interface GmpExportMenuProps {
  filters: GmpExportFilters;
  disabled?: boolean;
}

const exportOptions: Array<{
  format: GmpExportFormat;
  title: string;
  description: string;
  Icon: typeof FileSpreadsheet;
}> = [
  {
    format: "template",
    title: "Laporan Template (.xlsx)",
    description: "Format laporan resmi beserta bukti visual.",
    Icon: FileSpreadsheet,
  },
  {
    format: "table",
    title: "Tabel seperti di Web (.xlsx)",
    description: "Urutan 16 kolom mengikuti tabel Data GMP.",
    Icon: Table2,
  },
];

export function GmpExportMenu({ filters, disabled = false }: GmpExportMenuProps) {
  const rootRef = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);
  const [exporting, setExporting] = useState<GmpExportFormat | null>(null);

  useEffect(() => {
    if (!open) return;
    const closeOnOutsideClick = (event: MouseEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false);
    };
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", closeOnOutsideClick);
    document.addEventListener("keydown", closeOnEscape);
    return () => {
      document.removeEventListener("mousedown", closeOnOutsideClick);
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [open]);

  const handleExport = async (format: GmpExportFormat) => {
    if (exporting || disabled) return;
    setOpen(false);
    setExporting(format);
    try {
      const result = await exportGmp(format, filters);
      downloadGmpExport(result.blob, result.fileName);
      toast.success(format === "table" ? "Tabel Data GMP berhasil diekspor" : "Laporan GMP berhasil diekspor");
    } catch (error) {
      toast.error(await getGmpExportErrorMessage(error));
    } finally {
      setExporting(null);
    }
  };

  return (
    <div ref={rootRef} className="relative z-20 shrink-0">
      <Button
        type="button"
        onClick={() => setOpen((value) => !value)}
        disabled={disabled || exporting !== null}
        aria-haspopup="menu"
        aria-expanded={open}
        className="min-w-[150px] justify-between gap-2 bg-primary text-primary-foreground hover:bg-primary/90"
      >
        <span className="inline-flex items-center">
          {exporting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Download className="mr-2 h-4 w-4" />}
          {exporting ? "Mengekspor..." : "Export"}
        </span>
        <ChevronDown className={cn("h-4 w-4 transition-transform", open && "rotate-180")} />
      </Button>

      {open && (
        <div
          role="menu"
          className="absolute right-0 top-full z-[120] mt-2 w-[310px] overflow-hidden rounded-xl border border-border bg-card p-1.5 text-card-foreground opacity-100 shadow-2xl ring-1 ring-black/5 dark:ring-white/10"
        >
          {exportOptions.map(({ format, title, description, Icon }) => (
            <button
              key={format}
              type="button"
              role="menuitem"
              onClick={() => handleExport(format)}
              className="flex w-full items-start gap-3 rounded-lg px-3 py-2.5 text-left transition-colors hover:bg-muted focus:bg-muted focus:outline-none"
            >
              <span className="mt-0.5 rounded-lg bg-primary/10 p-2 text-primary">
                <Icon className="h-4 w-4" />
              </span>
              <span className="min-w-0">
                <span className="block text-sm font-semibold text-foreground">{title}</span>
                <span className="mt-0.5 block text-xs leading-relaxed text-muted-foreground">{description}</span>
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
