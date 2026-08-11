import { useState, useEffect } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Card } from "../ui/card";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import {
  Wrench,
  User,
  Calendar,
  Clock,
  MapPin,
  FileText,
  Tag,
  CheckCircle2,
  XCircle,
  Loader2,
  Lock,
} from "lucide-react";
import { Issue, IssuePhoto, issueApi, WOWRStatus } from "@/lib/api/issue.api";
import { cn } from "@/lib/utils";

interface DetailSpesifikasiTemuanCardProps {
  issue: Issue;
  photo?: IssuePhoto;
  dueDate?: Date | null;
  isAuditor: boolean;
  canEditWOWR?: boolean;
  onRefresh?: () => void;
}

export const DetailSpesifikasiTemuanCard = ({
  issue,
  photo,
  dueDate,
  isAuditor,
  canEditWOWR,
  onRefresh,
}: DetailSpesifikasiTemuanCardProps) => {
  const queryClient = useQueryClient();

  const activeTarget = photo || issue;

  const [type, setType] = useState<"WOWR" | "None" | "">(
    activeTarget.needs_wo_wr || activeTarget.wo_id || activeTarget.wr_id ? "WOWR" : "None"
  );
  const [inputValue, setInputValue] = useState("");
  const [validationError, setValidationError] = useState("");

  const wowrMutation = useMutation({
    mutationFn: (data: {
      needs_wo_wr: boolean;
      wo_id: string;
      wr_id: string;
      wowr_status?: WOWRStatus;
    }) => issueApi.update(issue.issue_id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issue", issue.issue_id] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      toast.success("Data Maintenance WO/WR berhasil disimpan. Harap tunggu konfirmasi Auditor.");
      if (onRefresh) onRefresh();
    },
    onError: (err: any) => {
      const msg =
        err?.response?.data?.message ||
        err?.response?.data?.error ||
        "Gagal menyimpan data WO/WR.";
      toast.error(msg);
    },
  });

  const photoWowrMutation = useMutation({
    mutationFn: ({ photoId, data }: { photoId: string; data: any }) =>
      issueApi.updatePhotoWOWR(photoId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issue-photos", issue.issue_id] });
      queryClient.invalidateQueries({ queryKey: ["issue", issue.issue_id] });
      toast.success("Data Maintenance WO/WR foto berhasil disimpan. Harap tunggu konfirmasi Auditor.");
      if (onRefresh) onRefresh();
    },
    onError: (err: any) => {
      const msg =
        err?.response?.data?.message ||
        err?.response?.data?.error ||
        "Gagal menyimpan data WO/WR foto.";
      toast.error(msg);
    },
  });

  const wowrValidationMutation = useMutation({
    mutationFn: (status: WOWRStatus) => {
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
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["issue-photos", issue.issue_id] });
      queryClient.invalidateQueries({ queryKey: ["issue", issue.issue_id] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      if (variables === "Verified") {
        toast.success("Bukti WO/WR terisolasi foto berhasil diverifikasi!");
      } else {
        toast.error("Bukti WO/WR ditolak.");
      }
      if (onRefresh) onRefresh();
    },
    onError: (err: any) => {
      const msg =
        err?.response?.data?.message ||
        err?.response?.data?.error ||
        "Gagal memperbarui status validasi WO/WR.";
      toast.error(msg);
    },
  });

  useEffect(() => {
    const target = photo || issue;
    if (target.needs_wo_wr || target.wo_id || target.wr_id) {
      setType("WOWR");
      setInputValue(target.wo_id || target.wr_id || "");
    } else {
      setType("None");
      setInputValue("");
    }
    setValidationError("");
  }, [issue, photo]);

  const handleSave = () => {
    if (!type) {
      setValidationError("Harap pilih jenis penggunaan Maintenance (WO/WR).");
      toast.error("Gagal: Harap pilih jenis penggunaan Maintenance (WO/WR).");
      return;
    }

    if (type === "WOWR" && !inputValue.trim()) {
      setValidationError("Nomor referensi WO / WR wajib diisi jika memilih Fu WO / WR.");
      toast.error("Gagal: Nomor referensi WO / WR wajib diisi jika memilih Fu WO / WR.");
      return;
    }

    setValidationError("");

    const upperValue = inputValue.trim().toUpperCase();

    if (photo && photo.issue_photo_id) {
      photoWowrMutation.mutate({
        photoId: photo.issue_photo_id,
        data: {
          needs_wo_wr: type === "WOWR",
          wo_id: type === "WOWR" ? upperValue : "",
          wr_id: "",
          wowr_status: type === "WOWR" ? "PendingValidation" : "None",
        },
      });
    } else {
      wowrMutation.mutate({
        needs_wo_wr: type === "WOWR",
        wo_id: type === "WOWR" ? upperValue : "",
        wr_id: "",
        wowr_status: type === "WOWR" ? "PendingValidation" : "None",
      });
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

  const habitName = photo?.habit_name || issue.hei?.habit?.habit_name || issue.habit_name;
  const equipmentName = photo?.equipment_name || issue.hei?.equipment?.equipment_name || issue.equipment_name;
  const infrastructureName = photo?.infrastructure_name || issue.hei?.infrastructure?.infrastructure_name || issue.infrastructure_name;

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
        {(habitName || equipmentName || infrastructureName || photo?.hei_category || issue.hei_category) && (
          <div className="pt-2 border-t border-border/60 space-y-1.5">
            <span className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground flex items-center gap-1">
             Kategori HEI (Hasil Inspeksi)
            </span>
            <div className="flex flex-wrap gap-2">
              {habitName && (
                <span className="px-2.5 py-1 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-semibold border border-emerald-500/20 text-[11px]">
                  [Habit] {habitName}
                </span>
              )}
              {equipmentName && (
                <span className="px-2.5 py-1 rounded-xl bg-blue-500/10 text-blue-600 dark:text-blue-400 font-semibold border border-blue-500/20 text-[11px]">
                  [Equipment] {equipmentName}
                </span>
              )}
              {infrastructureName && (
                <span className="px-2.5 py-1 rounded-xl bg-purple-500/10 text-purple-600 dark:text-purple-400 font-semibold border border-purple-500/20 text-[11px]">
                  [Infrastructure] {infrastructureName}
                </span>
              )}
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

      {/* ── Section 2: Maintenance (WO / WR) Selection & Validation (Isolated Per Photo) ──── */}
      <div className="pt-4 border-t border-border space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <h4 className="font-bold text-sm">
              Maintenance (WO / WR) {photo ? "Terisolasi Foto" : ""}
            </h4>
          </div>
          {isReadOnly ? (
            <span className="text-[10px] bg-amber-500/10 text-amber-600 dark:text-amber-400 px-2 py-0.5 rounded-full font-semibold border border-amber-500/20 flex items-center gap-1">
              Terkunci
            </span>
          ) : (
            <span className="text-[10px] bg-green-500/10 text-green-600 dark:text-green-400 px-2 py-0.5 rounded-full font-semibold border border-green-500/20">
              Dapat Diisi
            </span>
          )}
        </div>

        {/* Radio Options */}
        <div className="flex flex-col gap-2.5 text-xs">
          <label
            className={cn(
              "flex items-center gap-2.5 p-3 rounded-xl border border-border/80 transition-all",
              type === "WOWR" ? "bg-primary/5 border-primary/40 font-semibold" : "bg-card/40",
              !isReadOnly && "cursor-pointer hover:bg-muted/40",
              isReadOnly && "opacity-75 cursor-not-allowed"
            )}
          >
            <input
              type="radio"
              name="wowr_type_detail"
              value="WOWR"
              checked={type === "WOWR"}
              onChange={() => {
                setType("WOWR");
                if (!inputValue) setInputValue(activeTarget.wo_id || activeTarget.wr_id || "");
                setValidationError("");
              }}
              disabled={isReadOnly}
              className="accent-primary h-4 w-4"
            />
            <span className="text-xs font-medium">Fu WO / WR</span>
          </label>

          <label
            className={cn(
              "flex items-center gap-2.5 p-3 rounded-xl border border-border/80 transition-all",
              type === "None" ? "bg-muted/30 border-border font-semibold" : "bg-card/40",
              !isReadOnly && "cursor-pointer hover:bg-muted/40",
              isReadOnly && "opacity-75 cursor-not-allowed"
            )}
          >
            <input
              type="radio"
              name="wowr_type_detail"
              value="None"
              checked={type === "None"}
              onChange={() => {
                setType("None");
                setInputValue("");
                setValidationError("");
              }}
              disabled={isReadOnly}
              className="accent-primary h-4 w-4"
            />
            <span className="text-xs font-medium">Tidak Menggunakan WO / WR</span>
          </label>
        </div>

        {/* Input Text & Strict Validation Message */}
        {type === "WOWR" && (
          <div className="space-y-2 pt-1 animate-in fade-in slide-in-from-top-2 duration-300">
            <label className="text-xs font-semibold text-foreground flex items-center justify-between">
              <span>Nomor Referensi WO / WR <span className="text-red-500">*</span></span>
              {activeTarget.wo_id || activeTarget.wr_id ? (
                <span className="text-[11px] font-mono text-emerald-600 font-bold">
                  Tersimpan: {activeTarget.wo_id || activeTarget.wr_id}
                </span>
              ) : null}
            </label>
            <Input
              value={inputValue}
              onChange={(e) => {
                const upperVal = e.target.value.toUpperCase();
                setInputValue(upperVal);
                if (upperVal.trim()) setValidationError("");
              }}
              disabled={isReadOnly}
              placeholder="MASUKKAN NOMOR REFERENSI WO / WR UNTUK FOTO INI..."
              className={cn(
                "h-11 rounded-xl text-xs uppercase font-mono tracking-wider font-semibold",
                validationError ? "border-red-500 focus-visible:ring-red-500" : ""
              )}
            />
            {validationError && (
              <p className="text-red-500 text-xs font-medium pl-1 animate-in fade-in">
                {validationError}
              </p>
            )}
          </div>
        )}

        {/* Auditor Validation Controls */}
        {isAuditor && activeTarget.wowr_status === "PendingValidation" && (
          <div className="p-3.5 rounded-xl bg-amber-500/10 border border-amber-500/30 space-y-2">
            <div className="flex items-center justify-between text-xs font-semibold text-amber-700 dark:text-amber-400">
              <span>Validasi Bukti WO / WR Auditor</span>
              <span className="font-mono text-[11px]">{activeTarget.wo_id || activeTarget.wr_id}</span>
            </div>
            <div className="flex gap-2 pt-1">
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
                className="w-1/2 text-xs bg-green-600 hover:bg-green-700 text-white rounded-xl h-9 font-semibold"
                isLoading={wowrValidationMutation.isPending}
                onClick={() => wowrValidationMutation.mutate("Verified")}
              >
                <CheckCircle2 className="mr-1.5 h-3.5 w-3.5" /> Setujui
              </Button>
            </div>
          </div>
        )}

        {/* Save Button for Maintenance WO / WR */}
        {!isReadOnly && (
          <Button
            onClick={handleSave}
            isLoading={wowrMutation.isPending || photoWowrMutation.isPending}
            className="w-full bg-primary text-primary-foreground hover:bg-primary/90 font-semibold rounded-xl h-11 text-xs shadow-md transition-all mt-2"
          >
            {wowrMutation.isPending || photoWowrMutation.isPending ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <span>Simpan Data Maintenance Foto</span>
            )}
            
          </Button>
        )}
      </div>
    </Card>
  );
};
