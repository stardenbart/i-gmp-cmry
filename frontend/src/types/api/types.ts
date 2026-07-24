/**
 * API Types
 * All type definitions for API responses and requests
 */

// ─── Generic Types ─────────────────────────────────────────────────────────

/** Generic paginated response */
export interface PaginatedResponse<T> {
  data: {
    items: T[];
    pagination: {
      total: number;
      page: number;
      limit: number;
      total_pages: number;
    };
  };
  message: string;
}

/** Generic single item response */
export interface SingleItemResponse<T> {
  data: T;
  message: string;
}

/** Generic API error response */
export interface ApiError {
  message: string;
  error?: string;
  statusCode?: number;
}

// ─── Auth Types ─────────────────────────────────────────────────────────────

/** Login request */
export interface LoginRequest {
  email: string;
  password: string;
}

/** Login response */
export interface LoginResponse {
  token: string;
  user: User;
}

/** User object */
export interface User {
  user_id: string;
  username: string;
  full_name: string;
  email: string;
  department_id: string;
  role_id: string;
  user_status: "Active" | "Inactive" | "Suspended";
  created_at?: string;
  updated_at?: string;
  password?: string;
  pic_kawasan_ids?: string[];
  pic_kategori?: string;
  pic_mappings?: { kawasan_id: string; kategori_pic: string; pic_map_id: string }[];
}

/** Change password request */
export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

/** Update profile request */
export interface UpdateProfileRequest {
  full_name?: string;
  email?: string;
}

// ─── Department Types ───────────────────────────────────────────────────────

