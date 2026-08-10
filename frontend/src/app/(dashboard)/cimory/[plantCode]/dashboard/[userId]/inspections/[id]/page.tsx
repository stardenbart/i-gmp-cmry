"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams, useRouter, useSearchParams, usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import {
  ArrowLeft,
  ArrowUp,
  CheckCircle2,
  FileDown,
  Trash2,
  Edit,
  Play,
  AlertTriangle,
  Loader2,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  ChevronUp,
  Info,
  Lock,
  Unlock,
  Layers,
  ListChecks,
} from "lucide-react";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { useEffect, useState, useCallback, useRef } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { inspectionApi } from "@/lib/api/inspection.api";
import { issueApi } from "@/lib/api/issue.api";
import { picApi } from "@/lib/api/pic.api";
import { api } from "@/lib/api/axios";
import { usePermissions } from "@/lib/usePermissions";
import { useAuthStore } from "@/stores/authStore";
import {
  PhotoUploaderWithKeterangan,
  PhotoItem,
  dataURLtoFile,
} from "@/components/Inspection/PhotoUploaderWithKeterangan";
import { useDistributedDraft } from "@/hooks/useDistributedDraft";
import { useAspekLock } from "@/hooks/useAspekLock";
import { AspekStatusPanel } from "@/components/Inspection/AspekStatusPanel";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import {
  setActiveAspekIndex,
  setActiveDetailIndex,
  setHighlightedUraianId,
  setNgPhotosMap,
  updateItemPhotos,
  setShowCancelDialog,
  setShowScrollTop,
  resetSession,
} from "@/store/slices/inspectionSessionSlice";

