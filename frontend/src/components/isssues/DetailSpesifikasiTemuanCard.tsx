import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";
import { Card } from "../ui/card";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { CheckCircle2, Loader2, XCircle } from "lucide-react";
import { Issue, IssuePhoto, issueApi, WOWRStatus } from "@/lib/api/issue.api";
import { cn } from "@/lib/utils";

type MutationError = {
  response?: { data?: { message?: string; error?: string } };
};

const getMutationErrorMessage = (error: unknown, fallback: string) => {
  const mutationError = error as MutationError;
  return mutationError.response?.data?.message || mutationError.response?.data?.error ||
    (error instanceof Error ? error.message : fallback);
};

interface DetailSpesifikasiTemuanCardProps {
  issue: Issue;
  photo?: IssuePhoto;
  photos?: IssuePhoto[];
  dueDate?: Date | null;
  isAuditor: boolean;
  canEditWOWR?: boolean;
  onRefresh?: () => void;
}

export const DetailSpesifikasiTemuanCard = ({
  issue,
  photo,
  photos = [],
  dueDate,
  isAuditor,
  canEditWOWR,
  onRefresh,
}: DetailSpesifikasiTemuanCardProps) => {
  const queryClient = useQueryClient();
  const activeTarget = photo || issue;
  const [type, setType] = useState<"WOWR" | "None">(
    activeTarget.needs_wo_wr || activeTarget.wo_id || activeTarget.wr_id ? "WOWR" : "None"
  );
  const [inputValue, setInputValue] = useState(activeTarget.wo_id || activeTarget.wr_id || "");
  const [validationError, setValidationError] = useState("");
  // `issue.issue_pic_user_id` is always the inspector who ran the audit
  // (see inspection_usecases.go BulkSave), never the Auditee/PIC who
  // actually uploads WO/WR evidence — comparing the photo's uploader
  // against it here always failed, permanently blocking Auditor/Admin
  // from approving even when a valid photo existed. photo_type "WOWR"
  // plus the ref_photo_id match already scope this to the right evidence.
  const hasWOWRProofImage = photos.some((proof) =>
    proof.photo_type === "WOWR" &&
    Boolean(proof.image_url) &&
    (photo?.issue_photo_id ? proof.ref_photo_id === photo.issue_photo_id : true)
  );

  const refreshQueries = () => {
    queryClient.invalidateQueries({ queryKey: ["issue", issue.issue_id] });
    queryClient.invalidateQueries({ queryKey: ["issue-photos", issue.issue_id] });
    queryClient.invalidateQueries({ queryKey: ["issues"] });
    onRefresh?.();
  };

  const wowrMutation = useMutation({
    mutationFn: (data: { needs_wo_wr: boolean; wo_id: string; wr_id: string; wowr_status?: WOWRStatus }) =>
      issueApi.update(issue.issue_id, data),
    onSuccess: () => {
      refreshQueries();
      toast.success("Data Maintenance WO/WR berhasil disimpan.");
    },
    onError: (error: unknown) =>
      toast.error(getMutationErrorMessage(error, "Gagal menyimpan data WO/WR.")),
  });

  const photoWowrMutation = useMutation({
    mutationFn: ({ photoId, data }: {
      photoId: string;
      data: { needs_wo_wr: boolean; wo_id: string; wr_id: string; wowr_status?: WOWRStatus };
    }) => issueApi.updatePhotoWOWR(photoId, data),
    onSuccess: () => {
      refreshQueries();
      toast.success("Data Maintenance WO/WR foto berhasil disimpan.");
    },
    onError: (error: unknown) =>
      toast.error(getMutationErrorMessage(error, "Gagal menyimpan data WO/WR foto.")),
  });

  const wowrValidationMutation = useMutation({
    mutationFn: (status: WOWRStatus) => {
      if (status === "Verified" && !hasWOWRProofImage) {
        return Promise.reject(new Error("Auditee belum mengunggah foto bukti penyelesaian WO/WR."));
      }
      if (photo?.issue_photo_id) {
        return issueApi.updatePhotoWOWR(photo.issue_photo_id, {
          needs_wo_wr: true,
          wo_id: photo.wo_id || "",
          wr_id: photo.wr_id || "",
          wowr_status: status,
        });
      }
      return issueApi.update(issue.issue_id, {
        wowr_status: status,
        ...(status === "Verified" ? { issue_status: "Closed" } : {}),
      });
    },
    onSuccess: (_, status) => {
      refreshQueries();
      toast.success(status === "Verified" ? "Bukti WO/WR berhasil diverifikasi." : "Bukti WO/WR ditolak.");
    },
    onError: (error: unknown) =>
      toast.error(getMutationErrorMessage(error, "Gagal memperbarui status validasi WO/WR.")),
  });

  const handleSave = () => {
    if (type === "WOWR" && !inputValue.trim()) {
      setValidationError("Nomor referensi WO / WR wajib diisi.");
      return;
    }

    setValidationError("");
    const data = {
      needs_wo_wr: type === "WOWR",
      wo_id: type === "WOWR" ? inputValue.trim().toUpperCase() : "",
      wr_id: "",
      wowr_status: (type === "WOWR" ? "PendingValidation" : "None") as WOWRStatus,
    };

    if (photo?.issue_photo_id) {
      photoWowrMutation.mutate({ photoId: photo.issue_photo_id, data });
    } else {
      wowrMutation.mutate(data);
    }
  };

  const isReadOnly = canEditWOWR !== undefined ? !canEditWOWR : false;
  const isOverdue =
    issue.computed_status === "OpenOverdue" ||
    issue.computed_status === "ClosedOverdue" ||
    (dueDate &&
      dueDate < new Date() &&
      issue.issue_status !== "Closed" &&
      issue.issue_status !== "Verified");

  const heiSources = photo ? [photo] : photos.length > 0 ? photos : issue.photos || [];
  const heiLabels = new Map<string, { category: string; name: string }>();

  const addHEI = (category?: string, name?: string) => {
    if (!category?.trim() && !name?.trim()) return;
    const cleanCategory = category?.trim() || "HEI";
    const cleanName = name?.trim() || cleanCategory;
    heiLabels.set(`${cleanCategory.toLowerCase()}::${cleanName.toLowerCase()}`, {
      category: cleanCategory,
      name: cleanName,
    });
  };

  heiSources.forEach((source) => {
    addHEI(source.hei_category, source.hei_name);
    if (source.habit_name) addHEI("Habit", source.habit_name);
    if (source.equipment_name) addHEI("Equipment", source.equipment_name);
    if (source.infrastructure_name) addHEI("Infrastructure", source.infrastructure_name);
  });

  if (heiLabels.size === 0) {
    addHEI(issue.hei_category, issue.hei_name);
    if (issue.hei?.habit?.habit_name || issue.habit_name) addHEI("Habit", issue.hei?.habit?.habit_name || issue.habit_name);
    if (issue.hei?.equipment?.equipment_name || issue.equipment_name) addHEI("Equipment", issue.hei?.equipment?.equipment_name || issue.equipment_name);
    if (issue.hei?.infrastructure?.infrastructure_name || issue.infrastructure_name) addHEI("Infrastructure", issue.hei?.infrastructure?.infrastructure_name || issue.infrastructure_name);
  }

  return (
    <Card className="p-6 bg-card/60 backdrop-blur-md space-y-6 border-border/80 shadow-sm rounded-2xl">
      {/* Header Card */}
      <div className="flex items-center justify-between border-b border-border pb-3">
        <div className="flex items-center gap-2">
          <h3 className="font-bold text-base tracking-tight">Detail Spesifikasi Temuan Awal</h3>
        </div>
      </div>

      {/* ── Section 1: Information Details ──────────────────────────── */}
      <div className="space-y-4 text-xs">
        {/* Lokasi Audit Area */}
        <div className="space-y-1.5">
          <span className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground flex items-center gap-1">
            Lokasi Audit Area
          </span>
          <div className="grid grid-cols-3 gap-2">
            <div className="bg-muted/40 p-2.5 rounded-xl border border-border/60">
              <span className="text-muted-foreground block text-[10px] font-semibold">Area</span>
              <span className="font-bold text-foreground text-xs">{issue.area_name || "Tanpa Area"}</span>
            </div>
            <div className="bg-muted/40 p-2.5 rounded-xl border border-border/60">
              <span className="text-muted-foreground block text-[10px] font-semibold">Kawasan</span>
              <span className="font-bold text-foreground text-xs">{issue.kawasan_name || "Tanpa Kawasan"}</span>
            </div>
            <div className="bg-muted/40 p-2.5 rounded-xl border border-border/60">
              <span className="text-muted-foreground block text-[10px] font-semibold">Detail Kawasan</span>
              <span className="font-bold text-foreground text-xs">{issue.detail_kawasan_name || "Tanpa Detail Kawasan"}</span>
            </div>
          </div>
        </div>

        {/* Aspek & Detail Aspek */}
        {(issue.aspek_name || issue.detail_aspek_name) && (
          <div className="grid grid-cols-2 gap-2">
            <div className="bg-muted/40 p-2.5 rounded-xl border border-border/60">
              <span className="text-muted-foreground block text-[10px] font-semibold">Aspek</span>
              <span className="font-semibold text-foreground">{issue.aspek_name || "-"}</span>
            </div>
            <div className="bg-muted/40 p-2.5 rounded-xl border border-border/60">
              <span className="text-muted-foreground block text-[10px] font-semibold">Detail Aspek</span>
              <span className="font-semibold text-foreground">{issue.detail_aspek_name || "-"}</span>
            </div>
          </div>
        )}

        {/* Uraian Checklist / Keterangan Temuan */}
        <div>
          <span className="text-muted-foreground block text-[11px] font-medium mb-1">
            Uraian Checklist / Keterangan Temuan
          </span>
          <div className="bg-background/80 p-3 rounded-xl border border-border/80 font-medium text-foreground text-xs leading-relaxed">
            {photo?.keterangan || issue.uraian_text || issue.keterangan || "Tidak ada rincian keterangan"}
          </div>
        </div>

        {/* Output Klasifikasi HEI */}
        {heiLabels.size > 0 && (
          <div className="pt-2 border-t border-border/60 space-y-1.5">
            <span className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground flex items-center gap-1">
             Kategori HEI (Hasil Inspeksi)
            </span>
            <div className="flex flex-wrap gap-2">
              {Array.from(heiLabels.values()).map(({ category, name }) => {
                const normalizedCategory = category.toLowerCase();
                const color = normalizedCategory === "habit"
                  ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20"
                  : normalizedCategory === "equipment"
                    ? "bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20"
                    : normalizedCategory === "infrastructure"
                      ? "bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/20"
                      : "bg-slate-500/10 text-slate-600 dark:text-slate-300 border-slate-500/20";
                return (
                  <span
                    key={`${category}-${name}`}
                    className={cn("px-2.5 py-1 rounded-xl font-semibold border text-[11px]", color)}
                  >
                    [{category}] {name !== category ? name : ""}
                  </span>
                );
              })}
            </div>
          </div>
        )}

        {/* Metadata PIC, Due Date, Created At */}
        <div className="grid grid-cols-2 gap-3 pt-3 border-t border-border/60 text-xs">
          <div className="flex items-start gap-2">
            <div>
              <span className="text-[11px] text-muted-foreground block font-medium">PIC Penanggung Jawab</span>
              <span className="font-semibold text-foreground">{issue.pic_name || issue.issue_pic_user_id}</span>
            </div>
          </div>

          <div className="flex items-start gap-2">
            <div>
              <span className="text-[11px] text-muted-foreground block font-medium">Target Penyelesaian</span>
              {dueDate ? (
                <span className={cn("font-semibold", isOverdue ? "text-red-500 font-bold" : "text-foreground")}>
                  {dueDate.toLocaleDateString("id-ID", { day: "numeric", month: "short", year: "numeric" })}
                </span>
              ) : (
                <span className="text-muted-foreground italic">Tidak ditentukan</span>
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="pt-4 border-t border-border space-y-4">
        <div className="flex items-center justify-between">
          <h4 className="font-bold text-sm">
            Maintenance (WO / WR) {photo ? "Terisolasi Foto" : ""}
          </h4>
          <span className={cn(
            "text-[10px] px-2 py-0.5 rounded-full font-semibold border",
            isReadOnly
              ? "bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20"
              : "bg-green-500/10 text-green-600 dark:text-green-400 border-green-500/20"
          )}>
            {isReadOnly ? "Terkunci" : "Dapat Diisi"}
          </span>
        </div>

        <div className="flex flex-col gap-2.5 text-xs">
          <label className={cn(
            "flex items-center gap-2.5 p-3 rounded-xl border transition-all",
            type === "WOWR" ? "bg-primary/5 border-primary/40 font-semibold" : "bg-card/40 border-border/80",
            isReadOnly ? "opacity-75 cursor-not-allowed" : "cursor-pointer hover:bg-muted/40"
          )}>
            <input
              type="radio"
              name={`wowr_type_${photo?.issue_photo_id || issue.issue_id}`}
              checked={type === "WOWR"}
              onChange={() => {
                setType("WOWR");
                setInputValue(activeTarget.wo_id || activeTarget.wr_id || "");
                setValidationError("");
              }}
              disabled={isReadOnly}
              className="accent-primary h-4 w-4"
            />
            <span>Fu WO / WR</span>
          </label>

          <label className={cn(
            "flex items-center gap-2.5 p-3 rounded-xl border transition-all",
            type === "None" ? "bg-muted/30 border-border font-semibold" : "bg-card/40 border-border/80",
            isReadOnly ? "opacity-75 cursor-not-allowed" : "cursor-pointer hover:bg-muted/40"
          )}>
            <input
              type="radio"
              name={`wowr_type_${photo?.issue_photo_id || issue.issue_id}`}
              checked={type === "None"}
              onChange={() => {
                setType("None");
                setInputValue("");
                setValidationError("");
              }}
              disabled={isReadOnly}
              className="accent-primary h-4 w-4"
            />
            <span>Tidak Menggunakan WO / WR</span>
          </label>
        </div>

        {type === "WOWR" && (
          <div className="space-y-2">
            <label className="text-xs font-semibold text-foreground">
              Nomor Referensi WO / WR <span className="text-red-500">*</span>
            </label>
            <Input
              value={inputValue}
              onChange={(event) => {
                setInputValue(event.target.value.toUpperCase());
                setValidationError("");
              }}
              disabled={isReadOnly}
              placeholder="MASUKKAN NOMOR REFERENSI WO / WR..."
              className={cn(
                "h-11 rounded-xl text-xs uppercase font-mono tracking-wider font-semibold",
                validationError && "border-red-500 focus-visible:ring-red-500"
              )}
            />
            {validationError && <p className="text-red-500 text-xs font-medium">{validationError}</p>}
          </div>
        )}

        {isAuditor && activeTarget.wowr_status === "PendingValidation" && (
          <div className="p-3.5 rounded-xl bg-amber-500/10 border border-amber-500/30 space-y-2">
            <div className="flex items-center justify-between text-xs font-semibold text-amber-700 dark:text-amber-400">
              <span>Validasi Bukti WO / WR Auditor</span>
              <span className="font-mono text-[11px]">{activeTarget.wo_id || activeTarget.wr_id}</span>
            </div>
            <div className="flex gap-2">
              <Button
                size="sm"
                variant="destructive"
                className="w-1/2 text-xs rounded-xl h-9"
                isLoading={wowrValidationMutation.isPending}
                onClick={() => wowrValidationMutation.mutate("Rejected")}
              >
                <XCircle className="mr-1.5 h-3.5 w-3.5" /> Tolak
              </Button>
              <Button
                size="sm"
                className="w-1/2 text-xs bg-green-600 hover:bg-green-700 text-white rounded-xl h-9"
                isLoading={wowrValidationMutation.isPending}
                disabled={!hasWOWRProofImage}
                onClick={() => wowrValidationMutation.mutate("Verified")}
                title={!hasWOWRProofImage ? "Auditee belum mengunggah foto bukti WO/WR" : "Setujui WO/WR"}
              >
                <CheckCircle2 className="mr-1.5 h-3.5 w-3.5" /> Setujui
              </Button>
            </div>
            {!hasWOWRProofImage && (
              <p className="text-[11px] font-semibold text-red-500">
                Verifikasi terkunci: Auditee belum mengunggah foto bukti WO/WR.
              </p>
            )}
          </div>
        )}

        {!isReadOnly && (
          <Button
            onClick={handleSave}
            isLoading={wowrMutation.isPending || photoWowrMutation.isPending}
            className="w-full rounded-xl h-11 text-xs font-semibold"
          >
            {(wowrMutation.isPending || photoWowrMutation.isPending) && (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            )}
            Simpan Data Maintenance
          </Button>
        )}
      </div>

    </Card>
  );
};
