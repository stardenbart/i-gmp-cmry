"use client";

import { useState, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import Link from "next/link";
import {
  CheckCircle2,
  Clock,
  XCircle,
  Search,
  ArrowLeft,
  Loader2,
  FileSpreadsheet,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { issueApi } from "@/lib/api/issue.api";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { toast } from "sonner";

export default function WOWRReportPage() {
  const { plantCode, userId } = useParams() as { plantCode: string; userId: string };
  const user = useAuthStore(state => state.user);
  const mounted = useMounted();
  const [search, setSearch] = useState("");
  const [selectedArea, setSelectedArea] = useState<string>("ALL");

  const { data, isLoading } = useQuery({
    queryKey: ["wowr-report", userId, plantCode],
    queryFn: () => issueApi.getWOWRReport(),
    enabled: mounted && !!user,
  });

  const summary = data?.summary || {
    total: 0,
    verified: 0,
    pending: 0,
    rejected: 0,
    awaiting: 0,
    verified_rate: 0,
    pending_rate: 0,
    rejected_rate: 0,
    awaiting_rate: 0,
  };

  const byArea: Array<{
    area_name: string;
    total: number;
    verified: number;
    pending: number;
    rejected: number;
    awaiting: number;
  }> = data?.by_area || [];

  // Filter items by Area and Search text
  const filteredItems = useMemo(() => {
    const rawItems: Array<{
      issue_id: string;
      photo_id?: string;
      wo_id: string;
      wr_id: string;
      needs_wo_wr: boolean;
      wowr_status: string;
      issue_status: string;
      area_id: string;
      area_name: string;
      kawasan_id: string;
      kawasan_name: string;
      detail_kawasan_name: string;
      pic_name: string;
      keterangan: string;
      due_date?: string;
      created_at: string;
    }> = data?.items || [];
    let result = rawItems;
    if (selectedArea !== "ALL") {
      result = result.filter(i => i.area_name === selectedArea);
    }
    if (search.trim()) {
      const q = search.toLowerCase();
      result = result.filter(
        i =>
          i.issue_id.toLowerCase().includes(q) ||
          i.wo_id.toLowerCase().includes(q) ||
          i.wr_id.toLowerCase().includes(q) ||
          i.area_name.toLowerCase().includes(q) ||
          i.kawasan_name.toLowerCase().includes(q) ||
          i.pic_name.toLowerCase().includes(q) ||
          i.keterangan.toLowerCase().includes(q)
      );
    }
    return result;
  }, [data, selectedArea, search]);

  // Client-side Excel/CSV Export generator
  const handleExportCSV = () => {
    if (filteredItems.length === 0) {
      toast.error("Tidak ada data WO/WR yang tersedia untuk diexport.");
      return;
    }

    const headers = [
      "No",
      "Issue ID",
      "Nomor WO",
      "Nomor WR",
      "Status WO/WR",
      "Status Issue",
      "Area",
      "Kawasan",
      "Detail Kawasan",
      "PIC Name",
      "Keterangan",
      "Tanggal Due Date",
      "Tanggal Dibuat"
    ];

    const rows = filteredItems.map((item, idx) => [
      idx + 1,
      `"${item.issue_id}"`,
      `"${item.wo_id || "-"}"`,
      `"${item.wr_id || "-"}"`,
      `"${item.wowr_status}"`,
      `"${item.issue_status}"`,
      `"${item.area_name}"`,
      `"${item.kawasan_name}"`,
      `"${item.detail_kawasan_name}"`,
      `"${item.pic_name}"`,
      `"${(item.keterangan || "").replace(/"/g, '""')}"`,
      `"${item.due_date ? new Date(item.due_date).toLocaleDateString("id-ID") : "-"}"`,
      `"${new Date(item.created_at).toLocaleDateString("id-ID")}"`
    ]);

    const csvContent = "data:text/csv;charset=utf-8,\uFEFF" + [headers.join(","), ...rows.map(e => e.join(","))].join("\n");
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement("a");
    link.setAttribute("href", encodedUri);
    link.setAttribute("download", `Report_WOWR_${new Date().toISOString().slice(0, 10)}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    toast.success("File Laporan WO/WR (.csv) berhasil diunduh!");
  };

  if (!mounted || !user) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="space-y-6 pb-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      {/* ── Top Header ── */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 sm:gap-4 bg-card p-3.5 sm:p-5 rounded-2xl border shadow-sm min-w-0">
        <div className="flex items-center gap-3 sm:gap-4 min-w-0">
          <Link href={`/cimory/${plantCode}/dashboard/${userId}/wowr`}>
            <Button variant="outline" size="icon" className="h-10 w-10 rounded-xl">
              <ArrowLeft className="h-5 w-5" />
            </Button>
          </Link>
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-lg sm:text-xl font-bold tracking-tight break-words">Laporan & Analytics WO / WR</h2>
              <span className="bg-primary/10 text-primary border border-primary/20 text-[10px] font-bold px-2 py-0.5 rounded-full">
                Real-Time KPI
              </span>
            </div>
            <p className="text-xs sm:text-sm text-muted-foreground mt-0.5 break-words">
              Analisis tingkat verifikasi dan performa pengerjaan Maintenance Work Order & Work Request
            </p>
          </div>
        </div>

        <Button 
          onClick={handleExportCSV}
          className="w-full sm:w-auto rounded-xl gap-2 font-semibold shadow-sm bg-emerald-600 hover:bg-emerald-700 text-white h-10 sm:h-9 text-xs sm:text-sm"
        >
          <FileSpreadsheet className="h-4 w-4" />
          Export Laporan Excel (.csv)
        </Button>
      </div>

      {/* ── Summary Key Performance Cards ── */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
        {/* Total Card */}
        <div className="bg-card p-3 sm:p-5 rounded-2xl border shadow-sm space-y-2 relative overflow-hidden min-w-0">
          <div className="flex items-center justify-between">
            <span className="text-[10px] sm:text-xs text-muted-foreground font-semibold uppercase tracking-wider leading-tight">Total Temuan WO/WR</span>
          </div>
          <div className="flex items-end justify-between gap-2 pt-1">
            <span className="text-2xl sm:text-3xl font-bold font-mono">{summary.total}</span>
            <span className="text-[10px] sm:text-xs font-semibold text-muted-foreground">100% Total</span>
          </div>
          <p className="text-[10px] sm:text-[11px] text-muted-foreground line-clamp-2">Temuan aktif yang memerlukan Work Order / Work Request</p>
        </div>

        {/* Verified Rate Card */}
        <div className="bg-card p-3 sm:p-5 rounded-2xl border shadow-sm space-y-2 relative overflow-hidden min-w-0">
          <div className="flex items-center justify-between">
            <span className="text-[10px] sm:text-xs text-muted-foreground font-semibold uppercase tracking-wider leading-tight">Tingkat Verifikasi</span>
          </div>
          <div className="flex items-end justify-between gap-2 pt-1">
            <span className="text-2xl sm:text-3xl font-bold font-mono text-emerald-600 dark:text-emerald-400">
              {summary.verified_rate}%
            </span>
            <span className="text-[10px] sm:text-xs font-bold text-emerald-600 dark:text-emerald-400">
              {summary.verified} Disetujui
            </span>
          </div>
          <div className="w-full bg-secondary h-2 rounded-full overflow-hidden">
            <div className="bg-emerald-500 h-2 rounded-full transition-all" style={{ width: `${summary.verified_rate}%` }} />
          </div>
        </div>

        {/* Pending Rate Card */}
        <div className="bg-card p-3 sm:p-5 rounded-2xl border shadow-sm space-y-2 relative overflow-hidden min-w-0">
          <div className="flex items-center justify-between">
            <span className="text-[10px] sm:text-xs text-muted-foreground font-semibold uppercase tracking-wider leading-tight">Menunggu Validasi</span>
          </div>
          <div className="flex items-end justify-between gap-2 pt-1">
            <span className="text-2xl sm:text-3xl font-bold font-mono text-purple-600 dark:text-purple-400">
              {summary.pending_rate}%
            </span>
            <span className="text-[10px] sm:text-xs font-bold text-purple-600 dark:text-purple-400">
              {summary.pending} Menunggu
            </span>
          </div>
          <div className="w-full bg-secondary h-2 rounded-full overflow-hidden">
            <div className="bg-purple-500 h-2 rounded-full transition-all" style={{ width: `${summary.pending_rate}%` }} />
          </div>
        </div>

        {/* Rejected Rate Card */}
        <div className="bg-card p-3 sm:p-5 rounded-2xl border shadow-sm space-y-2 relative overflow-hidden min-w-0">
          <div className="flex items-center justify-between">
            <span className="text-[10px] sm:text-xs text-muted-foreground font-semibold uppercase tracking-wider leading-tight">Tingkat Penolakan</span>
          </div>
          <div className="flex items-end justify-between gap-2 pt-1">
            <span className="text-2xl sm:text-3xl font-bold font-mono text-red-600 dark:text-red-400">
              {summary.rejected_rate}%
            </span>
            <span className="text-[10px] sm:text-xs font-bold text-red-600 dark:text-red-400">
              {summary.rejected} Ditolak
            </span>
          </div>
          <div className="w-full bg-secondary h-2 rounded-full overflow-hidden">
            <div className="bg-red-500 h-2 rounded-full transition-all" style={{ width: `${summary.rejected_rate}%` }} />
          </div>
        </div>
      </div>

      {/* ── Area Breakdown Section ── */}
      <div className="bg-card p-3.5 sm:p-6 rounded-2xl border shadow-sm space-y-3 sm:space-y-4 min-w-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <h3 className="text-sm sm:text-base font-bold tracking-tight break-words">Breakdown Verifikasi per Area Operational</h3>
          </div>
          <span className="text-xs text-muted-foreground font-medium">Persentase Penyelesaian</span>
        </div>

        {byArea.length === 0 ? (
          <p className="text-xs text-muted-foreground italic text-center py-4">Belum ada data breakdown area.</p>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 sm:gap-4 pt-1 sm:pt-2">
            {byArea.map((area, idx) => {
              const rate = area.total > 0 ? (area.verified / area.total) * 100 : 0;
              return (
                <div key={idx} className="p-3 sm:p-4 rounded-xl bg-muted/30 border border-border/50 space-y-2 min-w-0">
                  <div className="flex items-center justify-between text-sm">
                    <span className="font-semibold">{area.area_name}</span>
                    <span className="font-mono font-bold text-primary">{rate.toFixed(1)}%</span>
                  </div>
                  <div className="w-full bg-secondary h-2.5 rounded-full overflow-hidden flex">
                    <div 
                      className="bg-emerald-500 h-2.5 transition-all" 
                      style={{ width: `${(area.verified / area.total) * 100}%` }}
                      title={`Verified: ${area.verified}`}
                    />
                    <div 
                      className="bg-purple-500 h-2.5 transition-all" 
                      style={{ width: `${(area.pending / area.total) * 100}%` }}
                      title={`Pending: ${area.pending}`}
                    />
                    <div 
                      className="bg-red-500 h-2.5 transition-all" 
                      style={{ width: `${(area.rejected / area.total) * 100}%` }}
                      title={`Rejected: ${area.rejected}`}
                    />
                  </div>
                  <div className="flex items-center justify-between text-[11px] text-muted-foreground pt-1">
                    <span>Total: <strong className="text-foreground">{area.total}</strong></span>
                    <div className="flex items-center gap-3">
                      <span className="text-emerald-600 font-medium">✓ {area.verified}</span>
                      <span className="text-purple-600 font-medium">{area.pending}</span>
                      <span className="text-red-600 font-medium">✗ {area.rejected}</span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* ── Table Section ── */}
      <div className="bg-card rounded-2xl border overflow-hidden shadow-sm space-y-3 sm:space-y-4 p-3.5 sm:p-5 min-w-0">
        <div className="flex flex-col lg:flex-row items-stretch lg:items-center justify-between gap-3 sm:gap-4">
          <h3 className="text-sm sm:text-base font-bold tracking-tight break-words">Detail Daftar Temuan WO / WR</h3>

          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-2.5 sm:gap-3 w-full lg:w-auto">
            {/* Area Filter */}
            <select
              value={selectedArea}
              onChange={(e) => setSelectedArea(e.target.value)}
              className="w-full sm:w-auto rounded-xl border bg-background px-3 py-2.5 text-xs focus:outline-none focus:ring-2 focus:ring-primary/40 font-medium"
            >
              <option value="ALL">Semua Area</option>
              {byArea.map((a, i) => (
                <option key={i} value={a.area_name}>{a.area_name}</option>
              ))}
            </select>

            {/* Search */}
            <div className="relative min-w-0 w-full sm:w-[220px] lg:w-[240px]">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
              <input
                type="search"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Cari WO/WR, PIC..."
                className="w-full rounded-xl border bg-background pl-8 pr-4 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary/40"
              />
            </div>
          </div>
        </div>

        <div className="overflow-x-auto rounded-xl border max-w-full">
          <table className="w-full min-w-[900px] text-xs text-left">
            <thead className="bg-muted/60 text-muted-foreground font-semibold uppercase">
              <tr>
                <th className="px-4 py-3">No</th>
                <th className="px-4 py-3">Issue ID</th>
                <th className="px-4 py-3">Nomor WO / WR</th>
                <th className="px-4 py-3">Area / Kawasan</th>
                <th className="px-4 py-3">PIC</th>
                <th className="px-4 py-3">Status Verifikasi</th>
                <th className="px-4 py-3">Keterangan</th>
                <th className="px-4 py-3">Tanggal Dibuat</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/50">
              {isLoading ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center">
                    <Loader2 className="h-6 w-6 animate-spin mx-auto text-primary" />
                    <p className="text-xs text-muted-foreground mt-2">Memuat data report WO/WR...</p>
                  </td>
                </tr>
              ) : filteredItems.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-muted-foreground italic">
                    Tidak ada data temuan WO/WR yang sesuai filter.
                  </td>
                </tr>
              ) : (
                filteredItems.map((item, idx) => (
                  <tr key={`wowr-row-${item.issue_id}-${item.photo_id || idx}`} className="hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3 font-mono text-muted-foreground">{idx + 1}</td>
                    <td className="px-4 py-3 font-mono font-semibold text-primary">
                      <span className="block">{item.issue_id}</span>
                      {item.photo_id && (
                        <span className="block text-[10px] font-normal text-muted-foreground">Foto: {item.photo_id}</span>
                      )}
                    </td>
                    <td className="px-4 py-3 font-bold text-foreground">
                      {item.wo_id || item.wr_id ? (
                        <span>{item.wo_id || item.wr_id}</span>
                      ) : (
                        <span className="text-muted-foreground font-normal italic">Belum diinput</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <div className="font-semibold">{item.area_name}</div>
                      <div className="text-[10px] text-muted-foreground">{item.kawasan_name}</div>
                    </td>
                    <td className="px-4 py-3 font-medium">{item.pic_name}</td>
                    <td className="px-4 py-3">
                      {item.wowr_status === "Verified" ? (
                        <span className="inline-flex items-center gap-1 font-bold px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">
                          <CheckCircle2 className="h-3 w-3" /> Terverifikasi
                        </span>
                      ) : item.wowr_status === "PendingValidation" ? (
                        <span className="inline-flex items-center gap-1 font-bold px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-500 border border-purple-500/20">
                          <Clock className="h-3 w-3 animate-spin" /> Menunggu Validasi
                        </span>
                      ) : item.wowr_status === "Rejected" ? (
                        <span className="inline-flex items-center gap-1 font-bold px-2 py-0.5 rounded-full bg-red-500/10 text-red-500 border border-red-500/20">
                          <XCircle className="h-3 w-3" /> Ditolak
                        </span>
                      ) : (
                        <span className="text-muted-foreground italic">Menunggu Bukti</span>
                      )}
                    </td>
                    <td className="px-4 py-3 max-w-xs truncate" title={item.keterangan}>
                      {item.keterangan || "-"}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground whitespace-nowrap">
                      {new Date(item.created_at).toLocaleDateString("id-ID", { day: "numeric", month: "short", year: "numeric" })}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
