"use client";

import { useState, useCallback, useRef, useEffect } from "react";
import { toast } from "sonner";
import { inspectionApi } from "@/lib/api/inspection.api";
import { useHeartbeat } from "@/hooks/useHeartbeat";

interface UseAspekLockProps {
  scopeId: string;
  aspekId: string | null;
  enabled?: boolean;
}

export function useAspekLock({ scopeId, aspekId, enabled = true }: UseAspekLockProps) {
  const [lockToken, setLockToken] = useState<string | null>(null);
  const [isLockedByMe, setIsLockedByMe] = useState(false);
  const [lockError, setLockError] = useState<string | null>(null);
  const [isAcquiring, setIsAcquiring] = useState(false);

  const activeAspekRef = useRef<string | null>(null);
  const lockTokenRef = useRef<string | null>(null);

  // Keep ref up to date for unmount cleanup
  useEffect(() => {
    lockTokenRef.current = lockToken;
  }, [lockToken]);

  // Acquire lock for an aspect
  const acquireLock = useCallback(
    async (targetAspekId: string) => {
      if (!scopeId || !targetAspekId || !enabled) return false;

      // Release previous lock if switching aspect
      if (activeAspekRef.current && activeAspekRef.current !== targetAspekId && lockTokenRef.current) {
        try {
          await inspectionApi.releaseLock(scopeId, activeAspekRef.current, lockTokenRef.current);
        } catch {
          // Silent release
        }
        setLockToken(null);
        setIsLockedByMe(false);
      }

      setIsAcquiring(true);
      setLockError(null);

      try {
        const res = await inspectionApi.acquireLock(scopeId, targetAspekId);
        if (res?.success && res?.data?.lock_token) {
          const newToken = res.data.lock_token;
          setLockToken(newToken);
          setIsLockedByMe(true);
          activeAspekRef.current = targetAspekId;
          setIsAcquiring(false);
          return true;
        }
        throw new Error(res?.error || "Gagal mengunci aspek");
      } catch (err: any) {
        const msg = err?.response?.data?.error || err?.message || "Aspek sedang dikunci oleh auditor lain";
        setLockError(msg);
        setLockToken(null);
        setIsLockedByMe(false);
        setIsAcquiring(false);
        return false;
      }
    },
    [scopeId, enabled]
  );

  // Release lock
  const releaseLock = useCallback(async () => {
    if (!scopeId || !activeAspekRef.current || !lockTokenRef.current) return;
    try {
      await inspectionApi.releaseLock(scopeId, activeAspekRef.current, lockTokenRef.current);
    } catch {
      // Silent release
    } finally {
      setLockToken(null);
      setIsLockedByMe(false);
      activeAspekRef.current = null;
      lockTokenRef.current = null;
    }
  }, [scopeId]);

  // Handle lock expired notification from heartbeat
  const handleLockExpired = useCallback(() => {
    setLockToken(null);
    setIsLockedByMe(false);
    setLockError("Sesi edit aspek Anda telah berakhir (expired)");
    toast.error("Waktu penguncian aspek habis. Klik tab aspek untuk mengunci kembali.");
  }, []);

  // Maintain heartbeat every 10s via existing useHeartbeat hook
  useHeartbeat({
    kawasanId: scopeId,
    aspekId: activeAspekRef.current || aspekId || "",
    lockToken,
    enabled: enabled && isLockedByMe && !!activeAspekRef.current,
    onLockExpired: handleLockExpired,
  });

  // Automatically acquire lock when active aspekId changes
  useEffect(() => {
    if (aspekId && enabled && aspekId !== activeAspekRef.current) {
      acquireLock(aspekId);
    }
  }, [aspekId, enabled, acquireLock]);

  // Release lock on component unmount
  useEffect(() => {
    return () => {
      if (scopeId && activeAspekRef.current && lockTokenRef.current) {
        inspectionApi.releaseLock(scopeId, activeAspekRef.current, lockTokenRef.current).catch(() => {});
      }
    };
  }, [scopeId]);

  return {
    lockToken,
    isLockedByMe,
    lockError,
    isAcquiring,
    acquireLock,
    releaseLock,
  };
}
