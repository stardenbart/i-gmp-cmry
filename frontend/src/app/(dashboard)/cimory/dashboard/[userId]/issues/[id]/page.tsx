"use client";

import { useState, useRef } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft, User, Calendar, FileText,
  Upload, Trash2, ImageIcon, CheckCircle2,
  Loader2, XCircle, AlertTriangle, Clock
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { issueApi, IssuePhoto, IssueStatus } from "@/lib/api/issue.api";
import { cn } from "@/lib/utils";

const statusConfig: Record<IssueStatus, { label: string; icon: React.ElementType; color: string; bg: string }> = {
  Open: { label: "Open", icon: AlertTriangle, color: "text-red-500", bg: "bg-red-500/10" },
  InProgress: { label: "In Progress", icon: Loader2, color: "text-blue-500", bg: "bg-blue-500/10" },
  Closed: { label: "Closed", icon: XCircle, color: "text-zinc-500", bg: "bg-zinc-500/10" },
  Verified: { label: "Verified", icon: CheckCircle2, color: "text-green-500", bg: "bg-green-500/10" },
};

const STATUS_TRANSITIONS: Record<IssueStatus, { label: string; next: IssueStatus; color: string }[]> = {
  Open: [{ label: "Mulai Proses", next: "InProgress", color: "bg-blue-600 hover:bg-blue-700" }],
  InProgress: [{ label: "Tutup Temuan", next: "Closed", color: "bg-zinc-600 hover:bg-zinc-700" }],
  Closed: [{ label: "Verifikasi", next: "Verified", color: "bg-green-600 hover:bg-green-700" }],
  Verified: [],
};

