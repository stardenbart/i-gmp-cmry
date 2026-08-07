"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft, Trash2, CheckCircle2,
  Loader2, XCircle, CircleDashed
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { issueApi, IssuePhoto, IssueStatus } from "@/lib/api/issue.api";
import { cn, formatImageUrl } from "@/lib/utils";
import { InfoCard } from "@/components/isssues/InfoCard";
import { PhotoSection } from "@/components/isssues/PhotosCard";
import { WOWRCard } from "@/components/isssues/WOWRCard";
import { useChunkedUpload } from "@/hooks/useChunkedUpload";
import { useAuthStore } from "@/stores/authStore";
import { usePermissions } from "@/lib/usePermissions";

const statusConfig: Record<IssueStatus, { label: string; icon: React.ElementType; color: string; bg: string }> = {
  Open: { label: "Open", icon: CircleDashed, color: "text-blue-500", bg: "bg-blue-500/10" },
  InProgress: { label: "In Progress", icon: Loader2, color: "text-amber-500", bg: "bg-amber-500/10" },
  PendingValidation: { label: "Pending Validation", icon: Loader2, color: "text-purple-500", bg: "bg-purple-500/10" },
  Closed: { label: "Closed", icon: CheckCircle2, color: "text-green-500", bg: "bg-green-500/10" },
  Verified: { label: "Verified", icon: CheckCircle2, color: "text-emerald-600", bg: "bg-emerald-500/10" },
  Overdue: { label: "Overdue", icon: CircleDashed, color: "text-red-500", bg: "bg-red-500/10" },
  OpenOverdue: { label: "Open Overdue", icon: CircleDashed, color: "text-red-500", bg: "bg-red-500/10" },
  ClosedOverdue: { label: "Closed Overdue", icon: CheckCircle2, color: "text-orange-500", bg: "bg-orange-500/10" },
};

const STATUS_TRANSITIONS: Record<IssueStatus, { label: string; next: IssueStatus; color: string }[]> = {
  Open: [{ label: "Mulai Kerjakan", next: "InProgress", color: "bg-amber-500 hover:bg-amber-600" }],
  InProgress: [
    { label: "Selesaikan Temuan", next: "Closed", color: "bg-green-600 hover:bg-green-700" }
  ],
  PendingValidation: [
    { label: "Selesaikan Temuan", next: "Closed", color: "bg-green-600 hover:bg-green-700" }
  ],
  Closed: [],
  Verified: [],
  Overdue: [{ label: "Mulai Kerjakan", next: "InProgress", color: "bg-amber-500 hover:bg-amber-600" }],
  OpenOverdue: [{ label: "Mulai Kerjakan", next: "InProgress", color: "bg-amber-500 hover:bg-amber-600" }],
  ClosedOverdue: [],
};

import { useRouter } from "next/navigation";

