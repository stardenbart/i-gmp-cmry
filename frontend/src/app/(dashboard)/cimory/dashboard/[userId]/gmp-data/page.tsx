"use client";

import { useState } from "react";
import { ClipboardCheck, Search, X, Loader2, Download, Eye, Calendar, MapPin, Filter, RotateCcw, Layers, ListChecks } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/axios";
import { dashboardApi } from "@/types/api/dashboard";
import { areaApi } from "@/types/api/master";
import { fetchItems } from "@/components/master/master.api";
import { useAdminGuard } from "@/lib/useAdminGuard";
import { useMounted } from "@/lib/useMounted";
import { cn, formatImageUrl } from "@/lib/utils";
import { SearchLatencyBadge } from "@/components/ui/SearchLatencyBadge";

export default function GmpDataAdminPage() {
  const mounted = useMounted();
  const { isAdmin, isLoading: isGuardLoading } = useAdminGuard();
  
  const [selectedArea, setSelectedArea] = useState<string>("");
  const [selectedKawasan, setSelectedKawasan] = useState<string>("");
  const [selectedDetailKawasan, setSelectedDetailKawasan] = useState<string>("");
  const [startDate, setStartDate] = useState<string>("");
  const [endDate, setEndDate] = useState<string>("");
  const [q, setQ] = useState("");
  const [isExporting, setIsExporting] = useState(false);
  const [previewImage, setPreviewImage] = useState<string | null>(null);

  // Lookups
  const { data: areasResponse } = useQuery({
    queryKey: ["areas-master-admin-gmp"],
    queryFn: () => areaApi.getAll(1, 100),
    enabled: mounted && isAdmin,
  });

  const { data: kawasanLookup } = useQuery({
    queryKey: ["kawasan-master-admin-gmp"],
    queryFn: () => fetchItems("/master/kawasan", 1, "", 500),
    enabled: mounted && isAdmin,
  });

  const { data: detailKawasanLookup } = useQuery({
    queryKey: ["detail-kawasan-master-admin-gmp"],
    queryFn: () => fetchItems("/master/detail-kawasan", 1, "", 500),
    enabled: mounted && isAdmin,
  });

  // Filtered dropdown options (cascading)
  const kawasanOptions = kawasanLookup?.items?.filter((k: any) => 
    !selectedArea || k.area_id === selectedArea || k.area?.area_id === selectedArea
  ) || [];

  const detailKawasanOptions = detailKawasanLookup?.items?.filter((dk: any) => 
    !selectedKawasan || dk.kawasan_id === selectedKawasan
  ) || [];

  // Main data preview query
  const { data: previewData, isLoading: isPreviewLoading } = useQuery({
    queryKey: ["dashboard-preview-export", selectedArea, selectedKawasan, selectedDetailKawasan, startDate, endDate],
    queryFn: () => dashboardApi.getPreviewExport({
      area_id: selectedArea,
      kawasan_id: selectedKawasan,
      detail_kawasan_id: selectedDetailKawasan,
      start_date: startDate,
      end_date: endDate,
    }),
    enabled: mounted && isAdmin,
  });

  const handleAreaChange = (val: string) => {
    setSelectedArea(val);
    setSelectedKawasan("");
    setSelectedDetailKawasan("");
  };

  const handleKawasanChange = (val: string) => {
    setSelectedKawasan(val);
    setSelectedDetailKawasan("");
  };

  const handleResetFilters = () => {
    setSelectedArea("");
    setSelectedKawasan("");
    setSelectedDetailKawasan("");
    setStartDate("");
    setEndDate("");
    setQ("");
  };

  const hasActiveFilters = Boolean(selectedArea || selectedKawasan || selectedDetailKawasan || startDate || endDate || q);

  const handleExport = async () => {
    try {
      setIsExporting(true);
      const response = await api.get("/dashboard/export", { 
        params: { 
          area_id: selectedArea,
          kawasan_id: selectedKawasan,
          detail_kawasan_id: selectedDetailKawasan,
          start_date: startDate,
          end_date: endDate,
        },
        responseType: 'blob' 
      });
      
      const areaName = areasResponse?.data?.items?.find((a: any) => a.area_id === selectedArea)?.area_name || "Semua Area";
      const datetime = new Date().toISOString().replace(/T/, '_').replace(/\..+/, '').replace(/:/g, '');
      const filename = `Report_${areaName.replace(/\s+/g, '_')}_${datetime}.xlsx`;
      
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', filename);
      document.body.appendChild(link);
      link.click();
      link.parentNode?.removeChild(link);
    } catch (error) {
      console.error("Failed to export report", error);
    } finally {
      setIsExporting(false);
    }
  };

  if (isGuardLoading) {
    return (
      <div className="flex justify-center p-8">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!isAdmin) return null;

  // Search logic (Client side across all fields)
  const filteredData = previewData?.data?.filter((row: any) => {
    if (!q) return true;
    const searchLower = q.toLowerCase();
    return (
      row.inspection_id?.toLowerCase().includes(searchLower) ||
      row.area?.toLowerCase().includes(searchLower) ||
      row.kawasan?.toLowerCase().includes(searchLower) ||
      row.detail_kawasan?.toLowerCase().includes(searchLower) ||
      row.pic?.toLowerCase().includes(searchLower) ||
      row.aspek?.toLowerCase().includes(searchLower) ||
      row.detail?.toLowerCase().includes(searchLower) ||
      row.uraian_id?.toLowerCase().includes(searchLower) ||
      row.keterangan?.toLowerCase().includes(searchLower)
    );
  }) || [];

  return (
    <div className="space-y-6 animate-in fade-in duration-500 pb-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-foreground">Data Inspeksi (GMP)</h2>
          <p className="text-muted-foreground text-sm">
            {isPreviewLoading ? "Memuat..." : `${filteredData.length} data inspeksi temuan (results)`}
          </p>
        </div>
        <Button 
          onClick={handleExport}
          disabled={isExporting}
          className="flex items-center gap-2 bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm disabled:opacity-70"
        >
          {isExporting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Download className="h-4 w-4" />}
          {isExporting ? "Mengekspor..." : "Export Report"}
        </Button>
      </div>

      {/* Filter Bar */}
      <Card className="p-4 bg-card/60 backdrop-blur-md border-border/50 space-y-3">
        {/* Search & Latency Row */}
        <div className="flex flex-col md:flex-row gap-4 items-stretch md:items-center">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
            <Input
              placeholder="Cari Inspeksi ID, Aspek, Detail, Uraian ID, Area, Kawasan, PIC, Keterangan..."
              value={q}
              onChange={e => setQ(e.target.value)}
              className="pl-9 bg-background/50"
            />
          </div>
          
          <SearchLatencyBadge
            searchQuery={q}
            isFetching={isPreviewLoading}
            pageName="Data Inspeksi (GMP)"
            apiPath="/dashboard/stats"
          />
        </div>

        {/* Multi-Filter Controls Row */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3 pt-2 border-t border-border/40">
          {/* Area Filter */}
          <div className="space-y-1">
            <label className="text-xs font-semibold text-muted-foreground flex items-center gap-1">
              <MapPin className="h-3 w-3 text-primary" /> Area:
            </label>
            <select 
              value={selectedArea} 
              onChange={(e) => handleAreaChange(e.target.value)}
              className="w-full px-3 py-2 border border-border rounded-md text-sm bg-background text-foreground h-9 focus:ring-1 focus:ring-primary"
            >
              <option value="">Semua Area</option>
              {areasResponse?.data?.items?.map((area: any) => (
                <option key={area.area_id} value={area.area_id}>
                  {area.area_name}
                </option>
              ))}
            </select>
          </div>

          {/* Kawasan Filter */}
          <div className="space-y-1">
            <label className="text-xs font-semibold text-muted-foreground flex items-center gap-1">
              <Layers className="h-3 w-3 text-primary" /> Kawasan:
            </label>
            <select 
              value={selectedKawasan} 
              onChange={(e) => handleKawasanChange(e.target.value)}
              disabled={!selectedArea && kawasanOptions.length === 0}
              className="w-full px-3 py-2 border border-border rounded-md text-sm bg-background text-foreground h-9 focus:ring-1 focus:ring-primary disabled:opacity-50"
            >
              <option value="">Semua Kawasan</option>
              {kawasanOptions.map((kawasan: any) => (
                <option key={kawasan.kawasan_id} value={kawasan.kawasan_id}>
                  {kawasan.kawasan_name}
                </option>
              ))}
            </select>
          </div>

          {/* Detail Kawasan Filter */}
          <div className="space-y-1">
            <label className="text-xs font-semibold text-muted-foreground flex items-center gap-1">
              <ListChecks className="h-3 w-3 text-primary" /> Detail Kawasan:
            </label>
            <select 
              value={selectedDetailKawasan} 
              onChange={(e) => setSelectedDetailKawasan(e.target.value)}
              disabled={!selectedKawasan && detailKawasanOptions.length === 0}
              className="w-full px-3 py-2 border border-border rounded-md text-sm bg-background text-foreground h-9 focus:ring-1 focus:ring-primary disabled:opacity-50"
            >
              <option value="">Semua Detail Kawasan</option>
              {detailKawasanOptions.map((dk: any) => (
                <option key={dk.detail_kawasan_id} value={dk.detail_kawasan_id}>
                  {dk.detail_kawasan_name}
                </option>
              ))}
            </select>
          </div>

          {/* Start Date Filter */}
          <div className="space-y-1">
            <label className="text-xs font-semibold text-muted-foreground flex items-center gap-1">
              <Calendar className="h-3 w-3 text-primary" /> Tanggal Mulai:
            </label>
            <Input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              className="h-9 text-xs bg-background"
            />
          </div>

          {/* End Date Filter & Reset Button */}
          <div className="space-y-1">
            <label className="text-xs font-semibold text-muted-foreground flex items-center gap-1">
              <Calendar className="h-3 w-3 text-primary" /> Tanggal Selesai:
            </label>
            <div className="flex gap-2">
              <Input
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                className="h-9 text-xs bg-background flex-1"
              />
              {hasActiveFilters && (
                <Button 
                  variant="outline" 
                  size="sm" 
                  onClick={handleResetFilters} 
                  title="Reset Semua Filter"
                  className="h-9 px-2 shrink-0 border-destructive/30 hover:bg-destructive/10 text-destructive"
                >
                  <RotateCcw className="h-3.5 w-3.5" />
                </Button>
              )}
            </div>
          </div>
        </div>
      </Card>

      {/* Table Section */}
      <Card className="overflow-hidden border-border/50 shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-sm text-left">
            <thead className="bg-muted/50 text-muted-foreground border-b border-border">
              <tr>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap">ID Inspeksi</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap">Kawasan / Detail</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap">Aspek</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap">Detail Aspek</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap">Uraian ID</th>
                <th className="px-4 py-3.5 font-semibold text-center whitespace-nowrap" title="Nilai per Detail Kawasan">Nilai</th>
                <th className="px-4 py-3.5 font-semibold text-center whitespace-nowrap" title="Total Nilai per Kawasan">Total Nilai (Kawasan)</th>
                <th className="px-4 py-3.5 font-semibold text-center whitespace-nowrap" title="Total Temuan per Kawasan">Total Temuan (Kawasan)</th>
                <th className="px-4 py-3.5 font-semibold text-center whitespace-nowrap" title="Visual Gambar per Detail Kawasan">Visual Gambar</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap" title="Keterangan per Detail Kawasan">Keterangan</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap" title="Follow Up Datetime per Detail Kawasan">Follow Up Datetime</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap" title="Due Date per Detail Kawasan">Due Date</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/50">
              {isPreviewLoading ? (
                <tr>
                  <td colSpan={11} className="px-6 py-12 text-center text-muted-foreground">
                    <Loader2 className="h-8 w-8 animate-spin mx-auto mb-3 text-primary/50" />
                    Memuat data inspeksi GMP...
                  </td>
                </tr>
              ) : filteredData.length === 0 ? (
                <tr>
                  <td colSpan={12} className="px-6 py-12">
                    <div className="w-full flex flex-col items-center justify-center text-center max-w-lg mx-auto py-4">
                      <ClipboardCheck className="h-12 w-12 text-muted-foreground/30 mb-3" />
                      <h3 className="text-lg font-semibold text-foreground">Tidak ada data</h3>
                      <p className="text-sm text-muted-foreground mt-1.5 w-full leading-relaxed">
                        Data inspeksi GMP untuk area yang dipilih tidak ditemukan, atau tidak cocok dengan kata kunci pencarian Anda.
                      </p>
                    </div>
                  </td>
                </tr>
              ) : (
                filteredData.map((row: any, idx: number) => (
                  <tr key={idx} className="hover:bg-muted/30 transition-colors">
                    {/* ID Inspeksi & Info Header */}
                    <td className="px-4 py-3 font-medium text-primary whitespace-nowrap">
                      <div className="flex flex-col">
                        <span className="font-semibold">{row.inspection_id}</span>
                        <span className="text-[11px] text-muted-foreground flex items-center gap-1 mt-0.5">
                          <MapPin className="h-3 w-3 inline" /> {row.area || "-"}
                        </span>
                      </div>
                    </td>

                    {/* Kawasan & Detail Kawasan */}
                    <td className="px-4 py-3 whitespace-nowrap">
                      <div className="flex flex-col">
                        <span className="font-medium text-foreground">{row.kawasan || "-"}</span>
                        <span className="text-[11px] text-muted-foreground">{row.detail_kawasan || "-"}</span>
                      </div>
                    </td>

                    {/* Aspek */}
                    <td className="px-4 py-3 font-medium whitespace-nowrap">{row.aspek || "-"}</td>

                    {/* Detail Aspek */}
                    <td className="px-4 py-3 whitespace-nowrap text-muted-foreground">{row.detail || "-"}</td>

                    {/* Uraian ID */}
                    <td className="px-4 py-3 font-mono text-xs text-muted-foreground whitespace-nowrap">{row.uraian_id || "-"}</td>

                    {/* Nilai */}
                    <td className="px-4 py-3 text-center whitespace-nowrap">
                      <span className={cn(
                        "px-2.5 py-1 rounded-full text-xs font-bold border",
                        row.nilai >= 80 ? "bg-green-500/10 text-green-600 border-green-500/20" :
                        row.nilai >= 60 ? "bg-yellow-500/10 text-yellow-600 border-yellow-500/20" :
                        "bg-red-500/10 text-red-600 border-red-500/20"
                      )}>
                        {row.nilai}
                      </span>
                    </td>

                    {/* Total Nilai */}
                    <td className="px-4 py-3 text-center whitespace-nowrap font-semibold">
                      <span className="px-2 py-0.5 rounded bg-muted/60 text-xs">
                        {row.total_nilai ?? "-"}
                      </span>
                    </td>

                    {/* Total Temuan */}
                    <td className="px-4 py-3 text-center whitespace-nowrap">
                      <span className={cn(
                        "px-2.5 py-0.5 rounded-full text-xs font-bold border",
                        (row.total_temuan ?? 0) > 0 
                          ? "bg-red-500/10 text-red-500 border-red-500/20" 
                          : "bg-muted text-muted-foreground border-transparent"
                      )}>
                        {row.total_temuan ?? 0}
                      </span>
                    </td>

                    {/* Visual Gambar (Issue) */}
                    <td className="px-4 py-3 text-center whitespace-nowrap">
                      {row.image_url ? (
                        <div 
                          className="relative group h-12 w-16 mx-auto overflow-hidden rounded-lg border border-border/60 shadow-sm cursor-pointer bg-muted"
                          onClick={() => setPreviewImage(formatImageUrl(row.image_url))}
                        >
                          {/* eslint-disable-next-line @next/next/no-img-element */}
                          <img 
                            src={formatImageUrl(row.image_url)} 
                            alt="Issue Photo" 
                            className="h-full w-full object-cover transition-transform group-hover:scale-110" 
                          />
                          <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center text-white">
                            <Eye className="h-3.5 w-3.5" />
                          </div>
                        </div>
                      ) : (
                        <span className="text-xs text-muted-foreground italic">-</span>
                      )}
                    </td>

                    {/* Keterangan */}
                    <td className="px-4 py-3 max-w-xs truncate" title={row.keterangan}>
                      {row.keterangan || <span className="text-muted-foreground italic">-</span>}
                    </td>

                    {/* Follow Up Datetime */}
                    <td className="px-4 py-3 whitespace-nowrap text-xs text-muted-foreground">
                      {row.follow_up_date ? (
                        <span className="flex items-center gap-1">
                          <Calendar className="h-3 w-3 text-primary" />
                          {new Date(row.follow_up_date).toLocaleString("id-ID", {
                            day: "numeric", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit"
                          })}
                        </span>
                      ) : "-"}
                    </td>

                    {/* Due Date */}
                    <td className="px-4 py-3 whitespace-nowrap text-xs text-muted-foreground">
                      {row.due_date ? (
                        <span className="flex items-center gap-1 font-medium text-foreground">
                          <Calendar className="h-3 w-3 text-destructive" />
                          {new Date(row.due_date).toLocaleDateString("id-ID", {
                            day: "numeric", month: "short", year: "numeric"
                          })}
                        </span>
                      ) : "-"}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Lightbox Image Preview Modal */}
      {previewImage && (
        <div 
          className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 animate-in fade-in duration-200"
          onClick={() => setPreviewImage(null)}
        >
          <div className="relative max-w-3xl max-h-[85vh] overflow-hidden rounded-2xl border border-white/20 bg-card shadow-2xl p-2" onClick={e => e.stopPropagation()}>
            <button
              onClick={() => setPreviewImage(null)}
              className="absolute top-4 right-4 z-10 p-2 rounded-full bg-black/60 text-white hover:bg-black/80 transition-colors"
            >
              <X className="h-5 w-5" />
            </button>
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img 
              src={formatImageUrl(previewImage)} 
              alt="Visual Issue Large" 
              className="max-h-[80vh] w-auto object-contain rounded-xl" 
            />
          </div>
        </div>
      )}
    </div>
  );
}