export default function IssueDetailPage() {
  const { id } = useParams() as { id: string };
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [photoType, setPhotoType] = useState<"Initial" | "FollowUp">("Initial");
  const [selectedImage, setSelectedImage] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["issue", id],
    queryFn: () => issueApi.getById(id),
  });

  const { data: photosData } = useQuery({
    queryKey: ["issue-photos", id],
    queryFn: () => issueApi.getPhotos(id),
  });

  const updateMutation = useMutation({
    mutationFn: (status: IssueStatus) => issueApi.update(id, { issue_status: status }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issue", id] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      toast.success("Status temuan diperbarui");
    },
    onError: () => toast.error("Gagal memperbarui status"),
  });

  const uploadMutation = useMutation({
    mutationFn: (file: File) => {
      const formData = new FormData();
      formData.append("photo", file);
      formData.append("photo_type", photoType);
      return issueApi.uploadPhoto(id, formData);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issue-photos", id] });
      toast.success("Foto berhasil diunggah");
    },
    onError: () => toast.error("Gagal mengunggah foto"),
  });

  const deleteMutation = useMutation({
    mutationFn: (photoId: string) => issueApi.deletePhoto(id, photoId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issue-photos", id] });
      toast.success("Foto dihapus");
    },
    onError: () => toast.error("Gagal menghapus foto"),
  });

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    uploadMutation.mutate(file);
    e.target.value = "";
  };

  const issue = data?.data;
  const photos: IssuePhoto[] = photosData?.data || [];
  const initialPhotos = photos.filter((p) => p.photo_type === "Initial");
  const followUpPhotos = photos.filter((p) => p.photo_type === "FollowUp");

  if (isLoading) {
    return (
      <div className="flex justify-center p-10">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!issue) return <div className="text-center p-10 text-muted-foreground">Temuan tidak ditemukan.</div>;

  const config = statusConfig[issue.issue_status as IssueStatus];
  const StatusIcon = config.icon;
  const transitions = STATUS_TRANSITIONS[issue.issue_status as IssueStatus];
  const dueDate = issue.due_date ? new Date(issue.due_date) : null;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <Link href="../">
            <Button variant="ghost" size="icon" className="rounded-full">
              <ArrowLeft className="h-5 w-5" />
            </Button>
          </Link>
          <div>
            <div className="flex items-center gap-3">
              <h2 className="text-2xl font-bold tracking-tight">Detail Temuan</h2>
              <span className={cn("inline-flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-full", config.bg, config.color)}>
                <StatusIcon className="h-3.5 w-3.5" />
                {config.label}
              </span>
            </div>
            <p className="text-muted-foreground text-sm font-mono mt-1">{id}</p>
          </div>
        </div>

        {/* Status Action Buttons */}
        {transitions.length > 0 && (
          <div className="flex gap-2">
            {transitions.map((t: { label: string; next: IssueStatus; color: string }) => (
              <Button
                key={t.next}
                className={cn("text-white", t.color)}
                isLoading={updateMutation.isPending}
                onClick={() => updateMutation.mutate(t.next)}
              >
                {t.label}
              </Button>
            ))}
          </div>
        )}
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Info Card */}
        <Card className="p-6 bg-card/60 backdrop-blur-md lg:col-span-1 space-y-4 h-fit">
          <h3 className="font-semibold border-b border-border pb-2">Informasi Temuan</h3>
          
          <div className="space-y-3 text-sm">
            <div className="flex items-start gap-3">
              <FileText className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0" />
              <div>
                <span className="text-xs text-muted-foreground block">Keterangan</span>
                <span className="font-medium">{issue.keterangan || "–"}</span>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <User className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0" />
              <div>
                <span className="text-xs text-muted-foreground block">PIC (Penanggung Jawab)</span>
                <span className="font-medium">{issue.issue_pic_user_id}</span>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <Calendar className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0" />
              <div>
                <span className="text-xs text-muted-foreground block">Target Penyelesaian</span>
                {dueDate ? (
                  <span className={cn("font-medium", dueDate < new Date() && issue.issue_status !== "Closed" && issue.issue_status !== "Verified" ? "text-red-500" : "")}>
                    {dueDate.toLocaleDateString("id-ID", { day: "numeric", month: "long", year: "numeric" })}
                  </span>
                ) : (
                  <span className="text-muted-foreground">Tidak ditentukan</span>
                )}
              </div>
            </div>
            <div className="flex items-start gap-3">
              <Clock className="h-4 w-4 text-muted-foreground mt-0.5 shrink-0" />
              <div>
                <span className="text-xs text-muted-foreground block">Dibuat</span>
                <span className="font-medium">
                  {new Date(issue.created_at).toLocaleDateString("id-ID", { day: "numeric", month: "long", year: "numeric" })}
                </span>
              </div>
            </div>
          </div>
        </Card>

        {/* Photos Card */}
        <Card className="p-6 bg-card/60 backdrop-blur-md lg:col-span-2 space-y-5">
          <div className="flex items-center justify-between border-b border-border pb-2">
            <h3 className="font-semibold">Dokumentasi Foto</h3>
            <div className="flex items-center gap-2">
              <select
                value={photoType}
                onChange={(e) => setPhotoType(e.target.value as "Initial" | "FollowUp")}
                className="h-9 rounded-xl border border-border bg-card px-3 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary"
              >
                <option value="Initial">Temuan Awal</option>
                <option value="FollowUp">Follow-Up</option>
              </select>
              <Button
                size="sm"
                variant="outline"
                isLoading={uploadMutation.isPending}
                onClick={() => fileInputRef.current?.click()}
              >
                <Upload className="mr-1.5 h-3.5 w-3.5" /> Upload
              </Button>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                className="hidden"
                onChange={handleFileChange}
              />
            </div>
          </div>

          {/* Initial Photos */}
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3">
              📌 Foto Temuan Awal ({initialPhotos.length})
            </p>
            {initialPhotos.length === 0 ? (
              <div className="flex items-center justify-center h-24 rounded-2xl border border-dashed border-border text-muted-foreground text-sm">
                <ImageIcon className="mr-2 h-4 w-4" /> Belum ada foto
              </div>
            ) : (
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                {initialPhotos.map((photo) => (
                  <PhotoCard
                    key={photo.issue_photo_id}
                    photo={photo}
                    onPreview={() => setSelectedImage(photo.image_url)}
                    onDelete={() => deleteMutation.mutate(photo.issue_photo_id)}
                    isDeleting={deleteMutation.isPending}
                  />
                ))}
              </div>
            )}
          </div>

          {/* Follow-Up Photos */}
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3">
              ✅ Foto Follow-Up ({followUpPhotos.length})
            </p>
            {followUpPhotos.length === 0 ? (
              <div className="flex items-center justify-center h-24 rounded-2xl border border-dashed border-border text-muted-foreground text-sm">
                <ImageIcon className="mr-2 h-4 w-4" /> Belum ada foto follow-up
              </div>
            ) : (
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                {followUpPhotos.map((photo) => (
                  <PhotoCard
                    key={photo.issue_photo_id}
                    photo={photo}
                    onPreview={() => setSelectedImage(photo.image_url)}
                    onDelete={() => deleteMutation.mutate(photo.issue_photo_id)}
                    isDeleting={deleteMutation.isPending}
                  />
                ))}
              </div>
            )}
          </div>
        </Card>
      </div>

      {/* Image Lightbox */}
      {selectedImage && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/90 backdrop-blur-sm p-4"
          onClick={() => setSelectedImage(null)}
        >
          <div className="relative max-w-3xl w-full" onClick={(e) => e.stopPropagation()}>
            <button
              onClick={() => setSelectedImage(null)}
              className="absolute -top-10 right-0 text-white hover:text-zinc-300"
            >
              <XCircle className="h-7 w-7" />
            </button>
            <img
              src={selectedImage}
              alt="Preview"
              className="rounded-2xl max-h-[80vh] w-full object-contain"
            />
          </div>
        </div>
      )}
    </div>
  );
}

function PhotoCard({
  photo,
  onPreview,
  onDelete,
  isDeleting,
}: {
  photo: IssuePhoto;
  onPreview: () => void;
  onDelete: () => void;
  isDeleting: boolean;
}) {
  return (
    <div className="relative group overflow-hidden rounded-2xl border border-border bg-muted aspect-square">
      <img
        src={photo.image_url || "/placeholder.png"}
        alt={photo.file_name}
        className="h-full w-full object-cover transition-transform group-hover:scale-105 cursor-pointer"
        onClick={onPreview}
      />
      <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
        <button
          onClick={onDelete}
          disabled={isDeleting}
          className="flex h-8 w-8 items-center justify-center rounded-full bg-red-500/80 text-white hover:bg-red-500 transition-colors"
        >
          <Trash2 className="h-4 w-4" />
        </button>
      </div>
      <div className="absolute bottom-0 left-0 right-0 p-2 bg-gradient-to-t from-black/60 to-transparent">
        <p className="text-white text-[10px] truncate">{photo.file_name}</p>
      </div>
    </div>
  );
}
