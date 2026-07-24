"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { ArrowLeft, Save, CheckCircle2, FileDown } from "lucide-react";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { inspectionApi } from "@/lib/api/inspection.api";
import { issueApi } from "@/lib/api/issue.api";
import { picApi } from "@/lib/api/pic.api";
import { useAuditorGuard } from "@/lib/useAdminGuard";
import { useAuthStore } from "@/stores/authStore";
import Image from "next/image";
import { useEffect, useState } from "react";

function ImagePreview({ file }: { file: File }) {
  const [preview, setPreview] = useState<string>("");

  useEffect(() => {
    if (!file) return;
    if (typeof file === "string") {
      setPreview(file);
      return;
    }
    const objectUrl = URL.createObjectURL(file);
    setPreview(objectUrl);
    return () => URL.revokeObjectURL(objectUrl);
  }, [file]);

  if (!preview) return null;
  return (
    <div className="mb-3 relative h-32 w-48 rounded-xl overflow-hidden border border-border shadow-sm">
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img src={preview} alt="Preview" className="w-full h-full object-cover" />
    </div>
  );
}

export default function InspectionDetailPage() {
  const { isAuditor, isLoading: isGuardLoading } = useAuditorGuard();
  const user = useAuthStore(state => state.user);
  const { id } = useParams() as { id: string };
  const queryClient = useQueryClient();

  const [isSaving, setIsSaving] = useState(false);

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
    queryFn: () => picApi.getAll({ area_id: inspection?.area_id, kawasan_id: inspection?.kawasan_id }),
    enabled: !!inspection?.area_id && !!inspection?.kawasan_id,
  });

  const { register, handleSubmit, watch, formState: { errors }, reset } = useForm();

  // Pre-fill form if checklist has results
  useEffect(() => {
    if (checklist?.aspeks) {
      const defaultValues: Record<string, any> = {};
      checklist.aspeks.forEach((aspek: any) => {
        aspek.details.forEach((detail: any) => {
          detail.uraians.forEach((uraian: any) => {
            if (uraian.result) {
              defaultValues[`nilai_${uraian.uraian_id}`] = uraian.result.checking === "OK" ? "2" : (uraian.result.checking === "NG" ? "0" : "");
              defaultValues[`ket_${uraian.uraian_id}`] = uraian.result.keterangan || "";
            }
          });
        });
      });
      reset(defaultValues);
    }
  }, [checklist, reset]);

  const processSubmit = async (formData: any, finalize: boolean) => {
    if (!checklist?.aspeks || !inspection) return;
    setIsSaving(true);
    
    // 1. Build Bulk Save Payload
    const resultsPayload: any[] = [];
    const ngUraians: any[] = []; // Collect NG items for issue creation

    checklist.aspeks.forEach((aspek: any) => {
      aspek.details.forEach((detail: any) => {
        detail.uraians.forEach((uraian: any) => {
          const val = formData[`nilai_${uraian.uraian_id}`];
          const ket = formData[`ket_${uraian.uraian_id}`];
          const photo = formData[`photo_${uraian.uraian_id}`];

          if (val) {
            const checking = val === "2" ? "OK" : "NG";
            resultsPayload.push({
              uraian_id: uraian.uraian_id,
              checking: checking,
              nilai: parseInt(val, 10),
              keterangan: ket || "",
            });

            if (checking === "NG" && !uraian.result) {
              // Only create issue if it wasn't already created (we assume no result means not created yet)
              ngUraians.push({
                uraian,
                keterangan: ket || "",
                photo: photo?.[0]
              });
            }
          }
        });
      });
    });

    try {
      if (resultsPayload.length > 0) {
        // Bulk save
        const bulkRes = await inspectionApi.bulkSaveResults(id, resultsPayload);

        // Find the PIC for issues
        const picUserId = picData?.items?.[0]?.user_id || "";

        // Wait, BulkSave doesn't return the new result IDs easily because they are generated.
        // Actually, if we just want to create an Issue, we need a result_id. 
        // Let's assume the backend will return the saved results or we just fetch again.
        // For now, if we create an issue, we can pass a dummy result_id if the API allows, or we need the real one.
        // Let's refetch to get real result IDs if we have NGs, or wait for the backend to return them.
        
        // As a workaround, we can fetch the checklist again to get the Result IDs, then create issues.
        if (ngUraians.length > 0 && picUserId) {
           const updatedChecklistRes = await inspectionApi.getChecklist(id);
           const updatedChecklist = updatedChecklistRes?.data;
           
           for (const ng of ngUraians) {
              let savedResultId = "";
              // Find result id
              updatedChecklist?.aspeks?.forEach((a: any) => {
                a.details?.forEach((d: any) => {
                  d.uraians?.forEach((u: any) => {
                    if (u.uraian_id === ng.uraian.uraian_id && u.result) {
                      savedResultId = u.result.result_id;
                    }
                  })
                })
              });

              if (savedResultId) {
                // Create Issue
                const issueRes = await issueApi.create({
                  result_id: savedResultId,
                  issue_pic_user_id: picUserId,
                  keterangan: ng.keterangan
                });
                
                // Upload Photo
                if (ng.photo && issueRes?.data?.issue_id) {
                  const fd = new FormData();
                  fd.append("photo", ng.photo);
                  fd.append("photo_type", "Initial");
                  await issueApi.uploadPhoto(issueRes.data.issue_id, fd);
                }
              }
           }
        }
      }

      if (finalize) {
        await inspectionApi.updateStatus(id, "Completed");
        toast.success("Inspeksi berhasil diselesaikan!");
        window.location.reload();
      } else {
        await queryClient.invalidateQueries({ queryKey: ["inspection", id] });
        await queryClient.invalidateQueries({ queryKey: ["inspection_checklist", id] });
        toast.success("Draft inspeksi berhasil disimpan!");
      }
    } catch (err: any) {
      toast.error(err.response?.data?.message || "Terjadi kesalahan saat menyimpan");
    } finally {
      setIsSaving(false);
    }
  };

  const onSubmitChecklist = async (formData: any) => {
    await processSubmit(formData, false);
  };

  const handleSubmitFinal = handleSubmit(async (formData: any) => {
    await processSubmit(formData, true);
  });

  const handleExport = async () => {
    try {
      const res = await fetch(`http://localhost:8080/api/v1/inspections/${id}/export`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("auth-storage") ? JSON.parse(localStorage.getItem("auth-storage")!).state.token : ""}`
        }
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
      
      toast.success("Berhasil mengekspor data");
    } catch (err) {
      toast.error("Gagal mengekspor laporan excel");
    }
  };

  // Show spinner while: guard is loading, user store not yet hydrated, or data is loading
  if (isGuardLoading || !user || isInspectionLoading || isChecklistLoading) {
    return (
      <div className="flex justify-center p-8">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!isAuditor) return (
    <div className="flex justify-center items-center p-12 text-muted-foreground">
      Anda tidak memiliki akses ke halaman ini.
    </div>
  );
  if (!inspection) return <div>Data tidak ditemukan</div>;

  const statusClass =
    inspection.status === "Completed"
      ? "bg-green-500/10 text-green-500"
      : inspection.status === "Ongoing"
      ? "bg-blue-500/10 text-blue-500"
      : "bg-zinc-500/10 text-zinc-500";

  const isCompleted = inspection.status === "Completed";

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <Link href="../">
            <Button variant="ghost" size="icon" className="rounded-full">
              <ArrowLeft className="h-5 w-5" />
            </Button>
          </Link>
          <div>
            <div className="flex items-center gap-3">
              <h2 className="text-2xl font-bold tracking-tight">Detail Inspeksi</h2>
              <span className={`text-xs font-medium px-2.5 py-1 rounded-full ${statusClass}`}>
                {inspection.status}
              </span>
            </div>
            <p className="text-muted-foreground text-sm font-mono mt-1">
              ID: {inspection.inspection_id}
            </p>
          </div>
        </div>

        <div className="flex gap-2">
          <Button onClick={handleExport} variant="outline" className="border-green-600 text-green-600 hover:bg-green-50">
            <FileDown className="mr-2 h-4 w-4" /> Export Excel
          </Button>

          {inspection.status !== "Completed" && (
            <Button 
              onClick={handleSubmitFinal} 
              disabled={isSaving}
              className="bg-green-600 hover:bg-green-700 text-white"
            >
              <CheckCircle2 className="mr-2 h-4 w-4" /> 
              {isSaving ? "Menyimpan..." : "Selesaikan Audit"}
            </Button>
          )}
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        <Card className="p-6 bg-card/60 backdrop-blur-md md:col-span-1 h-fit sticky top-24">
          <h3 className="font-semibold mb-4 border-b border-border pb-2">Informasi Area</h3>
          <div className="space-y-3 text-sm">
            <div>
              <span className="text-muted-foreground block text-xs">Area</span>
              <span className="font-medium">{inspection.area_name || inspection.area_id}</span>
            </div>
            <div>
              <span className="text-muted-foreground block text-xs">Kawasan</span>
              <span className="font-medium">{inspection.kawasan_name || inspection.kawasan_id}</span>
            </div>
            <div>
              <span className="text-muted-foreground block text-xs">Detail Kawasan</span>
              <span className="font-medium">{inspection.detail_kawasan_name || inspection.detail_kawasan_id}</span>
            </div>
            <div>
              <span className="text-muted-foreground block text-xs">Auditor</span>
              <span className="font-medium">{inspection.inspector_name || inspection.inspector_id}</span>
            </div>
          </div>
        </Card>

        <Card className="p-6 bg-card/60 backdrop-blur-md md:col-span-2">
          <h3 className="font-semibold mb-4 border-b border-border pb-2">Checklist Audit</h3>

          <form onSubmit={handleSubmit(onSubmitChecklist)} className="space-y-8">
            
            {checklist?.aspeks?.length === 0 && (
              <div className="p-8 text-center text-muted-foreground bg-muted/20 rounded-2xl border border-border">
                Belum ada Uraian (Checklist) yang diatur untuk Area ini.
              </div>
            )}

            {checklist?.aspeks?.map((aspek: any, aIndex: number) => (
              <div key={aspek.aspek_id} className="space-y-4">
                <div className="flex items-center gap-3">
                  <div className="h-8 w-8 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold">
                    {aIndex + 1}
                  </div>
                  <h3 className="text-xl font-bold tracking-tight">{aspek.aspek_name}</h3>
                </div>

                {aspek.details?.map((detail: any, dIndex: number) => (
                  <div key={detail.detail_id} className="ml-4 pl-4 border-l-2 border-border space-y-4">
                    <h4 className="font-semibold text-lg text-primary">{detail.detail_name}</h4>
                    
                    <div className="space-y-4 mt-4">
                      {detail.uraians?.map((uraian: any, uIndex: number) => {
                        const isNG = watch(`nilai_${uraian.uraian_id}`) === "0";
                        const hasExistingResult = !!uraian.result;
                        
                        return (
                          <div key={uraian.uraian_id} className="p-4 rounded-2xl border border-border bg-background/50 shadow-sm transition-all hover:shadow-md">
                            <div className="flex flex-col sm:flex-row justify-between gap-4 mb-4">
                              <div>
                                <h5 className="font-medium">
                                  {aIndex + 1}.{dIndex + 1}.{uIndex + 1} Uraian Pengecekan
                                </h5>
                                <p className="text-sm text-foreground/80 mt-1">
                                  {uraian.uraian_text}
                                </p>
                              </div>
                            </div>

                            <div className="grid grid-cols-1 sm:grid-cols-4 gap-6">
                              <div className="sm:col-span-2">
                                <label className="block text-xs font-medium text-muted-foreground mb-2">Uraian Penilaian (Nilai)</label>
                                <div className="flex items-center gap-3">
                                  <label className={`flex items-center gap-2 text-sm px-3 py-2 rounded-lg border border-border transition-all w-full ${isCompleted || hasExistingResult ? 'opacity-50 bg-muted/20' : 'cursor-pointer hover:border-green-500/50 hover:bg-green-500/5'}`}>
                                    {hasExistingResult || isCompleted ? (
                                      <input 
                                        key={`ro_ok_${uraian.uraian_id}`}
                                        type="radio" 
                                        name={`nilai_${uraian.uraian_id}`}
                                        value="2" 
                                        checked={uraian.result?.checking === "OK"} 
                                        readOnly
                                        disabled
                                        className="accent-green-600 h-4 w-4 disabled:cursor-not-allowed" 
                                      />
                                    ) : (
                                      <input 
                                        key={`act_ok_${uraian.uraian_id}`}
                                        type="radio" 
                                        value="2" 
                                        {...register(`nilai_${uraian.uraian_id}`, { required: true })}
                                        className="accent-green-600 h-4 w-4 cursor-pointer" 
                                      />
                                    )}
                                    <span className="font-medium text-green-600">2 - Aman</span>
                                  </label>
                                  <label className={`flex items-center gap-2 text-sm px-3 py-2 rounded-lg border border-border transition-all w-full ${isCompleted || hasExistingResult ? 'opacity-50 bg-muted/20' : 'cursor-pointer hover:border-red-500/50 hover:bg-red-500/5'}`}>
                                    {hasExistingResult || isCompleted ? (
                                      <input 
                                        key={`ro_ng_${uraian.uraian_id}`}
                                        type="radio" 
                                        name={`nilai_${uraian.uraian_id}`}
                                        value="0" 
                                        checked={uraian.result?.checking === "NG"} 
                                        readOnly
                                        disabled
                                        className="accent-red-600 h-4 w-4 disabled:cursor-not-allowed" 
                                      />
                                    ) : (
                                      <input 
                                        key={`act_ng_${uraian.uraian_id}`}
                                        type="radio" 
                                        value="0" 
                                        {...register(`nilai_${uraian.uraian_id}`, { required: true })}
                                        className="accent-red-600 h-4 w-4 cursor-pointer" 
                                      />
                                    )}
                                    <span className="font-medium text-red-600">0 - Ada Issue</span>
                                  </label>
                                </div>
                              </div>
                              <div className="sm:col-span-2">
                                <label className="block text-xs font-medium text-muted-foreground mb-2">Keterangan / Temuan</label>
                                {hasExistingResult || isCompleted ? (
                                  <Input 
                                    key={`ro_ket_${uraian.uraian_id}`}
                                    value={uraian.result?.keterangan || ""} 
                                    readOnly
                                    disabled
                                    className="h-10 rounded-xl disabled:opacity-50 disabled:cursor-not-allowed" 
                                    placeholder="Tulis catatan detail jika ada issue..." 
                                  />
                                ) : (
                                  <Input 
                                    key={`act_ket_${uraian.uraian_id}`}
                                    {...register(`ket_${uraian.uraian_id}`)}
                                    className="h-10 rounded-xl" 
                                    placeholder="Tulis catatan detail jika ada issue..." 
                                  />
                                )}
                              </div>
                            </div>

                            {/* Show File Uploader if NG and not yet saved */}
                            {(isNG && !hasExistingResult && !isCompleted) && (
                              <div className="mt-4 pt-4 border-t border-border/50 animate-in fade-in slide-in-from-top-2">
                                <label className="block text-xs font-medium text-muted-foreground mb-2">Foto Bukti Temuan (Wajib)</label>
                                <div className="w-full max-w-xs group">
                                  {watch(`photo_${uraian.uraian_id}`)?.[0] && (
                                    <ImagePreview file={watch(`photo_${uraian.uraian_id}`)[0]} />
                                  )}
                                  <input 
                                    type="file" 
                                    accept="image/*" 
                                    id={`photo_${uraian.uraian_id}`}
                                    {...register(`photo_${uraian.uraian_id}`, { required: "Foto bukti wajib diunggah" })} 
                                    className="hidden" 
                                  />
                                  <label htmlFor={`photo_${uraian.uraian_id}`} className={`flex items-center justify-center gap-3 px-4 py-2.5 border rounded-xl transition-colors w-full ${errors[`photo_${uraian.uraian_id}`] ? 'border-red-500 text-red-500 bg-red-500/10 hover:bg-red-500/20' : 'bg-primary/10 text-primary border-primary/20 hover:bg-primary/20 cursor-pointer'}`}>
                                    <svg className="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"></path>
                                    </svg>
                                    <span className="font-semibold text-sm whitespace-nowrap overflow-hidden text-ellipsis">
                                      {watch(`photo_${uraian.uraian_id}`)?.[0]?.name || "Pilih Foto Bukti"}
                                    </span>
                                  </label>
                                  {errors[`photo_${uraian.uraian_id}`] && (
                                    <p className="text-red-500 text-xs mt-2 ml-1">
                                      {errors[`photo_${uraian.uraian_id}`]?.message as string}
                                    </p>
                                  )}
                                </div>
                              </div>
                            )}

                            {/* If NG and already saved, show placeholder indicating issue created */}
                            {(isNG && hasExistingResult) && (
                               <div className="mt-4 pt-4 border-t border-border/50">
                                  <div className="bg-red-500/10 text-red-600 px-4 py-3 rounded-xl border border-red-500/20 text-sm font-medium flex items-center gap-2">
                                     <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
                                     Temuan (Issue) telah dicatat.
                                  </div>
                               </div>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  </div>
                ))}
              </div>
            ))}

            {inspection.status !== "Completed" && (
              <div className="flex justify-end pt-4">
                <Button type="submit" variant="outline" disabled={isSaving}>
                  <Save className="mr-2 h-4 w-4" /> 
                  {isSaving ? "Menyimpan..." : "Simpan Draft"}
                </Button>
              </div>
            )}
          </form>
        </Card>
      </div>
    </div>
  );
}
