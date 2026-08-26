import { useState } from "react";
import { ImageIcon, Eye, Trash2, Edit, Check, X, Loader2, User, Calendar, Tag } from "lucide-react";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import { Input } from "../ui/input";
import { cn, formatImageUrl } from "@/lib/utils";
import { issueApi } from "@/lib/api/issue.api";
import { toast } from "sonner";

// 1. Interface untuk tipe data foto
interface PhotoItem {
  issue_photo_id: string;
  ref_photo_id?: string;
  photo_type?: "Initial" | "FollowUp" | "WOWR";
  image_url: string;
  keterangan?: string;
  file_name?: string;
  pic_user_id?: string;
  uploader_name?: string;
  created_at?: string;
  follow_up_date?: string;
  hei_id?: string;
  hei_name?: string;
  hei_category?: string;
  habit_name?: string;
  equipment_name?: string;
  infrastructure_name?: string;
  habit_id?: string;
  equipment_id?: string;
  infrastructure_id?: string;
}

// 2. Props untuk komponen item foto satuan
interface PhotoCardProps {
  photo: PhotoItem;
  onPreview: () => void;
  onNavigateDetail?: () => void;
  onDelete: () => void;
  onUpdateSuccess?: () => void;
  isDeleting: boolean;
  isAuditor?: boolean;
  isInitialPhoto?: boolean;
  canDelete?: boolean;
  isClosed?: boolean;
  hasFollowUp?: boolean;
}

