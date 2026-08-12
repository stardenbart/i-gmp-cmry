"use client";

import { useState, useMemo, useRef, useEffect } from "react";
import { createPortal } from "react-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import Link from "next/link";
import {
  AlertTriangle,
  Search,
  CheckCircle2,
  XCircle,
  Loader2,
  Image as ImageIcon,
  UploadCloud,
  FileBox,
  Check,
  X,
  Eye,
  Filter,
  History,
  BarChart3,
  Percent,
  Clock,
  Wrench,
  Trash2
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { issueApi, Issue, IssuePhoto, WOWRStatus } from "@/lib/api/issue.api";
import { usePermissions } from "@/lib/usePermissions";
import { useMounted } from "@/lib/useMounted";
import { useChunkedUpload } from "@/hooks/useChunkedUpload";
import { usePolling } from "@/hooks/usePolling";
import { cn, formatImageUrl } from "@/lib/utils";
import { useAuthStore } from "@/stores/authStore";

// Interface untuk item WOWR yang terisolasi per foto
interface WOWRItem {
  id: string; // photo.issue_photo_id atau issue.issue_id
  issue_id: string;
  photo_id?: string;
  initial_photo?: IssuePhoto;
  wo_id: string;
  wr_id: string;
  wowr_status: WOWRStatus;
  needs_wo_wr: boolean;
  keterangan: string;
  uraian_text?: string;
  area_name?: string;
  kawasan_name?: string;
  detail_kawasan_name?: string;
  pic_name?: string;
  aspek_name?: string;
  detail_aspek_name?: string;
  due_date?: string;
  issue_status: string;
  habit_name?: string;
  equipment_name?: string;
  infrastructure_name?: string;
  initial_photos: IssuePhoto[];
  wowr_photos: IssuePhoto[];
  raw_issue: Issue;
}

// Modal Component for Uploading Photo & WO/WR Reference Number
function UploadProofModal({ 
  item, 
  onClose,
  onSuccess 
}: { 
  item: WOWRItem; 
  onClose: () => void;
  onSuccess: () => void;
}) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [woNumber, setWoNumber] = useState(item.wo_id || item.wr_id || "");
  const [validationError, setValidationError] = useState("");

  const { uploadMutation, uploadProgress } = useChunkedUpload({ issueId: item.issue_id });
  const queryClient = useQueryClient();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const updateStatusMutation = useMutation({
    mutationFn: async () => {
      const upperWo = woNumber.trim().toUpperCase();
      if (item.photo_id) {
        return issueApi.updatePhotoWOWR(item.photo_id, {
          needs_wo_wr: true,
          wo_id: upperWo,
          wr_id: "",
          wowr_status: "PendingValidation",
        });
      }
      return issueApi.update(item.issue_id, { 
        wo_id: upperWo,
        needs_wo_wr: true,
        wowr_status: "PendingValidation"
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wowr-issues"] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      toast.success("Bukti WO/WR & Nomor Referensi berhasil disimpan. Menunggu validasi Auditor.");
      onSuccess();
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Gagal memperbarui data WO/WR.");
    }
  });

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setSelectedFile(e.target.files[0]);
    }
  };

  const handleUpload = () => {
    if (!woNumber.trim()) {
      setValidationError("Nomor referensi WO / WR wajib diisi.");
      toast.error("Nomor referensi WO / WR wajib diisi.");
      return;
    }

    if (!selectedFile) {
      toast.error("Harap pilih foto bukti perbaikan.");
      return;
    }

    setValidationError("");

    uploadMutation.mutate(
      { file: selectedFile, type: "WOWR", refPhotoId: item.photo_id },
      {
        onSuccess: () => {
          updateStatusMutation.mutate();
        },
        onError: () => {
          toast.error("Gagal mengunggah foto bukti.");
        }
      }
    );
  };

  const isUploading = uploadMutation.isPending || updateStatusMutation.isPending;

  const modalContent = (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 sm:p-0">
      <div className="w-[90vw] sm:w-[480px] bg-background rounded-2xl p-6 shadow-xl border border-border/50 animate-in fade-in zoom-in-95 duration-200 flex flex-col space-y-4">
        <div className="flex items-center justify-between border-b border-border pb-3">
          <h3 className="text-base font-bold flex items-center gap-2">
            <Wrench className="h-5 w-5 text-amber-500" />
            Upload Bukti &amp; Input Nomor WO/WR
          </h3>
          <Button variant="ghost" size="icon" onClick={onClose} disabled={isUploading} className="rounded-full h-8 w-8">
            <X className="h-4 w-4" />
          </Button>
        </div>

        <p className="text-xs text-muted-foreground">
          Temuan ID: <span className="font-mono text-foreground font-semibold">{item.issue_id}</span> {item.photo_id ? `· Foto ID: ${item.photo_id}` : ""}
        </p>

        {/* Input Text Nomor Referensi WO / WR */}
        <div className="space-y-1.5">
          <label className="text-xs font-semibold text-foreground flex items-center justify-between">
            <span>Nomor Referensi WO / WR <span className="text-red-500">*</span></span>
          </label>
          <input
            type="text"
            value={woNumber}
            onChange={(e) => {
              const val = e.target.value.toUpperCase();
              setWoNumber(val);
              if (val.trim()) setValidationError("");
            }}
            disabled={isUploading}
            placeholder="MASUKKAN NOMOR REFERENSI WO / WR..."
            className={cn(
              "w-full h-11 rounded-xl border border-border bg-card px-3 py-2 text-xs uppercase font-mono tracking-wider font-semibold focus:outline-none focus:ring-2 focus:ring-primary/40",
              validationError ? "border-red-500 focus:ring-red-500" : ""
            )}
          />
          {validationError && (
            <p className="text-red-500 text-xs font-medium pl-1">⚠️ {validationError}</p>
          )}
        </div>

        {/* Upload Image Preview Box */}
        <div className="space-y-1.5">
          <label className="text-xs font-semibold text-foreground block">Foto Bukti Perbaikan <span className="text-red-500">*</span></label>
          {selectedFile ? (
            <div className="relative aspect-video w-full rounded-xl border border-border/80 overflow-hidden bg-muted flex items-center justify-center group">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img 
                src={URL.createObjectURL(selectedFile)} 
                alt="Preview" 
                className="object-cover max-h-full max-w-full"
              />
              <button
                type="button"
                onClick={() => setSelectedFile(null)}
                disabled={isUploading}
                className="absolute top-2 right-2 p-1.5 bg-red-600 text-white rounded-full opacity-80 hover:opacity-100 transition-opacity shadow-md"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          ) : (
            <div 
              className="border-2 border-dashed border-border/70 rounded-xl p-6 text-center cursor-pointer hover:bg-muted/50 transition-colors"
              onClick={() => fileInputRef.current?.click()}
            >
              <UploadCloud className="h-9 w-9 text-muted-foreground mx-auto mb-2" />
              <p className="text-xs font-semibold">Klik untuk memilih foto bukti</p>
              <p className="text-[11px] text-muted-foreground mt-0.5">Disarankan foto langsung dari lokasi perbaikan</p>
            </div>
          )}
        </div>

        <input 
          type="file" 
          accept="image/*"
          className="hidden"
          ref={fileInputRef}
          onChange={handleFileChange}
        />

        {uploadProgress > 0 && uploadProgress < 100 && (
          <div className="w-full bg-secondary rounded-full h-2 overflow-hidden">
            <div className="bg-primary h-2 rounded-full transition-all" style={{ width: `${uploadProgress}%` }} />
          </div>
        )}

        <div className="flex gap-3 justify-end pt-2 border-t border-border">
          <Button variant="outline" onClick={onClose} disabled={isUploading} className="rounded-xl text-xs">
            Batal
          </Button>
          <Button 
            onClick={handleUpload} 
            disabled={!selectedFile || !woNumber.trim() || isUploading} 
            isLoading={isUploading}
            className="rounded-xl text-xs bg-primary text-primary-foreground font-semibold"
          >
            Upload &amp; Selesaikan
          </Button>
        </div>
      </div>
    </div>
  );

  if (!mounted || typeof window === "undefined") return null;
  return createPortal(modalContent, document.body);
}

