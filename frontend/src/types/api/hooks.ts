/**
 * React Query Hooks for API calls
 * Type-safe data fetching with caching and automatic revalidation
 */

import {
  useQuery,
  useMutation,
  useQueryClient,
  QueryClient,
} from "@tanstack/react-query";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";

// API imports
import { authApi, userApi } from "./auth";
import {
  departmentApi,
  areaApi,
  kawasanApi,
  detailKawasanApi,
  aspekApi,
  detailApi,
  uraianApi,
} from "./master";
import { inspectionApi, inspectionResultApi } from "./inspection";
import { issueApi, issuePhotoApi } from "./issue";
import { dashboardApi, activityLogApi } from "./dashboard";
import { notificationApi } from "./notification";

// Types
import type {
  User,
  Department,
  Area,
  Kawasan,
  DetailKawasan,
  Aspek,
  Detail,
  Uraian,
  InspectionHeader,
  InspectionResult,
  Issue,
  IssuePhoto,
  Notification,
  DashboardStats,
  ActivityLog,
} from "./index";

// ─── Auth Hooks ─────────────────────────────────────────────────────────────

export function useCurrentUser() {
  const mounted = useMounted();
  const token = useAuthStore((state) => state.token);

  return useQuery({
    queryKey: ["current-user"],
    queryFn: authApi.me,
    enabled: mounted && !!token,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

export function useChangePassword() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: authApi.changePassword,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["current-user"] });
    },
  });
}

// ─── User Hooks ───────────────────────────────────────────────────────────────

export function useUsers(page = 1, limit = 10, search = "") {
  return useQuery({
    queryKey: ["users", page, limit, search],
    queryFn: () => userApi.getAll(page, limit, search),
    placeholderData: (previousData) => previousData,
  });
}

export function useUser(id: string) {
  return useQuery({
    queryKey: ["users", id],
    queryFn: () => userApi.getById(id),
    enabled: !!id,
  });
}

export function useCreateUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Parameters<typeof userApi.create>[0]) => userApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });
}

export function useUpdateUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Parameters<typeof userApi.update>[1] }) =>
      userApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });
}

export function useDeleteUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: userApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });
}

// ─── Department Hooks ────────────────────────────────────────────────────────

export function useDepartments(page = 1, limit = 10, search = "") {
  return useQuery({
    queryKey: ["departments", page, limit, search],
    queryFn: () => departmentApi.getAll(page, limit, search),
    placeholderData: (previousData) => previousData,
  });
}

export function useDepartment(id: string) {
  return useQuery({
    queryKey: ["departments", id],
    queryFn: () => departmentApi.getById(id),
    enabled: !!id,
  });
}

export function useCreateDepartment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: departmentApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["departments"] });
    },
  });
}

export function useUpdateDepartment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Parameters<typeof departmentApi.update>[1] }) =>
      departmentApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["departments"] });
    },
  });
}

export function useDeleteDepartment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: departmentApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["departments"] });
    },
  });
}

// ─── Area Hooks ─────────────────────────────────────────────────────────────

export function useAreas(page = 1, limit = 10, search = "", departmentId?: string) {
  return useQuery({
    queryKey: ["areas", page, limit, search, departmentId],
    queryFn: () => areaApi.getAll(page, limit, search, departmentId),
    placeholderData: (previousData) => previousData,
  });
}

export function useArea(id: string) {
  return useQuery({
    queryKey: ["areas", id],
    queryFn: () => areaApi.getById(id),
    enabled: !!id,
  });
}

export function useCreateArea() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: areaApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["areas"] });
    },
  });
}

export function useUpdateArea() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Parameters<typeof areaApi.update>[1] }) =>
      areaApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["areas"] });
    },
  });
}

export function useDeleteArea() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: areaApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["areas"] });
    },
  });
}

// ─── Kawasan Hooks ───────────────────────────────────────────────────────────

export function useKawasans(page = 1, limit = 10, search = "", areaId?: string) {
  return useQuery({
    queryKey: ["kawasans", page, limit, search, areaId],
    queryFn: () => kawasanApi.getAll(page, limit, search, areaId),
    placeholderData: (previousData) => previousData,
  });
}

export function useKawasan(id: string) {
  return useQuery({
    queryKey: ["kawasans", id],
    queryFn: () => kawasanApi.getById(id),
    enabled: !!id,
  });
}

export function useCreateKawasan() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: kawasanApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["kawasans"] });
    },
  });
}

export function useUpdateKawasan() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Parameters<typeof kawasanApi.update>[1] }) =>
      kawasanApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["kawasans"] });
    },
  });
}

export function useDeleteKawasan() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: kawasanApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["kawasans"] });
    },
  });
}

