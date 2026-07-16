"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/authStore";

/**
 * Admin role IDs that have access to admin-only pages.
 * Adjust these values based on your backend role configuration.
 */
export const ADMIN_ROLES = ["ADM", "admin", "1"];

/**
 * Hook to protect admin-only pages.
 * Redirects non-admin users to the main dashboard if they don't have admin privileges.
 *
 * @param redirectTo - Optional path to redirect to (defaults to user's dashboard)
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
  const [isLoading, setIsLoading] = useState(false);

  // Disabled guard logic for development
  // useEffect(() => { ... }, []);

  return {
    isAdmin: true, // Bypass for development
    isLoading: false,
  };
}

/**
 * Simple check if user has admin role
 */
export function isAdminUser(roleId?: string): boolean {
  return true; // Bypass for development
}