// Expandable Row Component (Dynamic & Photo-Isolated)
function IssueRow({ 
  item, 
  onUploadClick,
  onPreviewPhoto
}: { 
  item: WOWRItem; 
  onUploadClick: (item: WOWRItem) => void;
  onPreviewPhoto: (photo: { url: string; keterangan?: string; photoType?: string; uploaderName?: string; photoId?: string; issueId?: string }) => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const { hasPermission } = usePermissions();

  // Dynamic permission controls (Database driven)
  const canValidate = hasPermission("PERM-WOWR-U") || hasPermission("PERM-INSP-A");
  const canUploadProof = hasPermission("PERM-ISS-U") || hasPermission("PERM-WOWR-R") || hasPermission("PERM-WOWR-U");

  const queryClient = useQueryClient();
  
  const approveMutation = useMutation({
    mutationFn: async () => {
      if (item.photo_id) {
        return issueApi.updatePhotoWOWR(item.photo_id, {
          needs_wo_wr: true,
          wo_id: item.wo_id,
          wr_id: item.wr_id,
          wowr_status: "Verified"
        });
      }
      return issueApi.update(item.issue_id, { 
        wowr_status: "Verified",
        issue_status: "Closed"
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wowr-issues"] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      toast.success("Bukti WO/WR berhasil diverifikasi & temuan diselesaikan!");
    },
    onError: () => toast.error("Gagal memverifikasi bukti.")
  });

  const rejectMutation = useMutation({
    mutationFn: async () => {
      if (item.photo_id) {
        return issueApi.updatePhotoWOWR(item.photo_id, {
          needs_wo_wr: true,
          wo_id: item.wo_id,
          wr_id: item.wr_id,
          wowr_status: "Rejected"
        });
      }
      return issueApi.update(item.issue_id, { wowr_status: "Rejected" });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wowr-issues"] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      toast.error("Bukti WO/WR ditolak.");
    },
    onError: () => toast.error("Gagal menolak bukti.")
  });

  const deletePhotoMutation = useMutation({
    mutationFn: (targetPhotoId: string) => issueApi.deletePhoto(item.issue_id, targetPhotoId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wowr-issues"] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      toast.success("Foto berhasil dihapus");
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.message || "Gagal menghapus foto");
    },
  });

  const isClosed = item.issue_status === "Closed" || item.issue_status === "Verified";

  return (
    <>
      <tr 
        className={cn(
          "hover:bg-muted/30 transition-colors cursor-pointer",
          expanded && "bg-muted/20"
        )}
        onClick={() => setExpanded(!expanded)}
      >
        <td className="px-4 py-3 font-mono text-xs">
          <div>{item.issue_id}</div>
          {item.photo_id && <div className="text-[10px] text-muted-foreground font-mono">Foto: {item.photo_id}</div>}
        </td>
        <td className="px-4 py-3 font-semibold text-primary font-mono uppercase">
          {item.wo_id || item.wr_id || <span className="text-muted-foreground text-xs font-normal italic font-sans">Belum diinput</span>}
        </td>
        <td className="px-4 py-3 truncate max-w-[200px]" title={item.keterangan}>
          {item.keterangan || "-"}
        </td>
        <td className="px-4 py-3">
          {(!item.wowr_status || item.wowr_status === "None") ? (
            <span className="text-xs text-muted-foreground font-medium">Menunggu Bukti</span>
          ) : item.wowr_status === "PendingValidation" ? (
            <span className="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/20">
              <Loader2 className="h-3 w-3 animate-spin" /> Menunggu Validasi Auditor
            </span>
          ) : item.wowr_status === "Verified" ? (
            <span className="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
              <CheckCircle2 className="h-3 w-3" /> Terverifikasi
            </span>
          ) : item.wowr_status === "Rejected" ? (
            <span className="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full bg-red-500/10 text-red-600 dark:text-red-400 border border-red-500/20">
              <XCircle className="h-3 w-3" /> Ditolak
            </span>
          ) : (
            <span className="text-xs text-muted-foreground">{item.wowr_status}</span>
          )}
        </td>
        <td className="px-4 py-3 text-right">
          <div className="flex justify-end items-center gap-2">
            {/* Actions for Auditor (Dynamic permission: PERM-WOWR-U) */}
            {canValidate && item.wowr_status === "PendingValidation" && (
              <>
                <Button 
                  size="sm" 
                  variant="default"
                  className="text-xs h-8 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl font-semibold"
                  disabled={approveMutation.isPending || rejectMutation.isPending}
                  onClick={(e) => {
                    e.stopPropagation();
                    approveMutation.mutate();
                  }}
                >
                  <Check className="h-3.5 w-3.5 mr-1" />
                  Verifikasi
                </Button>
                <Button 
                  size="sm" 
                  variant="destructive"
                  className="text-xs h-8 rounded-xl font-semibold"
                  disabled={approveMutation.isPending || rejectMutation.isPending}
                  onClick={(e) => {
                    e.stopPropagation();
                    rejectMutation.mutate();
                  }}
                >
                  <X className="h-3.5 w-3.5 mr-1" />
                  Tolak
                </Button>
              </>
            )}

            {/* Actions for Auditee / PIC (Dynamic permission: PERM-ISS-U or PERM-WOWR-R) */}
            {canUploadProof && !isClosed && item.wowr_status !== "Verified" && (
              <Button 
                size="sm" 
                variant="default"
                className="text-xs h-8 rounded-xl bg-primary text-primary-foreground hover:bg-primary/90 font-semibold shadow-xs"
                onClick={(e) => {
                  e.stopPropagation();
                  onUploadClick(item);
                }}
              >
                <UploadCloud className="h-3.5 w-3.5 mr-1" />
                {item.wowr_status === "PendingValidation" ? "Re-upload Bukti" : "Upload Bukti"}
              </Button>
            )}
          </div>
        </td>
      </tr>
      {expanded && (
        <tr className="bg-muted/5 border-b">
          <td colSpan={5} className="p-0">
            <div className="p-4 bg-muted/10 border-t border-border/50 animate-in slide-in-from-top-2 fade-in duration-200">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <h4 className="font-semibold text-sm mb-3 border-b pb-1 border-border/50">Detail Temuan &amp; Spesifikasi</h4>
                  <div className="space-y-2.5 text-xs mb-4 bg-background/60 p-3.5 rounded-xl border border-border/60">
                    {/* Location & PIC */}
                    <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 border-b border-border/40 pb-2">
                      <div>
                        <span className="text-muted-foreground block text-[10px] uppercase font-semibold">Area</span>
                        <span className="font-semibold text-foreground">{item.area_name || "-"}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground block text-[10px] uppercase font-semibold">Kawasan</span>
                        <span className="font-semibold text-foreground">{item.kawasan_name || "-"}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground block text-[10px] uppercase font-semibold">Detail Kawasan / PIC</span>
                        <span className="font-semibold text-foreground">
                          {item.detail_kawasan_name || "-"} {item.pic_name ? `(${item.pic_name})` : ""}
                        </span>
                      </div>
                    </div>

                    {/* Aspek & Detail Aspek */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                      <div>
                        <span className="text-muted-foreground block text-[10px] uppercase font-semibold">Aspek Penilaian</span>
                        <span className="font-semibold text-foreground">{item.aspek_name || "-"}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground block text-[10px] uppercase font-semibold">Detail Aspek</span>
                        <span className="font-semibold text-foreground">{item.detail_aspek_name || "-"}</span>
                      </div>
                    </div>

                    {/* Uraian Checklist */}
                    <div>
                      <span className="text-muted-foreground block text-[10px] uppercase font-semibold">Uraian Checklist</span>
                      <span className="font-medium text-foreground">{item.uraian_text || item.keterangan || "-"}</span>
                    </div>

                    {/* HEI Specifications (Flexibel Optional - Render only if category is present) */}
                    {(item.habit_name || item.equipment_name || item.infrastructure_name) && (
                      <div className="pt-2 border-t border-border/40 grid grid-cols-1 sm:grid-cols-3 gap-2">
                        {item.habit_name && (
                          <div>
                            <span className="text-muted-foreground block text-[10px] uppercase font-semibold">Habit</span>
                            <span className="font-semibold text-emerald-600 dark:text-emerald-400">
                              {item.habit_name}
                            </span>
                          </div>
                        )}
                        {item.equipment_name && (
                          <div>
                            <span className="text-muted-foreground block text-[10px] uppercase font-semibold">Equipment</span>
                            <span className="font-semibold text-blue-600 dark:text-blue-400">
                              {item.equipment_name}
                            </span>
                          </div>
                        )}
                        {item.infrastructure_name && (
                          <div>
                            <span className="text-muted-foreground block text-[10px] uppercase font-semibold">Infrastructure</span>
                            <span className="font-semibold text-purple-600 dark:text-purple-400">
                              {item.infrastructure_name}
                            </span>
                          </div>
                        )}
                      </div>
                    )}

                    <div className="pt-2 border-t border-border/40 grid grid-cols-1 sm:grid-cols-2 gap-2">
                      <div>
                        <span className="text-muted-foreground block text-[10px] uppercase font-semibold">Target Penyelesaian (Due Date)</span>
                        <span className="font-medium">{item.due_date ? new Date(item.due_date).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }) : "-"}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground block text-[10px] uppercase font-semibold">Keterangan Foto / Temuan</span>
                        <span className="font-medium whitespace-pre-wrap">{item.keterangan || "-"}</span>
                      </div>
                    </div>
                  </div>
                  
                  <h5 className="text-xs font-semibold mb-2 text-muted-foreground uppercase tracking-wider">Foto Temuan Awal</h5>
                  {item.initial_photos.length > 0 ? (
                    <div className="flex gap-3 overflow-x-auto pb-2 scrollbar-none">
                      {item.initial_photos.map(p => (
                        <div key={p.issue_photo_id} className="flex flex-col w-40 shrink-0">
                          <div 
                            className="relative group h-28 w-40 overflow-hidden rounded-xl border border-border/60 shadow-xs cursor-pointer bg-muted"
                            onClick={(e) => {
                              e.stopPropagation();
                              onPreviewPhoto({
                                url: p.image_url,
                                keterangan: p.keterangan,
                                photoType: "Foto Temuan Awal",
                                uploaderName: (p as any).uploader_name,
                                photoId: p.issue_photo_id,
                                issueId: item.issue_id,
                              });
                            }}
                          >
                            {/* eslint-disable-next-line @next/next/no-img-element */}
                            <img 
                              src={formatImageUrl(p.image_url) || "/placeholder.png"} 
                              alt="Initial" 
                              onError={(e) => { e.currentTarget.src = "/placeholder.png"; }}
                              className="h-full w-full object-cover transition-transform group-hover:scale-105" 
                            />
                            <div className="absolute inset-0 bg-black/40 sm:bg-black/50 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2 text-white text-xs font-medium">
                              <Eye className="h-4 w-4" />
                              {!isClosed && canValidate && (
                                <button
                                  type="button"
                                  disabled={deletePhotoMutation.isPending}
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    toast.warning("Hapus foto temuan awal ini?", {
                                      description: "Tindakan ini tidak dapat dibatalkan.",
                                      action: {
                                        label: "Hapus",
                                        onClick: () => deletePhotoMutation.mutate(p.issue_photo_id),
                                      },
                                      cancel: {
                                        label: "Batal",
                                        onClick: () => {},
                                      },
                                    });
                                  }}
                                  className="p-1.5 bg-red-600 text-white rounded-full hover:bg-red-700 transition-colors shadow-md z-10"
                                  title="Hapus Foto"
                                >
                                  <Trash2 className="h-3.5 w-3.5" />
                                </button>
                              )}
                            </div>
                          </div>
                          <p className="text-[11px] font-medium text-foreground truncate mt-1.5 px-0.5" title={p.keterangan || "Tanpa keterangan"}>
                            {p.keterangan ? p.keterangan : <span className="text-muted-foreground italic text-[10px]">(Tanpa keterangan)</span>}
                          </p>
                        </div>
                      ))}
                    </div>
                  ) : <span className="text-xs text-muted-foreground italic">Tidak ada foto temuan awal.</span>}
                </div>
                <div>
                  <h4 className="font-semibold text-sm mb-3 border-b pb-1 border-border/50">Bukti Penyelesaian WO/WR</h4>
                  {item.wowr_photos.length > 0 ? (
                    <div className="flex gap-3 overflow-x-auto pb-2 scrollbar-none">
                      {item.wowr_photos.map(p => (
                        <div key={p.issue_photo_id} className="flex flex-col w-40 shrink-0">
                          <div 
                            className="relative group h-28 w-40 overflow-hidden rounded-xl border border-border/60 shadow-xs cursor-pointer bg-muted"
                            onClick={(e) => {
                              e.stopPropagation();
                              onPreviewPhoto({
                                url: p.image_url,
                                keterangan: p.keterangan,
                                photoType: "Bukti Penyelesaian WO/WR",
                                uploaderName: (p as any).uploader_name,
                                photoId: p.issue_photo_id,
                                issueId: item.issue_id,
                              });
                            }}
                          >
                            {/* eslint-disable-next-line @next/next/no-img-element */}
                            <img 
                              src={formatImageUrl(p.image_url) || "/placeholder.png"} 
                              alt="WOWR Evidence" 
                              onError={(e) => { e.currentTarget.src = "/placeholder.png"; }}
                              className="h-full w-full object-cover transition-transform group-hover:scale-105" 
                            />
                            <div className="absolute inset-0 bg-black/40 sm:bg-black/50 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2 text-white text-xs font-medium">
                              <Eye className="h-4 w-4" />
                              {!isClosed && (canUploadProof || canValidate) && (
                                <button
                                  type="button"
                                  disabled={deletePhotoMutation.isPending}
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    toast.warning("Hapus foto bukti penyelesaian ini?", {
                                      description: "Tindakan ini tidak dapat dibatalkan.",
                                      action: {
                                        label: "Hapus",
                                        onClick: () => deletePhotoMutation.mutate(p.issue_photo_id),
                                      },
                                      cancel: {
                                        label: "Batal",
                                        onClick: () => {},
                                      },
                                    });
                                  }}
                                  className="p-1.5 bg-red-600 text-white rounded-full hover:bg-red-700 transition-colors shadow-md z-10"
                                  title="Hapus Foto"
                                >
                                  <Trash2 className="h-3.5 w-3.5" />
                                </button>
                              )}
                            </div>
                          </div>
                          <p className="text-[11px] font-medium text-foreground truncate mt-1.5 px-0.5" title={p.keterangan || "Tanpa keterangan"}>
                            {p.keterangan ? p.keterangan : <span className="text-muted-foreground italic text-[10px]">(Tanpa keterangan)</span>}
                          </p>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="h-28 flex flex-col items-center justify-center border-2 border-dashed border-border/60 rounded-xl text-center bg-muted/30">
                      <ImageIcon className="h-6 w-6 text-muted-foreground/50 mb-1" />
                      <span className="text-xs text-muted-foreground italic">Belum ada foto bukti penyelesaian WO/WR.</span>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </td>
        </tr>
      )}
    </>
  );
}

/* ── Page ──────────────────────────────────────────────────────────────── */
export default function WOWRPage() {
  const { plantCode, userId } = useParams() as { plantCode: string; userId: string };
  const user = useAuthStore(state => state.user);
  const mounted = useMounted();
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("ALL");
  const [selectedItem, setSelectedItem] = useState<WOWRItem | null>(null);
  const [previewPhoto, setPreviewPhoto] = useState<{ url: string; keterangan?: string; photoType?: string; uploaderName?: string; photoId?: string; issueId?: string } | null>(null);

  usePolling();

  const { data, isLoading } = useQuery({
    queryKey: ["wowr-issues", userId],
    queryFn: () => issueApi.getAll({ needs_wo_wr: true, limit: 1000 }),
    enabled: mounted && !!user,
  });

  const allIssues: Issue[] = Array.isArray(data?.items) ? data.items : [];

  // Transform allIssues into WOWRItems isolated per initial photo (IssuePhoto)
  const wowrItems = useMemo(() => {
    const items: WOWRItem[] = [];
    allIssues.forEach((issue) => {
      const initialPhotos = issue.photos?.filter((p) => p.photo_type === "Initial") || [];
      if (initialPhotos.length > 0) {
        initialPhotos.forEach((photo) => {
          const specificWOWRPhotos = issue.photos?.filter(
            (p) =>
              p.photo_type === "WOWR" &&
              (p.ref_photo_id === photo.issue_photo_id || (!p.ref_photo_id && initialPhotos.length <= 1))
          ) || [];

          items.push({
            id: photo.issue_photo_id,
            issue_id: issue.issue_id,
            photo_id: photo.issue_photo_id,
            initial_photo: photo,
            wo_id: photo.wo_id || issue.wo_id || "",
            wr_id: photo.wr_id || issue.wr_id || "",
            wowr_status: photo.wowr_status || issue.wowr_status || "None",
            needs_wo_wr: photo.needs_wo_wr || issue.needs_wo_wr || false,
            keterangan: photo.keterangan || issue.keterangan || "",
            uraian_text: issue.uraian_text,
            area_name: issue.area_name,
            kawasan_name: issue.kawasan_name,
            detail_kawasan_name: issue.detail_kawasan_name,
            pic_name: issue.pic_name,
            aspek_name: issue.aspek_name,
            detail_aspek_name: issue.detail_aspek_name,
            due_date: issue.due_date,
            issue_status: issue.issue_status,
            habit_name: photo.habit_name || issue.habit_name || issue.hei?.habit?.habit_name,
            equipment_name: photo.equipment_name || issue.equipment_name || issue.hei?.equipment?.equipment_name,
            infrastructure_name: photo.infrastructure_name || issue.infrastructure_name || issue.hei?.infrastructure?.infrastructure_name,
            initial_photos: [photo],
            wowr_photos: specificWOWRPhotos,
            raw_issue: issue,
          });
        });
      } else {
        const specificWOWRPhotos = issue.photos?.filter(
          (p) => p.photo_type === "WOWR"
        ) || [];

        items.push({
          id: issue.issue_id,
          issue_id: issue.issue_id,
          wo_id: issue.wo_id || "",
          wr_id: issue.wr_id || "",
          wowr_status: issue.wowr_status || "None",
          needs_wo_wr: issue.needs_wo_wr || false,
          keterangan: issue.keterangan || "",
          uraian_text: issue.uraian_text,
          area_name: issue.area_name,
          kawasan_name: issue.kawasan_name,
          detail_kawasan_name: issue.detail_kawasan_name,
          pic_name: issue.pic_name,
          aspek_name: issue.aspek_name,
          detail_aspek_name: issue.detail_aspek_name,
          due_date: issue.due_date,
          issue_status: issue.issue_status,
          habit_name: issue.habit_name || issue.hei?.habit?.habit_name,
          equipment_name: issue.equipment_name || issue.hei?.equipment?.equipment_name,
          infrastructure_name: issue.infrastructure_name || issue.hei?.infrastructure?.infrastructure_name,
          initial_photos: [],
          wowr_photos: specificWOWRPhotos,
          raw_issue: issue,
        });
      }
    });
    return items;
  }, [allIssues]);

  // Summary statistics calculations based on isolated WOWRItems
  const totalCount = wowrItems.length;
  const verifiedCount = useMemo(() => wowrItems.filter(i => i.wowr_status === "Verified").length, [wowrItems]);
  const pendingCount = useMemo(() => wowrItems.filter(i => i.wowr_status === "PendingValidation").length, [wowrItems]);
  const rejectedCount = useMemo(() => wowrItems.filter(i => i.wowr_status === "Rejected").length, [wowrItems]);
  const awaitingCount = useMemo(() => wowrItems.filter(i => !i.wowr_status || i.wowr_status === "None").length, [wowrItems]);

  const verifiedRate = totalCount > 0 ? (verifiedCount / totalCount) * 100 : 0;
  const pendingRate = totalCount > 0 ? (pendingCount / totalCount) * 100 : 0;
  const rejectedRate = totalCount > 0 ? (rejectedCount / totalCount) * 100 : 0;

  const filteredItems = useMemo(() => {
    let result = wowrItems;

    if (statusFilter !== "ALL") {
      if (statusFilter === "None") {
        result = result.filter(i => !i.wowr_status || i.wowr_status === "None");
      } else {
        result = result.filter(i => i.wowr_status === statusFilter);
      }
    }

    if (!search.trim()) return result;
    const q = search.toLowerCase();
    return result.filter(
      (i) =>
        i.keterangan?.toLowerCase().includes(q) ||
        i.issue_id?.toLowerCase().includes(q) ||
        i.photo_id?.toLowerCase().includes(q) ||
        i.wo_id?.toLowerCase().includes(q) ||
        i.wr_id?.toLowerCase().includes(q) ||
        i.area_name?.toLowerCase().includes(q) ||
        i.kawasan_name?.toLowerCase().includes(q) ||
        i.detail_kawasan_name?.toLowerCase().includes(q) ||
        i.pic_name?.toLowerCase().includes(q) ||
        i.uraian_text?.toLowerCase().includes(q)
    );
  }, [wowrItems, statusFilter, search]);

  if (!mounted || !user) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <>
      <div className="space-y-6 pb-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
        {/* ── Page header ── */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-card p-5 rounded-2xl border shadow-sm">
          <div className="flex items-center gap-4">
            <div className="h-12 w-12 rounded-2xl bg-primary/10 border border-primary/20 flex items-center justify-center text-primary shadow-inner">
              <FileBox className="h-6 w-6" />
            </div>
            <div>
              <h2 className="text-xl font-bold tracking-tight">Manajemen WO / WR</h2>
              <p className="text-sm text-muted-foreground mt-0.5">
                Kelola, unggah bukti perbaikan, dan verifikasi Work Order &amp; Work Request per foto temuan
              </p>
            </div>
          </div>
          <Link href={`/cimory/${plantCode}/dashboard/${userId}/monitoring/wowr`}>
            <Button className="rounded-xl gap-2 font-semibold shadow-sm hover:shadow transition-all">
              <BarChart3 className="h-4 w-4" />
              Laporan &amp; Analytics WO/WR
            </Button>
          </Link>
        </div>

        {/* ── Summary Statistics Cards ── */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-card p-4 rounded-2xl border shadow-sm space-y-1">
            <span className="text-xs text-muted-foreground font-medium">Total Foto WO/WR</span>
            <div className="flex items-baseline justify-between">
              <span className="text-2xl font-bold font-mono">{totalCount}</span>
              <span className="text-xs text-muted-foreground font-semibold">100%</span>
            </div>
            <div className="w-full bg-secondary h-1.5 rounded-full overflow-hidden mt-2">
              <div className="bg-primary h-1.5 rounded-full w-full" />
            </div>
          </div>

          <div className="bg-card p-4 rounded-2xl border shadow-sm space-y-1">
            <span className="text-xs text-muted-foreground font-medium">Terverifikasi</span>
            <div className="flex items-baseline justify-between">
              <span className="text-2xl font-bold font-mono text-emerald-600 dark:text-emerald-400">{verifiedCount}</span>
              <span className="text-xs font-bold text-emerald-600 dark:text-emerald-400">{verifiedRate.toFixed(1)}%</span>
            </div>
            <div className="w-full bg-secondary h-1.5 rounded-full overflow-hidden mt-2">
              <div className="bg-emerald-500 h-1.5 rounded-full" style={{ width: `${verifiedRate}%` }} />
            </div>
          </div>

          <div className="bg-card p-4 rounded-2xl border shadow-sm space-y-1">
            <span className="text-xs text-muted-foreground font-medium">Menunggu Validasi</span>
            <div className="flex items-baseline justify-between">
              <span className="text-2xl font-bold font-mono text-purple-600 dark:text-purple-400">{pendingCount}</span>
              <span className="text-xs font-bold text-purple-600 dark:text-purple-400">{pendingRate.toFixed(1)}%</span>
            </div>
            <div className="w-full bg-secondary h-1.5 rounded-full overflow-hidden mt-2">
              <div className="bg-purple-500 h-1.5 rounded-full" style={{ width: `${pendingRate}%` }} />
            </div>
          </div>

          <div className="bg-card p-4 rounded-2xl border shadow-sm space-y-1">
            <span className="text-xs text-muted-foreground font-medium">Ditolak</span>
            <div className="flex items-baseline justify-between">
              <span className="text-2xl font-bold font-mono text-red-600 dark:text-red-400">{rejectedCount}</span>
              <span className="text-xs font-bold text-red-600 dark:text-red-400">{rejectedRate.toFixed(1)}%</span>
            </div>
            <div className="w-full bg-secondary h-1.5 rounded-full overflow-hidden mt-2">
              <div className="bg-red-500 h-1.5 rounded-full" style={{ width: `${rejectedRate}%` }} />
            </div>
          </div>
        </div>

        {/* ── Status Filter Tabs & Search ── */}
        <div className="flex flex-col md:flex-row items-stretch md:items-center justify-between gap-4">
          <div className="flex items-center gap-1.5 overflow-x-auto pb-1 md:pb-0 scrollbar-none">
            {[
              { id: "ALL", label: "Semua", count: totalCount },
              { id: "PendingValidation", label: "Menunggu Validasi", count: pendingCount, color: "text-purple-500 bg-purple-500/10 border-purple-500/20" },
              { id: "Verified", label: "Terverifikasi", count: verifiedCount, color: "text-emerald-500 bg-emerald-500/10 border-emerald-500/20" },
              { id: "Rejected", label: "Ditolak", count: rejectedCount, color: "text-red-500 bg-red-500/10 border-red-500/20" },
              { id: "None", label: "Menunggu Bukti", count: awaitingCount },
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setStatusFilter(tab.id)}
                className={cn(
                  "px-3.5 py-2 rounded-xl text-xs font-semibold whitespace-nowrap transition-all flex items-center gap-2 border",
                  statusFilter === tab.id
                    ? "bg-primary text-primary-foreground border-primary shadow-sm"
                    : "bg-card hover:bg-muted text-muted-foreground border-border/60"
                )}
              >
                <span>{tab.label}</span>
                <span className={cn(
                  "px-1.5 py-0.5 rounded-full text-[10px] font-bold font-mono",
                  statusFilter === tab.id
                    ? "bg-primary-foreground/20 text-primary-foreground"
                    : tab.color || "bg-muted text-muted-foreground"
                )}>
                  {tab.count}
                </span>
              </button>
            ))}
          </div>

          <div className="relative min-w-[240px]">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
            <input
              type="search"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Cari WO/WR, area, PIC..."
              className="w-full rounded-xl border bg-card pl-9 pr-4 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary/40 transition-all"
            />
          </div>
        </div>

        {/* ── Table / List ── */}
        <div className="bg-card rounded-2xl border overflow-hidden shadow-sm">
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground text-xs uppercase">
                <tr>
                  <th className="px-4 py-3 font-medium">Issue / Foto ID</th>
                  <th className="px-4 py-3 font-medium">Nomor WO / WR</th>
                  <th className="px-4 py-3 font-medium max-w-xs">Keterangan Foto</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium text-right">Aksi</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/50">
                {isLoading ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-8 text-center">
                      <Loader2 className="h-6 w-6 animate-spin mx-auto text-primary" />
                      <p className="text-xs text-muted-foreground mt-2">Memuat data WO/WR...</p>
                    </td>
                  </tr>
                ) : filteredItems.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="p-6 sm:p-10">
                      <div className="w-full rounded-3xl border border-dashed border-border/70 bg-gradient-to-b from-card/80 via-card/40 to-background p-8 sm:p-12 text-center shadow-sm">
                        <div className="mx-auto w-full max-w-md text-center space-y-4" style={{ width: "100%", maxWidth: "28rem", marginLeft: "auto", marginRight: "auto" }}>
                          <div className="inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-500/10 border border-amber-500/20 text-amber-500 shadow-inner mx-auto mb-2">
                            <AlertTriangle className="h-8 w-8 text-amber-500" />
                          </div>

                          <h3 className="w-full text-lg font-bold text-foreground tracking-tight text-center block">
                            {search ? "Tidak Ada Data WO/WR Ditemukan" : "Belum Ada Data WO/WR"}
                          </h3>

                          <p className="w-full text-sm text-muted-foreground leading-relaxed text-center block" style={{ wordBreak: "normal", overflowWrap: "break-word" }}>
                            {search
                              ? `Tidak ada data WO/WR yang cocok dengan kata kunci "${search}".`
                              : "Belum ada data Work Order / Work Request yang perlu ditindaklanjuti."}
                          </p>

                          {search && (
                            <div className="w-full flex items-center justify-center pt-2">
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => setSearch("")}
                                className="rounded-full px-5 h-9 text-xs font-semibold"
                              >
                                <X className="mr-1.5 h-3.5 w-3.5" /> Hapus Filter
                              </Button>
                            </div>
                          )}
                        </div>
                      </div>
                    </td>
                  </tr>
                ) : (
                  filteredItems.map((item) => (
                    <IssueRow 
                      key={item.id} 
                      item={item} 
                      onUploadClick={(it) => setSelectedItem(it)} 
                      onPreviewPhoto={(photo) => setPreviewPhoto(photo)}
                    />
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
      
      {/* Upload Modal */}
      {selectedItem && (
        <UploadProofModal 
          item={selectedItem} 
          onClose={() => setSelectedItem(null)}
          onSuccess={() => setSelectedItem(null)}
        />
      )}

      {/* Rich Image Lightbox Preview Modal with Caption */}
      {previewPhoto && (
        <div
          className="fixed inset-0 z-[110] flex items-center justify-center bg-black/90 backdrop-blur-md p-4 animate-in fade-in duration-200"
          onClick={() => setPreviewPhoto(null)}
        >
          <div className="relative max-w-3xl w-full bg-card border border-border/80 rounded-3xl overflow-hidden shadow-2xl flex flex-col" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between p-4 border-b border-border/60 bg-muted/40">
              <div>
                <span className="text-xs font-bold uppercase tracking-wider text-primary">
                  {previewPhoto.photoType || "Bukti Foto Temuan"}
                </span>
                {previewPhoto.uploaderName && (
                  <span className="text-xs text-muted-foreground block">
                    Diunggah oleh: <strong>{previewPhoto.uploaderName}</strong>
                  </span>
                )}
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setPreviewPhoto(null)}
                className="rounded-full hover:bg-muted"
              >
                <X className="h-5 w-5" />
              </Button>
            </div>

            <div className="relative bg-black flex items-center justify-center p-2 min-h-[300px] max-h-[65vh]">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={formatImageUrl(previewPhoto.url) || "/placeholder.png"}
                alt="Preview Foto"
                onError={(e) => { e.currentTarget.src = "/placeholder.png"; }}
                className="max-h-[60vh] w-auto max-w-full object-contain rounded-lg shadow-md"
              />
            </div>

            {/* Photo Caption Text Banner */}
            <div className="p-4 bg-card border-t border-border/60 space-y-1">
              <span className="text-[11px] font-bold text-muted-foreground uppercase tracking-wider block">Keterangan Foto:</span>
              <p className="text-sm font-medium text-foreground leading-relaxed whitespace-pre-wrap">
                {previewPhoto.keterangan ? previewPhoto.keterangan : <span className="text-muted-foreground italic text-xs">Tidak ada keterangan tertulis untuk foto ini.</span>}
              </p>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