export default function IssueDetailPage() {
  const { id, plantCode, userId } = useParams() as { id: string; plantCode?: string; userId?: string };
  const router = useRouter();
  const queryClient = useQueryClient();
  const [selectedImage, setSelectedImage] = useState<string | null>(null);
  
  const user = useAuthStore((state) => state.user);
  const { hasPermission } = usePermissions();

  const { uploadMutation, uploadProgress } = useChunkedUpload({ issueId: id });

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
    onError: (err: any) => {
      const msg = err?.response?.data?.message || err?.response?.data?.error || "Gagal memperbarui status";
      toast.error(msg);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (photoId: string) => issueApi.deletePhoto(id, photoId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issue-photos", id] });
      toast.success("Foto dihapus");
    },
    onError: (err: any) => {
      const msg = err?.response?.data?.message || err?.response?.data?.error || "Gagal menghapus foto";
      toast.error(msg);
    },
  });

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

  // Dynamic role calculation
  const isAuditor = hasPermission("PERM-INSP-C") || hasPermission("PERM-WOWR-U");
  const isPIC = !isAuditor;

  // Status calculation
  const rawStatus = issue.issue_status as IssueStatus;
  const currentStatus = (issue.computed_status || issue.issue_status) as IssueStatus;
  const isPendingValidation = rawStatus === "PendingValidation" || currentStatus === "PendingValidation";

  const config = statusConfig[currentStatus] || statusConfig[rawStatus] || statusConfig["Open"];
  const StatusIcon = config.icon;
  let transitions = STATUS_TRANSITIONS[currentStatus] || STATUS_TRANSITIONS[rawStatus] || [];

  // Filter transition buttons based on role (PIC vs Auditor)
  if (isPIC) {
    if (rawStatus === "Open" || rawStatus === "Overdue" || rawStatus === "OpenOverdue") {
      transitions = transitions.filter(t => t.next === "InProgress");
    } else if (rawStatus === "InProgress") {
      transitions = transitions.filter(t => t.next === "Closed");
    } else {
      // Closed, Verified: Auditee cannot perform status changes
      transitions = [];
    }
  } else if (isPendingValidation) {
    // For Auditor during PendingValidation, approval/rejection buttons are displayed in the dedicated purple banner below
    transitions = [];
  }

  // Hak akses edit untuk WO/WR & Foto FollowUp
  const isClosed = rawStatus === "Closed" || rawStatus === "Verified" || currentStatus === "Closed" || currentStatus === "Verified";
  const isWorkStarted = rawStatus === "InProgress" || rawStatus === "PendingValidation";
  const canAuditeeUploadFollowUp = !isPIC || isWorkStarted;
  const canEditWOWR = !isClosed && (isAuditor || isWorkStarted);

  const dueDate = issue.due_date ? new Date(issue.due_date) : null;

  const handleStatusTransition = (nextStatus: IssueStatus) => {
    // Validasi khusus untuk PIC/Auditee saat menyelesaikan temuan
    if (isPIC && (nextStatus === "Closed" || nextStatus === "PendingValidation")) {
      if (issue.needs_wo_wr || issue.wo_id || issue.wr_id) {
        if (!issue.wo_id && !issue.wr_id) {
          toast.error("Gagal: Anda harus menginput Nomor WO / WR dan menyimpannya terlebih dahulu.");
          return;
        }
        if (issue.wowr_status !== "Verified") {
          toast.error("Gagal: Temuan ini menggunakan WO/WR. Harap tunggu persetujuan (konfirmasi) WO/WR oleh Auditor terlebih dahulu.");
          return;
        }
        if (followUpPhotos.length === 0) {
          toast.error("Gagal: Anda harus mengunggah setidaknya 1 bukti Foto Follow-Up perbaikan WO/WR terlebih dahulu.");
          return;
        }
      } else {
        if (followUpPhotos.length === 0) {
          toast.error("Gagal: Anda harus mengunggah bukti Foto Follow-Up penyelesaian temuan terlebih dahulu.");
          return;
        }
      }
    }

    updateMutation.mutate(nextStatus);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <Link href="../issues">
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
                className={cn("text-white font-medium shadow-sm", t.color)}
                isLoading={updateMutation.isPending}
                onClick={() => handleStatusTransition(t.next)}
              >
                {t.label}
              </Button>
            ))}
          </div>
        )}
      </div>

      {/* ── Auditor: Pending Validation Banner ─────────────────────────── */}
      {!isPIC && isPendingValidation && (
        <div className="p-5 rounded-2xl bg-purple-500/10 border border-purple-500/40 flex flex-col sm:flex-row sm:items-center justify-between gap-4 shadow-sm">
          <div className="space-y-1">
            <h4 className="font-bold text-sm text-purple-700 dark:text-purple-300 flex items-center gap-2">
              <span className="relative flex h-2.5 w-2.5">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-purple-400 opacity-75" />
                <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-purple-500" />
              </span>
              Temuan Ini Membutuhkan Validasi Anda
            </h4>
            <p className="text-xs text-muted-foreground leading-relaxed">
              Auditee telah mengajukan bukti perbaikan. Tinjau foto &amp; informasi di bawah lalu pilih tindakan:
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2 shrink-0">
            <Button
              className="bg-red-600 hover:bg-red-700 text-white font-semibold shadow-sm"
              isLoading={updateMutation.isPending}
              onClick={() => updateMutation.mutate("InProgress")}
            >
              ✕ Tolak — Minta Perbaikan Ulang
            </Button>
            <Button
              className="bg-green-600 hover:bg-green-700 text-white font-bold shadow-md"
              isLoading={updateMutation.isPending}
              onClick={() => updateMutation.mutate("Closed")}
            >
              ✓ Setujui — Tutup Temuan
            </Button>
          </div>
        </div>
      )}

      {/* ── PIC/Auditee: Waiting for Validation Banner ───────────────── */}
      {isPIC && isPendingValidation && (
        <div className="p-4 rounded-xl bg-purple-500/10 border border-purple-500/30 text-purple-600 dark:text-purple-300">
          <strong className="font-semibold block mb-0.5 text-sm">⏳ Sedang Menunggu Validasi Auditor</strong>
          <p className="text-xs">Anda telah mengajukan bukti perbaikan. Saat ini temuan sedang ditinjau oleh Auditor.</p>
        </div>
      )}

      {/* ── PIC/Auditee: Work Not Started Warning ────────────────────── */}
      {isPIC && (rawStatus === "Open" || rawStatus === "Overdue" || rawStatus === "OpenOverdue") && (
        <div className="p-4 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-600 dark:text-amber-400">
          <p className="text-sm font-medium">
            Silakan klik tombol <strong>&quot;Mulai Kerjakan&quot;</strong> di kanan atas terlebih dahulu untuk mengaktifkan pengisian WO/WR dan pengunggahan Foto Follow-Up.
          </p>
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-1">
          {/* Info Card */}
          <InfoCard issue={issue} dueDate={dueDate} />

          {/* WO / WR Form Card */}
          <WOWRCard issue={issue} isAuditor={isAuditor} canEdit={canEditWOWR} />
        </div>

        {/* Photos Card */}
        <PhotoSection 
          initialPhotos={initialPhotos}
          followUpPhotos={followUpPhotos}
          uploadMutation={uploadMutation}
          deleteMutation={deleteMutation}
          setSelectedImage={setSelectedImage}
          onNavigateDetail={(photoId) =>
            router.push(`/cimory/${plantCode || "all"}/dashboard/${userId || user?.id}/issues/${id}/photos/${photoId}`)
          }
          uploadProgress={uploadProgress}
          isAuditor={isAuditor}
          canUploadFollowUp={canAuditeeUploadFollowUp}
          isClosed={rawStatus === "Closed" || rawStatus === "Verified" || currentStatus === "Closed" || currentStatus === "Verified"}
          onRefresh={() => queryClient.invalidateQueries({ queryKey: ["issue-photos", id] })}
        />
      </div>

      {/* Image Lightbox */}
      {selectedImage && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/90 backdrop-blur-sm p-4"
          onClick={() => setSelectedImage(null)}
        >
          <div className="relative max-w-3xl w-full" onClick={(e) => e.stopPropagation()}>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setSelectedImage(null)}
              className="absolute -top-10 right-0 text-white hover:text-zinc-300"
            >
              <XCircle className="h-7 w-7" />
            </Button>
            <img
              src={formatImageUrl(selectedImage) || "/placeholder.png"}
              alt="Preview"
              onError={(e) => { e.currentTarget.src = "/placeholder.png"; }}
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
        src={formatImageUrl(photo.image_url) || "/placeholder.png"}
        alt={photo.file_name}
        onError={(e) => { e.currentTarget.src = "/placeholder.png"; }}
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
      <div className="absolute bottom-0 left-0 right-0 p-2 bg-linear-to-t from-black/60 to-transparent">
        <p className="text-white text-[10px] truncate">{photo.file_name}</p>
      </div>
    </div>
  );
}