// Komponen Kecil: Menampilkan item foto satuan dengan keterangan, tombol Edit, dan tombol Delete
const PhotoCard = ({
  photo,
  onPreview,
  onNavigateDetail,
  onDelete,
  onUpdateSuccess,
  isDeleting,
  isAuditor = true,
  isInitialPhoto = false,
  canDelete = true,
  isClosed = false,
  hasFollowUp = false,
}: PhotoCardProps) => {
  const [isEditing, setIsEditing] = useState(false);
  const [keteranganText, setKeteranganText] = useState(photo.keterangan || "");
  const [isSaving, setIsSaving] = useState(false);

  const canEditDescription = !isClosed && (!isInitialPhoto || isAuditor);
  const canDeletePhoto = !isClosed && (!isInitialPhoto || isAuditor) && canDelete;

  const handleSaveKeterangan = async () => {
    if (!keteranganText.trim()) {
      toast.error("Keterangan foto tidak boleh kosong");
      return;
    }
    try {
      setIsSaving(true);
      await issueApi.updatePhoto(photo.issue_photo_id, keteranganText);
      toast.success("Keterangan foto berhasil diperbarui");
      setIsEditing(false);
      if (onUpdateSuccess) onUpdateSuccess();
    } catch {
      toast.error("Gagal memperbarui keterangan foto");
    } finally {
      setIsSaving(false);
    }
  };

  const handleImageClick = () => {
    if (isInitialPhoto && onNavigateDetail) {
      onNavigateDetail();
    } else {
      onPreview();
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      handleImageClick();
    }
  };

  return (
    <div className={cn(
      "group relative flex flex-col rounded-2xl border bg-card overflow-hidden shadow-xs transition-all focus-within:ring-2 focus-within:ring-primary/40 [content-visibility:auto] [contain-intrinsic-size:1px_280px]",
      isInitialPhoto && hasFollowUp
        ? "border-green-500/70 hover:border-green-500 ring-1 ring-green-500/10"
        : isInitialPhoto
          ? "border-red-500/70 hover:border-red-500 ring-1 ring-red-500/10"
          : "border-border hover:border-primary/50"
    )}>
      {/* Thumbnail Container with Keyboard Access & Performance Props */}
      <div
        className="relative aspect-square w-full overflow-hidden bg-muted cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary"
        onClick={handleImageClick}
        onKeyDown={handleKeyDown}
        tabIndex={0}
        role="button"
        aria-label={isInitialPhoto ? `Foto temuan: ${photo.keterangan || "Detail Spesifikasi Foto"}` : `Foto follow-up: ${photo.keterangan || "Lihat Full"}`}
      >
        {/* Dynamic authenticated upload URLs intentionally use a native image element. */}
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          src={formatImageUrl(photo.image_url) || "/placeholder.png"}
          alt={photo.keterangan || "Foto bukti temuan audit"}
          loading="lazy"
          decoding="async"
          onError={(e) => { e.currentTarget.src = "/placeholder.png"; }}
          className="h-full w-full object-cover transition-transform group-hover:scale-105"
        />
        {isInitialPhoto && (
          <span className={cn(
            "absolute left-2 top-2 z-10 rounded-full border px-2.5 py-1 text-[10px] font-bold shadow-sm backdrop-blur-sm",
            hasFollowUp
              ? "border-green-400/60 bg-green-600/90 text-white"
              : "border-red-400/60 bg-red-600/90 text-white"
          )}>
            {hasFollowUp ? "Sudah Follow Up" : "Belum Follow Up"}
          </span>
        )}
        {/* Overlay Menu Saat Hover / Focus (Selalu Terlihat di Mobile Touchscreen) */}
        <div className="absolute inset-0 flex items-center justify-center gap-2 bg-black/40 sm:bg-black/50 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 sm:group-focus-within:opacity-100 transition-opacity">
          <Button
            size="icon"
            variant="outline"
            className="h-9 w-9 min-h-[36px] min-w-[36px] bg-black/60 border-white/20 text-white hover:bg-white hover:text-black focus-visible:ring-2 focus-visible:ring-white"
            onClick={(e) => {
              e.stopPropagation();
              handleImageClick();
            }}
            aria-label={isInitialPhoto ? "Lihat Detail Spesifikasi Foto" : "Lihat Foto Full"}
            title={isInitialPhoto ? "Lihat Detail Spesifikasi Foto" : "Lihat Foto Full"}
          >
            <Eye className="h-4 w-4" />
          </Button>

          {/* Tombol Edit Keterangan */}
          {canEditDescription && (
            <Button
              size="icon"
              variant="outline"
              className="h-9 w-9 min-h-[36px] min-w-[36px] bg-black/60 border-white/20 text-white hover:bg-primary hover:text-white focus-visible:ring-2 focus-visible:ring-white"
              onClick={(e) => {
                e.stopPropagation();
                setIsEditing(!isEditing);
              }}
              aria-label="Edit Keterangan Foto"
              title="Edit Keterangan Foto"
            >
              <Edit className="h-4 w-4" />
            </Button>
          )}

          {/* Tombol Delete Foto */}
          {canDeletePhoto && (
            <Button
              size="icon"
              variant="destructive"
              className="h-9 w-9 min-h-[36px] min-w-[36px] shadow-md focus-visible:ring-2 focus-visible:ring-white"
              onClick={(e) => {
                e.stopPropagation();
                onDelete();
              }}
              disabled={isDeleting}
              aria-label="Hapus Foto Ini"
              title="Hapus Foto Ini"
            >
              {isDeleting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Trash2 className="h-4 w-4" />}
            </Button>
          )}
        </div>
      </div>

      {/* Keterangan, PIC Penginput & Tanggal Info */}
      <div className="p-3 bg-card border-t border-border/60 text-xs flex flex-col justify-between grow space-y-2">
        {isEditing && canEditDescription ? (
          <div className="space-y-2">
            <Input
              value={keteranganText}
              onChange={(e) => setKeteranganText(e.target.value)}
              placeholder="Tulis keterangan foto..."
              className="text-xs h-8"
              autoFocus
            />
            <div className="flex justify-end gap-1.5">
              <Button
                size="sm"
                variant="ghost"
                className="h-7 text-[11px] px-2"
                onClick={() => {
                  setIsEditing(false);
                  setKeteranganText(photo.keterangan || "");
                }}
                disabled={isSaving}
              >
                <X className="h-3 w-3 mr-1" /> Batal
              </Button>
              <Button
                size="sm"
                className="h-7 text-[11px] px-2 bg-primary text-primary-foreground"
                onClick={handleSaveKeterangan}
                disabled={isSaving}
              >
                {isSaving ? <Loader2 className="h-3 w-3 animate-spin mr-1" /> : <Check className="h-3 w-3 mr-1" />}
                Simpan
              </Button>
            </div>
          </div>
        ) : (
          <div className="flex items-start justify-between gap-2">
            <p className="text-muted-foreground text-xs leading-tight line-clamp-2">
              {photo.keterangan ? (
                <span className="text-foreground font-medium">{photo.keterangan}</span>
              ) : (
                <span className="italic text-muted-foreground/70">Tidak ada keterangan foto</span>
              )}
            </p>
            {canEditDescription && (
              <button
                onClick={() => setIsEditing(true)}
                className="text-primary hover:text-primary/80 shrink-0 p-1 rounded hover:bg-primary/10 transition-colors"
                title="Edit Keterangan"
              >
                <Edit className="h-3.5 w-3.5" />
              </button>
            )}
          </div>
        )}

        {/* Isolated HEI Badge per Photo */}
        {(photo.hei_name || photo.hei_category || photo.habit_name || photo.equipment_name || photo.infrastructure_name) && (
          <div className="flex flex-wrap gap-1 pt-1.5 border-t border-border/40">
            <span className="inline-flex items-center gap-1 text-[10px] font-semibold px-2 py-0.5 rounded-md bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20" title={`HEI: ${photo.hei_category ? `[${photo.hei_category}] ` : ""}${photo.hei_name || photo.habit_name || photo.equipment_name || photo.infrastructure_name || ""}`}>
              <Tag className="h-2.5 w-2.5 shrink-0" />
              <span className="truncate max-w-[150px]">
                {photo.hei_category && photo.hei_name
                  ? `[${photo.hei_category}] ${photo.hei_name}`
                  : photo.hei_category || photo.hei_name || photo.habit_name || photo.equipment_name || photo.infrastructure_name}
              </span>
            </span>
          </div>
        )}

        {/* Uploader PIC & Date Badge */}
        <div className="pt-2 border-t border-border/40 flex items-center justify-between text-[10px] text-muted-foreground">
          <span className="font-semibold flex items-center gap-1 truncate max-w-[125px] text-foreground/80" title={photo.uploader_name || photo.pic_user_id || "Pengguna"}>
            <User className="h-3 w-3 text-primary shrink-0" />
            <span className="truncate">{photo.uploader_name || photo.pic_user_id || "Pengguna"}</span>
          </span>
          {(photo.created_at || photo.follow_up_date) && (
            <span className="flex items-center gap-1 shrink-0 font-mono text-[10px]">
              <Calendar className="h-3 w-3 text-muted-foreground" />
              {new Date(photo.created_at || photo.follow_up_date!).toLocaleDateString("id-ID", {
                day: "numeric",
                month: "short",
              })}
            </span>
          )}
        </div>
      </div>
    </div>
  );
};

