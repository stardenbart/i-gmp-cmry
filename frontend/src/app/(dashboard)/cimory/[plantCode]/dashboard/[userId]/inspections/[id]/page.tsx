"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";
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
import { useAuditorGuard } from "@/lib/useAdminGuard";
import { useAuthStore } from "@/stores/authStore";
import {
  PhotoUploaderWithKeterangan,
  PhotoItem,
  dataURLtoFile,
} from "@/components/Inspection/PhotoUploaderWithKeterangan";
import { useInspectionDraft } from "@/hooks/useInspectionDraft";

export default function InspectionDetailPage() {
  const { isAuditor, isLoading: isGuardLoading } = useAuditorGuard();
  const user = useAuthStore((state) => state.user);
  const { id, userId, plantCode } = useParams() as { id: string; userId: string; plantCode?: string };
  const router = useRouter();
  const queryClient = useQueryClient();

  const [isSaving, setIsSaving] = useState(false);
  const [isCanceling, setIsCanceling] = useState(false);
  const [isReopening, setIsReopening] = useState(false);
  const [showCancelDialog, setShowCancelDialog] = useState(false);
  const [showScrollTop, setShowScrollTop] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      if (window.scrollY > 300) {
        setShowScrollTop(true);
      } else {
        setShowScrollTop(false);
      }
    };

    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const scrollToTop = () => {
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  // Per-uraian photo state for NG items: map of uraian_id => PhotoItem[]
  const [ngPhotosMap, setNgPhotosMap] = useState<Record<string, PhotoItem[]>>({});
  const ngPhotosMapRef = useRef<Record<string, PhotoItem[]>>({});

  // LocalStorage draft persistence hook
  const { draftData, saveDraft, clearDraft, hasDraft } = useInspectionDraft(id);

  const isFormInitialized = useRef(false);

  // Reset initialization flag when inspection ID changes
  useEffect(() => {
    isFormInitialized.current = false;
    ngPhotosMapRef.current = {};
  }, [id]);

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

  const { data: picData } = useQuery({
    queryKey: ["pic_mapping", inspection?.area_id, inspection?.kawasan_id],
    queryFn: () =>
      picApi.getAll({ area_id: inspection?.area_id, kawasan_id: inspection?.kawasan_id }),
    enabled: !!inspection?.area_id && !!inspection?.kawasan_id,
  });

  const { register, handleSubmit, watch, setValue, formState: { errors }, reset } = useForm();

  // Populate form state ONCE when checklist is available (server DB data OR localStorage draft recovery)
  useEffect(() => {
    if (!checklist?.aspeks || isFormInitialized.current) return;

    const defaultValues: Record<string, any> = {};
    const photoStateMap: Record<string, PhotoItem[]> = {};

    checklist.aspeks.forEach((aspek: any) => {
      aspek.details?.forEach((detail: any) => {
        detail.uraians?.forEach((uraian: any) => {
          const uId = uraian.uraian_id;
          
          // 1. First check if LocalStorage draft exists for this uraian
          const localDraft = draftData?.results?.[uId];
          if (localDraft) {
            if (localDraft.checking === "OK") defaultValues[`nilai_${uId}`] = "2";
            if (localDraft.checking === "NG") defaultValues[`nilai_${uId}`] = "0";
            defaultValues[`ket_${uId}`] = localDraft.keterangan || "";
            if (localDraft.photos && localDraft.photos.length > 0) {
              const validPhotos = localDraft.photos
                .filter((p) => p.previewUrl && !p.previewUrl.startsWith("blob:"))
                .map((p) => ({
                  id: p.id,
                  previewUrl: p.previewUrl || "",
                  keterangan: p.keterangan || "",
                }));
              if (validPhotos.length > 0) {
                photoStateMap[uId] = validPhotos;
              }
            }
          } 
          // 2. Otherwise use existing server result if available
          else if (uraian.result) {
            defaultValues[`nilai_${uId}`] =
              uraian.result.checking === "OK"
                ? "2"
                : uraian.result.checking === "NG"
                ? "0"
                : "";
            defaultValues[`ket_${uId}`] = uraian.result.keterangan || "";
          }
        });
      });
    });

    reset(defaultValues);
    if (Object.keys(photoStateMap).length > 0) {
      ngPhotosMapRef.current = photoStateMap;
      setNgPhotosMap(photoStateMap);
    }
    isFormInitialized.current = true;
  }, [checklist, draftData, reset]);

  // Auto-save form changes to LocalStorage for recovery
  const handleFormValueChange = useCallback(
    (formValues: any) => {
      if (!isFormInitialized.current || !checklist?.aspeks || inspection?.status === "Completed") return;

      const draftResults: Record<string, any> = {};
      checklist.aspeks.forEach((aspek: any) => {
        aspek.details?.forEach((detail: any) => {
          detail.uraians?.forEach((uraian: any) => {
            const uId = uraian.uraian_id;
            const val = formValues[`nilai_${uId}`];
            const ket = formValues[`ket_${uId}`];
            const photos = ngPhotosMapRef.current[uId] || [];

            if (val) {
              draftResults[uId] = {
                checking: val === "2" ? "OK" : "NG",
                nilai: parseInt(val, 10),
                keterangan: ket || "",
                photos: photos.map((p) => ({
                  id: p.id,
                  previewUrl: p.previewUrl,
                  keterangan: p.keterangan,
                })),
              };
            }
          });
        });
      });

      saveDraft(draftResults);
    },
    [checklist, inspection?.status, saveDraft]
  );

  // Subscribe to form input changes for auto-save (prevents infinite re-renders)
  useEffect(() => {
    if (!isFormInitialized.current) return;
    const subscription = watch((values) => {
      handleFormValueChange(values);
    });
    return () => subscription.unsubscribe();
  }, [watch, handleFormValueChange]);

  // Handler for photo updates per NG item
  const handlePhotosChange = (uraianId: string, photos: PhotoItem[]) => {
    ngPhotosMapRef.current = {
      ...ngPhotosMapRef.current,
      [uraianId]: photos,
    };
    setNgPhotosMap({ ...ngPhotosMapRef.current });
    const currentFormValues = watch();
    handleFormValueChange(currentFormValues);
  };

  // Submit Final ("Selesaikan Audit")
  const onFinalSubmit = async (formData: any) => {
    if (!checklist?.aspeks || !inspection) return;

    // Validate that all uraians are answered
    let unansweredCount = 0;
    let missingPhotoInfoCount = 0;

    const resultsPayload: any[] = [];
    const ngUraianTasks: { uraian: any; keterangan: string; photos: PhotoItem[] }[] = [];

    checklist.aspeks.forEach((aspek: any) => {
      aspek.details?.forEach((detail: any) => {
        detail.uraians?.forEach((uraian: any) => {
          const uId = uraian.uraian_id;
          const val = formData[`nilai_${uId}`];
          const ket = formData[`ket_${uId}`];
          const photos = ngPhotosMap[uId] || [];

          if (!val) {
            unansweredCount++;
          } else {
            const checking = val === "2" ? "OK" : "NG";
            resultsPayload.push({
              uraian_id: uId,
              checking,
              nilai: parseInt(val, 10),
              keterangan: ket || "",
            });

            if (checking === "NG") {
              // Validate photos and per-photo keterangan
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
      // 1. Bulk Save Results to DB
      await inspectionApi.bulkSaveResults(id, resultsPayload);

      // 2. Process NG Issues & Upload Photos with Per-Photo Keterangan
      const picUserId = picData?.items?.[0]?.user_id || user?.id || "";

      if (ngUraianTasks.length > 0) {
        // Re-fetch checklist to obtain generated Result IDs
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
            // Create Issue
            const issueRes = await issueApi.create({
              result_id: savedResultId,
              issue_pic_user_id: picUserId,
              keterangan: task.keterangan || "Temuan NG pada inspeksi",
            });

            const createdIssueId = issueRes?.data?.issue_id;

            // Upload each photo with its dedicated keterangan
            if (createdIssueId && task.photos.length > 0) {
              for (const photoItem of task.photos) {
                let photoFile = photoItem.file;
                if (!photoFile && photoItem.previewUrl?.startsWith("data:")) {
                  photoFile = dataURLtoFile(photoItem.previewUrl, `photo_${Date.now()}.jpg`);
                }
                if (photoFile && photoFile.size > 0) {
                  const fd = new FormData();
                  fd.append("photo", photoFile);
                  fd.append("photo_type", "Initial");
                  fd.append("keterangan", photoItem.keterangan || "");
                  await issueApi.uploadPhoto(createdIssueId, fd);
                }
              }
            }
          }
        }
      }

      // 3. Update Inspection Status to Completed
      await inspectionApi.updateStatus(id, "Completed");

      // 4. Clear LocalStorage draft
      clearDraft();

      toast.success("Inspeksi berhasil diselesaikan!");
      await queryClient.invalidateQueries({ queryKey: ["inspection", id] });
      await queryClient.invalidateQueries({ queryKey: ["inspection_checklist", id] });
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
      await inspectionApi.delete(id);
      clearDraft();
      toast.success("Inspeksi telah dibatalkan dan kunci lokasi dilepas.");
      router.push(`/cimory/${plantCode || "all"}/dashboard/${userId}/inspections`);
    } catch (err: any) {
      toast.error(err.response?.data?.message || "Gagal membatalkan inspeksi");
    } finally {
      setIsCanceling(false);
      setShowCancelDialog(false);
    }
  };

  // Handler for "Edit Inspeksi" (Re-open Completed inspection to Ongoing)
  const handleReopenForEdit = async () => {
    try {
      setIsReopening(true);
      await inspectionApi.updateStatus(id, "Ongoing");
      // Reset form init flag so useEffect repopulates from server data (not stale draft)
      isFormInitialized.current = false;
      ngPhotosMapRef.current = {};
      clearDraft(); // clear any stale localStorage draft
      toast.success("Inspeksi dibuka kembali untuk diedit.");
      await queryClient.invalidateQueries({ queryKey: ["inspection", id] });
      await queryClient.invalidateQueries({ queryKey: ["inspection_checklist", id] });
    } catch (err: any) {
      toast.error(err.response?.data?.message || "Gagal mengedit inspeksi");
    } finally {
      setIsReopening(false);
    }
  };

  // Excel Export Handler
  const handleExport = async () => {
    try {
      const backendUrl = process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080";
      const token = localStorage.getItem("auth-storage")
        ? JSON.parse(localStorage.getItem("auth-storage")!).state.token
        : "";

      const res = await fetch(`${backendUrl}/api/v1/inspections/${id}/export`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!res.ok) throw new Error("Export failed");

      const blob = await res.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `report_inspeksi_${id}.xlsx`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);

      toast.success("Berhasil mengekspor laporan Excel");
    } catch (err) {
      toast.error("Gagal mengekspor laporan Excel");
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

  if (!inspection) return <div className="p-8 text-center text-muted-foreground">Data inspeksi tidak ditemukan</div>;

  const currentUserId = user?.id;
  const isInspectorOwner = !!currentUserId && (inspection.inspector_id === currentUserId);
  const isCompleted = inspection.status === "Completed" || inspection.status === "Approved";
  const isOngoing = inspection.status === "Ongoing";
  const isReadOnly = isCompleted || !isInspectorOwner;

  const statusClass = isCompleted
    ? "bg-green-500/10 text-green-600 border border-green-500/20"
    : isOngoing
    ? "bg-amber-500/10 text-amber-600 border border-amber-500/20 animate-pulse"
    : "bg-zinc-500/10 text-zinc-500 border border-zinc-500/20";

  // Check if checklist has any saved results (meaning it was submitted/completed before)
  const hasExistingResults = checklist?.aspeks?.some((a: any) =>
    a.details?.some((d: any) =>
      d.uraians?.some((u: any) => !!u.result)
    )
  );

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-12">
      {/* Header Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <Link href={`/cimory/${plantCode || "all"}/dashboard/${userId}/inspections`}>
            <Button variant="ghost" size="icon" className="rounded-full">
              <ArrowLeft className="h-5 w-5" />
            </Button>
          </Link>
          <div>
            <div className="flex items-center gap-3">
              <h2 className="text-2xl font-bold tracking-tight">Detail Inspeksi</h2>
              <span className={`text-xs font-semibold px-3 py-1 rounded-full ${statusClass}`}>
                {isOngoing ? "● Ongoing (Sedang Berlangsung)" : inspection.status}
              </span>
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              ID: <span className="font-mono">{inspection.inspection_id}</span> · Dibuat oleh{" "}
              <strong className="text-foreground">{inspection.inspector_name || inspection.inspector_id}</strong>
            </p>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center gap-2">
          {/* If Completed, show "Edit Inspeksi" button */}
          {isCompleted && isInspectorOwner && (
            <Button
              onClick={handleReopenForEdit}
              disabled={isReopening}
              variant="outline"
              className="border-primary text-primary hover:bg-primary/5"
            >
              {isReopening ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Edit className="mr-2 h-4 w-4" />
              )}
              Edit Inspeksi
            </Button>
          )}

          {/* If Ongoing and is owner */}
          {isOngoing && isInspectorOwner && (
            <>
              {/* Show "Batal Edit" if it has existing results, otherwise show "Batalkan Inspeksi" (delete) */}
              {hasExistingResults ? (
                <Button
                  variant="outline"
                  onClick={async () => {
                    try {
                      setIsSaving(true);
                      await inspectionApi.updateStatus(id, "Completed");
                      clearDraft();
                      toast.info("Perubahan dibatalkan. Status dikembalikan ke Selesai.");
                      await queryClient.invalidateQueries({ queryKey: ["inspection", id] });
                    } catch (err: any) {
                      toast.error("Gagal membatalkan edit");
                    } finally {
                      setIsSaving(false);
                    }
                  }}
                  disabled={isSaving}
                  className="border-muted-foreground/40 text-muted-foreground hover:bg-muted"
                >
                  Batal Edit
                </Button>
              ) : (
                <Button
                  variant="destructive"
                  onClick={() => setShowCancelDialog(true)}
                  disabled={isSaving || isCanceling}
                  className="bg-red-600 hover:bg-red-700 text-white font-medium"
                >
                  <Trash2 className="mr-2 h-4 w-4" /> Batalkan Inspeksi
                </Button>
              )}

              <Button
                onClick={handleSubmit(onFinalSubmit)}
                disabled={isSaving || isCanceling}
                className="bg-green-600 hover:bg-green-700 text-white font-semibold shadow-md px-5"
              >
                {isSaving ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" /> Menyimpan...
                  </>
                ) : (
                  <>
                    <CheckCircle2 className="mr-2 h-4 w-4" /> Selesaikan Audit
                  </>
                )}
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Ongoing Lock Status Notice Banner */}
      {isOngoing && isInspectorOwner && (
        <div className="rounded-2xl border border-amber-500/30 bg-amber-500/10 p-4 flex items-start gap-3 shadow-xs">
          <div className="rounded-full bg-amber-500/20 p-2 text-amber-600 shrink-0">
            <Play className="h-5 w-5 fill-current" />
          </div>
          <div className="space-y-1 text-xs">
            <div className="font-bold text-amber-700 dark:text-amber-400 text-sm flex items-center gap-2">
              <span>Inspeksi Sedang Berlangsung & Lokasi Terkunci</span>
              {hasDraft && (
                <span className="rounded bg-amber-500/20 px-2 py-0.5 text-label-sm text-amber-700 dark:text-amber-300 font-mono">
                  [Draft lokal dipulihkan]
                </span>
              )}
            </div>
            <p className="text-muted-foreground leading-relaxed">
              Lokasi <strong className="text-foreground">{inspection.kawasan_name}</strong> -{" "}
              <strong className="text-foreground">{inspection.detail_kawasan_name}</strong> sedang dikunci untuk Anda. 
              Isian Anda otomatis tersimpan di memori browser. Klik tombol <strong>"Selesaikan Audit"</strong> untuk menyimpan hasil secara permanen ke database.
            </p>
          </div>
        </div>
      )}

      {/* Read-Only Notice Banner if opened by another user */}
      {isOngoing && !isInspectorOwner && (
        <div className="rounded-2xl border border-blue-500/30 bg-blue-500/10 p-4 flex items-start gap-3 shadow-xs">
          <div className="rounded-full bg-blue-500/20 p-2 text-blue-600 shrink-0">
            <AlertTriangle className="h-5 w-5" />
          </div>
          <div className="space-y-1 text-xs">
            <div className="font-bold text-blue-700 dark:text-blue-400 text-sm flex items-center gap-2">
              <span>Inspeksi Sedang Dikerjakan oleh {inspection.inspector_name || "Auditor Lain"} (Mode Read-Only)</span>
            </div>
            <p className="text-muted-foreground leading-relaxed">
              Lokasi <strong className="text-foreground">{inspection.kawasan_name}</strong> -{" "}
              <strong className="text-foreground">{inspection.detail_kawasan_name}</strong> saat ini sedang diinspeksi oleh{" "}
              <strong className="text-foreground">{inspection.inspector_name || "Auditor Lain"}</strong>. Anda hanya memiliki akses lihat data dalam mode Read-Only.
            </p>
          </div>
        </div>
      )}

      {/* Grid Content */}
      <div className="grid gap-6 md:grid-cols-3">
        {/* Info Card */}
        <Card className="p-6 bg-card/60 backdrop-blur-md md:col-span-1 h-fit sticky top-24 shadow-sm border-border/80">
          <h3 className="font-bold text-base mb-4 border-b border-border pb-2 flex items-center justify-between">
            <span>Informasi Area</span>
            <span className="text-xs text-muted-foreground font-normal">Audit Session</span>
          </h3>
          <div className="space-y-3.5 text-sm">
            <div>
              <span className="text-muted-foreground block text-xs font-medium">Area</span>
              <span className="font-semibold text-foreground">{inspection.area_name || inspection.area_id}</span>
            </div>
            <div>
              <span className="text-muted-foreground block text-xs font-medium">Kawasan</span>
              <span className="font-semibold text-foreground">{inspection.kawasan_name || inspection.kawasan_id}</span>
            </div>
            <div>
              <span className="text-muted-foreground block text-xs font-medium">Detail Kawasan</span>
              <span className="font-semibold text-foreground">{inspection.detail_kawasan_name || inspection.detail_kawasan_id}</span>
            </div>
            <div>
              <span className="text-muted-foreground block text-xs font-medium">Auditor Pelaksana</span>
              <span className="font-semibold text-foreground">{inspection.inspector_name || inspection.inspector_id}</span>
            </div>
            {inspection.score !== undefined && inspection.score !== null && (
              <div className="pt-2 border-t border-border">
                <span className="text-muted-foreground block text-xs font-medium">Skor Hasil Inspeksi</span>
                <span className={`text-lg font-bold ${inspection.score >= 80 ? "text-green-600" : "text-amber-600"}`}>
                  {inspection.score.toFixed(1)}%
                </span>
              </div>
            )}
          </div>
        </Card>

        {/* Checklist Form */}
        <Card className="p-6 bg-card/60 backdrop-blur-md md:col-span-2 shadow-sm border-border/80">
          <h3 className="font-bold text-base mb-4 border-b border-border pb-2 flex items-center justify-between">
            <span>Checklist Penilaian Audit</span>
            <span className="text-xs text-muted-foreground font-normal">
              {checklist?.aspeks?.reduce(
                (acc: number, a: any) =>
                  acc + a.details?.reduce((dAcc: number, d: any) => dAcc + (d.uraians?.length || 0), 0),
                0
              ) || 0}{" "}
              Item Pengecekan
            </span>
          </h3>

          <form className="space-y-8">
            {checklist?.aspeks?.length === 0 && (
              <div className="p-8 text-center text-muted-foreground bg-muted/20 rounded-2xl border border-border">
                Belum ada Uraian (Checklist) yang diatur untuk Area ini.
              </div>
            )}

            {checklist?.aspeks?.map((aspek: any, aIndex: number) => (
              <div key={aspek.aspek_id} className="space-y-4">
                <div className="flex items-center gap-3">
                  <div className="h-8 w-8 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-sm">
                    {aIndex + 1}
                  </div>
                  <h3 className="text-lg font-bold tracking-tight text-foreground">{aspek.aspek_name}</h3>
                </div>

                {aspek.details?.map((detail: any, dIndex: number) => (
                  <div key={detail.detail_id} className="ml-4 pl-4 border-l-2 border-primary/20 space-y-4">
                    <h4 className="font-semibold text-base text-primary">{detail.detail_name}</h4>

                    <div className="space-y-4 mt-4">
                      {detail.uraians?.map((uraian: any, uIndex: number) => {
                        const uId = uraian.uraian_id;
                        const isNG = watch(`nilai_${uId}`) === "0";

                        return (
                          <div
                            key={uId}
                            className={`p-4 rounded-2xl border transition-all ${
                              isNG
                                ? "border-destructive/30 bg-destructive/5"
                                : watch(`nilai_${uId}`) === "2"
                                ? "border-green-500/30 bg-green-500/5"
                                : "border-border bg-card"
                            }`}
                          >
                            <div className="flex flex-col sm:flex-row justify-between gap-4 mb-3">
                              <div>
                                <h5 className="font-semibold text-sm">
                                  {aIndex + 1}.{dIndex + 1}.{uIndex + 1} Uraian Pengecekan
                                </h5>
                                <p className="text-sm text-foreground/80 mt-1 leading-relaxed">
                                  {uraian.uraian_text}
                                </p>
                              </div>
                            </div>

                            {/* Radio Options: 2 (Aman / OK) vs 0 (Ada Issue / NG) */}
                            <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
                              <div className="sm:col-span-2">
                                <label className="block text-xs font-semibold text-muted-foreground mb-2">
                                  Penilaian <span className="text-destructive">*</span>
                                </label>
                                <div className="flex items-center gap-3">
                                  <label
                                    className={`flex items-center gap-2 text-sm px-3.5 py-2.5 rounded-xl border transition-all w-full ${
                                      isReadOnly
                                        ? "opacity-60 cursor-not-allowed bg-muted/20"
                                        : "cursor-pointer hover:border-green-500/50 hover:bg-green-500/5"
                                    } ${
                                      watch(`nilai_${uId}`) === "2"
                                        ? "border-green-500 bg-green-500/10 font-bold"
                                        : "border-border"
                                    }`}
                                  >
                                    <input
                                      type="radio"
                                      value="2"
                                      disabled={isReadOnly}
                                      {...register(`nilai_${uId}`, { required: true })}
                                      className="accent-green-600 h-4 w-4 cursor-pointer"
                                    />
                                    <span className="font-semibold text-green-600">2 - Aman (OK)</span>
                                  </label>

                                  <label
                                    className={`flex items-center gap-2 text-sm px-3.5 py-2.5 rounded-xl border transition-all w-full ${
                                      isReadOnly
                                        ? "opacity-60 cursor-not-allowed bg-muted/20"
                                        : "cursor-pointer hover:border-red-500/50 hover:bg-red-500/5"
                                    } ${
                                      watch(`nilai_${uId}`) === "0"
                                        ? "border-red-500 bg-red-500/10 font-bold"
                                        : "border-border"
                                    }`}
                                  >
                                    <input
                                      type="radio"
                                      value="0"
                                      disabled={isReadOnly}
                                      {...register(`nilai_${uId}`, { required: true })}
                                      className="accent-red-600 h-4 w-4 cursor-pointer"
                                    />
                                    <span className="font-semibold text-red-600">0 - Ada Issue (NG)</span>
                                  </label>
                                </div>
                              </div>
                            </div>

                            {/* Multi-Photo Uploader Component for NG Items */}
                            {isNG && (
                              <PhotoUploaderWithKeterangan
                                photos={ngPhotosMap[uId] || []}
                                onChange={(photos) => handlePhotosChange(uId, photos)}
                                disabled={isReadOnly}
                              />
                            )}
                          </div>
                        );
                      })}
                    </div>
                  </div>
                ))}
              </div>
            ))}
          </form>
        </Card>
      </div>

      {/* Confirmation Dialog for Batalkan Inspeksi */}
      {showCancelDialog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-md p-4 animate-in fade-in duration-200">
          <div className="w-[90vw] sm:w-112.5 shrink-0 rounded-2xl border border-destructive/30 bg-background p-6 shadow-2xl space-y-6 text-foreground">
            <div className="flex items-start gap-4">
              <div className="rounded-full bg-destructive/10 p-3 text-destructive shrink-0 mt-0.5">
                <AlertTriangle className="h-6 w-6" />
              </div>
              <div className="space-y-1.5 flex-1 min-w-0">
                <h3 className="text-lg font-bold tracking-tight text-foreground">
                  Batalkan Inspeksi Ini?
                </h3>
                <p className="text-xs text-muted-foreground leading-relaxed">
                  Tindakan ini akan <strong className="text-foreground">menghapus sesi inspeksi</strong> dari database, menghapus seluruh draf lokal di browser Anda, dan <strong className="text-foreground">melepas kunci penguncian lokasi</strong> sehingga kawasan ini dapat diinspeksi kembali.
                </p>
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 pt-3 border-t border-border/60">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setShowCancelDialog(false)}
                disabled={isCanceling}
                className="px-4 text-xs font-semibold h-10 rounded-xl"
              >
                Batal
              </Button>
              <Button
                type="button"
                variant="destructive"
                size="sm"
                onClick={handleCancelInspection}
                disabled={isCanceling}
                className="px-4 bg-destructive hover:bg-destructive/90 text-white text-xs font-semibold h-10 rounded-xl gap-2"
              >
                {isCanceling ? (
                  <Loader2 className="h-4 w-4 animate-spin shrink-0" />
                ) : (
                  <Trash2 className="h-4 w-4 shrink-0" />
                )}
                Ya, Batalkan
              </Button>
            </div>
          </div>
        </div>
      )}
      {/* Floating Scroll to Top Button */}
      {showScrollTop && (
        <button
          type="button"
          onClick={scrollToTop}
          className="fixed bottom-28 right-6 z-50 flex h-11 w-11 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-lg border border-primary/20 hover:bg-primary/90 hover:scale-105 active:scale-95 transition-all duration-200"
          title="Ke Atas"
          aria-label="Scroll ke atas"
        >
          <ArrowUp className="h-5 w-5" />
        </button>
      )}
    </div>
  );
}
