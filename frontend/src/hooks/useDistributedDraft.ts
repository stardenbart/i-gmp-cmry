"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { inspectionApi } from "@/lib/api/inspection.api";
import { PhotoItem } from "@/components/Inspection/PhotoUploaderWithKeterangan";

export interface RedisDraftValue {
  checking?: "OK" | "NG" | "NA";
  nilai?: number;
  keterangan?: string;
  photos?: PhotoItem[];
}

// Utility function to split an array into chunks/batches
function chunkArray<T>(items: T[], chunkSize: number): T[][] {
  if (chunkSize <= 0) return [items];
  const chunks: T[][] = [];
  for (let i = 0; i < items.length; i += chunkSize) {
    chunks.push(items.slice(i, i + chunkSize));
  }
  return chunks;
}

export function useDistributedDraft(kawasanId: string | undefined, batchSize = 5) {
  const [draftsByAspek, setDraftsByAspek] = useState<Record<string, any>>({});
  const [mergedUraianValues, setMergedUraianValues] = useState<Record<string, any>>({});
  const [mergedPhotosMap, setMergedPhotosMap] = useState<Record<string, PhotoItem[]>>({});
  const [isLoadingDrafts, setIsLoadingDrafts] = useState(false);
  const debounceTimersRef = useRef<Record<string, NodeJS.Timeout>>({});

  // Fetch all draft states from Redis using Batching + Chunking
  const fetchAllDrafts = useCallback(
    async (aspekIds?: string[]) => {
      if (!kawasanId) return;
      setIsLoadingDrafts(true);
      try {
        const res = await inspectionApi.getAllDrafts(kawasanId);
        const rawDrafts: Record<string, any> = res?.data || {};
        setDraftsByAspek(rawDrafts);

        const draftEntries = Object.entries(rawDrafts);
        if (draftEntries.length === 0) {
          setMergedUraianValues({});
          setMergedPhotosMap({});
          setIsLoadingDrafts(false);
          return;
        }

        // Filter entries if specific aspekIds are provided
        const targetEntries =
          aspekIds && aspekIds.length > 0
            ? draftEntries.filter(([aspekId]) => aspekIds.includes(aspekId))
            : draftEntries;

        // Apply Chunking: Split entries into batches of batchSize
        const entryChunks = chunkArray(targetEntries, batchSize);
        const formDefaults: Record<string, any> = {};
        const photosMap: Record<string, PhotoItem[]> = {};

        // Process each chunk asynchronously to avoid main-thread blocking
        for (const chunk of entryChunks) {
          await new Promise<void>((resolve) => {
            chunk.forEach(([_aspekId, aspekDraft]) => {
              if (!aspekDraft?.data) return;
              try {
                const parsedData =
                  typeof aspekDraft.data === "string"
                    ? JSON.parse(aspekDraft.data)
                    : aspekDraft.data;

                Object.entries(parsedData).forEach(([uId, val]: [string, any]) => {
                  if (val && typeof val === "object") {
                    if (val.checking === "OK") {
                      formDefaults[`nilai_${uId}`] = "2";
                    } else if (val.checking === "NG") {
                      formDefaults[`nilai_${uId}`] = "0";
                    } else if (val.nilai !== undefined && val.nilai !== null) {
                      formDefaults[`nilai_${uId}`] = (val.nilai === 2 || val.nilai >= 80) ? "2" : "0";
                    }
                    if (val.keterangan !== undefined) {
                      formDefaults[`ket_${uId}`] = val.keterangan;
                    }
                    if (Array.isArray(val.photos) && val.photos.length > 0) {
                      photosMap[uId] = val.photos;
                    }
                  }
                });
              } catch (e) {
                console.warn("Error parsing draft JSON chunk entry:", e);
              }
            });

            // Yield to browser main-thread event loop
            setTimeout(resolve, 0);
          });
        }

        setMergedUraianValues(formDefaults);
        setMergedPhotosMap(photosMap);
      } catch (e) {
        console.warn("Failed to fetch Redis drafts:", e);
      } finally {
        setIsLoadingDrafts(false);
      }
    },
    [kawasanId, batchSize]
  );

  useEffect(() => {
    fetchAllDrafts();
  }, [fetchAllDrafts]);

  // Save single aspect draft to Redis with debounce
  const saveAspekDraft = useCallback(
    (
      aspekId: string,
      lockToken: string,
      aspekDataPayload: Record<string, RedisDraftValue>,
      skor: number = 0
    ) => {
      if (!kawasanId || !aspekId || !lockToken) return;

      // Clear existing debounce timer for this aspek
      if (debounceTimersRef.current[aspekId]) {
        clearTimeout(debounceTimersRef.current[aspekId]);
      }

      // Optimistically update local state immediately
      setDraftsByAspek((prev) => ({
        ...prev,
        [aspekId]: {
          ...(prev[aspekId] || {}),
          data: JSON.stringify(aspekDataPayload),
          skor,
          last_saved_at: Math.floor(Date.now() / 1000),
        },
      }));

      // Debounce call to Redis API (500ms)
      debounceTimersRef.current[aspekId] = setTimeout(async () => {
        try {
          await inspectionApi.saveAspekDraft(kawasanId, aspekId, lockToken, {
            data: aspekDataPayload,
            skor,
          });
        } catch (e) {
          console.warn(`Background save to Redis failed for aspek ${aspekId}:`, e);
        }
      }, 500);
    },
    [kawasanId]
  );

  return {
    draftsByAspek,
    mergedUraianValues,
    mergedPhotosMap,
    isLoadingDrafts,
    fetchAllDrafts,
    saveAspekDraft,
  };
}