export default function InspectionDetailPage() {
  const { hasPermission, isLoading: isGuardLoading } = usePermissions();
  const isAuditor = hasPermission("PERM-INSP-R");
  const user = useAuthStore((state) => state.user);
  const { id, userId, plantCode } = useParams() as { id: string; userId: string; plantCode?: string };
  const router = useRouter();
  const searchParams = useSearchParams();
  const pathname = usePathname();
  const queryClient = useQueryClient();

  const dispatch = useAppDispatch();
  const {
    activeAspekIndex,
    activeDetailIndex,
    highlightedUraianId,
    ngPhotosMap,
    showCancelDialog,
    showScrollTop,
  } = useAppSelector((state) => state.inspectionSession);

  const targetAspekParam = searchParams.get("aspek") || searchParams.get("activeAspek");
  const targetUraianParam = searchParams.get("uraian") || searchParams.get("uraian_id");

  const [isSaving, setIsSaving] = useState(false);
  const [isCanceling, setIsCanceling] = useState(false);
  const [isReopening, setIsReopening] = useState(false);
  const [isInfoOpen, setIsInfoOpen] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      if (window.scrollY > 300) {
        dispatch(setShowScrollTop(true));
      } else {
        dispatch(setShowScrollTop(false));
      }
    };

    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, [dispatch]);

  // Cleanup Redux session state when switching inspections or unmounting
  useEffect(() => {
    return () => {
      dispatch(resetSession());
    };
  }, [dispatch, id]);

  const scrollToTop = () => {
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const ngPhotosMapRef = useRef<Record<string, PhotoItem[]>>({});
  const isFormInitialized = useRef(false);

  // Data queries
  const { data: inspectionRes, isLoading: isInspectionLoading } = useQuery({
    queryKey: ["inspection", id],
    queryFn: () => inspectionApi.getById(id),
  });

  const { data: checklistRes, isLoading: isChecklistLoading } = useQuery({
    queryKey: ["inspection_checklist", id],
    queryFn: () => inspectionApi.getChecklist(id),
  });

  const inspection = inspectionRes?.data;
  const checklist = checklistRes?.data;

  // Distributed Redis draft state hook
  // IMPORTANT: scope by inspection_id (not kawasan_id) to isolate each inspection's draft
  // Multiple inspections under the same kawasan but different detail_kawasan share
  // the same kawasan_id and uraian_ids, so using kawasan_id causes data bleeding.
  const inspectionScopeId = inspection?.inspection_id;
  const kawasanId = inspection?.kawasan_id; // kept for PIC query only
  const {
    draftsByAspek,
    mergedUraianValues,
    mergedPhotosMap,
    isLoadingDrafts,
    fetchAllDrafts,
    saveAspekDraft,
  } = useDistributedDraft(inspectionScopeId);

  const aspeksList = checklist?.aspeks || [];
  const currentAspek = aspeksList[activeAspekIndex] || aspeksList[0];
  const currentAspekId = currentAspek?.aspek_id || null;

  // Plant & Auditor Authorization
  const isSamePlant =
    !user?.plant_id ||
    !inspection?.plant_id ||
    user.plant_id === inspection.plant_id;

  const isOngoing = inspection?.status === "Ongoing" || inspection?.status === "Draft";
  const isCompleted = inspection?.status === "Completed" || inspection?.status === "Approved";
  const canEditInspection = isAuditor && isSamePlant && isOngoing;

  // Distributed Aspect Lock Hook
  const {
    lockToken,
    isLockedByMe,
    lockError,
    isAcquiring,
    acquireLock,
    releaseLock,
  } = useAspekLock({
    scopeId: inspectionScopeId || "",
    aspekId: currentAspekId,
    enabled: canEditInspection,
  });

  const { data: picData } = useQuery({
    queryKey: ["pic_mapping", inspection?.area_id, inspection?.kawasan_id],
    queryFn: () =>
      picApi.getAll({ area_id: inspection?.area_id, kawasan_id: inspection?.kawasan_id }),
    enabled: !!inspection?.area_id && !!inspection?.kawasan_id,
  });

  const { register, handleSubmit, watch, setValue, formState: { errors }, reset } = useForm();

  // Synchronize URL search params when active aspect or uraian changes
  const updateUrlParams = useCallback(
    (aspekId?: string | null, uraianId?: string | null) => {
      if (typeof window === "undefined") return;
      const params = new URLSearchParams(window.location.search);
      if (aspekId) {
        params.set("aspek", aspekId);
      }
      if (uraianId) {
        params.set("uraian", uraianId);
      } else if (uraianId === null) {
        params.delete("uraian");
        params.delete("uraian_id");
      }
      const newSearch = params.toString();
      const newUrl = newSearch ? `${pathname}?${newSearch}` : pathname;
      window.history.replaceState(null, "", newUrl);
    },
    [pathname]
  );

  // Deep-linking: auto-switch active aspect and scroll into view when URL params exist
  useEffect(() => {
    if (!checklist?.aspeks || checklist.aspeks.length === 0) return;

    let foundAspekIndex = -1;
    let foundDetailIndex: number | "all" = "all";

    if (targetUraianParam) {
      checklist.aspeks.forEach((aspek: any, aIdx: number) => {
        aspek.details?.forEach((detail: any, dIdx: number) => {
          detail.uraians?.forEach((uraian: any) => {
            if (uraian.uraian_id === targetUraianParam) {
              foundAspekIndex = aIdx;
              foundDetailIndex = dIdx;
            }
          });
        });
      });
    }

    if (foundAspekIndex === -1 && targetAspekParam) {
      const aIdx = checklist.aspeks.findIndex(
        (a: any) =>
          a.aspek_id === targetAspekParam ||
          a.aspek_name?.toLowerCase() === targetAspekParam.toLowerCase()
      );
      if (aIdx !== -1) {
        foundAspekIndex = aIdx;
      }
    }

    if (foundAspekIndex !== -1 && foundAspekIndex !== activeAspekIndex) {
      dispatch(setActiveAspekIndex(foundAspekIndex));
      if (foundDetailIndex !== "all") {
        dispatch(setActiveDetailIndex(foundDetailIndex));
      }
    }

    if (targetUraianParam) {
      dispatch(setHighlightedUraianId(targetUraianParam));
      const scrollTimer = setTimeout(() => {
        const el = document.getElementById(`uraian-${targetUraianParam}`);
        if (el) {
          el.scrollIntoView({ behavior: "smooth", block: "center" });
        }
      }, 500);

      const unhighlightTimer = setTimeout(() => {
        setHighlightedUraianId(null);
      }, 3500);
      return () => {
        clearTimeout(scrollTimer);
        clearTimeout(unhighlightTimer);
      };
    }
  }, [checklist, targetAspekParam, targetUraianParam]);

  // Helper to generate scoped photo map keys (isolated per Detail Aspek and Uraian)
  const getPhotoKey = useCallback((detailId: string | undefined | null, uraianId: string) => {
    return detailId ? `${detailId}_${uraianId}` : uraianId;
  }, []);

  // Helper to calculate completion progress for a single Aspek
  const getAspekProgress = useCallback(
    (aspek: any) => {
      let total = 0;
      let answered = 0;
      const formValues = watch();

      aspek?.details?.forEach((detail: any) => {
        detail?.uraians?.forEach((uraian: any) => {
          total++;
          const pKey = getPhotoKey(detail.detail_id, uraian.uraian_id);
          const val = formValues[`nilai_${pKey}`] ?? formValues[`nilai_${uraian.uraian_id}`];
          if (val !== undefined && val !== null && val !== "") {
            answered++;
          }
        });
      });

      return {
        total,
        answered,
        isComplete: total > 0 && answered === total,
      };
    },
    [watch, getPhotoKey]
  );

  // Helper to calculate completion progress for a single Detail Aspek
  const getDetailProgress = useCallback(
    (detail: any) => {
      let total = 0;
      let answered = 0;
      const formValues = watch();

      detail?.uraians?.forEach((uraian: any) => {
        total++;
        const pKey = getPhotoKey(detail?.detail_id, uraian.uraian_id);
        const val = formValues[`nilai_${pKey}`] ?? formValues[`nilai_${uraian.uraian_id}`];
        if (val !== undefined && val !== null && val !== "") {
          answered++;
        }
      });

      return {
        total,
        answered,
        isComplete: total > 0 && answered === total,
      };
    },
    [watch, getPhotoKey]
  );

  // Helper to calculate total overall progress
  const getOverallProgress = useCallback(() => {
    let total = 0;
    let answered = 0;
    const formValues = watch();

    checklist?.aspeks?.forEach((aspek: any) => {
      aspek?.details?.forEach((detail: any) => {
        detail?.uraians?.forEach((uraian: any) => {
          total++;
          const pKey = getPhotoKey(detail.detail_id, uraian.uraian_id);
          const val = formValues[`nilai_${pKey}`] ?? formValues[`nilai_${uraian.uraian_id}`];
          if (val !== undefined && val !== null && val !== "") {
            answered++;
          }
        });
      });
    });

    return {
      total,
      answered,
      percent: total > 0 ? Math.round((answered / total) * 100) : 0,
    };
  }, [checklist, watch, getPhotoKey]);

  // Populate form state when checklist and Redis drafts are loaded
  useEffect(() => {
    if (!checklist?.aspeks || isLoadingDrafts) return;

    const defaultValues: Record<string, any> = {};
    const photoStateMap: Record<string, PhotoItem[]> = {};

    checklist.aspeks.forEach((aspek: any) => {
      aspek.details?.forEach((detail: any) => {
        detail.uraians?.forEach((uraian: any) => {
          const uId = uraian.uraian_id;
          const pKey = getPhotoKey(detail.detail_id, uId);

          // 1. Check existing server DB result first (PostgreSQL)
          if (uraian.result) {
            const serverVal =
              uraian.result.checking === "OK"
                ? "2"
                : uraian.result.checking === "NG"
                ? "0"
                : "";
            defaultValues[`nilai_${pKey}`] = serverVal;
            defaultValues[`nilai_${uId}`] = serverVal;
          }

          // 2. Redis draft override (shared across auditors)
          const draftVal = mergedUraianValues[`nilai_${pKey}`] ?? mergedUraianValues[`nilai_${uId}`];
          if (draftVal !== undefined && draftVal !== null && draftVal !== "") {
            defaultValues[`nilai_${pKey}`] = draftVal;
            defaultValues[`nilai_${uId}`] = draftVal;
          }

          // 3. Photos (scoped pKey takes precedence; un-scoped uId only used if detail_id is not set)
          const draftPhotos = mergedPhotosMap[pKey] || (!detail.detail_id ? mergedPhotosMap[uId] : undefined);
          if (draftPhotos && draftPhotos.length > 0) {
            photoStateMap[pKey] = draftPhotos;
          }
        });
      });
    });

    // Directly set defaultValues from DB & Redis (do NOT merge prev to avoid preserving empty fields)
    reset(defaultValues);
    if (Object.keys(photoStateMap).length > 0) {
      ngPhotosMapRef.current = photoStateMap;
      dispatch(setNgPhotosMap(photoStateMap));
    }
    
    // Enable form sync after a micro-delay to prevent watch() from sending empty resets
    setTimeout(() => {
      isFormInitialized.current = true;
    }, 100);
  }, [checklist, mergedUraianValues, mergedPhotosMap, isLoadingDrafts, reset, getPhotoKey]);

  // Save current active aspect changes to Redis
  const syncCurrentAspekToRedis = useCallback(
    (formValues: any) => {
      if (
        !isFormInitialized.current ||
        !currentAspek ||
        !currentAspekId ||
        !lockToken ||
        !isLockedByMe ||
        !isOngoing
      )
        return;

      const aspekPayload: Record<string, any> = {};

      currentAspek.details?.forEach((detail: any) => {
        detail.uraians?.forEach((uraian: any) => {
          const uId = uraian.uraian_id;
          const pKey = getPhotoKey(detail.detail_id, uId);
          const val = formValues[`nilai_${pKey}`] ?? formValues[`nilai_${uId}`];
          const photos = ngPhotosMapRef.current[pKey] || (!detail.detail_id ? (ngPhotosMapRef.current[uId] || []) : []);

          if (val !== undefined && val !== null && val !== "") {
            const checking = val === "2" ? "OK" : "NG";
            aspekPayload[pKey] = {
              checking,
              nilai: checking === "OK" ? (uraian.standard_score || 100) : 0,
              keterangan: "",
              photos: photos.map((p) => ({
                id: p.id,
                previewUrl: p.previewUrl?.startsWith("data:") ? "" : p.previewUrl,
                keterangan: p.keterangan,
                hei_id: p.hei_id,
                hei_category: p.hei_category,
                habit_id: p.habit_id,
                equipment_id: p.equipment_id,
                infrastructure_id: p.infrastructure_id,
              })),
            };
          }
        });
      });

      // Prevent sending empty payload that wipes existing Redis draft
      if (Object.keys(aspekPayload).length === 0) return;

      saveAspekDraft(currentAspekId, lockToken, aspekPayload, 0);
    },
    [currentAspek, currentAspekId, lockToken, isLockedByMe, isOngoing, saveAspekDraft, getPhotoKey]
  );

  // Subscribe to form changes for active aspect
  useEffect(() => {
    if (!isFormInitialized.current) return;
    const subscription = watch((values) => {
      syncCurrentAspekToRedis(values);
    });
    return () => subscription.unsubscribe();
  }, [watch, syncCurrentAspekToRedis]);

  // Handler for photo updates per NG item (Upload file instantly to MinIO if new)
  const handlePhotosChange = async (key: string, photos: PhotoItem[]) => {
    const processedPhotos: PhotoItem[] = await Promise.all(
      photos.map(async (photo) => {
        if (photo.file && !photo.previewUrl?.includes("http")) {
          try {
            const fd = new FormData();
            fd.append("file", photo.file);
            fd.append("file_type", "image");
            fd.append("inspection_id", id);
            const res = await api.post("/uploads/file", fd, {
              headers: { "Content-Type": "multipart/form-data" },
            });
            const uploadedUrl = res.data?.data?.file_url || res.data?.data?.processed_url;
            if (uploadedUrl) {
              return { ...photo, previewUrl: uploadedUrl, file: undefined };
            }
          } catch (e) {
            console.warn("Direct MinIO upload for draft photo failed:", e);
          }
        }
        return photo;
      })
    );

    ngPhotosMapRef.current = {
      ...ngPhotosMapRef.current,
      [key]: processedPhotos,
    };
    dispatch(updateItemPhotos({ pKey: key, photos: processedPhotos }));

    const currentFormValues = watch();
    syncCurrentAspekToRedis(currentFormValues);
  };

  // Submit Final ("Selesaikan Audit")
  const onFinalSubmit = async (formData: any) => {
    if (!checklist?.aspeks || !inspection) return;

    let unansweredCount = 0;
    let missingPhotoInfoCount = 0;

    const resultsPayload: any[] = [];
    const ngUraianTasks: { uraian: any; keterangan: string; photos: PhotoItem[] }[] = [];

    checklist.aspeks.forEach((aspek: any) => {
      aspek.details?.forEach((detail: any) => {
        detail.uraians?.forEach((uraian: any) => {
          const uId = uraian.uraian_id;
          const pKey = getPhotoKey(detail.detail_id, uId);
          const val = formData[`nilai_${pKey}`] ?? formData[`nilai_${uId}`];
          const ket = formData[`ket_${pKey}`] ?? formData[`ket_${uId}`];
          const photos = ngPhotosMap[pKey] || (!detail.detail_id ? (ngPhotosMap[uId] || []) : []);

          if (!val) {
            unansweredCount++;
          } else {
            const checking = val === "2" ? "OK" : "NG";
            resultsPayload.push({
              uraian_id: uId,
              checking,
              nilai: checking === "OK" ? (uraian.standard_score || 100) : 0,
              keterangan: ket || "",
            });

            if (checking === "NG") {
              if (photos.length === 0) {
                missingPhotoInfoCount++;
              }
              const hasEmptyPhotoKet = photos.some((p) => !p.keterangan.trim());
              if (hasEmptyPhotoKet) {
                missingPhotoInfoCount++;
              }

              ngUraianTasks.push({
                uraian,
                keterangan: photos[0]?.keterangan || ket || "Temuan NG pada inspeksi",
                photos,
              });
            }
          }
        });
      });
    });

    if (unansweredCount > 0) {
      toast.error(`Harap isi semua uraian penilaian (${unansweredCount} uraian belum dinilai).`);
      return;
    }

    if (missingPhotoInfoCount > 0) {
      toast.error("Setiap item NG wajib memiliki minimal 1 foto bukti DAN keterangan spesifik per foto.");
      return;
    }

    setIsSaving(true);

    try {
      // 1. Bulk Save Results to PostgreSQL DB Core
      await inspectionApi.bulkSaveResults(id, resultsPayload);

      // 2. Process NG Issues & Upload Photos with Per-Photo Keterangan
      const picUserId = picData?.items?.[0]?.user_id || user?.id || "";

      if (ngUraianTasks.length > 0) {
        const updatedChecklistRes = await inspectionApi.getChecklist(id);
        const updatedChecklist = updatedChecklistRes?.data;

        for (const task of ngUraianTasks) {
          let savedResultId = "";
          updatedChecklist?.aspeks?.forEach((a: any) => {
            a.details?.forEach((d: any) => {
              d.uraians?.forEach((u: any) => {
                if (u.uraian_id === task.uraian.uraian_id && u.result) {
                  savedResultId = u.result.result_id;
                }
              });
            });
          });

          if (savedResultId) {
            const issueRes = await issueApi.create({
              result_id: savedResultId,
              issue_pic_user_id: picUserId,
              keterangan: task.keterangan || "Temuan NG pada inspeksi",
            });

            const createdIssueId = issueRes?.data?.issue_id;

            if (createdIssueId && task.photos.length > 0) {
              for (const photoItem of task.photos) {
                let photoFile = photoItem.file;
                if (!photoFile && photoItem.previewUrl) {
                  if (photoItem.previewUrl.startsWith("data:")) {
                    photoFile = dataURLtoFile(photoItem.previewUrl, `photo_${Date.now()}.jpg`);
                  } else {
                    try {
                      const resp = await fetch(photoItem.previewUrl);
                      const blob = await resp.blob();
                      photoFile = new File([blob], `photo_${Date.now()}.jpg`, { type: blob.type || "image/jpeg" });
                    } catch (e) {
                      console.warn("Failed to convert photo previewUrl to File blob:", e);
                    }
                  }
                }
                if (photoFile && photoFile.size > 0) {
                  const fd = new FormData();
                  fd.append("photo", photoFile);
                  fd.append("photo_type", "Initial");
                  fd.append("keterangan", photoItem.keterangan || "");
                  if (photoItem.hei_id) fd.append("hei_id", photoItem.hei_id);
                  if (photoItem.habit_id) fd.append("habit_id", photoItem.habit_id);
                  if (photoItem.equipment_id) fd.append("equipment_id", photoItem.equipment_id);
                  if (photoItem.infrastructure_id) fd.append("infrastructure_id", photoItem.infrastructure_id);
                  await issueApi.uploadPhoto(createdIssueId, fd);
                }
              }
            }
          }
        }
      }

      // 3. Update Inspection Status to Completed
      await inspectionApi.updateStatus(id, "Completed");

      // 4. Release active aspect lock
      await releaseLock();

      toast.success("Inspeksi berhasil diselesaikan dan disinkronkan ke DB Core!");
      await queryClient.invalidateQueries({ queryKey: ["inspection", id] });
      await queryClient.invalidateQueries({ queryKey: ["inspection_checklist", id] });
      await queryClient.invalidateQueries({ queryKey: ["inspections-filter"] });
      await queryClient.invalidateQueries({ queryKey: ["my-ongoing-inspections"] });
    } catch (err: any) {
      toast.error(err.response?.data?.message || "Gagal menyelesaikan inspeksi");
    } finally {
      setIsSaving(false);
    }
  };

  // Handler for "Batalkan Inspeksi"
  const handleCancelInspection = async () => {
    try {
      setIsCanceling(true);
      await releaseLock();
      await inspectionApi.delete(id);
      await queryClient.invalidateQueries({ queryKey: ["inspections-filter"] });
      await queryClient.invalidateQueries({ queryKey: ["my-ongoing-inspections"] });
      toast.success("Inspeksi telah dibatalkan dan kunci lokasi dilepas.");
      router.push(`/cimory/${plantCode || "all"}/dashboard/${userId}/inspections`);
    } catch (err: any) {
      toast.error(err.response?.data?.message || "Gagal membatalkan inspeksi");
    } finally {
      setIsCanceling(false);
      dispatch(setShowCancelDialog(false));
    }
  };

  // Handler for "Edit Inspeksi" (Re-open Completed inspection)
  const handleReopenForEdit = async () => {
    try {
      setIsReopening(true);
      await inspectionApi.updateStatus(id, "Ongoing");
      isFormInitialized.current = false;
      ngPhotosMapRef.current = {};
      await fetchAllDrafts();
      toast.success("Inspeksi dibuka kembali untuk diedit.");
      await queryClient.invalidateQueries({ queryKey: ["inspection", id] });
      await queryClient.invalidateQueries({ queryKey: ["inspection_checklist", id] });
      await queryClient.invalidateQueries({ queryKey: ["inspections-filter"] });
      await queryClient.invalidateQueries({ queryKey: ["my-ongoing-inspections"] });
    } catch (err: any) {
      toast.error(err.response?.data?.message || "Gagal mengedit inspeksi");
    } finally {
      setIsReopening(false);
    }
  };

  if (isGuardLoading || !user || isInspectionLoading || isChecklistLoading) {
    return (
      <div className="flex justify-center p-8">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!isAuditor) {
    return (
      <div className="flex justify-center items-center p-12 text-muted-foreground">
        Anda tidak memiliki akses ke halaman ini.
      </div>
    );
  }

  if (!inspection) {
    return (
      <div className="flex flex-col items-center justify-center p-12 text-center space-y-4 max-w-md mx-auto my-12 bg-card rounded-2xl border border-border shadow-sm">
        <div className="p-3 bg-amber-500/10 text-amber-600 rounded-full">
          <AlertTriangle className="w-8 h-8" />
        </div>
        <div className="space-y-1">
          <h3 className="font-bold text-lg">Inspeksi Tidak Ditemukan</h3>
          <p className="text-sm text-muted-foreground">
            Data inspeksi dengan ID <span className="font-mono font-semibold">{id}</span> tidak ditemukan atau telah dibatalkan.
          </p>
        </div>
        <Link href={`/cimory/${plantCode || "all"}/dashboard/${userId}/inspections`}>
          <Button variant="outline" className="gap-2 mt-2">
            <ArrowLeft className="w-4 h-4" /> Kembali ke Daftar Inspeksi
          </Button>
        </Link>
      </div>
    );
  }

  const statusClass = isCompleted
    ? "bg-green-500/10 text-green-600 border border-green-500/20"
    : isOngoing
    ? "bg-amber-500/10 text-amber-600 border border-amber-500/20 animate-pulse"
    : "bg-zinc-500/10 text-zinc-500 border border-zinc-500/20";

  const hasExistingResults = checklist?.aspeks?.some((a: any) =>
    a.details?.some((d: any) =>
      d.uraians?.some((u: any) => !!u.result)
    )
  );

  const overallProgress = getOverallProgress();

  // Filter displayed Detail Aspeks based on activeDetailIndex (Level 2 -> Level 3 selection)
  const detailsList = currentAspek?.details || [];
  const displayedDetails =
    activeDetailIndex === "all"
      ? detailsList
      : detailsList[activeDetailIndex as number]
      ? [detailsList[activeDetailIndex as number]]
      : detailsList;

  return (
    <div className="space-y-3 sm:space-y-6 w-full max-w-5xl mx-auto px-2 sm:px-0 pb-36 sm:pb-12 overflow-x-hidden">
      {/* Header Bar */}
      <div className="flex flex-col gap-2.5 sm:gap-3">
        <div className="flex items-center gap-2.5 sm:gap-3 min-w-0">
          <Link href={`/cimory/${plantCode || "all"}/dashboard/${userId}/inspections`}>
            <Button variant="ghost" size="icon" className="rounded-full shrink-0 h-9 w-9">
              <ArrowLeft className="h-4 w-4" />
            </Button>
          </Link>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 flex-wrap">
              <h2 className="text-lg sm:text-2xl font-bold tracking-tight">Detail Inspeksi</h2>
              <span className={`text-[10px] sm:text-xs font-semibold px-2 py-0.5 rounded-full shrink-0 ${statusClass}`}>
                {isOngoing ? "● Ongoing" : inspection.status}
              </span>
            </div>
            <p className="text-[11px] sm:text-xs text-muted-foreground mt-0.5 truncate">
              ID: <span className="font-mono">{inspection.inspection_id}</span> · <strong className="text-foreground">{inspection.inspector_name || inspection.inspector_id}</strong>
            </p>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="grid grid-cols-2 sm:flex sm:flex-wrap items-stretch sm:items-center gap-2 w-full">
          {isCompleted && isAuditor && isSamePlant && (
            <Button
              onClick={handleReopenForEdit}
              disabled={isReopening}
              variant="outline"
              className="w-full sm:w-auto sm:flex-none h-10 sm:h-9 px-2.5 sm:px-4 text-xs border-primary min-w-0 text-primary hover:bg-primary/5 rounded-xl font-semibold"
            >
              {isReopening ? (
                <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />
              ) : (
                <Edit className="mr-1.5 h-3.5 w-3.5" />
              )}
              Edit Inspeksi
            </Button>
          )}

          {isOngoing && isAuditor && isSamePlant && (
            <>
              {hasExistingResults ? (
                <Button
                  variant="outline"
                  onClick={async () => {
                    try {
                      setIsSaving(true);
                      await releaseLock();
                      await inspectionApi.updateStatus(id, "Completed");
                      toast.info("Perubahan dibatalkan. Status dikembalikan ke Selesai.");
                      await queryClient.invalidateQueries({ queryKey: ["inspection", id] });
                    } catch (err: any) {
                      toast.error("Gagal membatalkan edit");
                    } finally {
                      setIsSaving(false);
                    }
                  }}
                  disabled={isSaving}
                  className="w-full sm:w-auto sm:flex-none h-10 sm:h-9 px-2.5 sm:px-3 text-xs border-muted-foreground/40 min-w-0 text-muted-foreground hover:bg-muted rounded-xl font-medium"
                >
                  Batal Edit
                </Button>
              ) : (
                <Button
                  variant="destructive"
                  onClick={() => dispatch(setShowCancelDialog(true))}
                  disabled={isSaving || isCanceling}
                  className="w-full sm:w-auto sm:flex-none h-10 sm:h-9 px-2.5 sm:px-3 text-xs bg-red-600 min-w-0 hover:bg-red-700 text-white font-semibold rounded-xl shadow-xs whitespace-nowrap"
                >
                  <Trash2 className="mr-1.5 h-3.5 w-3.5 shrink-0" /> Batalkan
                </Button>
              )}

              <Button
                onClick={handleSubmit(onFinalSubmit)}
                disabled={isSaving || isCanceling}
                className="w-full sm:w-auto sm:flex-none h-10 sm:h-9 px-2.5 sm:px-3 text-xs bg-green-600 min-w-0 hover:bg-green-700 text-white font-semibold rounded-xl shadow-xs whitespace-nowrap"
              >
                {isSaving ? (
                  <>
                    <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin shrink-0" /> Simpan...
                  </>
                ) : (
                  <>
                    <CheckCircle2 className="mr-1.5 h-3.5 w-3.5 shrink-0" /> Selesaikan Audit
                  </>
                )}
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Plant Scope Warning if viewing from different plant */}
      {!isSamePlant && (
        <div className="rounded-2xl border border-amber-500/30 bg-amber-500/10 p-4 flex items-start gap-3">
          <AlertTriangle className="h-5 w-5 text-amber-600 shrink-0 mt-0.5" />
          <div className="text-xs text-muted-foreground">
            <span className="font-bold text-amber-700 dark:text-amber-400 block text-sm">
              Perhatian: Plant Berbeda
            </span>
            Anda terdaftar di Plant <strong>{user?.plant_id}</strong>, sedangkan inspeksi ini milik Plant <strong>{inspection.plant_id || "Lain"}</strong>. Anda hanya dapat melihat data dalam mode Read-Only.
          </div>
        </div>
      )}

      {/* Real-Time Kawasan Aspect Locking Panel */}
      {kawasanId && (
        <AspekStatusPanel
          kawasanId={kawasanId}
          aspeks={aspeksList.map((a: any) => ({ aspek_id: a.aspek_id, aspek_name: a.aspek_name }))}
          currentUserId={user?.id}
          activeAspekId={currentAspekId}
          onSelectAspek={(selectedAspekId) => {
            const idx = aspeksList.findIndex((a: any) => a.aspek_id === selectedAspekId);
            if (idx !== -1) {
              setActiveAspekIndex(idx);
              setActiveDetailIndex(0);
            }
          }}
        />
      )}

      {/* Grid Content: Info Card first (order-1), Checklist Card second (order-2) */}
      <div className="grid gap-4 sm:gap-6 md:grid-cols-3">
        {/* Info Card — Collapsible on mobile, sticky sidebar on desktop */}
        <Card className="p-3 sm:p-5 bg-card/70 backdrop-blur-md md:col-span-1 min-w-0 h-fit md:sticky md:top-24 shadow-xs border-border/80 rounded-2xl order-1 md:order-1 transition-all">
          {/* Header Bar with Collapsible Toggle for Mobile */}
          <div className="flex items-center justify-between gap-2 min-w-0 border-b border-border pb-2 md:pb-2.5">
            <button
              type="button"
              onClick={() => setIsInfoOpen(!isInfoOpen)}
              className="flex items-center gap-2 text-left w-full justify-between md:cursor-default"
            >
              <div className="flex items-center gap-1.5">
                <Info className="w-4 h-4 text-primary shrink-0" />
                <h3 className="font-bold text-xs sm:text-sm text-foreground">Informasi Area</h3>
              </div>
              <div className="flex items-center gap-1.5">
                <span className="text-[10px] text-muted-foreground font-normal hidden sm:inline">Audit Session</span>
                <div className="md:hidden p-1 rounded-md hover:bg-muted text-muted-foreground">
                  {isInfoOpen ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                </div>
              </div>
            </button>
          </div>

          {/* Compact summary pill when collapsed on mobile */}
          {!isInfoOpen && (
            <div className="md:hidden pt-2 flex items-center justify-between gap-2 min-w-0 text-[11px] text-muted-foreground">
              <span className="truncate max-w-[200px]">
                <strong className="text-foreground">{inspection.area_name || inspection.area_id}</strong> · {inspection.kawasan_name || inspection.kawasan_id}
              </span>
              <span className="text-primary font-bold shrink-0">{overallProgress.percent}%</span>
            </div>
          )}

          {/* Expanded Content Grid */}
          <div className={cn(
            "grid grid-cols-2 md:grid-cols-1 gap-2.5 sm:gap-3 text-xs pt-2.5 md:pt-3",
            !isInfoOpen && "hidden md:grid"
          )}>
            <div>
              <span className="text-muted-foreground block text-[10px] font-medium uppercase tracking-wide">Area</span>
              <span className="font-semibold text-foreground text-xs sm:text-sm leading-snug break-words">{inspection.area_name || inspection.area_id}</span>
            </div>
            <div>
              <span className="text-muted-foreground block text-[10px] font-medium uppercase tracking-wide">Kawasan</span>
              <span className="font-semibold text-foreground text-xs sm:text-sm leading-snug break-words">{inspection.kawasan_name || inspection.kawasan_id}</span>
            </div>
            <div>
              <span className="text-muted-foreground block text-[10px] font-medium uppercase tracking-wide">Detail Kawasan</span>
              <span className="font-semibold text-foreground text-xs sm:text-sm leading-snug break-words">{inspection.detail_kawasan_name || inspection.detail_kawasan_id}</span>
            </div>
            <div>
              <span className="text-muted-foreground block text-[10px] font-medium uppercase tracking-wide">Auditor</span>
              <span className="font-semibold text-foreground text-xs sm:text-sm leading-snug break-words">{inspection.inspector_name || inspection.inspector_id}</span>
            </div>
            {inspection.score !== undefined && inspection.score !== null && (
              <div className="col-span-2 md:col-span-1 pt-2 border-t border-border">
                <span className="text-muted-foreground block text-[10px] font-medium uppercase tracking-wide">Skor Inspeksi</span>
                <span className={`text-lg font-bold ${inspection.score >= 80 ? "text-green-600" : "text-amber-600"}`}>
                  {inspection.score.toFixed(1)}%
                </span>
              </div>
            )}

            {/* Overall Progress Bar */}
            <div className="col-span-2 md:col-span-1 pt-2 border-t border-border space-y-1.5">
              <div className="flex items-center justify-between gap-2 min-w-0 text-[11px]">
                <span className="text-muted-foreground font-medium">Progres Penilaian</span>
                <span className="font-bold text-primary">{overallProgress.answered}/{overallProgress.total} ({overallProgress.percent}%)</span>
              </div>
              <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
                <div
                  className="h-full bg-primary transition-all duration-300 rounded-full"
                  style={{ width: `${overallProgress.percent}%` }}
                />
              </div>
            </div>
          </div>
        </Card>

        {/* Checklist Form Card — Second on mobile (order-2) */}
        <Card className="p-3 sm:p-6 bg-card/60 backdrop-blur-md md:col-span-2 min-w-0 shadow-xs border-border/80 space-y-4 sm:space-y-6 rounded-2xl sm:rounded-3xl order-2 md:order-2">
          {/* Header Title */}
          <div className="flex flex-wrap items-center justify-between gap-1.5 border-b border-border pb-2.5 sm:pb-3">
            <div className="flex items-center gap-2">
              <h3 className="font-bold text-sm sm:text-base">Checklist Penilaian Audit</h3>
            </div>
            <span className="text-[10px] sm:text-xs text-muted-foreground font-medium text-right sm:text-left shrink-0">
              {aspeksList.length} Aspek Pengecekan
            </span>
          </div>

          {aspeksList.length === 0 ? (
            <div className="p-6 sm:p-8 text-center text-xs sm:text-sm text-muted-foreground bg-muted/20 rounded-2xl border border-border">
              Belum ada Uraian (Checklist) yang diatur untuk Area ini.
            </div>
          ) : (
            <form className="space-y-4 sm:space-y-6">
              {/* LEVEL 1: Interactive Horizontal Aspek Tabs Bar with Scroll Hint */}
              <div className="space-y-1.5">
                <span className="text-[10px] sm:text-[11px] font-bold uppercase tracking-wider text-muted-foreground block">
                  Langkah 1: Pilih Aspek Audit
                </span>
                <div className="relative">
                  <div className="flex items-center gap-1.5 sm:gap-2 overflow-x-auto pb-2 border-b border-border scrollbar-none snap-x snap-mandatory">
                    {aspeksList.map((aspek: any, idx: number) => {
                      const isCurrent = idx === activeAspekIndex;
                      const { answered, total, isComplete } = getAspekProgress(aspek);

                      return (
                        <button
                          key={aspek.aspek_id || idx}
                          type="button"
                          onClick={() => {
                            dispatch(setActiveAspekIndex(idx));
                            dispatch(setActiveDetailIndex(0));
                            updateUrlParams(aspek.aspek_id, null);
                          }}
                          className={`flex items-center gap-1.5 sm:gap-2 px-2.5 sm:px-3.5 py-2 sm:py-2.5 rounded-xl border text-[11px] sm:text-xs font-semibold whitespace-nowrap transition-all shrink-0 snap-start min-h-[40px] ${
                            isCurrent
                              ? "bg-primary text-primary-foreground border-primary shadow-xs"
                              : isComplete
                              ? "bg-green-500/10 text-green-600 border-green-500/30 hover:bg-green-500/20"
                              : "bg-card text-muted-foreground border-border hover:bg-muted hover:text-foreground"
                          }`}
                        >
                          <span
                            className={`h-5 w-5 rounded-full flex items-center justify-center text-[10px] font-bold shrink-0 ${
                              isCurrent
                                ? "bg-primary-foreground text-primary"
                                : isComplete
                                ? "bg-green-600 text-white"
                                : "bg-muted text-muted-foreground"
                            }`}
                          >
                            {isComplete ? "✓" : idx + 1}
                          </span>
                          <span className="truncate max-w-[120px] sm:max-w-[150px]">{aspek.aspek_name}</span>
                          <span
                            className={`text-[9px] sm:text-[10px] px-1.5 py-0.5 rounded-full font-mono ${
                              isCurrent
                                ? "bg-primary-foreground/20 text-primary-foreground font-bold"
                                : isComplete
                                ? "bg-green-500/20 text-green-700 font-bold"
                                : "bg-muted text-muted-foreground"
                            }`}
                          >
                            {answered}/{total}
                          </span>
                        </button>
                      );
                    })}
                  </div>
                  {/* Visual gradient overlay hint for horizontal scroll on mobile */}
                  <div className="pointer-events-none absolute right-0 top-0 bottom-2 w-6 bg-gradient-to-l from-card to-transparent sm:hidden" />
                </div>
              </div>

              {/* LEVEL 2 & 3: Current Active Aspek & Selected Detail Aspek */}
              {currentAspek && (
                <div className="space-y-4 sm:space-y-6 animate-in fade-in duration-200">
                  {/* Active Aspek Lock & Info Banner */}
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 p-3 sm:p-4 bg-muted/40 rounded-2xl border border-border min-w-0">
                    <div>
                      <span className="text-[10px] sm:text-[11px] font-bold uppercase tracking-wider text-primary">
                        Aspek {activeAspekIndex + 1} dari {aspeksList.length}
                      </span>
                      <h3 className="text-base sm:text-lg font-bold text-foreground mt-0.5 flex flex-wrap items-center gap-2">
                        {currentAspek.aspek_name}
                        {isLockedByMe ? (
                          <span className="text-[11px] sm:text-xs font-normal text-emerald-600 bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 rounded-full flex items-center gap-1">
                            <Lock className="w-3 h-3" /> Dikunci oleh Anda
                          </span>
                        ) : lockError ? (
                          <span className="text-[11px] sm:text-xs font-normal text-destructive bg-destructive/10 border border-destructive/20 px-2 py-0.5 rounded-full flex items-center gap-1">
                            <Lock className="w-3 h-3" /> {lockError}
                          </span>
                        ) : null}
                      </h3>
                    </div>
                    <div className="text-left sm:text-right">
                      <span className="text-[10px] sm:text-xs text-muted-foreground block font-medium">Progres Aspek</span>
                      <span className="text-xs font-bold text-foreground font-mono">
                        {getAspekProgress(currentAspek).answered} / {getAspekProgress(currentAspek).total} Uraian Terisi
                      </span>
                    </div>
                  </div>

                  {/* Lock error overlay if locked by another user */}
                  {!isLockedByMe && canEditInspection && (
                    <div className="p-3 sm:p-4 rounded-xl border border-amber-500/30 bg-amber-500/10 text-xs text-amber-700 dark:text-amber-300 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2">
                      <span className="flex items-center gap-2">
                        <AlertTriangle className="w-4 h-4 shrink-0" />
                        {lockError || "Klik tombol di samping untuk mengambil alih penguncian aspek ini."}
                      </span>
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        onClick={() => acquireLock(currentAspekId!)}
                        disabled={isAcquiring}
                        className="text-xs w-full sm:w-auto self-stretch sm:self-auto"
                      >
                        {isAcquiring ? "Mengunci..." : "Kunci Aspek Ini"}
                      </Button>
                    </div>
                  )}

                  {/* LEVEL 2: Interactive Detail Aspek Selection Sub-Tabs */}
                  {detailsList.length > 0 && (
                    <div className="space-y-2 bg-muted/20 p-2.5 sm:p-3.5 rounded-2xl border border-border/80 min-w-0">
                      <div className="flex items-center justify-between gap-2 min-w-0">
                        <span className="text-xs font-bold text-foreground flex items-center gap-1.5 min-w-0">
                          <Layers className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-primary" /> Langkah 2: Detail Aspek
                        </span>
                        <button
                          type="button"
                          onClick={() => dispatch(setActiveDetailIndex("all"))}
                          className={`text-[10px] sm:text-[11px] font-semibold px-2 py-0.5 sm:py-1 rounded-lg border transition-all ${
                            activeDetailIndex === "all"
                              ? "bg-primary text-primary-foreground border-primary shadow-xs"
                              : "bg-card text-muted-foreground hover:bg-muted"
                          }`}
                        >
                          Semua ({detailsList.length})
                        </button>
                      </div>

                      <div className="relative">
                        <div className="flex items-center gap-1.5 sm:gap-2 overflow-x-auto pb-1 scrollbar-none snap-x snap-mandatory">
                          {detailsList.map((detail: any, dIdx: number) => {
                            const isSelected = activeDetailIndex === dIdx;
                            const { answered, total, isComplete } = getDetailProgress(detail);

                            return (
                              <button
                                key={detail.detail_id || dIdx}
                                type="button"
                                onClick={() => dispatch(setActiveDetailIndex(dIdx))}
                                className={`flex items-center gap-1.5 sm:gap-2 px-2.5 sm:px-3 py-1.5 sm:py-2 rounded-xl border text-[11px] sm:text-xs font-semibold whitespace-nowrap transition-all shrink-0 snap-start min-h-[36px] ${
                                  isSelected
                                    ? "bg-primary/10 text-primary border-primary font-bold shadow-xs ring-1 ring-primary/30"
                                    : isComplete
                                    ? "bg-green-500/10 text-green-600 border-green-500/30 hover:bg-green-500/20"
                                    : "bg-card text-muted-foreground border-border hover:bg-muted"
                                }`}
                              >
                                <span className="truncate max-w-[140px] sm:max-w-[180px]">{detail.detail_name}</span>
                                <span
                                  className={`text-[9px] sm:text-[10px] px-1.5 py-0.5 rounded-full font-mono font-bold ${
                                    isComplete
                                      ? "bg-green-600 text-white"
                                      : "bg-muted text-muted-foreground"
                                  }`}
                                >
                                  {answered}/{total}
                                </span>
                              </button>
                            );
                          })}
                        </div>
                        {/* Visual gradient hint for horizontal scroll on mobile */}
                        <div className="pointer-events-none absolute right-0 top-0 bottom-1 w-6 bg-gradient-to-l from-muted/80 to-transparent sm:hidden" />
                      </div>
                    </div>
                  )}

                  {/* LEVEL 3: Render Uraian Checklist Items under Selected Detail Aspek */}
                  {displayedDetails.map((detail: any, dIndex: number) => (
                    <div
                      key={detail.detail_id || dIndex}
                      className="space-y-3 sm:space-y-4 bg-card/40 p-2.5 sm:p-5 rounded-2xl border border-border/80 animate-in fade-in duration-150 min-w-0"
                    >
                      <div className="flex items-center justify-between gap-2 min-w-0 border-b border-border/60 pb-2">
                        <div className="flex items-center gap-1.5">
                          <div className="h-2 w-2 rounded-full bg-primary shrink-0" />
                          <h4 className="font-bold text-xs sm:text-sm text-primary tracking-tight break-words min-w-0">
                            Detail Aspek: {detail.detail_name}
                          </h4>
                        </div>
                        <span className="text-[10px] sm:text-xs font-mono font-medium text-muted-foreground">
                          {getDetailProgress(detail).answered} / {getDetailProgress(detail).total} Uraian
                        </span>
                      </div>

                      {/* Empty state fallback when detail has no uraians */}
                      {(!detail.uraians || detail.uraians.length === 0) && (
                        <div className="p-4 text-center text-xs text-muted-foreground bg-muted/20 rounded-xl border border-border/60">
                          Belum ada Uraian Pengecekan di Detail Aspek ini.
                        </div>
                      )}

                      <div className="space-y-3 sm:space-y-4">
                        {detail.uraians?.map((uraian: any, uIndex: number) => {
                          const uId = uraian.uraian_id;
                          const pKey = getPhotoKey(detail.detail_id, uId);
                          const currentNilai = watch(`nilai_${pKey}`) ?? watch(`nilai_${uId}`);
                          const isNG = currentNilai === "0";
                          const formDisabled =
                            isCompleted || (!isLockedByMe && canEditInspection) || !isSamePlant;
                          const isHighlighted = highlightedUraianId === uId;

                          return (
                            <div
                              key={uId || uIndex}
                              id={`uraian-${uId}`}
                              onClick={() => updateUrlParams(currentAspekId, uId)}
                              className={cn(
                                "p-3 sm:p-4 rounded-xl border bg-card/70 space-y-2.5 sm:space-y-3 transition-all duration-500 cursor-pointer",
                                isHighlighted
                                  ? "border-primary ring-2 ring-primary/40 bg-primary/10 shadow-lg shadow-primary/10"
                                  : "border-border/60 hover:border-border"
                              )}
                            >
                              {/* Row: text + OK/NG — Touch-friendly min-h 44px buttons */}
                              <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-2.5">
                                <div className="space-y-0.5 flex-1 min-w-0">
                                  <div className="flex items-start gap-1.5">
                                    <span className="text-[10px] sm:text-[11px] font-bold text-muted-foreground font-mono shrink-0 mt-0.5">
                                      #{uIndex + 1}
                                    </span>
                                    <h5 className="font-semibold text-xs sm:text-sm text-foreground leading-snug">
                                      {uraian.uraian_text}
                                    </h5>
                                  </div>
                                  {uraian.standard_score !== undefined && (
                                    <span className="text-[10px] sm:text-[11px] text-muted-foreground block font-medium ml-4">
                                      Bobot: <strong className="text-foreground">{uraian.standard_score}</strong>
                                    </span>
                                  )}
                                </div>

                                {/* Rating Radio Options — Touch Target Minimum 44px height */}
                                <div className="grid grid-cols-2 sm:flex sm:items-center gap-1.5 bg-muted/40 p-1 rounded-xl border border-border/60 w-full sm:w-auto shrink-0 self-stretch sm:self-start">
                                  <label
                                    className={`flex items-center justify-center min-h-[44px] w-full sm:min-w-[52px] px-3 sm:px-4 py-2 rounded-lg text-xs font-bold cursor-pointer transition-all ${
                                      currentNilai === "2"
                                        ? "bg-green-600 text-white shadow-xs ring-1 ring-green-600/30"
                                        : "hover:bg-muted text-muted-foreground"
                                    } ${formDisabled ? "opacity-60 cursor-not-allowed" : "active:scale-95"}`}
                                  >
                                    <input
                                      type="radio"
                                      value="2"
                                      disabled={formDisabled}
                                      {...register(`nilai_${pKey}`)}
                                      className="sr-only"
                                    />
                                    <span>OK</span>
                                  </label>

                                  <label
                                    className={`flex items-center justify-center min-h-[44px] w-full sm:min-w-[52px] px-3 sm:px-4 py-2 rounded-lg text-xs font-bold cursor-pointer transition-all ${
                                      currentNilai === "0"
                                        ? "bg-red-600 text-white shadow-xs ring-1 ring-red-600/30"
                                        : "hover:bg-muted text-muted-foreground"
                                    } ${formDisabled ? "opacity-60 cursor-not-allowed" : "active:scale-95"}`}
                                  >
                                    <input
                                      type="radio"
                                      value="0"
                                      disabled={formDisabled}
                                      {...register(`nilai_${pKey}`)}
                                      className="sr-only"
                                    />
                                    <span>NG</span>
                                  </label>
                                </div>
                              </div>

                              {/* Photo Uploader for NG item */}
                              {isNG && (
                                <div className="pt-3 border-t border-red-500/20 bg-red-500/5 p-2.5 sm:p-3 rounded-xl space-y-2 min-w-0">
                                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-1">
                                    <span className="text-xs font-bold text-red-600 dark:text-red-400 flex items-center gap-1.5">
                                      Foto Bukti Temuan NG & Keterangan
                                    </span>
                                    <span className="text-[10px] text-muted-foreground">
                                      Wajib upload foto & Keterangan (Max 3 foto)
                                    </span>
                                  </div>

                                  <PhotoUploaderWithKeterangan
                                    photos={ngPhotosMap[pKey] || (!detail.detail_id ? (ngPhotosMap[uId] || []) : [])}
                                    onChange={(updatedPhotos) => handlePhotosChange(pKey, updatedPhotos)}
                                    // maxPhotos={3}
                                    disabled={formDisabled}
                                  />
                                </div>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  ))}

                  {/* Navigation Controls */}
                  <div className="grid grid-cols-2 gap-2 pt-3 border-t border-border w-full">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled={activeDetailIndex === 0 || activeDetailIndex === "all"}
                      onClick={() => {
                        if (typeof activeDetailIndex === "number" && activeDetailIndex > 0) {
                          setActiveDetailIndex(activeDetailIndex - 1);
                          scrollToTop();
                        }
                      }}
                      className="w-full text-xs h-10 sm:h-9"
                    >
                      <ChevronLeft className="w-3.5 h-3.5 mr-1" />
                      <span>Sebelumnya</span>
                    </Button>

                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled={
                        activeDetailIndex === "all" ||
                        typeof activeDetailIndex !== "number" ||
                        activeDetailIndex >= detailsList.length - 1
                      }
                      onClick={() => {
                        if (typeof activeDetailIndex === "number" && activeDetailIndex < detailsList.length - 1) {
                          setActiveDetailIndex(activeDetailIndex + 1);
                          scrollToTop();
                        }
                      }}
                      className="w-full text-xs h-10 sm:h-9"
                    >
                      <span>Berikutnya</span>
                      <ChevronRight className="w-3.5 h-3.5 ml-1" />
                    </Button>
                  </div>
                </div>
              )}
            </form>
          )}
        </Card>
      </div>

      {/* Floating Scroll-to-Top Button (positioned above mobile BottomNav) */}
      {showScrollTop && (
        <Button
          onClick={scrollToTop}
          size="icon"
          className="fixed bottom-20 sm:bottom-6 right-4 sm:right-6 rounded-full shadow-lg bg-primary hover:bg-primary/90 text-primary-foreground z-50 transition-all duration-300 h-11 w-11 flex items-center justify-center"
        >
          <ArrowUp className="h-5 w-5" />
        </Button>
      )}

      {/* Modal Dialog Confirm Batalkan Inspeksi */}
      {showCancelDialog && (
        <div className="fixed inset-0 z-[100] bg-black/80 backdrop-blur-md flex items-center justify-center p-2.5 sm:p-6 overflow-y-auto">
          <div className="bg-card text-card-foreground border border-border/80 rounded-2xl sm:rounded-3xl p-3.5 sm:p-7 w-full max-w-[480px] max-h-[calc(100dvh-1.25rem)] sm:max-h-none shadow-2xl space-y-4 sm:space-y-6 animate-in fade-in zoom-in-95 duration-200 overflow-y-auto box-border shrink-0">
            <div className="flex items-start gap-3 sm:gap-4">
              <div className="p-3 sm:p-3.5 rounded-xl sm:rounded-2xl bg-red-500/10 text-red-600 dark:text-red-400 shrink-0 border border-red-500/20">
                <AlertTriangle className="h-5 w-5 sm:h-6 sm:w-6" />
              </div>
              <div className="space-y-1 flex-1 min-w-0">
                <h3 className="text-base sm:text-xl font-bold text-foreground tracking-tight">Batalkan Inspeksi Ini?</h3>
                <p className="text-xs sm:text-sm text-muted-foreground leading-relaxed">
                  Tindakan ini akan <strong className="text-foreground font-semibold">menghapus sesi inspeksi</strong> dari database, menghapus seluruh draf di Redis, dan <strong className="text-foreground font-semibold">melepas kunci penguncian lokasi</strong> sehingga kawasan ini dapat diinspeksi kembali.
                </p>
              </div>
            </div>

            <div className="flex flex-col-reverse sm:flex-row items-stretch sm:items-center justify-end gap-2.5 pt-3 sm:pt-4 border-t border-border/60">
              <Button
                variant="outline"
                type="button"
                onClick={() => dispatch(setShowCancelDialog(false))}
                disabled={isCanceling}
                className="flex-1 sm:flex-none h-9 sm:h-10 px-3.5 sm:px-5 text-xs sm:text-sm font-semibold rounded-xl border-border/80 hover:bg-muted"
              >
                Kembali
              </Button>
              <Button
                variant="destructive"
                type="button"
                onClick={handleCancelInspection}
                disabled={isCanceling}
                className="flex-1 sm:flex-none h-9 sm:h-10 px-3.5 sm:px-5 text-xs sm:text-sm font-semibold rounded-xl bg-red-600 hover:bg-red-700 text-white shadow-xs shrink-0 whitespace-nowrap"
              >
                {isCanceling ? <Loader2 className="h-3.5 w-3.5 sm:h-4 sm:w-4 animate-spin mr-1.5" /> : null}
                {isCanceling ? "Membatalkan..." : "Ya, Batalkan & Hapus Sesi"}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}