// ─── Detail Kawasan Hooks ────────────────────────────────────────────────────

export function useDetailKawasans(
  page = 1,
  limit = 10,
  search = "",
  kawasanId?: string
) {
  return useQuery({
    queryKey: ["detail-kawasans", page, limit, search, kawasanId],
    queryFn: () => detailKawasanApi.getAll(page, limit, search, kawasanId),
    placeholderData: (previousData) => previousData,
  });
}

export function useDetailKawasan(id: string) {
  return useQuery({
    queryKey: ["detail-kawasans", id],
    queryFn: () => detailKawasanApi.getById(id),
    enabled: !!id,
  });
}

export function useCreateDetailKawasan() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: detailKawasanApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["detail-kawasans"] });
    },
  });
}

export function useUpdateDetailKawasan() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: Parameters<typeof detailKawasanApi.update>[1];
    }) => detailKawasanApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["detail-kawasans"] });
    },
  });
}

export function useDeleteDetailKawasan() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: detailKawasanApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["detail-kawasans"] });
    },
  });
}

// ─── Aspek Hooks ────────────────────────────────────────────────────────────

export function useAspeks(page = 1, limit = 10, search = "", areaId?: string) {
  return useQuery({
    queryKey: ["aspeks", page, limit, search, areaId],
    queryFn: () => aspekApi.getAll(page, limit, search, areaId),
    placeholderData: (previousData) => previousData,
  });
}

export function useAspek(id: string) {
  return useQuery({
    queryKey: ["aspeks", id],
    queryFn: () => aspekApi.getById(id),
    enabled: !!id,
  });
}

export function useCreateAspek() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: aspekApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["aspeks"] });
    },
  });
}

export function useUpdateAspek() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Parameters<typeof aspekApi.update>[1] }) =>
      aspekApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["aspeks"] });
    },
  });
}

export function useDeleteAspek() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: aspekApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["aspeks"] });
    },
  });
}

// ─── Detail Hooks ───────────────────────────────────────────────────────────

export function useDetails(page = 1, limit = 10, search = "", aspekId?: string) {
  return useQuery({
    queryKey: ["details", page, limit, search, aspekId],
    queryFn: () => detailApi.getAll(page, limit, search, aspekId),
    placeholderData: (previousData) => previousData,
  });
}

export function useDetail(id: string) {
  return useQuery({
    queryKey: ["details", id],
    queryFn: () => detailApi.getById(id),
    enabled: !!id,
  });
}

export function useCreateDetail() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: detailApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["details"] });
    },
  });
}

export function useUpdateDetail() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Parameters<typeof detailApi.update>[1] }) =>
      detailApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["details"] });
    },
  });
}

export function useDeleteDetail() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: detailApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["details"] });
    },
  });
}

// ─── Uraian Hooks ───────────────────────────────────────────────────────────

export function useUrains(page = 1, limit = 10, search = "", detailId?: string) {
  return useQuery({
    queryKey: ["urains", page, limit, search, detailId],
    queryFn: () => uraianApi.getAll(page, limit, search, detailId),
    placeholderData: (previousData) => previousData,
  });
}

export function useUraian(id: string) {
  return useQuery({
    queryKey: ["urains", id],
    queryFn: () => uraianApi.getById(id),
    enabled: !!id,
  });
}

export function useCreateUraian() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: uraianApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["urains"] });
    },
  });
}

export function useUpdateUraian() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Parameters<typeof uraianApi.update>[1] }) =>
      uraianApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["urains"] });
    },
  });
}

export function useDeleteUraian() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: uraianApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["urains"] });
    },
  });
}

// ─── Inspection Hooks ───────────────────────────────────────────────────────

export function useInspections(
  page = 1,
  limit = 10,
  status?: string,
  areaId?: string,
  search = ""
) {
  return useQuery({
    queryKey: ["inspections", page, limit, status, areaId, search],
    queryFn: () => inspectionApi.getAll(page, limit, status, areaId, search),
    placeholderData: (previousData) => previousData,
  });
}

export function useInspection(id: string) {
  return useQuery({
    queryKey: ["inspections", id],
    queryFn: () => inspectionApi.getById(id),
    enabled: !!id,
  });
}

export function useCreateInspection() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: inspectionApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inspections"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] });
    },
  });
}

export function useUpdateInspection() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: Parameters<typeof inspectionApi.update>[1];
    }) => inspectionApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inspections"] });
    },
  });
}

export function useDeleteInspection() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: inspectionApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inspections"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] });
    },
  });
}

export function useStartInspection() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: inspectionApi.start,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inspections"] });
    },
  });
}

export function useCompleteInspection() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: inspectionApi.complete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inspections"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] });
    },
  });
}

