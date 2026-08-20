"use client";

import { useState, useEffect } from "react";
import { ClipboardCheck, Search, X, Loader2, Download, Eye, Calendar, MapPin, Filter, RotateCcw, Layers, ListChecks, AlertTriangle, Building2 } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/axios";
import { dashboardApi } from "@/types/api/dashboard";
import { areaApi, plantApi } from "@/types/api/master";
import { fetchItems } from "@/components/master/master.api";
import { usePermissions } from "@/lib/usePermissions";
import { useMounted } from "@/lib/useMounted";
import { cn, formatImageUrl } from "@/lib/utils";
import { SearchLatencyBadge } from "@/components/ui/SearchLatencyBadge";
import { useAuthStore } from "@/stores/authStore";
import { useDebounce } from "@/hooks/useDebounce";

import { useAppDispatch, useAppSelector } from "@/store/hooks";
import {
  setSelectedArea,
  setSelectedKawasan,
  setSelectedDetailKawasan,
  setStartDate,
  setEndDate,
  setQ,
  resetFilters,
} from "@/store/slices/gmpFilterSlice";

import { useParams } from "next/navigation";

export default function GmpDataAdminPage() {
  const mounted = useMounted();
  const params = useParams();
  const plantCode = (params?.plantCode as string) || "";
  const effectivePlantId = (plantCode && plantCode !== "all" && plantCode !== "global") ? plantCode : undefined;

  const user = useAuthStore((state) => state.user);
  const isSuperAdmin = user?.role_id === "ROLE-000" || user?.role_id === "SUPERADMIN" || user?.role?.role_name === "Super Admin";

  const { hasPermission, isLoading: isGuardLoading } = usePermissions();
  const isAdmin = hasPermission("PERM-MSTR-R");
  
  const dispatch = useAppDispatch();
  const {
    selectedArea,
    selectedKawasan,
    selectedDetailKawasan,
    startDate,
    endDate,
    q,
  } = useAppSelector((state) => state.gmpFilter);

  const [searchInputValue, setSearchInputValue] = useState(q || "");
  const debouncedSearchValue = useDebounce(searchInputValue, 400);

  useEffect(() => {
    dispatch(setQ(debouncedSearchValue));
  }, [debouncedSearchValue, dispatch]);

  const [selectedPlant, setSelectedPlant] = useState<string>(effectivePlantId || "");
  const [isExporting, setIsExporting] = useState(false);
  const [previewImage, setPreviewImage] = useState<string | null>(null);

  const userPlantId = user?.plant_id;
  const activePlantId = isSuperAdmin ? (selectedPlant || effectivePlantId) : (userPlantId || effectivePlantId);

  // Lookups
  const { data: plantsResponse } = useQuery({
    queryKey: ["plants-master-admin-gmp"],
    queryFn: () => plantApi.getAll(1, 500),
    enabled: mounted && isAdmin && isSuperAdmin,
  });

  const plantItems = Array.isArray(plantsResponse?.items)
    ? plantsResponse.items
    : Array.isArray(plantsResponse?.data?.items)
    ? plantsResponse.data.items
    : Array.isArray(plantsResponse)
    ? plantsResponse
    : [];

  const { data: areasResponse } = useQuery({
    queryKey: ["areas-master-admin-gmp", activePlantId],
    queryFn: () => areaApi.getAll(1, 100),
    enabled: mounted && isAdmin,
  });

  const allAreaItems = areasResponse?.data?.items || (areasResponse as any)?.items || [];
  const filteredAreaOptions = allAreaItems.filter((area: any) =>
    !activePlantId || area.plant_id === activePlantId
  );

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
    queryKey: ["dashboard-preview-export", selectedArea, selectedKawasan, selectedDetailKawasan, startDate, endDate, activePlantId],
    queryFn: () => dashboardApi.getPreviewExport({
      area_id: selectedArea,
      kawasan_id: selectedKawasan,
      detail_kawasan_id: selectedDetailKawasan,
      start_date: startDate,
      end_date: endDate,
      plant_id: activePlantId,
    }),
    enabled: mounted && isAdmin,
  });

  const handlePlantChange = (val: string) => {
    setSelectedPlant(val);
    dispatch(setSelectedArea(""));
    dispatch(setSelectedKawasan(""));
    dispatch(setSelectedDetailKawasan(""));
  };

  const handleAreaChange = (val: string) => {
    dispatch(setSelectedArea(val));
  };

  const handleKawasanChange = (val: string) => {
    dispatch(setSelectedKawasan(val));
  };

  const handleResetFilters = () => {
    setSelectedPlant("");
    setSearchInputValue("");
    dispatch(resetFilters());
  };

  const hasActiveFilters = Boolean(selectedPlant || selectedArea || selectedKawasan || selectedDetailKawasan || startDate || endDate || q);

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
          plant_id: activePlantId,
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
      console.error("Export error:", error);
    } finally {
      setIsExporting(false);
    }
  };

  if (!mounted || isGuardLoading) return <div className="h-64 flex justify-center items-center"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div></div>;

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
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
            <ClipboardCheck className="h-6 w-6 text-primary" /> Data Inspeksi (GMP)
          </h1>
          <p className="text-sm text-muted-foreground">
            Laporan lengkap hasil inspeksi kebersihan & kelayakan pabrik
          </p>
        </div>
        
        <Button 
          onClick={handleExport} 
          disabled={isExporting || isPreviewLoading}
          className="bg-primary text-primary-foreground hover:bg-primary/90"
        >
          {isExporting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Download className="mr-2 h-4 w-4" />}
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
              value={searchInputValue}
              onChange={e => setSearchInputValue(e.target.value)}
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
        <div className={cn(
          "grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3 pt-2 border-t border-border/40",
          isSuperAdmin ? "xl:grid-cols-6" : "xl:grid-cols-5"
        )}>
          {/* Plant Filter (Super Admin Only) */}
          {isSuperAdmin && (
            <div className="space-y-1">
              <label className="text-xs font-semibold text-muted-foreground flex items-center gap-1">
                <Building2 className="h-3 w-3 text-primary" /> Pabrik / Plant:
              </label>
              <select 
                value={selectedPlant} 
                onChange={(e) => handlePlantChange(e.target.value)}
                className="w-full px-3 py-2 border border-border rounded-md text-sm bg-background text-foreground h-9 focus:ring-1 focus:ring-primary"
              >
                <option value="">Semua Plant (Global)</option>
                {plantItems.map((plant: any) => (
                  <option key={plant.plant_id || plant.plant_code} value={plant.plant_id || plant.plant_code}>
                    {plant.plant_name || plant.plant_code || plant.plant_id}
                  </option>
                ))}
              </select>
            </div>
          )}

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
              {filteredAreaOptions.map((area: any) => (
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
              onChange={(e) => dispatch(setSelectedDetailKawasan(e.target.value))}
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
              onChange={(e) => dispatch(setStartDate(e.target.value))}
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
                onChange={(e) => dispatch(setEndDate(e.target.value))}
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
          <table className="w-full min-w-[1400px] text-sm text-left border-collapse">
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
                  <td colSpan={12} className="px-6 py-12 text-center text-muted-foreground">
                    <Loader2 className="h-8 w-8 animate-spin mx-auto mb-3 text-primary/50" />
                    Memuat data inspeksi GMP...
                  </td>
                </tr>
              ) : filteredData.length === 0 ? (
                <tr>
                  <td colSpan={12} className="p-6 sm:p-10">
                    <div className="w-full rounded-3xl border border-dashed border-border/70 bg-gradient-to-b from-card/80 via-card/40 to-background p-8 sm:p-12 text-center shadow-sm">
                      <div className="mx-auto w-full max-w-md text-center space-y-4" style={{ width: "100%", maxWidth: "28rem", marginLeft: "auto", marginRight: "auto" }}>
                        <div className="inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-500/10 border border-amber-500/20 text-amber-500 shadow-inner mx-auto mb-2">
                          <AlertTriangle className="h-8 w-8 text-amber-500" />
                        </div>

                        <h3 className="w-full text-lg font-bold text-foreground tracking-tight text-center block">
                          {q || hasActiveFilters ? "Tidak Ada Data Inspeksi Ditemukan" : "Belum Ada Data Inspeksi"}
                        </h3>

                        <p className="w-full text-sm text-muted-foreground leading-relaxed text-center block" style={{ wordBreak: "normal", overflowWrap: "break-word" }}>
                          {hasActiveFilters
                            ? "Data inspeksi GMP untuk area atau filter yang dipilih tidak ditemukan, atau tidak cocok dengan kata kunci pencarian Anda."
                            : "Data inspeksi GMP akan secara otomatis tercatat ketika ada sesi inspeksi yang telah dilaksanakan."}
                        </p>

                        {hasActiveFilters && (
                          <div className="w-full flex items-center justify-center pt-2">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={handleResetFilters}
                              className="rounded-full px-5 h-9 text-xs font-semibold"
                            >
                              <X className="mr-1.5 h-3.5 w-3.5" /> Hapus Filter
                            </Button>
                          </div>
                        )}
                      </div>
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
                            src={formatImageUrl(row.image_url) || "/placeholder.png"} 
                            alt="Issue Photo" 
                            onError={(e) => { e.currentTarget.src = "/placeholder.png"; }}
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
              src={formatImageUrl(previewImage) || "/placeholder.png"} 
              alt="Visual Issue Large" 
              onError={(e) => { e.currentTarget.src = "/placeholder.png"; }}
              className="max-h-[80vh] w-auto object-contain rounded-xl" 
            />
          </div>
        </div>
      )}
    </div>
  );
}
