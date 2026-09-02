import axios from "axios";
import { api } from "@/lib/api/axios";

export type GmpExportFormat = "template" | "table";

export interface GmpExportFilters {
  plant_id?: string;
  area_id?: string;
  kawasan_id?: string;
  detail_kawasan_id?: string;
  start_date?: string;
  end_date?: string;
  q?: string;
}

export async function exportGmp(format: GmpExportFormat, filters: GmpExportFilters) {
  const response = await api.get<Blob>("/dashboard/export", {
    params: { ...filters, format },
    responseType: "blob",
  });
  const disposition = response.headers["content-disposition"] as string | undefined;
  const fallbackPrefix = format === "table" ? "Tabel_Data_GMP" : "Laporan_GMP";
  const filename = disposition?.match(/filename\*?=(?:UTF-8''|\")?([^";]+)/i)?.[1];
  return {
    blob: response.data,
    fileName: filename ? decodeURIComponent(filename.trim()) : `${fallbackPrefix}_${new Date().toISOString().slice(0, 10)}.xlsx`,
  };
}

export function downloadGmpExport(blob: Blob, fileName: string) {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = fileName;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  window.setTimeout(() => URL.revokeObjectURL(url), 1_000);
}

export async function getGmpExportErrorMessage(error: unknown): Promise<string> {
  if (axios.isAxiosError(error)) {
    const responseData = error.response?.data;
    if (responseData instanceof Blob) {
      try {
        const parsed = JSON.parse(await responseData.text()) as { message?: string; error?: string };
        return parsed.message || parsed.error || "Gagal mengekspor Data GMP";
      } catch {
        // Fall through to the regular Axios message below.
      }
    }
    if (responseData && typeof responseData === "object") {
      const parsed = responseData as { message?: string; error?: string };
      return parsed.message || parsed.error || error.message;
    }
    return error.message || "Gagal mengekspor Data GMP";
  }
  return error instanceof Error ? error.message : "Gagal mengekspor Data GMP";
}
