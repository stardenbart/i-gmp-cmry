export type MasterImportType = "aspek" | "detail" | "uraian" | "hei";

export interface ImportFieldError {
  field: string;
  code: string;
  message: string;
}

export interface ImportPreviewRow {
  row: number;
  status: "valid" | "invalid";
  normalized_data: Record<string, string | number>;
  errors: ImportFieldError[];
}

export interface ImportSummary {
  total: number;
  valid: number;
  invalid: number;
  duplicate_in_file: number;
  duplicate_in_database: number;
}

export interface ImportPreview {
  import_id: string;
  type: MasterImportType;
  template_version: string;
  file_hash: string;
  reference_generated_at?: string;
  summary: ImportSummary;
  rows: ImportPreviewRow[];
  can_commit: boolean;
  validation_token?: string;
}

export interface ImportCreatedRow {
  row: number;
  entity_id: string;
  data: Record<string, string | number>;
}

export interface ImportCommitResult {
  import_id: string;
  type: MasterImportType;
  inserted: number;
  created_rows: ImportCreatedRow[];
  next_import_type?: MasterImportType;
}

export const IMPORT_TYPE_BY_TAB: Partial<Record<string, MasterImportType>> = {
  aspeks: "aspek",
  details: "detail",
  urains: "uraian",
  hei: "hei",
};

export const TAB_BY_IMPORT_TYPE: Record<MasterImportType, string> = {
  aspek: "aspeks",
  detail: "details",
  uraian: "urains",
  hei: "hei",
};

export const IMPORT_LABEL: Record<MasterImportType, string> = {
  aspek: "Aspek Audit",
  detail: "Detail Aspek",
  uraian: "Uraian",
  hei: "Klasifikasi HEI",
};
