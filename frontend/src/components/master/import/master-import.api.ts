import axios from "axios";
import { api } from "@/lib/api/axios";
import type { ImportCommitResult, ImportPreview, MasterImportType } from "./master-import.types";

interface APIEnvelope<T> {
  data?: T;
  message?: string;
}

function createImportForm(file: File, plantId?: string, validationToken?: string) {
  const form = new FormData();
  form.append("file", file);
  if (plantId) form.append("plant_id", plantId);
  if (validationToken) form.append("validation_token", validationToken);
  return form;
}

export async function downloadImportTemplate(type: MasterImportType, plantId?: string) {
  const response = await api.get(`/master/import/${type}/template`, {
    params: plantId ? { plant_id: plantId } : undefined,
    responseType: "blob",
  });
  const disposition = response.headers["content-disposition"] as string | undefined;
  const match = disposition?.match(/filename="?([^";]+)"?/i);
  downloadBlob(response.data as Blob, match?.[1] || `template-import-${type}.xlsx`);
}

export async function validateMasterImport(type: MasterImportType, file: File, plantId?: string) {
  try {
    const response = await api.post<APIEnvelope<ImportPreview>>(
      `/master/import/${type}/validate`,
      createImportForm(file, plantId),
      { headers: { "Content-Type": "multipart/form-data" } },
    );
    return response.data.data!;
  } catch (error) {
    if (axios.isAxiosError<APIEnvelope<ImportPreview>>(error) && error.response?.data?.data) {
      return error.response.data.data;
    }
    throw error;
  }
}

export async function commitMasterImport(
  type: MasterImportType,
  file: File,
  validationToken: string,
  plantId?: string,
) {
  const response = await api.post<APIEnvelope<ImportCommitResult>>(
    `/master/import/${type}/commit`,
    createImportForm(file, plantId, validationToken),
    { headers: { "Content-Type": "multipart/form-data" } },
  );
  return response.data.data!;
}

export function downloadCSV(fileName: string, rows: Array<Record<string, string | number>>) {
  if (rows.length === 0) return;
  const headers = Array.from(new Set(rows.flatMap((row) => Object.keys(row))));
  const escape = (value: string | number | undefined) => {
    const safe = String(value ?? "");
    const formulaSafe = /^[=+\-@]/.test(safe) ? `'${safe}` : safe;
    return `"${formulaSafe.replaceAll('"', '""')}"`;
  };
  const csv = [headers.map(escape).join(","), ...rows.map((row) => headers.map((header) => escape(row[header])).join(","))].join("\r\n");
  downloadBlob(new Blob(["\uFEFF", csv], { type: "text/csv;charset=utf-8" }), fileName);
}

function downloadBlob(blob: Blob, fileName: string) {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = fileName;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
}
