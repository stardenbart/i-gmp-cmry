"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/authStore";

/**
 * Admin role IDs that have access to admin-only pages.
 * Adjust these values based on your backend role configuration.
 */
export const ADMIN_ROLES = ["ROLE-000", "ROLE-001", "SUPERADMIN", "ADM", "admin", "1"];
export const AUDITOR_ROLES = ["ROLE-002", "AUDITOR", "auditor", "2", ...ADMIN_ROLES];

/**
 * Hook to protect admin-only pages.
 * Redirects non-admin users to the main dashboard if they don't have admin privileges.
 *
 * @param redirectTo - Optional pah to redirect to (defaults to user's dashboard)
 * @returns { isAdmin: boolean, isLoading: boolean }
 *
 * @example
 * // In an admin page component
 * export default function AdminPage() {
 *   const { isAdmin, isLoading } = useAdminGuard();
 *
 *   if (isLoading) return <LoadingSpinner />;
 *   if (!isAdmin) return null; // Will redirect
 *
 *   return <AdminContent />;
 * }
 */
export function useAdminGuard(redirectTo?: string) {
  const router = useRouter();
  const user = useAuthStore((state) => state.user);
  const isAdmin = user ? isAdminUser(user.role_id) : false;

  useEffect(() => {
    if (user && !isAdmin) {
      const plantCode = user.plant_id || "global";
      const target = redirectTo || `/cimory/${plantCode}/dashboard/${user.id}`;
      router.replace(target);
    }
  }, [user, isAdmin, router, redirectTo]);

  return {
    isAdmin,
    isLoading: Boolean(user && !isAdmin),
  };
}

export function isSuperAdminUser(roleId?: string, roleName?: string): boolean {
  if (roleName) {
    const lower = roleName.toLowerCase();
    if (lower.includes("super admin") || lower.includes("superadmin")) return true;
  }
  if (!roleId) return false;
  const lowerId = roleId.toLowerCase();
  return lowerId === "role-000" || lowerId === "superadmin";
}

/**
 * Simple check if user has admin role
 */
export function isAdminUser(roleId?: string, roleName?: string): boolean {
  if (roleName) {
    const lower = roleName.toLowerCase();
    if (lower.includes("admin") || lower.includes("administrator")) return true;
  }
  if (!roleId) return false;
  const lowerId = roleId.toLowerCase();
  return ADMIN_ROLES.includes(roleId) || lowerId.includes("admin") || lowerId.includes("adm") || lowerId.includes("role-000") || lowerId.includes("role-001");
}

export function isAuditorUser(roleId?: string, roleName?: string, username?: string): boolean {
  if (username) {
    const lowerU = username.toLowerCase();
    if (lowerU.includes("auditor") || lowerU.includes("admin")) return true;
  }
  if (roleName) {
    const lower = roleName.toLowerCase();
    if (lower.includes("auditor") || lower.includes("admin") || lower.includes("administrator")) return true;
  }
  if (!roleId) return false;
  const lowerId = roleId.toLowerCase();
  return AUDITOR_ROLES.includes(roleId) || lowerId.includes("auditor") || lowerId.includes("admin") || lowerId.includes("adm") || lowerId.includes("role-000") || lowerId.includes("role-001") || lowerId.includes("role-002");
}

/**
 * Hook to protect auditor-only pages.
 * Redirects non-auditor users (e.g. auditee/PIC) to the issues page or dashboard.
 */
export function useAuditorGuard(redirectTo?: string) {
  const router = useRouter();
  const user = useAuthStore((state) => state.user);
  const isAuditor = user ? isAuditorUser(user.role_id) : false;

  useEffect(() => {
    if (user && !isAuditor) {
      const plantCode = user.plant_id || "global";
      const target = redirectTo || `/cimory/${plantCode}/dashboard/${user.id}/issues`;
      router.replace(target);
    }
  }, [user, isAuditor, router, redirectTo]);

  return {
    isAuditor,
    isLoading: Boolean(user && !isAuditor),
  };
}

/**
 * Hook to protect auditee-only pages (e.g. WO/WR).
 * Redirects auditor users to the issues page or dashboard.
 */
export function useAuditeeGuard(redirectTo?: string) {
  const router = useRouter();
  const user = useAuthStore((state) => state.user);
  const isAuditee = user ? !isAuditorUser(user.role_id) : false;

  useEffect(() => {
    if (user && !isAuditee) {
      const plantCode = user.plant_id || "global";
      const target = redirectTo || `/cimory/${plantCode}/dashboard/${user.id}/issues`;
      router.replace(target);
    }
  }, [user, isAuditee, router, redirectTo]);

  return {
    isAuditee,
    isLoading: Boolean(user && !isAuditee),
  };
}
