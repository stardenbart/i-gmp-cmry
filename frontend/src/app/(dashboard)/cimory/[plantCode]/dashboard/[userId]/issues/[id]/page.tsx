"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams, useSearchParams } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft, CheckCircle2,
  Loader2, XCircle, CircleDashed,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { issueApi, Issue, IssuePhoto, IssueStatus } from "@/lib/api/issue.api";
import { cn, formatImageUrl } from "@/lib/utils";
import { InfoCard } from "@/components/isssues/InfoCard";
import { PhotoSection } from "@/components/isssues/PhotosCard";
import { useAuthStore } from "@/stores/authStore";
import { usePermissions } from "@/lib/usePermissions";
import { usePolling } from "@/hooks/usePolling";
import { filterApi } from "@/lib/api/filter.api";

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
    { label: "Ajukan Validasi", next: "PendingValidation", color: "bg-purple-600 hover:bg-purple-700" }
  ],
  PendingValidation: [],
  Closed: [],
  Verified: [],
  Overdue: [{ label: "Mulai Kerjakan", next: "InProgress", color: "bg-amber-500 hover:bg-amber-600" }],
  OpenOverdue: [{ label: "Mulai Kerjakan", next: "InProgress", color: "bg-amber-500 hover:bg-amber-600" }],
  ClosedOverdue: [],
};

function getMutationErrorMessage(error: unknown, fallback: string): string {
  if (typeof error === "object" && error !== null && "response" in error) {
    const response = (error as { response?: { data?: { message?: string; error?: string } } }).response;
    return response?.data?.message || response?.data?.error || fallback;
  }
  return fallback;
}

import { useRouter } from "next/navigation";