// 3. Props untuk komponen kontainer utama
interface PhotoSectionProps {
  initialPhotos: PhotoItem[];
  proofPhotos?: PhotoItem[];
  deleteMutation: { isPending: boolean; mutate: (id: string) => void };
  setSelectedImage: (url: string) => void;
  onNavigateDetail?: (photoId: string) => void;
  isAuditor?: boolean;
  isClosed?: boolean;
  onRefresh?: () => void;
}

// Komponen Utama: Kontainer Dokumentasi Foto
export const PhotoSection = ({
  initialPhotos = [],
  proofPhotos = [],
  deleteMutation,
  setSelectedImage,
  onNavigateDetail,
  isAuditor = true,
  isClosed = false,
  onRefresh,
}: PhotoSectionProps) => {
  const followedUpPhotoIds = new Set(
    proofPhotos.map((photo) => photo.ref_photo_id).filter((photoId): photoId is string => Boolean(photoId))
  );

  return (
    <Card className="p-6 bg-card/60 backdrop-blur-md lg:col-span-2 space-y-5 shadow-sm border-border/80">
      <div className="flex items-center justify-between border-b border-border pb-3">
        <h3 className="font-bold text-base">Dokumentasi Bukti Foto & Keterangan</h3>
      </div>

      {/* Initial Photos */}
      <div className="space-y-3">
        <p className="text-xs font-bold uppercase tracking-wider text-muted-foreground flex items-center justify-between">
          <span>Foto Bukti Temuan Awal ({initialPhotos.length})</span>
          <span className="text-label-sm font-normal normal-case text-muted-foreground">
            Klik foto untuk melihat detail spesifikasi lokasi &amp; follow-up
          </span>
        </p>

        {initialPhotos.length === 0 ? (
          <div className="flex items-center justify-center h-24 rounded-2xl border border-dashed border-border text-muted-foreground text-xs">
            <ImageIcon className="mr-2 h-4 w-4" /> Belum ada foto bukti temuan awal
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {initialPhotos.map((photo) => (
              <PhotoCard
                key={photo.issue_photo_id}
                photo={photo}
                onPreview={() => setSelectedImage(formatImageUrl(photo.image_url))}
                onNavigateDetail={() => onNavigateDetail?.(photo.issue_photo_id)}
                onDelete={() => deleteMutation.mutate(photo.issue_photo_id)}
                onUpdateSuccess={onRefresh}
                isDeleting={deleteMutation.isPending}
                isAuditor={isAuditor}
                isInitialPhoto={true}
                hasFollowUp={followedUpPhotoIds.has(photo.issue_photo_id)}
                isClosed={isClosed}
              />
            ))}
          </div>
        )}
      </div>

    </Card>
  );
};
