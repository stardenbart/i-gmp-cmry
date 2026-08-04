"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Key, Plus, Copy, Check, Trash2, ShieldAlert, BookOpen, ExternalLink, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";
import { apiKeyApi, CreateAPIKeyResponse } from "@/types/api/apikey";

import { masterApi } from "@/lib/api/master.api";

interface ApiKeyManagerProps {
  plantFilter?: string;
}

export function ApiKeyManager({ plantFilter = "ALL" }: ApiKeyManagerProps) {
  const queryClient = useQueryClient();
  const [isCreating, setIsCreating] = useState(false);
  const [newKeyName, setNewKeyName] = useState("");
  const [isSingleUse, setIsSingleUse] = useState(false);
  const [targetPlant, setTargetPlant] = useState(plantFilter !== "ALL" ? plantFilter : "GLOBAL");
  const [createdKeyData, setCreatedKeyData] = useState<CreateAPIKeyResponse | null>(null);
  const [copied, setCopied] = useState(false);

  // Interactive Datetime Filter Configurator
  const [sinceDate, setSinceDate] = useState<string>(
    new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split("T")[0] // default 30 days ago
  );
  const [copiedUrl, setCopiedUrl] = useState(false);

  const { data: plantsList = [] } = useQuery({
    queryKey: ["master-plants"],
    queryFn: () => masterApi.getPlants({ limit: 1000 }),
  });

  const { data: keys = [], isLoading } = useQuery({
    queryKey: ["api-keys", plantFilter],
    queryFn: () => apiKeyApi.list(plantFilter),
  });

  const createMutation = useMutation({
    mutationFn: () => apiKeyApi.create(newKeyName, isSingleUse, targetPlant),
    onSuccess: (data) => {
      setCreatedKeyData(data);
      setNewKeyName("");
      setIsCreating(false);
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
      toast.success("API Key berhasil dibuat!");
    },
    onError: (err: any) => {
      toast.error(err.response?.data?.message || "Gagal membuat API Key");
    },
  });

  const revokeMutation = useMutation({
    mutationFn: (id: string) => apiKeyApi.revoke(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
      toast.success("API Key berhasil dicabut");
    },
    onError: (err: any) => {
      toast.error(err.response?.data?.message || "Gagal mencabut API Key");
    },
  });

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    toast.success("Token disalin ke clipboard!");
    setTimeout(() => setCopied(false), 2000);
  };

  const getPowerBIUrl = (dateStr: string) => {
    const origin = typeof window !== "undefined" ? window.location.origin : "http://localhost:3000";
    if (!dateStr) return `${origin}/api/v1/public/powerbi/data`;
    return `${origin}/api/v1/public/powerbi/data?since=${dateStr}T00:00:00Z`;
  };

  const handleCopyUrl = (url: string) => {
    navigator.clipboard.writeText(url);
    setCopiedUrl(true);
    toast.success("URL Endpoint Power BI berhasil disalin!");
    setTimeout(() => setCopiedUrl(false), 2000);
  };

  const setPresetDate = (daysAgo: number | null) => {
    if (daysAgo === null) {
      setSinceDate("");
      return;
    }
    const d = new Date(Date.now() - daysAgo * 24 * 60 * 60 * 1000);
    setSinceDate(d.toISOString().split("T")[0]);
  };

  return (
    <div className="space-y-6">
      {/* Header & Action */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-5 bg-card rounded-2xl border border-border/80 shadow-sm">
        <div>
          <h3 className="text-lg font-bold flex items-center gap-2">
            <Key className="w-5 h-5 text-primary" /> Integrasi API Key Power BI
          </h3>
          <p className="text-xs text-muted-foreground mt-1 max-w-xl">
            Buat API Key Publik (Bearer Token) untuk menarik data <strong>inspeksi, temuan, follow-up, & photo</strong> secara aman ke Power BI. Key dapat dikonfigurasi <em>Single-Use (1x penarikan)</em> demi keamanan maksimal.
          </p>
        </div>
        <Button onClick={() => setIsCreating(true)} className="rounded-xl shrink-0">
          <Plus className="w-4 h-4 mr-2" /> Buat API Key Baru
        </Button>
      </div>

      {/* Form Dialog Create */}
      {isCreating && (
        <div className="p-5 bg-muted/30 border border-primary/30 rounded-2xl space-y-4 animate-in fade-in duration-200">
          <h4 className="font-semibold text-sm">Buat API Key Baru</h4>
          <div className="space-y-3">
            <div>
              <label className="text-xs font-medium block mb-1">Nama Key / Deskripsi Integrasi</label>
              <Input
                placeholder="Misal: Power BI Monthly Audit Report"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
              />
            </div>

            <div>
              <label className="text-xs font-medium block mb-1">Cakupan Data Pabrik (Target Plant)</label>
              <select
                className="w-full text-xs bg-background border border-border rounded-xl px-3 py-2 outline-none h-10 font-medium"
                value={targetPlant}
                onChange={(e) => setTargetPlant(e.target.value)}
              >
                <option value="GLOBAL">Semua Plant / Akses Global (SuperAdmin)</option>
                {plantsList.map((p: any) => (
                  <option key={p.plant_id} value={p.plant_id}>Pabrik {p.plant_name} ({p.plant_id})</option>
                ))}
              </select>
            </div>

            <div className="flex items-center justify-between p-3 bg-card rounded-xl border border-border">
              <div>
                <span className="text-xs font-semibold block">Sekali Pakai (Single-Use Pull)</span>
                <span className="text-[11px] text-muted-foreground block">Key otomatis non-aktif setelah 1 kali penarikan data Power BI selesai.</span>
              </div>
              <button
                type="button"
                role="switch"
                aria-checked={isSingleUse}
                onClick={() => setIsSingleUse(!isSingleUse)}
                className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out ${
                  isSingleUse ? "bg-primary" : "bg-muted-foreground/30"
                }`}
              >
                <span
                  className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow-lg transition duration-200 ease-in-out ${
                    isSingleUse ? "translate-x-5" : "translate-x-0"
                  }`}
                />
              </button>
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" size="sm" onClick={() => setIsCreating(false)} className="rounded-xl">
              Batal
            </Button>
            <Button
              size="sm"
              onClick={() => createMutation.mutate()}
              disabled={!newKeyName.trim() || createMutation.isPending}
              className="rounded-xl"
            >
              {createMutation.isPending && <RefreshCw className="w-3.5 h-3.5 mr-2 animate-spin" />}
              Buat Key
            </Button>
          </div>
        </div>
      )}

      {/* One-time Key Reveal Dialog */}
      {createdKeyData && (
        <div className="p-6 bg-amber-500/10 border-2 border-amber-500/30 rounded-2xl space-y-4 animate-in zoom-in-95 duration-200">
          <div className="flex items-start gap-3">
            <ShieldAlert className="w-6 h-6 text-amber-600 shrink-0 mt-0.5" />
            <div>
              <h4 className="font-bold text-amber-700 dark:text-amber-400">Salin Token API Key Anda Sekarang</h4>
              <p className="text-xs text-amber-600/90 dark:text-amber-400/90 mt-0.5">
                Token ini <strong>hanya ditampilkan sekali saja</strong>. Harap langsung disalin dan disimpan dengan aman.
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2 bg-card p-3 rounded-xl border border-amber-500/30">
            <code className="text-xs font-mono break-all flex-1 text-foreground font-semibold">
              {createdKeyData.raw_token}
            </code>
            <Button size="sm" variant="outline" onClick={() => handleCopy(createdKeyData.raw_token)} className="shrink-0 rounded-lg">
              {copied ? <Check className="w-4 h-4 text-green-500" /> : <Copy className="w-4 h-4" />}
            </Button>
          </div>

          <div className="bg-card/60 p-3.5 rounded-xl border border-border text-xs space-y-1.5">
            <span className="font-semibold block text-foreground">Cara Menggunakan di Power BI:</span>
            <p className="text-muted-foreground text-[11px]">
              1. Pilih <strong>Get Data -&gt; Web -&gt; Advanced</strong>
            </p>
            <p className="text-muted-foreground text-[11px]">
              2. Masukkan URL: <code className="bg-muted px-1 rounded text-foreground font-mono">{getPowerBIUrl(sinceDate)}</code>
            </p>
            <p className="text-muted-foreground text-[11px]">
              3. Tambahkan Header: <code className="bg-muted px-1 rounded text-foreground font-mono">Authorization</code> = <code className="bg-muted px-1 rounded text-foreground font-mono">Bearer {createdKeyData.raw_token}</code>
            </p>
          </div>

          <div className="flex justify-end">
            <Button size="sm" onClick={() => setCreatedKeyData(null)} className="rounded-xl">
              Saya Sudah Menyimpan Token Ini
            </Button>
          </div>
        </div>
      )}

      {/* Table of Keys */}
      <div className="bg-card rounded-2xl border border-border/80 overflow-hidden shadow-sm">
        <div className="px-5 py-4 border-b border-border flex items-center justify-between">
          <h4 className="font-semibold text-sm">Daftar API Key Publik</h4>
          <span className="text-xs text-muted-foreground">{keys.length} Key terdaftar</span>
        </div>

        {isLoading ? (
          <div className="p-8 text-center text-xs text-muted-foreground">Memuat daftar API Key...</div>
        ) : keys.length === 0 ? (
          <div className="p-8 text-center text-xs text-muted-foreground">Belum ada API Key yang dibuat. Klik "+ Buat API Key Baru" untuk memulai.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-xs text-left">
              <thead className="bg-muted/50 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider text-[10px]">
                <tr>
                  <th className="px-4 py-3">Nama Key</th>
                  <th className="px-4 py-3">Cakupan Plant</th>
                  <th className="px-4 py-3">Prefix Token</th>
                  <th className="px-4 py-3">Tipe Expiry</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Terakhir Digunakan</th>
                  <th className="px-4 py-3 text-right">Aksi</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {keys.map((k) => (
                  <tr key={k.key_id} className="hover:bg-muted/20 transition-colors">
                    <td className="px-4 py-3 font-medium text-foreground">{k.name}</td>
                    <td className="px-4 py-3 font-medium">
                      {k.plant_id ? (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold bg-primary/10 text-primary border border-primary/20">
                          {plantsList.find((p: any) => p.plant_id === k.plant_id)?.plant_name || k.plant_id}
                        </span>
                      ) : (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold bg-purple-500/10 text-purple-600 border border-purple-500/20">
                          Global (All Plants)
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3 font-mono text-muted-foreground">{k.prefix}...</td>
                    <td className="px-4 py-3">
                      {k.is_single_use ? (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold bg-amber-500/10 text-amber-600 border border-amber-500/20">
                          Single-Use (1x Pull)
                        </span>
                      ) : (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold bg-blue-500/10 text-blue-600 border border-blue-500/20">
                          Multi-Use
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {k.is_active ? (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold bg-green-500/10 text-green-600 border border-green-500/20">
                          Aktif
                        </span>
                      ) : k.used_at ? (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold bg-slate-500/10 text-slate-500 border border-slate-500/20">
                          Sudah Digunakan (Expired)
                        </span>
                      ) : (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold bg-red-500/10 text-red-500 border border-red-500/20">
                          Dicabut (Revoked)
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">
                      {k.used_at ? new Date(k.used_at).toLocaleString("id-ID") : "-"}
                    </td>
                    <td className="px-4 py-3 text-right">
                      {k.is_active && (
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => revokeMutation.mutate(k.key_id)}
                          className="h-7 w-7 p-0 text-red-500 hover:text-red-600 hover:bg-red-500/10 rounded-lg"
                          title="Cabut Akses Key"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </Button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Interactive Datetime Filter Configurator for Power BI */}
      <div className="p-5 bg-card rounded-2xl border border-border shadow-sm space-y-4">
        <h4 className="font-semibold text-xs flex items-center gap-2 text-primary">
          <BookOpen className="w-4 h-4" /> Pengaturan Datetime Filter Penarikan (Power BI URL Configurator)
        </h4>
        <p className="text-xs text-muted-foreground leading-relaxed">
          Atur batas tanggal minimal (parameter <code className="bg-muted px-1.5 py-0.5 rounded text-foreground font-mono">since</code>) untuk penarikan data incremental. URL di bawah akan terisi otomatis dengan tanggal yang Anda pilih.
        </p>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 items-end">
          <div>
            <label className="text-xs font-semibold block mb-1">Tarik Data Sejak Tanggal (Since):</label>
            <Input
              type="date"
              value={sinceDate}
              onChange={(e) => setSinceDate(e.target.value)}
              className="w-full"
            />
          </div>

          <div className="flex flex-wrap gap-1.5">
            <Button size="sm" variant="outline" onClick={() => setPresetDate(7)} className="text-[11px] h-9 rounded-xl">
              7 Hari Terakhir
            </Button>
            <Button size="sm" variant="outline" onClick={() => setPresetDate(30)} className="text-[11px] h-9 rounded-xl">
              30 Hari Terakhir
            </Button>
            <Button size="sm" variant="outline" onClick={() => setPresetDate(90)} className="text-[11px] h-9 rounded-xl">
              90 Hari Terakhir
            </Button>
            <Button size="sm" variant="outline" onClick={() => setPresetDate(null)} className="text-[11px] h-9 rounded-xl">
              Semua Data
            </Button>
          </div>
        </div>

        <div className="space-y-1.5 pt-2">
          <label className="text-xs font-semibold block">URL Endpoint Hasil Generator Power BI:</label>
          <div className="flex items-center gap-2 bg-muted/60 p-3 rounded-xl border border-border">
            <code className="text-xs font-mono break-all flex-1 text-foreground font-semibold">
              {getPowerBIUrl(sinceDate)}
            </code>
            <Button size="sm" variant="outline" onClick={() => handleCopyUrl(getPowerBIUrl(sinceDate))} className="shrink-0 rounded-lg">
              {copiedUrl ? <Check className="w-4 h-4 text-green-500" /> : <Copy className="w-4 h-4" />}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
