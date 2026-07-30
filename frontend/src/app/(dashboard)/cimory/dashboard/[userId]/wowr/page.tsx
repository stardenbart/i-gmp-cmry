"use client";

import { useState, useMemo, useRef, useEffect } from "react";
import { createPortal } from "react-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams } from "next/navigation";
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
  History
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { issueApi, Issue } from "@/lib/api/issue.api";
import { useAuthStore } from "@/stores/authStore";
import { isAuditorUser } from "@/lib/useAdminGuard";
import { usePermissions } from "@/lib/usePermissions";
import { useMounted } from "@/lib/useMounted";
import { useChunkedUpload } from "@/hooks/useChunkedUpload";
import { useSSE } from "@/hooks/useSSE";
import { cn, formatImageUrl } from "@/lib/utils";

// Modal Component for Uploading Photo (For Auditee)
function UploadProofModal({ 
  issue, 
  onClose,
  onSuccess 
}: { 
  issue: Issue; 
  onClose: () => void;
  onSuccess: () => void;
}) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const { uploadMutation, uploadProgress } = useChunkedUpload({ issueId: issue.issue_id });
  const queryClient = useQueryClient();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const updateStatusMutation = useMutation({
    mutationFn: () => issueApi.update(issue.issue_id, { 
      wowr_status: "PendingValidation"
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wowr-issues"] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      toast.success("Bukti WO/WR berhasil diunggah. Menunggu validasi Auditor.");
      onSuccess();
    },
    onError: () => {
      toast.error("Gagal memperbarui status menjadi Pending Validation.");
    }
  });

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setSelectedFile(e.target.files[0]);
    }
  };

  const handleUpload = () => {
    if (!selectedFile) return;
    
    uploadMutation.mutate(
      { file: selectedFile, type: "WOWR" },
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
      <div className="w-[90vw] sm:w-[450px] bg-background rounded-2xl p-6 shadow-xl border border-border/50 animate-in fade-in zoom-in-95 duration-200 flex flex-col">
        <h3 className="text-lg font-bold mb-1">Upload Bukti WO/WR</h3>
        <p className="text-sm text-muted-foreground mb-4">
          ID: <span className="font-mono text-foreground">{issue.issue_id}</span>
        </p>

        {selectedFile ? (
          <div className="relative aspect-video w-full rounded-xl border overflow-hidden mb-4 bg-muted flex items-center justify-center">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img 
              src={URL.createObjectURL(selectedFile)} 
              alt="Preview" 
              className="object-cover max-h-full max-w-full"
            />
          </div>
        ) : (
          <div 
            className="border-2 border-dashed border-border/60 rounded-xl p-8 mb-4 text-center cursor-pointer hover:bg-muted/50 transition-colors"
            onClick={() => fileInputRef.current?.click()}
          >
            <UploadCloud className="h-10 w-10 text-muted-foreground mx-auto mb-2" />
            <p className="text-sm font-medium">Klik untuk memilih foto</p>
            <p className="text-xs text-muted-foreground">Disarankan menggunakan kamera langsung</p>
          </div>
        )}

        <input 
          type="file" 
          accept="image/*"
          className="hidden"
          ref={fileInputRef}
          onChange={handleFileChange}
        />

        {uploadProgress > 0 && uploadProgress < 100 && (
          <div className="w-full bg-secondary rounded-full h-2 mb-4 overflow-hidden">
            <div className="bg-primary h-2 rounded-full transition-all" style={{ width: `${uploadProgress}%` }} />
          </div>
        )}

        <div className="flex gap-3 justify-end mt-6">
          <Button variant="outline" onClick={onClose} disabled={isUploading}>
            Batal
          </Button>
          <Button 
            onClick={handleUpload} 
            disabled={!selectedFile || isUploading} 
            isLoading={isUploading}
          >
            Upload & Selesaikan
          </Button>
        </div>
      </div>
    </div>
  );

  if (!mounted) return null;
  return createPortal(modalContent, document.body);
}