export interface Department {
  department_id: string;
  department_name: string;
  department_code: string;
  is_active: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface CreateDepartmentRequest {
  department_name: string;
  department_code: string;
}

export interface UpdateDepartmentRequest extends Partial<CreateDepartmentRequest> {
  is_active?: boolean;
}

// ─── Area Types ─────────────────────────────────────────────────────────────

export interface Area {
  area_id: string;
  area_name: string;
  area_code: string;
  department_id?: string;
  department_name?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CreateAreaRequest {
  area_name: string;
  area_code: string;
  department_id?: string;
}

export interface UpdateAreaRequest extends Partial<CreateAreaRequest> {
  is_active?: boolean;
}

// ─── Kawasan Types ──────────────────────────────────────────────────────────

export interface Kawasan {
  kawasan_id: string;
  kawasan_name: string;
  kawasan_code: string;
  area_id: string;
  area_name?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CreateKawasanRequest {
  kawasan_name: string;
  kawasan_code: string;
  area_id: string;
}

export interface UpdateKawasanRequest extends Partial<CreateKawasanRequest> {}

// ─── Detail Kawasan Types ───────────────────────────────────────────────────

export interface DetailKawasan {
  detail_kawasan_id: string;
  detail_kawasan_name: string;
  detail_kawasan_code: string;
  kawasan_id: string;
  kawasan_name?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CreateDetailKawasanRequest {
  detail_kawasan_name: string;
  detail_kawasan_code: string;
  kawasan_id: string;
}

export interface UpdateDetailKawasanRequest extends Partial<CreateDetailKawasanRequest> {}

// ─── Aspek Types ───────────────────────────────────────────────────────────

export interface Aspek {
  aspek_id: string;
  aspek_name: string;
  aspek_code: string;
  aspek_weight: number;
  area_id?: string;
  area_name?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CreateAspekRequest {
  aspek_name: string;
  aspek_code: string;
  aspek_weight: number;
  area_id?: string;
}

export interface UpdateAspekRequest extends Partial<CreateAspekRequest> {}

// ─── Detail Types ──────────────────────────────────────────────────────────

export interface Detail {
  detail_id: string;
  detail_name: string;
  detail_code: string;
  aspek_id: string;
  aspek_name?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CreateDetailRequest {
  detail_name: string;
  detail_code: string;
  aspek_id: string;
}

export interface UpdateDetailRequest extends Partial<CreateDetailRequest> {}

// ─── Uraian Types ──────────────────────────────────────────────────────────

export interface Uraian {
  uraian_id: string;
  uraian_name: string;
  uraian_code: string;
  detail_id: string;
  detail_name?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CreateUraianRequest {
  uraian_name: string;
  uraian_code: string;
  detail_id: string;
}

export interface UpdateUraianRequest extends Partial<CreateUraianRequest> {}

// ─── Inspection Types ───────────────────────────────────────────────────────

export interface InspectionHeader {
  inspection_header_id: string;
  inspection_code: string;
  inspection_date: string;
  inspection_status: "draft" | "scheduled" | "in_progress" | "completed" | "cancelled";
  area_id: string;
  area_name?: string;
  user_id: string;
  inspector_name?: string;
  total_issues: number;
  created_at?: string;
  updated_at?: string;
}

export interface InspectionResult {
  result_id: string;
  inspection_header_id: string;
  uraian_id: string;
  uraian_name?: string;
  result_status: "pass" | "fail" | "na";
  finding_notes?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CreateInspectionRequest {
  inspection_date: string;
  area_id: string;
  inspection_status?: "draft" | "scheduled";
}

export interface UpdateInspectionRequest {
  inspection_date?: string;
  inspection_status?: InspectionHeader["inspection_status"];
}

export interface SubmitInspectionResultRequest {
  uraian_id: string;
  result_status: "pass" | "fail" | "na";
  finding_notes?: string;
}

// ─── Issue Types ────────────────────────────────────────────────────────────

export type IssueStatus = "Open" | "InProgress" | "PendingValidation" | "Closed" | "Verified" | "Overdue";

export interface Issue {
  issue_id: string;
  result_id: string;
  issue_pic_user_id: string;
  pic_name?: string;
  issue_status: IssueStatus;
  due_date?: string;
  keterangan?: string;
  created_at?: string;
  updated_at?: string;
  photos?: IssuePhoto[];
}

export interface IssuePhoto {
  issue_photo_id: string;
  issue_id: string;
  pic_user_id: string;
  photo_type: "before" | "after";
  image_url: string;
  follow_up_date?: string;
  jumlah_follow_up?: number;
  created_at?: string;
}

export interface CreateIssueRequest {
  result_id: string;
  issue_pic_user_id: string;
  due_date?: string;
  keterangan?: string;
}

export interface UpdateIssueRequest {
  issue_status?: IssueStatus;
  issue_pic_user_id?: string;
  due_date?: string;
  keterangan?: string;
}

export interface UploadIssuePhotoRequest {
  issue_id: string;
  photo_type: "before" | "after";
  follow_up_date?: string;
  jumlah_follow_up?: number;
}

// ─── Activity Log Types ─────────────────────────────────────────────────────

export interface ActivityLog {
  log_id: string;
  user_id: string;
  user_name?: string;
  action: string;
  module: string;
  entity_id?: string;
  entity_name?: string;
  ip_address?: string;
  user_agent?: string;
  old_values?: Record<string, unknown>;
  new_values?: Record<string, unknown>;
  created_at: string;
}

// ─── Notification Types ─────────────────────────────────────────────────────

export type NotificationType = "info" | "warning" | "error" | "success";

export interface Notification {
  id: string;
  type: NotificationType;
  title: string;
  message: string;
  is_read: boolean;
  link?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  user_id?: string;
}

export interface PaginatedNotifications {
  items: Notification[];
  pagination: {
    total: number;
    page: number;
    limit: number;
    total_pages: number;
  };
}

// ─── Dashboard Types ────────────────────────────────────────────────────────

export interface DashboardStats {
  total_inspections: number;
  open_issues: number;
  closed_issues: number;
  total_pics: number;
  inspections_this_month: number;
  issues_resolved_this_month: number;
  issues_by_month: Array<{ month: string; count: number }>;
  issues_trend: Array<{ date: string; open: number; closed: number }>;
  recent_inspections: Array<{
    id: string;
    name: string;
    status: string;
    created_at: string;
    area_name?: string;
  }>;
  top_issues_by_area: Array<{ area_name: string; count: number }>;
}

// ─── SSE Event Types ────────────────────────────────────────────────────────

export interface SSEStatsEvent {
  type: "stats";
  data: DashboardStats;
  timestamp: string;
}

export interface SSENotificationEvent {
  type: "notification";
  data: Notification;
  timestamp: string;
}

export interface SSEClearEvent {
  type: "clear";
  timestamp: string;
}

export type SSEEvent = SSEStatsEvent | SSENotificationEvent | SSEClearEvent;

// ─── Setting Types ──────────────────────────────────────────────────────────

export interface Setting {
  setting_key: string;
  setting_value: string;
  setting_description?: string;
  updated_at?: string;
}

export interface UpdateSettingRequest {
  setting_value: string;
}
