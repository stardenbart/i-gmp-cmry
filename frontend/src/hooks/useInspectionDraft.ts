"use client";

import { useState, useEffect, useCallback } from "react";

export interface LocalPhotoDraft {
  id: string;
  previewUrl?: string;
  keterangan: string;
  fileName?: string;
}

export interface LocalUraianDraft {
  checking?: "OK" | "NG" | "NA";
  nilai?: number;
  keterangan?: string;
  photos?: LocalPhotoDraft[];
}

export interface InspectionDraftData {
  inspectionId: string;
  updatedAt: string;
  results: Record<string, LocalUraianDraft>; // keyed by uraian_id
}

const STORAGE_PREFIX = "inspection_draft_";

export function useInspectionDraft(inspectionId: string | undefined) {
  const storageKey = inspectionId ? `${STORAGE_PREFIX}${inspectionId}` : null;
  
  const [draftData, setDraftData] = useState<InspectionDraftData | null>(() => {
    if (!storageKey || typeof window === "undefined") return null;
    try {
      const raw = localStorage.getItem(storageKey);
      if (raw) {
        return JSON.parse(raw) as InspectionDraftData;
      }
    } catch (e) {
      console.warn("Failed to load local inspection draft:", e);
    }
    return null;
  });
  const [isRestored, setIsRestored] = useState(false);

  // Sync draft on storageKey change
  useEffect(() => {
    if (!storageKey || typeof window === "undefined") return;

    const timer = window.setTimeout(() => {
      try {
        const raw = localStorage.getItem(storageKey);
        if (raw) {
          const parsed = JSON.parse(raw) as InspectionDraftData;
          setDraftData(parsed);
        } else {
          setDraftData(null);
        }
      } catch (e) {
        console.warn("Failed to load local inspection draft:", e);
      }
    }, 0);
    return () => window.clearTimeout(timer);
  }, [storageKey]);

  // Save draft
  const saveDraft = useCallback(
    (results: Record<string, LocalUraianDraft>) => {
      if (!storageKey || typeof window === "undefined") return;

      try {
        const payload: InspectionDraftData = {
          inspectionId: inspectionId!,
          updatedAt: new Date().toISOString(),
          results,
        };
        localStorage.setItem(storageKey, JSON.stringify(payload));
        setDraftData(payload);
      } catch (e) {
        console.warn("Failed to save local inspection draft:", e);
      }
    },
    [storageKey, inspectionId]
  );

  // Clear draft
  const clearDraft = useCallback(() => {
    if (!storageKey || typeof window === "undefined") return;
    try {
      localStorage.removeItem(storageKey);
      setDraftData(null);
    } catch (e) {
      console.warn("Failed to clear local inspection draft:", e);
    }
  }, [storageKey]);

  return {
    draftData,
    saveDraft,
    clearDraft,
    hasDraft: !!draftData && Object.keys(draftData.results || {}).length > 0,
    isRestored,
    setIsRestored,
  };
}
