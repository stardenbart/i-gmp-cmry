"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/authStore";

/**
 * Admin role IDs that have access to admin-only pages.
 * Adjust these values based on your backend role configuration.
 */
export const ADMIN_ROLES = ["ROLE-001", "ADM", "admin", "1"];
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
  const [isLoading, setIsLoading] = useState(true);
  const [isAdmin, setIsAdmin] = useState(false);
  const router = useRouter();
  const user = useAuthStore((state) => state.user);

  useEffect(() => {
    if (!user) {
      // User store not loaded yet or completely logged out, assume loading/false
      setIsLoading(false);
      return;
    }

    const hasAdminRole = isAdminUser(user.role_id);
    setIsAdmin(hasAdminRole);
    
    if (!hasAdminRole) {
      const target = redirectTo || `/cimory/dashboard/${user.id}`;
      router.replace(target);
    } else {
      setIsLoading(false);
    }
  }, [user, router, redirectTo]);

  return {
    isAdmin,
    isLoading,
  };
}

/**
 * Simple check if user has admin role
 */
export function isAdminUser(roleId?: string): boolean {
  if (!roleId) return false;
  return ADMIN_ROLES.includes(roleId);
}

/**
 * Simple check if user has auditor role (or admin)
 */
export function isAuditorUser(roleId?: string): boolean {
  if (!roleId) return false;
  return AUDITOR_ROLES.includes(roleId);
}

/**
 * Hook to protect auditor-only pages.
 * Redirects non-auditor users (e.g. auditee/PIC) to the issues page or dashboard.
 */
export function useAuditorGuard(redirectTo?: string) {
  const [isLoading, setIsLoading] = useState(true);
  const [isAuditor, setIsAuditor] = useState(false);
  const router = useRouter();
  const user = useAuthStore((state) => state.user);

  useEffect(() => {
    if (!user) {
      setIsLoading(false);
      return;
    }

    const hasAuditorRole = isAuditorUser(user.role_id);
    setIsAuditor(hasAuditorRole);
    
    if (!hasAuditorRole) {
      const target = redirectTo || `/cimory/dashboard/${user.id}/issues`;
      router.replace(target);
    } else {
      setIsLoading(false);
    }
  }, [user, router, redirectTo]);

  return {
    isAuditor,
    isLoading,
  };
}

/**
 * Hook to protect auditee-only pages (e.g. WO/WR).
 * Redirects auditor users to the issues page or dashboard.
 */
export function useAuditeeGuard(redirectTo?: string) {
  const [isLoading, setIsLoading] = useState(true);
  const [isAuditee, setIsAuditee] = useState(false);
  const router = useRouter();
  const user = useAuthStore((state) => state.user);

  useEffect(() => {
    if (!user) {
      setIsLoading(false);
      return;
    }

    const hasAuditorRole = isAuditorUser(user.role_id);
    // If they are NOT an auditor, they are an auditee (PIC)
    const hasAuditeeRole = !hasAuditorRole;
    
    setIsAuditee(hasAuditeeRole);
    
    if (!hasAuditeeRole) {
      const target = redirectTo || `/cimory/dashboard/${user.id}/issues`;
      router.replace(target);
    } else {
      setIsLoading(false);
    }
  }, [user, router, redirectTo]);

  return {
    isAuditee,
    isLoading,
  };
}
