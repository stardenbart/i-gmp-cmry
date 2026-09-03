"use client";

import { useMemo, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { AlertTriangle, CheckCircle2, Download, FileSpreadsheet, Upload, X } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { fetchItems } from "../master.api";
import { getApiErrorMessage } from "@/lib/api/error";
import { isSuperAdminUser } from "@/lib/useAdminGuard";
import { useAuthStore } from "@/stores/authStore";
import type { Plant } from "@/types/api";
import {
  commitMasterImport,
  downloadCSV,
  downloadImportTemplate,
  validateMasterImport,
} from "./master-import.api";
import {
  IMPORT_LABEL,
  type ImportCommitResult,
  type ImportPreview,
  type MasterImportType,
} from "./master-import.types";

interface MasterImportDialogProps {
  isOpen: boolean;
  importType: MasterImportType | null;
  onClose: () => void;
  onCommitted: (type: MasterImportType) => void;
  onContinue: (type: MasterImportType) => void;
}

export function MasterImportDialog({ isOpen, importType, onClose, onCommitted, onContinue }: MasterImportDialogProps) {
  const user = useAuthStore((state) => state.user);
  const isSuperAdmin = isSuperAdminUser(user?.role_id, user?.role?.role_name);
  const inputRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File | null>(null);
  const [plantId, setPlantId] = useState(user?.plant_id || "");
  const [preview, setPreview] = useState<ImportPreview | null>(null);
  const [result, setResult] = useState<ImportCommitResult | null>(null);
  const [isDownloading, setIsDownloading] = useState(false);
  const [isValidating, setIsValidating] = useState(false);
  const [isCommitting, setIsCommitting] = useState(false);

  const requiresPlant = importType !== "hei";
  const { data: plants } = useQuery({
    queryKey: ["master", "plants", "import-lookup"],
    queryFn: () => fetchItems<Plant>("/master/plants", 1, "", 500),
    enabled: isOpen && isSuperAdmin && requiresPlant,
  });

  const canUsePlant = !requiresPlant || Boolean(plantId);
  const isBusy = isDownloading || isValidating || isCommitting;
  const title = importType ? IMPORT_LABEL[importType] : "Master Data";
  const invalidRows = useMemo(() => preview?.rows.filter((row) => row.status === "invalid") || [], [preview]);

  if (!isOpen || !importType) return null;

  const selectFile = (selected: File | undefined) => {
    if (!selected) return;
    if (!selected.name.toLowerCase().endsWith(".xlsx")) {
      toast.error("Hanya file .xlsx yang diperbolehkan");
      return;
    }
    if (selected.size > 5 * 1024 * 1024) {
      toast.error("Ukuran file maksimal 5 MB");
      return;
    }
    setFile(selected);
    setPreview(null);
    setResult(null);
  };

  const handleDownloadTemplate = async (type: MasterImportType = importType) => {
    if (type !== "hei" && !plantId) {
      toast.error("Pilih plant target terlebih dahulu");
      return false;
    }
    setIsDownloading(true);
    try {
      await downloadImportTemplate(type, type === "hei" ? undefined : plantId);
      toast.success("Template terbaru berhasil diunduh");
      return true;
    } catch (error) {
      toast.error(getApiErrorMessage(error, "Gagal mengunduh template"));
      return false;
    } finally {
      setIsDownloading(false);
    }
  };

  const handleValidate = async () => {
    if (!file || !canUsePlant) return;
    setIsValidating(true);
    setResult(null);
    try {
      const validated = await validateMasterImport(importType, file, requiresPlant ? plantId : undefined);
      setPreview(validated);
      if (validated.can_commit) toast.success("Seluruh data valid dan siap diimport");
      else toast.error("File masih memiliki data invalid atau duplikat");
    } catch (error) {
      toast.error(getApiErrorMessage(error, "Validasi import gagal"));
    } finally {
      setIsValidating(false);
    }
  };

  const handleCommit = async () => {
    if (!file || !preview?.can_commit || !preview.validation_token) return;
    setIsCommitting(true);
    try {
      const committed = await commitMasterImport(importType, file, preview.validation_token, requiresPlant ? plantId : undefined);
      setResult(committed);
      onCommitted(importType);
      toast.success(`${committed.inserted} data berhasil diimport`);
    } catch (error) {
      toast.error(getApiErrorMessage(error, "Commit import gagal"));
      setPreview(null);
    } finally {
      setIsCommitting(false);
    }
  };

  const downloadErrors = () => {
    downloadCSV(
      `hasil-validasi-${importType}.csv`,
      invalidRows.map((row) => ({
        row: row.row,
        status: row.status,
        ...row.normalized_data,
        errors: row.errors.map((error) => `${error.field || "data"}: ${error.message}`).join(" | "),
      })),
    );
  };

  const downloadCreatedIDs = () => {
    if (!result) return;
    downloadCSV(
      `mapping-id-import-${result.type}-${result.import_id}.csv`,
      result.created_rows.map((row) => ({ source_row: row.row, ...row.data })),
    );
  };

  const continueToNext = async () => {
    if (!result?.next_import_type) return;
    const downloaded = await handleDownloadTemplate(result.next_import_type);
    if (downloaded) onContinue(result.next_import_type);
  };

  return (
    <div
      className="fixed inset-0 z-[120] flex items-center justify-center bg-black/60 p-3 backdrop-blur-sm sm:p-6"
      onMouseDown={(event) => event.target === event.currentTarget && !isBusy && onClose()}
    >
      <div className="flex max-h-[92vh] w-full max-w-5xl flex-col overflow-hidden rounded-2xl border border-border bg-card shadow-2xl">
        <div className="flex items-start justify-between gap-4 border-b border-border px-5 py-4 sm:px-6">
          <div>
            <h2 className="text-lg font-semibold text-foreground">Import {title}</h2>
            <p className="mt-1 text-xs text-muted-foreground">
              Validasi tidak menyimpan data. Commit hanya aktif jika seluruh baris valid.
            </p>
          </div>
          <button type="button" onClick={onClose} disabled={isBusy} className="rounded-lg p-2 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50">
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="flex-1 space-y-5 overflow-y-auto p-5 sm:p-6">
          {!result && (
            <>
              {isSuperAdmin && requiresPlant && (
                <div className="space-y-1.5">
                  <label className="text-sm font-medium">Plant target</label>
                  <select
                    value={plantId}
                    onChange={(event) => { setPlantId(event.target.value); setFile(null); setPreview(null); }}
                    className="h-11 w-full rounded-xl border border-border bg-card px-3 text-sm sm:max-w-md"
                  >
                    <option value="">Pilih plant...</option>
                    {plants?.items.map((plant) => (
                      <option key={plant.plant_id} value={plant.plant_id}>{plant.plant_name} ({plant.plant_code})</option>
                    ))}
                  </select>
                </div>
              )}

              <div className="flex flex-col gap-3 rounded-xl border border-blue-500/20 bg-blue-500/5 p-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <p className="text-sm font-medium">Gunakan template terbaru</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {importType === "hei"
                      ? "Kategori pada sheet REFERENSI hanya contoh. Anda dapat mengisi category_name dengan kategori baru."
                      : "Sheet REFERENSI adalah snapshot. Download ulang setelah import parent berhasil."}
                  </p>
                </div>
                <Button type="button" variant="outline" size="sm" onClick={() => void handleDownloadTemplate()} isLoading={isDownloading} disabled={!canUsePlant}>
                  <Download className="h-4 w-4" /> Download Template
                </Button>
              </div>

              <div
                className="flex min-h-40 cursor-pointer flex-col items-center justify-center rounded-2xl border-2 border-dashed border-border p-6 text-center transition-colors hover:border-primary/60 hover:bg-muted/30"
                onClick={() => inputRef.current?.click()}
                onDragOver={(event) => event.preventDefault()}
                onDrop={(event) => { event.preventDefault(); selectFile(event.dataTransfer.files[0]); }}
              >
                <input
                  ref={inputRef}
                  type="file"
                  accept=".xlsx"
                  className="hidden"
                  onChange={(event) => {
                    selectFile(event.target.files?.[0]);
                    event.currentTarget.value = "";
                  }}
                />
                <FileSpreadsheet className="mb-3 h-9 w-9 text-primary" />
                <p className="text-sm font-medium">{file ? file.name : "Pilih atau jatuhkan file XLSX di sini"}</p>
                <p className="mt-1 text-xs text-muted-foreground">Maksimal 5 MB dan 5.000 baris</p>
              </div>

              {preview && <ImportPreviewPanel preview={preview} onDownloadErrors={downloadErrors} />}
            </>
          )}

          {result && (
            <div className="space-y-5">
              <div className="flex items-start gap-3 rounded-xl border border-emerald-500/20 bg-emerald-500/5 p-4">
                <CheckCircle2 className="mt-0.5 h-5 w-5 text-emerald-600" />
                <div>
                  <p className="font-semibold">Import berhasil</p>
                  <p className="mt-1 text-sm text-muted-foreground">{result.inserted} data tersimpan. ID baru dapat dilihat dan diunduh di bawah.</p>
                </div>
              </div>
              <div className="overflow-x-auto rounded-xl border border-border">
                <table className="w-full min-w-[620px] text-left text-sm">
                  <thead className="bg-muted/50 text-xs uppercase text-muted-foreground">
                    <tr><th className="px-4 py-3">Baris sumber</th><th className="px-4 py-3">ID baru</th><th className="px-4 py-3">Data</th></tr>
                  </thead>
                  <tbody>
                    {result.created_rows.slice(0, 200).map((row) => (
                      <tr key={`${row.row}-${row.entity_id}`} className="border-t border-border">
                        <td className="px-4 py-3">{row.row}</td>
                        <td className="px-4 py-3 font-mono text-xs font-semibold">{row.entity_id}</td>
                        <td className="px-4 py-3 text-xs text-muted-foreground">{Object.entries(row.data).filter(([key]) => !key.endsWith("_id")).map(([, value]) => value).join(" • ")}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {result.created_rows.length > 200 && <p className="text-xs text-muted-foreground">Preview dibatasi 200 baris. File mapping tetap berisi seluruh hasil.</p>}
            </div>
          )}
        </div>

        <div className="flex flex-col-reverse gap-3 border-t border-border bg-muted/10 px-5 py-4 sm:flex-row sm:justify-end sm:px-6">
          <Button type="button" variant="outline" onClick={onClose} disabled={isBusy}>Tutup</Button>
          {!result && (
            <>
              <Button type="button" variant="outline" onClick={() => void handleValidate()} isLoading={isValidating} disabled={!file || !canUsePlant || isCommitting}>
                <Upload className="h-4 w-4" /> Periksa Data
              </Button>
              <Button type="button" onClick={() => void handleCommit()} isLoading={isCommitting} disabled={!preview?.can_commit || !preview.validation_token || isValidating}>
                Import Sekarang
              </Button>
            </>
          )}
          {result && (
            <>
              <Button type="button" variant="outline" onClick={downloadCreatedIDs}><Download className="h-4 w-4" /> Unduh Mapping ID</Button>
              {result.next_import_type && <Button type="button" onClick={() => void continueToNext()} isLoading={isDownloading}>Lanjut Import {IMPORT_LABEL[result.next_import_type]}</Button>}
            </>
          )}
        </div>
      </div>
    </div>
  );
}

function ImportPreviewPanel({ preview, onDownloadErrors }: { preview: ImportPreview; onDownloadErrors: () => void }) {
  const cards = [
    ["Total", preview.summary.total],
    ["Valid", preview.summary.valid],
    ["Invalid", preview.summary.invalid],
    ["Duplikat File", preview.summary.duplicate_in_file],
    ["Duplikat Database", preview.summary.duplicate_in_database],
  ];
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-5">
        {cards.map(([label, value]) => <div key={String(label)} className="rounded-xl border border-border p-3"><p className="text-xs text-muted-foreground">{label}</p><p className="mt-1 text-xl font-semibold">{value}</p></div>)}
      </div>
      {preview.can_commit ? (
        <div className="flex items-center gap-2 rounded-xl border border-emerald-500/20 bg-emerald-500/5 p-4 text-sm text-emerald-700 dark:text-emerald-300"><CheckCircle2 className="h-5 w-5" />Seluruh baris valid. File siap di-commit.</div>
      ) : (
        <div className="space-y-3">
          <div className="flex flex-col gap-3 rounded-xl border border-amber-500/20 bg-amber-500/5 p-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex items-start gap-2 text-sm"><AlertTriangle className="mt-0.5 h-5 w-5 shrink-0 text-amber-600" /><span>Perbaiki seluruh error lalu unggah dan validasi ulang. Tidak ada data yang disimpan.</span></div>
            <Button type="button" size="sm" variant="outline" onClick={onDownloadErrors}><Download className="h-4 w-4" /> Unduh Error</Button>
          </div>
          <div className="max-h-72 overflow-auto rounded-xl border border-border">
            <table className="w-full min-w-[720px] text-left text-xs">
              <thead className="sticky top-0 bg-muted"><tr><th className="px-3 py-2">Baris</th><th className="px-3 py-2">Data</th><th className="px-3 py-2">Kesalahan</th></tr></thead>
              <tbody>{preview.rows.filter((row) => row.status === "invalid").map((row) => <tr key={row.row} className="border-t border-border align-top"><td className="px-3 py-2 font-medium">{row.row}</td><td className="px-3 py-2">{Object.values(row.normalized_data).join(" • ")}</td><td className="px-3 py-2 text-destructive">{row.errors.map((error) => `${error.field ? `${error.field}: ` : ""}${error.message}`).join(" | ")}</td></tr>)}</tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
