/**
 * API Types & Services
 * Barrel export for all types, API functions, and React Query hooks
 *
 * Usage:
 * import { User, LoginRequest } from "@/types/api";
 * import { authApi, userApi } from "@/types/api";
 * import { useUsers, useCreateUser } from "@/types/api";
 */

// ─── Types ─────────────────────────────────────────────────────────────────

export type {
  // Generic
  PaginatedResponse,
  SingleItemResponse,
  ApiError,

  // Auth
  LoginRequest,
  LoginResponse,
  User,
  ChangePasswordRequest,
  UpdateProfileRequest,

  // Master Data
  Department,
  CreateDepartmentRequest,
  UpdateDepartmentRequest,
  Area,
  CreateAreaRequest,
  UpdateAreaRequest,
  Kawasan,
  CreateKawasanRequest,
  UpdateKawasanRequest,
  DetailKawasan,
  CreateDetailKawasanRequest,
  UpdateDetailKawasanRequest,
  Aspek,
  CreateAspekRequest,
  UpdateAspekRequest,
  Detail,
  CreateDetailRequest,
  UpdateDetailRequest,
  Uraian,
  CreateUraianRequest,
  UpdateUraianRequest,

  // Inspection
  InspectionHeader,
  InspectionResult,
  CreateInspectionRequest,
  UpdateInspectionRequest,
  SubmitInspectionResultRequest,

  // Issue
  IssueStatus,
  Issue,
  IssuePhoto,
  CreateIssueRequest,
  UpdateIssueRequest,
  UploadIssuePhotoRequest,

  // Activity Log
  ActivityLog,

  // Notification
  NotificationType,
  Notification,
  PaginatedNotifications,

  // Dashboard
  DashboardStats,

  // SSE
  SSEStatsEvent,
  SSENotificationEvent,
  SSEClearEvent,
  SSEEvent,

  // Settings
  Setting,
  UpdateSettingRequest,
} from "./types";

// ─── API Services ────────────────────────────────────────────────────────────

export { authApi, userApi } from "./auth";
export {
  departmentApi,
  areaApi,
  kawasanApi,
  detailKawasanApi,
  aspekApi,
  detailApi,
  uraianApi,
} from "./master";
export { inspectionApi, inspectionResultApi } from "./inspection";
export { issueApi, issuePhotoApi } from "./issue";
export { dashboardApi, activityLogApi } from "./dashboard";
export { notificationApi } from "./notification";

// ─── React Query Hooks ───────────────────────────────────────────────────────

export {
  // Auth
  useCurrentUser,
  useChangePassword,

  // Users
  useUsers,
  useUser,
  useCreateUser,
  useUpdateUser,
  useDeleteUser,

  // Departments
  useDepartments,
  useDepartment,
  useCreateDepartment,
  useUpdateDepartment,
  useDeleteDepartment,

  // Areas
  useAreas,
  useArea,
  useCreateArea,
  useUpdateArea,
  useDeleteArea,

  // Kawasans
  useKawasans,
  useKawasan,
  useCreateKawasan,
  useUpdateKawasan,
  useDeleteKawasan,

  // Detail Kawasans
  useDetailKawasans,
  useDetailKawasan,
  useCreateDetailKawasan,
  useUpdateDetailKawasan,
  useDeleteDetailKawasan,

  // Aspeks
  useAspeks,
  useAspek,
  useCreateAspek,
  useUpdateAspek,
  useDeleteAspek,

  // Details
  useDetails,
  useDetail,
  useCreateDetail,
  useUpdateDetail,
  useDeleteDetail,

  // Uraians
  useUrains,
  useUraian,
  useCreateUraian,
  useUpdateUraian,
  useDeleteUraian,

  // Inspections
  useInspections,
  useInspection,
  useCreateInspection,
  useUpdateInspection,
  useDeleteInspection,
  useStartInspection,
  useCompleteInspection,
  useInspectionResults,
  useSubmitInspectionResult,

  // Issues
  useIssues,
  useIssue,
  useCreateIssue,
  useUpdateIssue,
  useDeleteIssue,
  useCloseIssue,
  useVerifyIssue,
  useIssuePhotos,
  useUploadIssuePhoto,

  // Dashboard
  useDashboardStats,
  useRecentActivity,
  useActivityLogs,

  // Notifications
  useNotifications,
  useUnreadNotificationCount,
  useMarkNotificationAsRead,
  useMarkAllNotificationsAsRead,
} from "./hooks";
