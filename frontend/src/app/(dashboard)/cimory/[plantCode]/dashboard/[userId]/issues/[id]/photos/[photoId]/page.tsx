"use client";

import { useState, useRef } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft,
  Upload,
  Trash2,
  Edit,
  Check,
  X,
  Loader2,
  Eye,
  Tag,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { issueApi, IssuePhoto, IssueStatus, IssueCategory } from "@/lib/api/issue.api";
import { formatImageUrl } from "@/lib/utils";
import { useChunkedUpload } from "@/hooks/useChunkedUpload";
import { useAuthStore } from "@/stores/authStore";
import { usePermissions } from "@/lib/usePermissions";

export default function InitialPhotoDetailPage() {
  const { plantCode, userId, id, photoId } = useParams() as {
    plantCode: string;
    userId: string;
    id: string;
    photoId: string;
  };
  const router = useRouter();
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [selectedImage, setSelectedImage] = useState<string | null>(null);

  const user = useAuthStore((state) => state.user);
  const { hasPermission } = usePermissions();

  const { uploadMutation, uploadProgress } = useChunkedUpload({ issueId: id });

  // Fetch issue details
  const { data: issueRes, isLoading: isIssueLoading } = useQuery({
    queryKey: ["issue", id],
    queryFn: () => issueApi.getById(id),
  });

  // Fetch all photos for this issue
  const { data: photosRes, isLoading: isPhotosLoading } = useQuery({
    queryKey: ["issue-photos", id],
    queryFn: () => issueApi.getPhotos(id),
  });

  const deleteMutation = useMutation({
    mutationFn: (targetPhotoId: string) => issueApi.deletePhoto(id, targetPhotoId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issue-photos", id] });
      toast.success("Foto dihapus");
    },
    onError: () => toast.error("Gagal menghapus foto"),
  });

  const issue = issueRes?.data;
  const allPhotos: IssuePhoto[] = photosRes?.data || [];

  // Identify current initial photo & count total initial photos
  const currentPhoto = allPhotos.find((p) => p.issue_photo_id === photoId) || allPhotos.find((p) => p.photo_type === "Initial");
  const initialPhotosCount = allPhotos.filter((p) => p.photo_type === "Initial").length;

  // Filter follow-up photos specific to this initial photo (or photos without ref_photo_id)
  const followUpPhotos = allPhotos.filter(
    (p) =>
      p.photo_type === "FollowUp" &&
      (p.ref_photo_id === photoId || !p.ref_photo_id || initialPhotosCount <= 1)
  );

  if (isIssueLoading || isPhotosLoading) {
    return (
      <div className="flex justify-center p-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!issue) {
    return <div className="text-center p-12 text-muted-foreground">Temuan tidak ditemukan.</div>;
  }

  const isAuditor = hasPermission("PERM-INSP-C") || hasPermission("PERM-WOWR-U");
  const isPIC = !isAuditor;

  const rawStatus = issue.issue_status as IssueStatus;
  const isClosed = rawStatus === "Closed" || rawStatus === "Verified";
  const isWorkStarted = rawStatus === "InProgress" || rawStatus === "PendingValidation";
  const canUploadFollowUp = !isClosed && isWorkStarted;

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      uploadMutation.mutate({ file, type: "FollowUp", refPhotoId: photoId });
      e.target.value = "";
    }
  };

  const backUrl = `/cimory/${plantCode || "all"}/dashboard/${userId}/issues/${id}`;

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-12">
      {/* Header Bar */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <Link href={backUrl}>
            <Button variant="ghost" size="icon" className="rounded-full">
              <ArrowLeft className="h-5 w-5" />
            </Button>
          </Link>
          <div>
            <h2 className="text-base font-bold tracking-tight">Detail Spesifikasi Temuan Awal</h2>
            <p className="text-muted-foreground text-xs font-mono mt-0.5">
              Issue ID: <span className="font-semibold text-foreground">{id}</span> · Foto ID: {photoId}
            </p>
          </div>
        </div>

        <Link href={backUrl}>
          <Button variant="outline" size="sm" className="rounded-xl">
            Kembali ke Temuan
          </Button>
        </Link>
      </div>

      {/* Warning Banner if work has not been started */}
      {!isWorkStarted && !isClosed && (
        <div className="p-4 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-700 dark:text-amber-400 text-xs flex flex-col sm:flex-row sm:items-center justify-between gap-3 shadow-xs">
          <div className="flex items-center gap-2">
            <span>
              Pengunggahan Foto Follow-Up perbaikan belum dapat dilakukan karena status temuan masih <strong>{rawStatus}</strong>. Silakan klik tombol <strong>&quot;Mulai Kerjakan&quot;</strong> di halaman utama detail temuan terlebih dahulu.
            </span>
          </div>
          <Link href={backUrl} className="shrink-0">
            <Button size="sm" variant="outline" className="border-amber-500/50 text-amber-700 dark:text-amber-300 hover:bg-amber-500/10 rounded-xl">
              Ke Halaman Utama Temuan
            </Button>
          </Link>
        </div>
      )}

      {/* Grid Content */}
      <div className="grid gap-6 md:grid-cols-2">
        {/* Left Column: Location & Issue Specification Info Cards */}
        <div className="space-y-6">
          {/* Card 1: Lokasi Audit */}
          <Card className="p-6 bg-card/60 backdrop-blur-md shadow-sm border-border/80 space-y-3">
            <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-muted-foreground border-b border-border pb-2">
              <span>Lokasi Audit Area</span>
            </div>
            <div className="grid grid-cols-3 gap-3 pt-1 text-xs">
              <div className="bg-muted/40 p-3 rounded-xl border border-border/60">
                <span className="text-muted-foreground block text-sm font-bold">Area</span>
                <span className="font-bold text-foreground text-xs">{issue.area_name || "Tanpa Area"}</span>
              </div>
              <div className="bg-muted/40 p-3 rounded-xl border border-border/60">
                <span className="text-muted-foreground block text-sm font-bold">Kawasan</span>
                <span className="font-bold text-foreground text-xs">{issue.kawasan_name || "Tanpa Kawasan"}</span>
              </div>
              <div className="bg-muted/40 p-2 rounded-xl border border-border/60">
                <span className="text-muted-foreground block text-xs font-bold">Detail Kawasan</span>
                <span className="font-bold text-foreground text-xs">{issue.detail_kawasan_name || "Tanpa Detail Kawasan"}</span>
              </div>
            </div>
          </Card>

          {/* Card 2: Detail Issue & Checklist */}
          <Card className="p-6 bg-card/60 backdrop-blur-md shadow-sm border-border/80 space-y-4">
            <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-muted-foreground border-b border-border pb-2">
              <span>Detail Issue &amp; Uraian Temuan</span>
            </div>

            <div className="space-y-3 text-xs">
              <div className="grid grid-cols-2 gap-3">
                <div className="bg-muted/40 p-3 rounded-xl border border-border/60">
                  <span className="text-muted-foreground block text-[11px] font-medium">Aspek</span>
                  <span className="font-semibold text-foreground">{issue.aspek_name || "Tanpa Aspek"}</span>
                </div>
                <div className="bg-muted/40 p-3 rounded-xl border border-border/60">
                  <span className="text-muted-foreground block text-[11px] font-medium">Detail Aspek</span>
                  <span className="font-semibold text-foreground">{issue.detail_aspek_name || "Tanpa Detail Aspek"}</span>
                </div>
              </div>

              <div>
                <span className="text-muted-foreground block text-[11px] font-medium mb-1 flex items-center gap-1">
                  Uraian Checklist / Keterangan Temuan
                </span>
                <div className="bg-background/80 p-3 rounded-xl border border-border/80 font-medium text-foreground text-xs leading-relaxed">
                  {issue.uraian_text || issue.keterangan || "Tidak ada rincian keterangan"}
                </div>
              </div>

              {/* Output Kategori HEI (Hasil Inspeksi) */}
              <div className="pt-3 border-t border-border/60 space-y-1.5">
                <span className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground flex items-center gap-1">
                  <Tag className="h-3 w-3 text-primary" /> Kategori HEI (Hasil Inspeksi)
                </span>
                {(() => {
                  const heiCat = currentPhoto?.hei_category || issue?.hei_category;
                  const heiName = currentPhoto?.hei_name || issue?.hei_name || currentPhoto?.habit_name || currentPhoto?.equipment_name || currentPhoto?.infrastructure_name || issue?.habit_name || issue?.equipment_name || issue?.infrastructure_name;
                  
                  if (heiCat || heiName) {
                    const displayLabel = [
                      heiCat ? `[${heiCat}]` : "",
                      heiName && heiName.toLowerCase() !== heiCat?.toLowerCase() ? heiName : ""
                    ].filter(Boolean).join(" ");

                    return (
                      <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-xl bg-primary/10 text-primary font-semibold text-xs border border-primary/20">
                        {displayLabel || heiCat || heiName}
                      </div>
                    );
                  }

                  return (
                    <div className="text-xs text-muted-foreground italic bg-muted/20 px-3 py-1.5 rounded-xl border border-dashed border-border/60 w-fit">
                      Tanpa Klasifikasi HEI (Hasil Inspeksi)
                    </div>
                  );
                })()}
              </div>

              <div className="flex items-center justify-between pt-2 border-t border-border/60 text-muted-foreground">
                <span className="flex items-center gap-1">
                  PIC: <strong className="text-foreground">{issue.pic_name || issue.issue_pic_user_id}</strong>
                </span>
                <span className="flex items-center gap-1 font-mono">
                  {new Date(issue.created_at).toLocaleDateString("id-ID", { day: "numeric", month: "short", year: "numeric" })}
                </span>
              </div>
            </div>
          </Card>
        </div>

        {/* Right Column: Large Photo Preview */}
        <Card className="p-6 bg-card/60 backdrop-blur-md shadow-sm border-border/80 space-y-4 flex flex-col justify-between">
          <div className="flex items-center justify-between border-b border-border pb-2">
            <h3 className="font-bold text-base flex items-center gap-2">
              Foto Bukti Temuan Awal
            </h3>
            <span className="text-xs font-mono text-muted-foreground">High Resolution</span>
          </div>

          {currentPhoto ? (
            <div className="relative aspect-4/3 w-full rounded-2xl overflow-hidden bg-black/90 group border border-border shadow-inner">
              <img
                src={formatImageUrl(currentPhoto.image_url)}
                alt="Foto Temuan Awal"
                className="h-full w-full object-contain cursor-pointer transition-transform group-hover:scale-105"
                onClick={() => setSelectedImage(formatImageUrl(currentPhoto.image_url))}
              />
              <div
                className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center cursor-pointer"
                onClick={() => setSelectedImage(formatImageUrl(currentPhoto.image_url))}
              >
                <Button size="sm" variant="outline" className="gap-2 font-medium shadow-md bg-background/90 text-foreground">
                  <Eye className="h-4 w-4" /> Perbesar Foto
                </Button>
              </div>
            </div>
          ) : (
            <div className="h-64 flex items-center justify-center rounded-2xl border border-dashed border-border text-muted-foreground text-xs">
              Tidak ada foto temuan awal
            </div>
          )}

          {currentPhoto?.keterangan && (
            <div className="bg-amber-500/10 border border-amber-500/30 rounded-xl p-3 text-xs text-amber-700 dark:text-amber-300 flex items-start gap-2">
              <div>
                <span className="font-bold block text-[11px]">Keterangan Foto Audit:</span>
                <span>{currentPhoto.keterangan}</span>
              </div>
            </div>
          )}
        </Card>
      </div>

      {/* Section Bottom: Upload & List Foto Follow-Up (Perbaikan) */}
      <Card className="p-6 bg-card/60 backdrop-blur-md shadow-sm border-border/80 space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border pb-4">
          <div>
            <h3 className="font-bold text-lg flex items-center gap-2">
              Dokumentasi Foto Follow-Up Perbaikan
            </h3>
            <p className="text-xs text-muted-foreground mt-0.5">
              Unggah bukti perbaikan fisik beserta keterangan lengkap oleh PIC penanggung jawab.
            </p>
          </div>

          {!isClosed && (
            <div className="flex items-center gap-2">
              <Button
                size="sm"
                onClick={() => fileInputRef.current?.click()}
                disabled={uploadMutation.isPending || !canUploadFollowUp}
                className="bg-green-600 hover:bg-green-700 text-white font-semibold shadow-sm rounded-xl px-4"
              >
                {uploadMutation.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" /> Mengunggah...
                  </>
                ) : (
                  <>
                    <Upload className="mr-2 h-4 w-4" /> Upload Foto Follow-Up
                  </>
                )}
              </Button>

              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                className="hidden"
                onChange={handleFileChange}
                disabled={uploadMutation.isPending || !canUploadFollowUp}
              />
            </div>
          )}
        </div>

        {/* Upload Progress Bar */}
        {uploadMutation.isPending && uploadProgress > 0 && (
          <div className="w-full bg-muted rounded-full h-2.5 overflow-hidden">
            <div
              className="bg-green-600 h-2.5 rounded-full transition-all duration-300"
              style={{ width: `${uploadProgress}%` }}
            />
            <p className="text-[10px] text-right mt-1 text-muted-foreground">{uploadProgress}%</p>
          </div>
        )}

        {/* List Foto Follow-up */}
        {followUpPhotos.length === 0 ? (
          <div className="p-12 text-center text-muted-foreground rounded-2xl border border-dashed border-border bg-muted/20 flex flex-col items-center gap-2">
            <p className="font-medium text-sm">Belum ada foto bukti follow-up perbaikan.</p>
            <p className="text-xs text-muted-foreground">Klik tombol di atas untuk mengunggah bukti perbaikan baru.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-5">
            {followUpPhotos.map((photo) => (
              <FollowUpPhotoCard
                key={photo.issue_photo_id}
                photo={photo}
                onPreview={() => setSelectedImage(formatImageUrl(photo.image_url))}
                onDelete={() => deleteMutation.mutate(photo.issue_photo_id)}
                isDeleting={deleteMutation.isPending}
                isClosed={isClosed}
                canDelete={isAuditor || canUploadFollowUp}
                onRefresh={() => queryClient.invalidateQueries({ queryKey: ["issue-photos", id] })}
              />
            ))}
          </div>
        )}
      </Card>

      {/* Image Lightbox */}
      {selectedImage && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/90 backdrop-blur-sm p-4"
          onClick={() => setSelectedImage(null)}
        >
          <div className="relative max-w-4xl w-full" onClick={(e) => e.stopPropagation()}>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setSelectedImage(null)}
              className="absolute -top-10 right-0 text-white hover:text-zinc-300"
            >
              <X className="h-7 w-7" />
            </Button>
            <img
              src={selectedImage}
              alt="Preview"
              className="rounded-2xl max-h-[85vh] w-full object-contain"
            />
          </div>
        </div>
      )}
    </div>
  );
}

