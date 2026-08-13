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

export function useDistributedDraft(inspectionId: string | undefined, batchSize = 5) {
  const [draftsByAspek, setDraftsByAspek] = useState<Record<string, any>>({});
  const [mergedUraianValues, setMergedUraianValues] = useState<Record<string, any>>({});
  const [mergedPhotosMap, setMergedPhotosMap] = useState<Record<string, PhotoItem[]>>({});
  const [isLoadingDrafts, setIsLoadingDrafts] = useState(false);
  const debounceTimersRef = useRef<Record<string, NodeJS.Timeout>>({});
  // Foto debounce terpisah (2s) agar base64 besar tidak ikut setiap keystroke
  const photoDebounceRef = useRef<Record<string, NodeJS.Timeout>>({});
  const lastPhotoHashRef = useRef<Record<string, string>>({});

  // Fetch all draft states from Redis using Batching + Chunking
  const fetchAllDrafts = useCallback(
    async (aspekIds?: string[]) => {
      if (!inspectionId) return;
      setIsLoadingDrafts(true);
      try {
        const res = await inspectionApi.getAllDrafts(inspectionId);
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

                Object.entries(parsedData).forEach(([key, val]: [string, any]) => {
                  if (val && typeof val === "object") {
                    if (val.checking === "OK") {
                      formDefaults[`nilai_${key}`] = "2";
                    } else if (val.checking === "NG") {
                      formDefaults[`nilai_${key}`] = "0";
                    } else if (val.nilai !== undefined && val.nilai !== null) {
                      formDefaults[`nilai_${key}`] = (val.nilai === 2 || val.nilai >= 80) ? "2" : "0";
                    }
                    if (val.keterangan !== undefined) {
                      formDefaults[`ket_${key}`] = val.keterangan;
                    }
                    if (Array.isArray(val.photos) && val.photos.length > 0) {
                      // Filter out placeholder yang belum ada foto penuhnya
                      // __base64_pending__ = foto sedang menunggu sync penuh (2s debounce)
                      const validPhotos = val.photos.filter(
                        (p: any) => p.previewUrl !== "__base64_pending__"
                      );
                      if (validPhotos.length > 0) {
                        photosMap[key] = validPhotos;
                      } else if (val.photos.length > 0) {
                        // Tetap simpan metadata foto (id, keterangan, hei) walau previewUrl pending
                        photosMap[key] = val.photos.map((p: any) => ({ ...p, previewUrl: "" }));
                      }
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
    [inspectionId, batchSize]
  );

  useEffect(() => {
    fetchAllDrafts();
  }, [fetchAllDrafts]);

  // Save single aspect draft to Redis with debounce
  // Foto base64 hanya di-sync saat benar-benar berubah (debounce 2s) untuk hemat bandwidth
  const saveAspekDraft = useCallback(
    (
      aspekId: string,
      lockToken: string,
      aspekDataPayload: Record<string, RedisDraftValue>,
      skor: number = 0
    ) => {
      if (!inspectionId || !aspekId || !lockToken) return;

      // Hitung hash foto sederhana untuk deteksi perubahan
      const photoHash = JSON.stringify(
        Object.fromEntries(
          Object.entries(aspekDataPayload).map(([k, v]) => [
            k,
            (v.photos || []).map((p) => p.id + (p.previewUrl?.length || 0)).join(","),
          ])
        )
      );
      const photosChanged = photoHash !== lastPhotoHashRef.current[aspekId];

      // Payload tanpa base64 untuk sync metadata cepat (per keystroke)
      const metadataPayload = Object.fromEntries(
        Object.entries(aspekDataPayload).map(([k, v]) => [
          k,
          {
            ...v,
            photos: (v.photos || []).map((p) => ({
              ...p,
              // Strip base64 dari metadata payload — hanya kirim saat foto berubah
              previewUrl: p.previewUrl?.startsWith("data:") ? "__base64_pending__" : p.previewUrl || "",
            })),
          },
        ])
      );

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

      // Debounce metadata (tanpa base64) ke Redis (500ms per keystroke)
      debounceTimersRef.current[aspekId] = setTimeout(async () => {
        try {
          await inspectionApi.saveAspekDraft(inspectionId, aspekId, lockToken, {
            data: metadataPayload,
            skor,
          });
        } catch (e) {
          console.warn(`Background save metadata to Redis failed for aspek ${aspekId}:`, e);
        }
      }, 500);

      // Sync foto penuh (dengan base64) hanya saat ada perubahan foto (debounce 2s)
      if (photosChanged) {
        lastPhotoHashRef.current[aspekId] = photoHash;
        if (photoDebounceRef.current[aspekId]) {
          clearTimeout(photoDebounceRef.current[aspekId]);
        }
        photoDebounceRef.current[aspekId] = setTimeout(async () => {
          try {
            await inspectionApi.saveAspekDraft(inspectionId, aspekId, lockToken, {
              data: aspekDataPayload, // Full payload dengan base64
              skor,
            });
          } catch (e) {
            console.warn(`Background save photos to Redis failed for aspek ${aspekId}:`, e);
          }
        }, 2000);
      }
    },
    [inspectionId]
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