export function useInspectionResults(inspectionId: string) {
  return useQuery({
    queryKey: ["inspection-results", inspectionId],
    queryFn: () => inspectionResultApi.getByInspectionId(inspectionId),
    enabled: !!inspectionId,
  });
}

export function useSubmitInspectionResult() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      inspectionId,
      data,
    }: {
      inspectionId: string;
      data: Parameters<typeof inspectionResultApi.submit>[1];
    }) => inspectionResultApi.submit(inspectionId, data),
    onSuccess: (_, { inspectionId }) => {
      queryClient.invalidateQueries({ queryKey: ["inspection-results", inspectionId] });
      queryClient.invalidateQueries({ queryKey: ["inspections", inspectionId] });
    },
  });
}

// ─── Issue Hooks ─────────────────────────────────────────────────────────────

export function useIssues(
  page = 1,
  limit = 10,
  status?: string,
  picUserId?: string,
  search = ""
) {
  return useQuery({
    queryKey: ["issues", page, limit, status, picUserId, search],
    queryFn: () => issueApi.getAll(page, limit, status, picUserId, search),
    placeholderData: (previousData) => previousData,
  });
}

export function useIssue(id: string) {
  return useQuery({
    queryKey: ["issues", id],
    queryFn: () => issueApi.getById(id),
    enabled: !!id,
  });
}

export function useCreateIssue() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: issueApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] });
    },
  });
}

export function useUpdateIssue() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Parameters<typeof issueApi.update>[1] }) =>
      issueApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issues"] });
    },
  });
}

export function useDeleteIssue() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: issueApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] });
    },
  });
}

export function useCloseIssue() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: issueApi.close,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] });
    },
  });
}

export function useVerifyIssue() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: issueApi.verify,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] });
    },
  });
}

export function useIssuePhotos(issueId: string) {
  return useQuery({
    queryKey: ["issue-photos", issueId],
    queryFn: () => issuePhotoApi.getByIssueId(issueId),
    enabled: !!issueId,
  });
}

export function useUploadIssuePhoto() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      issueId,
      file,
      data,
    }: {
      issueId: string;
      file: File;
      data: Parameters<typeof issuePhotoApi.upload>[2];
    }) => issuePhotoApi.upload(issueId, file, data),
    onSuccess: (_, { issueId }) => {
      queryClient.invalidateQueries({ queryKey: ["issue-photos", issueId] });
      queryClient.invalidateQueries({ queryKey: ["issues", issueId] });
    },
  });
}

// ─── Dashboard Hooks ────────────────────────────────────────────────────────

export function useDashboardStats() {
  return useQuery({
    queryKey: ["dashboard-stats"],
    queryFn: dashboardApi.getStats,
    staleTime: 30 * 1000, // 30 seconds
    refetchInterval: 30 * 1000, // Auto-refetch every 30 seconds
  });
}

export function useRecentActivity(limit = 10) {
  return useQuery({
    queryKey: ["recent-activity", limit],
    queryFn: () => dashboardApi.getRecentActivity(limit),
    staleTime: 60 * 1000, // 1 minute
  });
}

export function useActivityLogs(
  page = 1,
  limit = 20,
  userId?: string,
  module?: string,
  startDate?: string,
  endDate?: string
) {
  return useQuery({
    queryKey: ["activity-logs", page, limit, userId, module, startDate, endDate],
    queryFn: () => activityLogApi.getAll(page, limit, userId, module, startDate, endDate),
    placeholderData: (previousData) => previousData,
  });
}

// ─── Notification Hooks ──────────────────────────────────────────────────────

export function useNotifications(page = 1, limit = 20) {
  return useQuery({
    queryKey: ["notifications", page, limit],
    queryFn: () => notificationApi.getAll(page, limit),
    placeholderData: (previousData) => previousData,
    refetchInterval: 60 * 1000, // Refetch every minute
  });
}

export function useUnreadNotificationCount() {
  return useQuery({
    queryKey: ["notifications-unread-count"],
    queryFn: notificationApi.getUnreadCount,
    refetchInterval: 30 * 1000, // Refetch every 30 seconds
  });
}

export function useMarkNotificationAsRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: notificationApi.markAsRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      queryClient.invalidateQueries({ queryKey: ["notifications-unread-count"] });
    },
  });
}

export function useMarkAllNotificationsAsRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: notificationApi.markAllAsRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      queryClient.invalidateQueries({ queryKey: ["notifications-unread-count"] });
    },
  });
}

// Export types for convenience
export type {
  User,
  Department,
  Area,
  Kawasan,
  DetailKawasan,
  Aspek,
  Detail,
  Uraian,
  InspectionHeader,
  InspectionResult,
  Issue,
  IssuePhoto,
  Notification,
  DashboardStats,
  ActivityLog,
};