function FollowUpPhotoCard({
  photo,
  onPreview,
  onDelete,
  isDeleting,
  isClosed,
  canDelete,
  onRefresh,
}: {
  photo: IssuePhoto;
  onPreview: () => void;
  onDelete: () => void;
  isDeleting: boolean;
  isClosed: boolean;
  canDelete: boolean;
  onRefresh: () => void;
}) {
  const [isEditing, setIsEditing] = useState(false);
  const [keteranganText, setKeteranganText] = useState(photo.keterangan || "");
  const [isSaving, setIsSaving] = useState(false);

  const handleSave = async () => {
    if (!keteranganText.trim()) {
      toast.error("Keterangan foto tidak boleh kosong");
      return;
    }
    try {
      setIsSaving(true);
      await issueApi.updatePhoto(photo.issue_photo_id, keteranganText);
      toast.success("Keterangan foto berhasil diperbarui");
      setIsEditing(false);
      onRefresh();
    } catch {
      toast.error("Gagal memperbarui keterangan foto");
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="group relative flex flex-col rounded-2xl border border-border bg-card overflow-hidden shadow-xs hover:border-primary/50 transition-all">
      {/* Thumbnail */}
      <div className="relative aspect-square w-full overflow-hidden bg-muted">
        <img
          src={formatImageUrl(photo.image_url)}
          alt="Foto Follow up"
          className="h-full w-full object-cover transition-transform group-hover:scale-105"
        />
        <div className="absolute inset-0 flex items-center justify-center gap-2 bg-black/50 opacity-0 transition-opacity group-hover:opacity-100">
          <Button
            size="icon"
            variant="outline"
            className="h-8 w-8 bg-black/60 border-white/20 text-white hover:bg-white hover:text-black"
            onClick={onPreview}
          >
            <Eye className="h-4 w-4" />
          </Button>

          {!isClosed && (
            <Button
              size="icon"
              variant="outline"
              className="h-8 w-8 bg-black/60 border-white/20 text-white hover:bg-primary hover:text-white"
              onClick={() => setIsEditing(!isEditing)}
            >
              <Edit className="h-4 w-4" />
            </Button>
          )}

          {!isClosed && canDelete && (
            <Button
              size="icon"
              variant="destructive"
              className="h-8 w-8 shadow-md"
              onClick={onDelete}
              disabled={isDeleting}
            >
              {isDeleting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Trash2 className="h-4 w-4" />}
            </Button>
          )}
        </div>
      </div>

      {/* Info Content */}
      <div className="p-3 bg-card border-t border-border/60 text-xs flex flex-col justify-between grow space-y-3">
        {isEditing && !isClosed ? (
          <div className="space-y-2">
            <Input
              value={keteranganText}
              onChange={(e) => setKeteranganText(e.target.value)}
              placeholder="Tulis keterangan foto..."
              className="text-xs h-8"
              autoFocus
            />
            <div className="flex justify-end gap-1.5">
              <Button size="sm" variant="ghost" className="h-7 text-[11px] px-2" onClick={() => setIsEditing(false)}>
                <X className="h-3 w-3 mr-1" /> Batal
              </Button>
              <Button size="sm" className="h-7 text-[11px] px-2 bg-primary" onClick={handleSave} disabled={isSaving}>
                {isSaving ? <Loader2 className="h-3 w-3 animate-spin mr-1" /> : <Check className="h-3 w-3 mr-1" />} Simpan
              </Button>
            </div>
          </div>
        ) : (
          <p className="text-muted-foreground text-xs leading-tight line-clamp-2">
            {photo.keterangan ? (
              <span className="text-foreground font-medium">{photo.keterangan}</span>
            ) : (
              <span className="italic text-muted-foreground/70">Tidak ada keterangan foto</span>
            )}
          </p>
        )}

        {/* Meta Info: PIC Penginput & Tanggal */}
        <div className="pt-2 border-t border-border/40 flex items-center justify-between text-[10px] text-muted-foreground">
          <span className="font-semibold flex items-center gap-1 truncate max-w-[125px] text-foreground/80" title={photo.uploader_name || photo.pic_user_id}>
            <span className="truncate">{photo.uploader_name || photo.pic_user_id || "Pengguna"}</span>
          </span>
          {(photo.created_at || photo.follow_up_date) && (
            <span className="flex items-center gap-1 shrink-0 font-mono text-[10px]">
              {new Date(photo.created_at || photo.follow_up_date!).toLocaleDateString("id-ID", {
                day: "numeric",
                month: "short",
                hour: "2-digit",
                minute: "2-digit",
              })}
            </span>
          )}
        </div>
      </div>
    </div>
  );
}
