"use client";

import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Check, Copy, ExternalLink, Link2, Loader2, RefreshCw, ShieldAlert, Trash2, X } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { getApiErrorMessage } from "@/lib/api/error";
import { kpiShareApi, type KPIShareFilter, type KPIShareRecord } from "@/lib/api/kpi-share.api";
import type { Plant } from "@/types/api";

interface KPIShareDialogProps {
  open: boolean;
  onClose: () => void;
  plantId: string;
  plants?: Plant[];
  filter: KPIShareFilter;
}

function publicURL(token: string) {
  return `${window.location.origin}/shared/kpi/${encodeURIComponent(token)}`;
}

function shareStatus(item: KPIShareRecord, now: number) {
  if (item.revoked_at) return { label: "Dicabut", className: "bg-red-500/10 text-red-600" };
  if (item.expires_at && new Date(item.expires_at).getTime() <= now) return { label: "Kedaluwarsa", className: "bg-amber-500/10 text-amber-600" };
  return { label: "Aktif", className: "bg-emerald-500/10 text-emerald-600" };
}

export function KPIShareDialog({ open, onClose, plantId, plants, filter }: KPIShareDialogProps) {
  const queryClient = useQueryClient();
  const [shareName, setShareName] = useState("Dashboard KPI");
  const [publicTitle, setPublicTitle] = useState("Dashboard KPI Cimory");
  const [expiryDays, setExpiryDays] = useState("7");
  const [allowPeriodChange, setAllowPeriodChange] = useState(false);
  const [createdLink, setCreatedLink] = useState("");
  const [copied, setCopied] = useState(false);
  const [renderTimestamp] = useState(() => Date.now());
  const selectedPlantName = useMemo(() => plants?.find((plant) => plant.plant_id === plantId)?.plant_name || plantId, [plantId, plants]);

  const sharesQuery = useQuery({
    queryKey: ["kpi-public-shares"],
    queryFn: kpiShareApi.list,
    enabled: open,
    staleTime: 15_000,
  });

  const createMutation = useMutation({
    mutationFn: () => kpiShareApi.create({
      share_name: shareName.trim(),
      public_title: publicTitle.trim(),
      plant_id: plantId,
      filter,
      allow_period_change: allowPeriodChange,
      expires_at: new Date(Date.now() + Number(expiryDays) * 86_400_000).toISOString(),
    }),
    onSuccess: (created) => {
      setCreatedLink(publicURL(created.raw_token));
      queryClient.invalidateQueries({ queryKey: ["kpi-public-shares"] });
      toast.success("Link publik berhasil dibuat");
    },
    onError: (error) => toast.error(getApiErrorMessage(error, "Gagal membuat link publik")),
  });

  const revokeMutation = useMutation({
    mutationFn: kpiShareApi.revoke,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["kpi-public-shares"] });
      toast.success("Link publik telah dicabut");
    },
    onError: (error) => toast.error(getApiErrorMessage(error, "Gagal mencabut link")),
  });

  const rotateMutation = useMutation({
    mutationFn: kpiShareApi.rotate,
    onSuccess: (created) => {
      setCreatedLink(publicURL(created.raw_token));
      queryClient.invalidateQueries({ queryKey: ["kpi-public-shares"] });
      toast.success("Link baru dibuat dan link lama dinonaktifkan");
    },
    onError: (error) => toast.error(getApiErrorMessage(error, "Gagal memperbarui link")),
  });

  if (!open) return null;

  const copyLink = async () => {
    await navigator.clipboard.writeText(createdLink);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 2000);
    toast.success("Link disalin");
  };

  return (
    <div className="fixed inset-0 z-[150] flex items-center justify-center bg-black/65 p-3 backdrop-blur-sm sm:p-6" role="dialog" aria-modal="true" aria-labelledby="kpi-share-title" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <div className="flex max-h-[92vh] w-full max-w-3xl flex-col overflow-hidden rounded-2xl border border-border bg-card shadow-2xl">
        <header className="flex items-start justify-between gap-4 border-b border-border px-5 py-4 sm:px-6">
          <div>
            <h2 id="kpi-share-title" className="flex items-center gap-2 text-lg font-semibold"><Link2 className="h-5 w-5 text-primary" /> Bagikan Dashboard KPI</h2>
            <p className="mt-1 text-xs text-muted-foreground">Buat tampilan publik read-only yang tetap dibatasi ke satu plant.</p>
          </div>
          <button type="button" onClick={onClose} className="rounded-lg p-2 text-muted-foreground hover:bg-muted hover:text-foreground" aria-label="Tutup"><X className="h-5 w-5" /></button>
        </header>

        <div className="flex-1 space-y-6 overflow-y-auto p-5 sm:p-6">
          {!plantId ? (
            <div className="flex gap-3 rounded-xl border border-amber-500/30 bg-amber-500/10 p-4 text-sm text-amber-700 dark:text-amber-300">
              <ShieldAlert className="mt-0.5 h-5 w-5 shrink-0" />
              <div><p className="font-semibold">Pilih satu plant terlebih dahulu</p><p className="mt-1 text-xs">Dashboard Semua Plant tidak dapat dibagikan. Tutup dialog lalu pilih plant pada header KPI.</p></div>
            </div>
          ) : (
            <section className="space-y-4">
              <div className="grid gap-4 sm:grid-cols-2">
                <label className="space-y-1.5 text-sm font-medium">Nama link<Input value={shareName} maxLength={100} onChange={(event) => setShareName(event.target.value)} placeholder="Contoh: KPI Meeting Mingguan" /></label>
                <label className="space-y-1.5 text-sm font-medium">Judul yang tampil<Input value={publicTitle} maxLength={150} onChange={(event) => setPublicTitle(event.target.value)} placeholder="Dashboard KPI Plant" /></label>
                <label className="space-y-1.5 text-sm font-medium">Plant<Input value={selectedPlantName} disabled /></label>
                <label className="space-y-1.5 text-sm font-medium">Masa berlaku
                  <select value={expiryDays} onChange={(event) => setExpiryDays(event.target.value)} className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                    <option value="1">1 hari</option><option value="7">7 hari</option><option value="30">30 hari</option><option value="90">90 hari</option>
                  </select>
                </label>
              </div>
              <label className="flex items-start gap-3 rounded-xl border border-border p-3">
                <input type="checkbox" checked={allowPeriodChange} onChange={(event) => setAllowPeriodChange(event.target.checked)} className="mt-1 h-4 w-4 accent-primary" />
                <span><span className="block text-sm font-medium">Izinkan pengunjung mengganti periode</span><span className="mt-0.5 block text-xs text-muted-foreground">Pengunjung tetap tidak dapat mengganti plant, layout, maupun konfigurasi visualisasi.</span></span>
              </label>
              <div className="flex justify-end"><Button onClick={() => createMutation.mutate()} isLoading={createMutation.isPending} disabled={!shareName.trim() || !publicTitle.trim()}>Buat Link Publik</Button></div>
            </section>
          )}

          {createdLink && (
            <section className="rounded-xl border border-emerald-500/30 bg-emerald-500/5 p-4">
              <p className="text-sm font-semibold text-emerald-700 dark:text-emerald-300">Link siap digunakan</p>
              <p className="mt-1 text-xs text-muted-foreground">Simpan sekarang. Demi keamanan, token lengkap tidak disimpan dalam daftar.</p>
              <div className="mt-3 flex gap-2"><Input readOnly value={createdLink} className="font-mono text-xs" /><Button variant="outline" size="sm" onClick={() => void copyLink()}>{copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}</Button><Button variant="outline" size="sm" onClick={() => window.open(createdLink, "_blank", "noopener,noreferrer")}><ExternalLink className="h-4 w-4" /></Button></div>
            </section>
          )}

          <section>
            <h3 className="mb-3 text-sm font-semibold">Link yang pernah dibuat</h3>
            {sharesQuery.isLoading ? <div className="flex justify-center py-8"><Loader2 className="h-5 w-5 animate-spin text-muted-foreground" /></div> : !sharesQuery.data?.length ? <p className="rounded-xl border border-dashed border-border p-6 text-center text-xs text-muted-foreground">Belum ada link publik.</p> : (
              <div className="space-y-2">
                {sharesQuery.data.map((item) => { const status = shareStatus(item, renderTimestamp); const canRotate = !item.expires_at || new Date(item.expires_at).getTime() > renderTimestamp; return (
                  <article key={item.share_id} className="flex flex-col gap-3 rounded-xl border border-border p-3 sm:flex-row sm:items-center sm:justify-between">
                    <div className="min-w-0"><div className="flex flex-wrap items-center gap-2"><p className="truncate text-sm font-semibold">{item.share_name}</p><span className={`rounded-full px-2 py-0.5 text-[10px] font-bold ${status.className}`}>{status.label}</span></div><p className="mt-1 text-xs text-muted-foreground">{item.plant_name} · {item.token_prefix}… · berakhir {item.expires_at ? new Date(item.expires_at).toLocaleString("id-ID") : "-"}</p><p className="mt-0.5 text-[11px] text-muted-foreground">{item.access_count} akses{item.last_accessed_at ? ` · terakhir ${new Date(item.last_accessed_at).toLocaleString("id-ID")}` : ""}</p></div>
                    <div className="flex shrink-0 gap-2">{canRotate && <Button variant="outline" size="sm" disabled={rotateMutation.isPending} onClick={() => rotateMutation.mutate(item.share_id)} title="Buat token baru"><RefreshCw className="h-3.5 w-3.5" /> Rotasi</Button>}{!item.revoked_at && canRotate && <Button variant="outline" size="sm" disabled={revokeMutation.isPending} onClick={() => revokeMutation.mutate(item.share_id)} className="text-red-600"><Trash2 className="h-3.5 w-3.5" /> Cabut</Button>}</div>
                  </article>
                ); })}
              </div>
            )}
          </section>
        </div>
      </div>
    </div>
  );
}