// Expandable Row Component
function IssueRow({ 
  issue, 
  isAuditor,
  onUploadClick,
  onPreviewImage
}: { 
  issue: Issue; 
  isAuditor: boolean;
  onUploadClick: (issue: Issue) => void;
  onPreviewImage: (url: string) => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const initialPhotos = issue.photos?.filter(p => p.photo_type === "Initial") || [];
  const wowrPhotos = issue.wowr_status === "Rejected"
    ? []
    : issue.photos?.filter(p => p.photo_type === "WOWR" || p.photo_type === "FollowUp") || [];
  
  const queryClient = useQueryClient();
  
  const approveMutation = useMutation({
    mutationFn: () => issueApi.update(issue.issue_id, { 
      wowr_status: "Verified",
      issue_status: "Closed"
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wowr-issues"] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      toast.success("Bukti WO/WR berhasil diverifikasi & temuan diselesaikan!");
    },
    onError: () => toast.error("Gagal memverifikasi bukti.")
  });

  const rejectMutation = useMutation({
    mutationFn: () => issueApi.update(issue.issue_id, { wowr_status: "Rejected" }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wowr-issues"] });
      queryClient.invalidateQueries({ queryKey: ["issues"] });
      queryClient.invalidateQueries({ queryKey: ["issues-filter"] });
      toast.error("Bukti WO/WR ditolak. Foto bukti lama dibersihkan agar Auditee mengunggah ulang.");
    },
    onError: () => toast.error("Gagal menolak bukti.")
  });

  return (
    <>
      <tr 
        className={cn(
          "hover:bg-muted/30 transition-colors cursor-pointer",
          expanded && "bg-muted/20"
        )}
        onClick={() => setExpanded(!expanded)}
      >
        <td className="px-4 py-3 font-mono text-xs">{issue.issue_id}</td>
        <td className="px-4 py-3 font-semibold text-primary">
          {issue.wo_id || issue.wr_id || <span className="text-muted-foreground text-xs font-normal italic">Belum diinput</span>}
        </td>
        <td className="px-4 py-3 truncate max-w-[200px]" title={issue.keterangan}>
          {issue.keterangan || "-"}
        </td>
        <td className="px-4 py-3">
          {(!issue.wowr_status || issue.wowr_status === "None") ? (
            <span className="text-xs text-muted-foreground">Menunggu Bukti</span>
          ) : issue.wowr_status === "PendingValidation" ? (
            <span className="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-500 border border-purple-500/20">
              <Loader2 className="h-3 w-3 animate-spin" /> Menunggu Validasi Auditor
            </span>
          ) : issue.wowr_status === "Verified" ? (
            <span className="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">
              <CheckCircle2 className="h-3 w-3" /> Terverifikasi
            </span>
          ) : issue.wowr_status === "Rejected" ? (
            <span className="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full bg-red-500/10 text-red-500 border border-red-500/20">
              <XCircle className="h-3 w-3" /> Ditolak
            </span>
          ) : (
            <span className="text-xs text-muted-foreground">{issue.wowr_status}</span>
          )}
        </td>
        <td className="px-4 py-3 text-right">
          {/* Actions for Auditor */}
          {isAuditor && issue.wowr_status === "PendingValidation" && (
            <div className="flex justify-end gap-2">
              <Button 
                size="sm" 
                variant="default"
                className="text-xs h-8 bg-emerald-600 hover:bg-emerald-700"
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
                className="text-xs h-8"
                disabled={approveMutation.isPending || rejectMutation.isPending}
                onClick={(e) => {
                  e.stopPropagation();
                  rejectMutation.mutate();
                }}
              >
                <X className="h-3.5 w-3.5 mr-1" />
                Tolak
              </Button>
            </div>
          )}

          {/* Actions for Auditee */}
          {!isAuditor && (!issue.wowr_status || issue.wowr_status === "None" || issue.wowr_status === "Rejected") && (
            <Button 
              size="sm" 
              variant="default"
              className="text-xs h-8"
              onClick={(e) => {
                e.stopPropagation();
                onUploadClick(issue);
              }}
            >
              <UploadCloud className="h-3.5 w-3.5 mr-1" />
              Upload Bukti
            </Button>
          )}
        </td>
      </tr>
      {expanded && (
        <tr className="bg-muted/5 border-b">
          <td colSpan={5} className="p-0">
            <div className="p-4 bg-muted/10 border-t border-border/50 animate-in slide-in-from-top-2 fade-in duration-200">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <h4 className="font-semibold text-sm mb-3 border-b pb-1 border-border/50">Detail Temuan</h4>
                  <div className="space-y-2 text-sm mb-4">
                    <div className="flex gap-2">
                      <span className="text-muted-foreground w-24">Due Date:</span>
                      <span className="font-medium">{issue.due_date ? new Date(issue.due_date).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }) : "-"}</span>
                    </div>
                    <div className="flex gap-2">
                      <span className="text-muted-foreground w-24">Deskripsi:</span>
                      <span className="font-medium whitespace-pre-wrap">{issue.keterangan || "-"}</span>
                    </div>
                  </div>
                  
                  <h5 className="text-xs font-semibold mb-2 text-muted-foreground uppercase tracking-wider">Foto Temuan Awal</h5>
                  {initialPhotos.length > 0 ? (
                    <div className="flex gap-2 overflow-x-auto pb-2">
                      {initialPhotos.map(p => (
                        <div 
                          key={p.issue_photo_id} 
                          className="relative group h-28 w-40 shrink-0 overflow-hidden rounded-xl border border-border/60 shadow-sm cursor-pointer bg-muted"
                          onClick={(e) => {
                            e.stopPropagation();
                            onPreviewImage(p.image_url);
                          }}
                        >
                          {/* eslint-disable-next-line @next/next/no-img-element */}
                          <img 
                            src={formatImageUrl(p.image_url)} 
                            alt="Initial" 
                            className="h-full w-full object-cover transition-transform group-hover:scale-105" 
                          />
                          <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-1.5 text-white text-xs font-medium">
                            <Eye className="h-4 w-4" />
                            <span>Lihat Foto</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : <span className="text-xs text-muted-foreground italic">Tidak ada foto temuan awal.</span>}
                </div>
                <div>
                  <h4 className="font-semibold text-sm mb-3 border-b pb-1 border-border/50">Bukti Penyelesaian WO/WR</h4>
                  {wowrPhotos.length > 0 ? (
                    <div className="flex gap-2 overflow-x-auto pb-2">
                      {wowrPhotos.map(p => (
                        <div 
                          key={p.issue_photo_id} 
                          className="relative group h-28 w-40 shrink-0 overflow-hidden rounded-xl border border-border/60 shadow-sm cursor-pointer bg-muted"
                          onClick={(e) => {
                            e.stopPropagation();
                            onPreviewImage(p.image_url);
                          }}
                        >
                          {/* eslint-disable-next-line @next/next/no-img-element */}
                          <img 
                            src={formatImageUrl(p.image_url)} 
                            alt="WOWR Evidence" 
                            className="h-full w-full object-cover transition-transform group-hover:scale-105" 
                          />
                          <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-1.5 text-white text-xs font-medium">
                            <Eye className="h-4 w-4" />
                            <span>Lihat Foto</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="h-28 flex flex-col items-center justify-center border-2 border-dashed border-border/60 rounded-xl text-center bg-muted/30">
                      <ImageIcon className="h-6 w-6 text-muted-foreground/50 mb-1" />
                      <span className="text-xs text-muted-foreground italic">Belum ada foto bukti penyelesaian WO/WR dari Auditee.</span>
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
  const { userId } = useParams() as { userId: string };
  const user = useAuthStore(state => state.user);
  const mounted = useMounted();
  const [search, setSearch] = useState("");
  const [selectedIssue, setSelectedIssue] = useState<Issue | null>(null);
  const [previewImage, setPreviewImage] = useState<string | null>(null);

  useSSE();

  const { hasPermission } = usePermissions();
  const isAuditor = isAuditorUser(user?.role_id, user?.role?.role_name, user?.username) || hasPermission("PERM-WOWR-U");

  const { data, isLoading } = useQuery({
    queryKey: ["wowr-issues", userId],
    queryFn: () => issueApi.getAll({ needs_wo_wr: true, limit: 1000 }),
    enabled: mounted && !!user, // Fetch when user is available
  });

  const allIssues: Issue[] = Array.isArray(data?.items) ? data.items : [];

  const issues = useMemo(() => {
    if (!search.trim()) return allIssues;
    const q = search.toLowerCase();
    return allIssues.filter(
      (i) =>
        i.keterangan?.toLowerCase().includes(q) ||
        i.issue_id?.toLowerCase().includes(q) ||
        i.wo_id?.toLowerCase().includes(q) ||
        i.wr_id?.toLowerCase().includes(q) ||
        i.area_name?.toLowerCase().includes(q) ||
        i.kawasan_name?.toLowerCase().includes(q) ||
        i.detail_kawasan_name?.toLowerCase().includes(q) ||
        i.pic_name?.toLowerCase().includes(q)
    );
  }, [allIssues, search]);

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
        <div className="flex items-center justify-between bg-card p-5 rounded-2xl border shadow-sm">
          <div className="flex items-center gap-4">
            <div className="h-12 w-12 rounded-full bg-primary/10 flex items-center justify-center text-primary">
              <FileBox className="h-6 w-6" />
            </div>
            <div>
              <h2 className="text-xl font-bold tracking-tight">Manajemen WO / WR</h2>
              <p className="text-sm text-muted-foreground mt-0.5">
                {isAuditor 
                  ? "Dashboard Auditor untuk memeriksa bukti penyelesaian WO/WR"
                  : "Kelola dan unggah bukti penyelesaian Work Order & Work Request Anda"}
              </p>
            </div>
          </div>
        </div>

        {/* ── Search ── */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
          <input
            type="search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cari Nomor WO/WR atau keterangan..."
            className="w-full rounded-xl border bg-card pl-9 pr-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/40 transition-all"
          />
        </div>

        {/* ── Table / List ── */}
        <div className="bg-card rounded-2xl border overflow-hidden shadow-sm">
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground text-xs uppercase">
                <tr>
                  <th className="px-4 py-3 font-medium">Issue ID</th>
                  <th className="px-4 py-3 font-medium">Nomor WO / WR</th>
                  <th className="px-4 py-3 font-medium max-w-xs">Keterangan</th>
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
                ) : issues.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-8 text-center text-muted-foreground">
                      {isAuditor 
                        ? "Tidak ada data WO / WR yang perlu Anda tinjau saat ini."
                        : "Belum ada data WO / WR yang perlu dikerjakan."}
                    </td>
                  </tr>
                ) : (
                  issues.map((issue) => (
                    <IssueRow 
                      key={issue.issue_id} 
                      issue={issue} 
                      isAuditor={isAuditor}
                      onUploadClick={(iss) => setSelectedIssue(iss)} 
                      onPreviewImage={(url) => setPreviewImage(url)}
                    />
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
      
      {/* Upload Modal */}
      {selectedIssue && (
        <UploadProofModal 
          issue={selectedIssue} 
          onClose={() => setSelectedIssue(null)}
          onSuccess={() => setSelectedIssue(null)}
        />
      )}

      {/* Image Lightbox Preview Modal */}
      {previewImage && (
        <div
          className="fixed inset-0 z-[110] flex items-center justify-center bg-black/90 backdrop-blur-sm p-4 animate-in fade-in duration-200"
          onClick={() => setPreviewImage(null)}
        >
          <div className="relative max-w-4xl w-full flex flex-col items-center" onClick={(e) => e.stopPropagation()}>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setPreviewImage(null)}
              className="absolute -top-12 right-0 text-white hover:text-zinc-300 rounded-full"
            >
              <X className="h-7 w-7" />
            </Button>
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={formatImageUrl(previewImage)}
              alt="Preview Bukti WO/WR"
              className="rounded-2xl max-h-[85vh] w-full object-contain shadow-2xl border border-white/10"
            />
          </div>
        </div>
      )}
    </>
  );
}