export default function IssueDetailPage() {
  const { id, plantCode, userId } = useParams() as { id: string; plantCode?: string; userId?: string };
  const searchParams = useSearchParams();
  const detailKawasanId = searchParams.get("detail_kawasan_id") || "";
  const router = useRouter();
  const queryClient = useQueryClient();
  const [selectedImage, setSelectedImage] = useState<string | null>(null);
  
  usePolling();

  const user = useAuthStore((state) => state.user);
  const { hasPermission } = usePermissions();

  const { data, isLoading } = useQuery({
    queryKey: ["issue", id],
    queryFn: () => issueApi.getById(id),
    staleTime: 0,
    refetchOnMount: "always",
    refetchInterval: 5000,
  });

  const issue = data?.data;
  const resolvedDetailKawasanId = detailKawasanId || issue?.detail_kawasan_id || "";

  const { data: photosData } = useQuery({
    queryKey: ["issue-photos", id],
    queryFn: () => issueApi.getPhotos(id),
    staleTime: 0,
    refetchOnMount: "always",
    refetchInterval: 5000,
  });

  const { data: locationIssuesData, isLoading: isLocationIssuesLoading } = useQuery({
    queryKey: ["issues-location", resolvedDetailKawasanId],
    queryFn: () => filterApi.issues({
      detail_kawasan_id: resolvedDetailKawasanId,
      page: 1,
      limit: 100,
      sort_by: "created_at",
      sort_order: "desc",
    }),
    enabled: Boolean(resolvedDetailKawasanId),
    staleTime: 0,
    refetchOnMount: "always",
    refetchInterval: 5000,
  });

  const updateMutation = useMutation({
    mutationFn: (status: IssueStatus) => issueApi.update(id, { issue_status: status }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issue", id] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      toast.success("Status temuan diperbarui");
    },
    onError: (error: unknown) => {
      toast.error(getMutationErrorMessage(error, "Gagal memperbarui status"));
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (photoId: string) => issueApi.deletePhoto(id, photoId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issue-photos", id] });
      toast.success("Foto dihapus");
    },
    onError: (error: unknown) => {
      toast.error(getMutationErrorMessage(error, "Gagal menghapus foto"));
    },
  });

  const locationIssues: Issue[] = resolvedDetailKawasanId
    ? (locationIssuesData?.items || [])
    : (issue ? [issue] : []);
  const photos: IssuePhoto[] = photosData?.data || [];
  const initialPhotos = photos.filter((p) => p.photo_type === "Initial");
  const followUpPhotos = photos.filter((p) => p.photo_type === "FollowUp");
  const wowrPhotos = photos.filter((p) => p.photo_type === "WOWR");

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
      transitions = transitions.filter(t => t.next === "PendingValidation");
    } else {
      // Closed, Verified: Auditee cannot perform status changes
      transitions = [];
    }
  } else if (isPendingValidation) {
    // For Auditor during PendingValidation, approval/rejection buttons are displayed in the dedicated purple banner below
    transitions = [];
  }

  const dueDate = issue.due_date ? new Date(issue.due_date) : null;

  const handleStatusTransition = (nextStatus: IssueStatus) => {
    // Validasi universal penyelesaian temuan (Closed, Verified, PendingValidation)
    if (nextStatus === "Closed" || nextStatus === "PendingValidation" || nextStatus === "Verified") {
      const initialCount = initialPhotos.length;
      const allProofPhotos = [...followUpPhotos, ...wowrPhotos];
      const proofCount = allProofPhotos.length;

      if (initialCount > 0) {
        const initialPhotosWithProof = initialPhotos.filter((initPhoto) =>
          allProofPhotos.some((fu) => fu.ref_photo_id === initPhoto.issue_photo_id)
        );
        const missingProofCount = initialCount - initialPhotosWithProof.length;

        if (proofCount < initialCount || missingProofCount > 0) {
          toast.error(
            `Gagal: Terdapat ${initialCount} foto bukti temuan awal, namun baru ${proofCount} foto perbaikan yang diunggah. Seluruh ${initialCount} foto temuan awal wajib memiliki bukti foto perbaikan (Follow-Up / WO-WR).`
          );
          return;
        }
      } else if (proofCount === 0) {
        toast.error("Gagal: Anda harus mengunggah setidaknya 1 bukti Foto Perbaikan terlebih dahulu sebelum menyelesaikan temuan.");
        return;
      }

      const initialPhotosUsingWOWR = initialPhotos.filter(
        (photo) => photo.needs_wo_wr || photo.wo_id || photo.wr_id
      );
      const initialPhotoWithoutWOWRNumber = initialPhotosUsingWOWR.find(
        (photo) => !photo.wo_id && !photo.wr_id
      );
      if (initialPhotoWithoutWOWRNumber) {
        toast.error("Gagal: Foto temuan menggunakan WO/WR tetapi Nomor WO/WR belum diisi.");
        return;
      }

      const isFinalApproval = nextStatus === "Closed" || nextStatus === "Verified";
      const unverifiedWOWRPhoto = initialPhotosUsingWOWR.find(
        (photo) => photo.wowr_status !== "Verified"
      );
      if (isFinalApproval && unverifiedWOWRPhoto) {
        toast.error("Gagal menutup temuan: masih ada WO/WR foto yang belum diverifikasi Auditor/Admin.");
        return;
      }

      if (issue.needs_wo_wr || issue.wo_id || issue.wr_id) {
        if (!issue.wo_id && !issue.wr_id) {
          toast.error("Gagal: Temuan ini menggunakan WO/WR. Harap input Nomor WO atau WR dan simpan terlebih dahulu.");
          return;
        }
        if (isFinalApproval && issue.wowr_status !== "Verified") {
          toast.error("Gagal: Temuan ini menggunakan WO/WR. Harap tunggu persetujuan (konfirmasi) WO/WR oleh Auditor terlebih dahulu.");
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
              onClick={() => handleStatusTransition("Closed")}
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

      {/* Singleton location card expands into every independent NG finding. */}
      <Card className="p-5 bg-card/60 backdrop-blur-md border-border/80 shadow-sm space-y-4">
          <div className="flex items-center justify-between gap-3 border-b border-border/60 pb-3">
            <div>
              <h3 className="font-bold text-base">Semua Temuan di Lokasi Ini</h3>
              <p className="text-xs text-muted-foreground mt-0.5">
                {issue.area_name || "Tanpa Area"} · {issue.kawasan_name || "Tanpa Kawasan"} · {issue.detail_kawasan_name || "Tanpa Detail Kawasan"}
              </p>
            </div>
            <span className="shrink-0 rounded-full border border-primary/30 bg-primary/10 px-3 py-1 text-xs font-bold text-primary">
              {locationIssues.length} temuan
            </span>
          </div>

          {isLocationIssuesLoading ? (
            <div className="flex justify-center py-5">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <div className="grid gap-2 sm:grid-cols-2">
              {locationIssues.map((locationIssue, index) => {
                const locationStatus = locationIssue.computed_status || locationIssue.issue_status;
                const locationConfig = statusConfig[locationStatus] || statusConfig.Open;
                const isSelected = locationIssue.issue_id === id;
                return (
                  <Link
                    key={locationIssue.issue_id}
                    href={`/cimory/${plantCode || "all"}/dashboard/${userId || user?.id}/issues/${locationIssue.issue_id}${resolvedDetailKawasanId ? `?detail_kawasan_id=${encodeURIComponent(resolvedDetailKawasanId)}` : ""}`}
                    className={cn(
                      "rounded-xl border p-3 transition-colors",
                      isSelected
                        ? "border-primary bg-primary/10"
                        : "border-border/70 bg-background/50 hover:border-primary/40 hover:bg-muted/40"
                    )}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <span className="text-[10px] font-mono text-muted-foreground">Temuan {index + 1}</span>
                      <span className={cn("text-[10px] font-bold", locationConfig.color)}>{locationConfig.label}</span>
                    </div>
                    <p className="mt-1.5 line-clamp-2 text-xs font-semibold leading-relaxed text-foreground">
                      {locationIssue.detail_aspek_name || "Tanpa uraian temuan"}
                    </p>
                    <p className="mt-2 text-[10px] text-muted-foreground">
                      {locationIssue.photos?.length || 0} foto · {locationIssue.issue_id}
                    </p>
                  </Link>
                );
              })}
            </div>
          )}
      </Card>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-1">
          {/* Information Card */}
          <InfoCard issue={issue} dueDate={dueDate} />
        </div>

        <div className="space-y-6 lg:col-span-2">
          {/* Photos Card */}
          <PhotoSection
            initialPhotos={initialPhotos}
            proofPhotos={[...followUpPhotos, ...wowrPhotos]}
            deleteMutation={deleteMutation}
            setSelectedImage={setSelectedImage}
            onNavigateDetail={(photoId) =>
              router.push(`/cimory/${plantCode || "all"}/dashboard/${userId || user?.id}/issues/${id}/photos/${photoId}`)
            }
            isAuditor={isAuditor}
            isClosed={rawStatus === "Closed" || rawStatus === "Verified" || currentStatus === "Closed" || currentStatus === "Verified"}
            onRefresh={() => queryClient.invalidateQueries({ queryKey: ["issue-photos", id] })}
          />
        </div>
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
            {/* Dynamic authenticated upload URLs intentionally use a native image element. */}
            {/* eslint-disable-next-line @next/next/no-img-element */}
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

