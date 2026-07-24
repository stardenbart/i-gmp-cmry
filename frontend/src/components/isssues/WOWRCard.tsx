import { useState, useEffect } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Card } from "../ui/card";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Wrench, CheckCircle2, XCircle, Loader2 } from "lucide-react";
import { Issue, issueApi, WOWRStatus } from "@/lib/api/issue.api";
import { cn } from "@/lib/utils";

interface WOWRCardProps {
  issue: Issue;
  isAuditor: boolean;
  canEdit?: boolean;
}

export const WOWRCard = ({ issue, isAuditor, canEdit }: WOWRCardProps) => {
  const queryClient = useQueryClient();
  const [type, setType] = useState<"WOWR" | "None" | "">(
    issue.needs_wo_wr || issue.wo_id || issue.wr_id ? "WOWR" : "None"
  );
  
  const [inputValue, setInputValue] = useState("");

  const wowrMutation = useMutation({
    mutationFn: (data: { needs_wo_wr: boolean; wo_id: string; wr_id: string }) => 
      issueApi.update(issue.issue_id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issue", issue.issue_id] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      toast.success("Data Maintenance WO/WR berhasil disimpan.");
    },
    onError: () => toast.error("Gagal menyimpan data WO/WR."),
  });

  const wowrValidationMutation = useMutation({
    mutationFn: (status: WOWRStatus) => 
      issueApi.update(issue.issue_id, { wowr_status: status }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["issue", issue.issue_id] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      queryClient.invalidateQueries({ queryKey: ["wowr-issues"] });
      toast.success(`Status validasi WO/WR diperbarui.`);
    },
    onError: () => toast.error("Gagal memperbarui status validasi WO/WR."),
  });

  useEffect(() => {
    // Determine initial state based on data
    if (issue.needs_wo_wr || issue.wo_id || issue.wr_id) {
      setType("WOWR");
      setInputValue(issue.wo_id || issue.wr_id || "");
    } else {
      setType("None");
      setInputValue("");
    }
  }, [issue]);

  const handleSave = () => {
    if (!type) return;

    wowrMutation.mutate({
      needs_wo_wr: type === "WOWR",
      wo_id: type === "WOWR" ? inputValue : "",
      wr_id: "", // Since we use 1 input, we can just save it into wo_id
    });
  };

  // Only Auditees (PIC) in "InProgress" status can edit. Auditors or Auditees not yet started are read-only.
  const isReadOnly = canEdit !== undefined ? !canEdit : isAuditor;

  return (
    <Card className="p-6 bg-card/60 backdrop-blur-md lg:col-span-1 space-y-4 h-fit">
      <div className="flex items-center justify-between border-b border-border pb-2">
        <div className="flex items-center gap-2">
          <Wrench className="h-4 w-4 text-muted-foreground" />
          <h3 className="font-semibold">Maintenance (WO / WR)</h3>
        </div>
        {!isAuditor && !canEdit && (
          <span className="text-[10px] bg-amber-500/10 text-amber-500 px-2 py-0.5 rounded-full font-medium">
            Terkunci
          </span>
        )}
      </div>

      <div className="space-y-4">
        {/* Radio Buttons */}
        <div className="flex flex-col gap-3">
          <label className={cn("flex items-center gap-2", !isReadOnly && "cursor-pointer", isReadOnly && "opacity-70 cursor-not-allowed")}>
            <input 
              type="radio" 
              name="wowr_type" 
              value="WOWR" 
              checked={type === "WOWR"}
              onChange={() => {
                  setType("WOWR");
                  if (!inputValue) setInputValue(issue.wo_id || issue.wr_id || "");
              }}
              disabled={isReadOnly}
              className="accent-primary h-4 w-4 disabled:cursor-not-allowed"
            />
            <span className="text-sm font-medium">Membutuhkan WO / WR</span>
          </label>
          <label className={cn("flex items-center gap-2", !isReadOnly && "cursor-pointer", isReadOnly && "opacity-70 cursor-not-allowed")}>
            <input 
              type="radio" 
              name="wowr_type" 
              value="None" 
              checked={type === "None"}
              onChange={() => {
                  setType("None");
                  setInputValue("");
              }}
              disabled={isReadOnly}
              className="accent-primary h-4 w-4 disabled:cursor-not-allowed"
            />
            <span className="text-sm font-medium">Tidak Membutuhkan WO / WR</span>
          </label>
        </div>

        {/* Input Text - Only show if WOWR is selected */}
        {type === "WOWR" && (
          <div className="space-y-2 mt-4 animate-in fade-in slide-in-from-top-2 duration-300">
            <label className="text-xs font-medium text-muted-foreground">
              Nomor WO / WR
            </label>
            <Input 
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              disabled={isReadOnly}
              placeholder="Masukkan nomor referensi WO / WR..."
              className="disabled:opacity-50 disabled:cursor-not-allowed"
            />
          </div>
        )}

        {/* Save Button */}
        {!isReadOnly && (
          <Button 
            className="w-full mt-2" 
            onClick={handleSave} 
            disabled={!type || (type === "WOWR" && !inputValue) || wowrMutation.isPending}
            isLoading={wowrMutation.isPending}
          >
            Simpan Data Maintenance
          </Button>
        )}

        {/* Auditor Validation Actions */}
        {isAuditor && issue.needs_wo_wr && issue.wowr_status === "PendingValidation" && (
          <div className="mt-4 pt-4 border-t border-border/50 animate-in fade-in zoom-in-95 duration-300">
            <h4 className="text-xs font-bold text-muted-foreground uppercase mb-3 tracking-wider">Validasi Bukti WO/WR</h4>
            <div className="flex gap-2">
              <Button 
                variant="default" 
                className="flex-1 bg-emerald-600 hover:bg-emerald-700 text-white" 
                size="sm"
                onClick={() => wowrValidationMutation.mutate("Verified")}
                isLoading={wowrValidationMutation.isPending}
              >
                <CheckCircle2 className="h-4 w-4 mr-1.5" /> Setujui
              </Button>
              <Button 
                variant="destructive" 
                className="flex-1" 
                size="sm"
                onClick={() => wowrValidationMutation.mutate("Rejected")}
                isLoading={wowrValidationMutation.isPending}
              >
                <XCircle className="h-4 w-4 mr-1.5" /> Tolak
              </Button>
            </div>
          </div>
        )}
        
        {/* WOWR Status Badge */}
        {issue.needs_wo_wr && issue.wowr_status && issue.wowr_status !== "None" && (
          <div className="mt-4 p-3 bg-muted/40 rounded-xl border border-border/50 flex items-center justify-between">
            <span className="text-xs font-semibold text-muted-foreground">Status Validasi</span>
            {issue.wowr_status === "PendingValidation" ? (
              <span className="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-500 border border-purple-500/20">
                <Loader2 className="h-3 w-3 animate-spin" /> Menunggu Validasi
              </span>
            ) : issue.wowr_status === "Verified" ? (
               <span className="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">
                <CheckCircle2 className="h-3 w-3" /> Terverifikasi
              </span>
            ) : issue.wowr_status === "Rejected" ? (
               <span className="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full bg-red-500/10 text-red-500 border border-red-500/20">
                <XCircle className="h-3 w-3" /> Ditolak
              </span>
            ) : null}
          </div>
        )}
      </div>
    </Card>
  );
}
