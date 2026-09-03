"use client";

import { useState, useEffect } from "react";
import { ClipboardCheck, Search, X, Loader2, Calendar, MapPin, RotateCcw, Layers, ListChecks, AlertTriangle, Building2 } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { useQuery } from "@tanstack/react-query";
import { dashboardApi } from "@/types/api/dashboard";
import { areaApi, plantApi } from "@/types/api/master";
import type { Kawasan, DetailKawasan } from "@/types/api";
import { fetchItems } from "@/components/master/master.api";
import { usePermissions } from "@/lib/usePermissions";
import { useMounted } from "@/lib/useMounted";
import { cn, formatImageUrl } from "@/lib/utils";
import { SearchLatencyBadge } from "@/components/ui/SearchLatencyBadge";
import { useAuthStore } from "@/stores/authStore";
import { useDebounce } from "@/hooks/useDebounce";
import { GmpEvidenceImages, GmpFollowUpDescriptions } from "@/components/gmp/GmpEvidenceImages";
import { GmpExportMenu } from "@/components/gmp/GmpExportMenu";

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
import type { PreviewExportRow } from "@/types/api/dashboard";

/** Per-row rowSpan for each hierarchical "path" column — 0 means this row
 * is a continuation of the group started above it and must render NO <td>
 * for that column at all (that's what lets the cell above visually span
 * down over it); a value > 0 means this row starts a new group and should
 * render the cell with `rowSpan={value}`.
 *
 * Relies on the backend's fixed sort order (gmpDataRelationOrder in
 * gmp_export_helpers.go: InspectionID, then AspekName, DetailName,
 * Uraian) already keeping every row of one group physically adjacent —
 * this only merges rows that are already consecutive, never reorders them.
 *
 * `inspeksi` covers the "ID Inspeksi", "Kawasan", and "Detail Kawasan" columns:
 * Kawasan/DetailKawasan come from Inspection_Header, so they're constant
 * for every row sharing one inspection_id and always span identically.
 */
interface GmpRowSpans {
  inspeksi: number;
  aspek: number;
  detailAspek: number;
  uraian: number;
  totalNilaiKawasan: number;
  complianceDetailKawasan: number;
}

function computeGmpRowSpans(rows: PreviewExportRow[]): GmpRowSpans[] {
  const spans: GmpRowSpans[] = rows.map(() => ({
    inspeksi: 0,
    aspek: 0,
    detailAspek: 0,
    uraian: 0,
    totalNilaiKawasan: 0,
    complianceDetailKawasan: 0,
  }));

  const fillLevel = (key: keyof GmpRowSpans, sameGroup: (a: PreviewExportRow, b: PreviewExportRow) => boolean) => {
    let start = 0;
    for (let i = 1; i <= rows.length; i++) {
      if (i === rows.length || !sameGroup(rows[i], rows[start])) {
        spans[start][key] = i - start;
        start = i;
      }
    }
  };

  fillLevel("inspeksi", (a, b) => a.inspection_id === b.inspection_id);
  fillLevel("aspek", (a, b) => a.inspection_id === b.inspection_id && a.aspek === b.aspek);
  fillLevel("detailAspek", (a, b) => a.inspection_id === b.inspection_id && a.aspek === b.aspek && a.detail === b.detail);
  fillLevel("uraian", (a, b) => a.inspection_id === b.inspection_id && a.aspek === b.aspek && a.detail === b.detail && a.uraian_id === b.uraian_id);
  fillLevel("totalNilaiKawasan", (a, b) => (a.kawasan_id || a.kawasan) === (b.kawasan_id || b.kawasan));
  fillLevel("complianceDetailKawasan", (a, b) => (
    (a.kawasan_id || a.kawasan) === (b.kawasan_id || b.kawasan)
    && (a.detail_kawasan_id || a.detail_kawasan) === (b.detail_kawasan_id || b.detail_kawasan)
  ));

  return spans;
}

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
  const [previewImage, setPreviewImage] = useState<string | null>(null);

  const userPlantId = user?.plant_id;
  const activePlantId = isSuperAdmin ? (selectedPlant || effectivePlantId) : (userPlantId || effectivePlantId);

  // Lookups
  const { data: plantsResponse } = useQuery({
    queryKey: ["plants-master-admin-gmp"],
    queryFn: () => plantApi.getAll(1, 500),
    enabled: mounted && isAdmin && isSuperAdmin,
  });

  const plantItems = plantsResponse?.data?.items || [];

  const { data: areasResponse } = useQuery({
    queryKey: ["areas-master-admin-gmp", activePlantId],
    queryFn: () => areaApi.getAll(1, 100),
    enabled: mounted && isAdmin,
  });

  const allAreaItems = areasResponse?.data?.items || [];
  const filteredAreaOptions = allAreaItems.filter((area) =>
    !activePlantId || area.plant_id === activePlantId
  );

  const { data: kawasanLookup } = useQuery({
    queryKey: ["kawasan-master-admin-gmp"],
    queryFn: () => fetchItems<Kawasan>("/master/kawasan", 1, "", 500),
    enabled: mounted && isAdmin,
  });

  const { data: detailKawasanLookup } = useQuery({
    queryKey: ["detail-kawasan-master-admin-gmp"],
    queryFn: () => fetchItems<DetailKawasan>("/master/detail-kawasan", 1, "", 500),
    enabled: mounted && isAdmin,
  });

  // Filtered dropdown options (cascading)
  const kawasanOptions = kawasanLookup?.items?.filter((k) =>
    !selectedArea || k.area_id === selectedArea
  ) || [];

  const detailKawasanOptions = detailKawasanLookup?.items?.filter((dk) =>
    !selectedKawasan || dk.kawasan_id === selectedKawasan
  ) || [];

  // Main data preview query
  const { data: previewData, isLoading: isPreviewLoading } = useQuery({
    queryKey: ["dashboard-preview-export", selectedArea, selectedKawasan, selectedDetailKawasan, startDate, endDate, activePlantId, q],
    queryFn: () => dashboardApi.getPreviewExport({
      area_id: selectedArea,
      kawasan_id: selectedKawasan,
      detail_kawasan_id: selectedDetailKawasan,
      start_date: startDate,
      end_date: endDate,
      plant_id: activePlantId,
      q,
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

  if (!mounted || isGuardLoading) return <div className="h-64 flex justify-center items-center"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div></div>;

  if (!isAdmin) return null;

  // Search is applied by the backend so preview and both XLSX export modes
  // always contain the same rows.
  const filteredData = previewData?.data || [];
  const rowSpans = computeGmpRowSpans(filteredData);

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

        <GmpExportMenu
          disabled={isPreviewLoading || filteredData.length === 0}
          filters={{
            area_id: selectedArea || undefined,
            kawasan_id: selectedKawasan || undefined,
            detail_kawasan_id: selectedDetailKawasan || undefined,
            start_date: startDate || undefined,
            end_date: endDate || undefined,
            plant_id: activePlantId || undefined,
            q: q || undefined,
          }}
        />
      </div>

      {/* Filter Bar */}
      <Card className="p-4 bg-card/60 backdrop-blur-md border-border/50 space-y-3">
        {/* Search & Latency Row */}
        <div className="flex flex-col md:flex-row gap-4 items-stretch md:items-center">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
            <Input
              placeholder="Cari Inspeksi ID, Aspek, Detail, Uraian, Area, Kawasan, PIC, Keterangan..."
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
                {plantItems.map((plant) => (
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
              {filteredAreaOptions.map((area) => (
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
              {kawasanOptions.map((kawasan) => (
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
              {detailKawasanOptions.map((dk) => (
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
          <table className="w-full min-w-[2050px] text-sm text-left border-collapse">
            <thead className="bg-muted/50 text-muted-foreground border-b border-border">
              <tr>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap">ID Inspeksi</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap">Kawasan</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap">Detail Kawasan</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap">Aspek</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap">Detail Aspek</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap">Uraian</th>
                <th className="px-4 py-3.5 font-semibold text-center whitespace-nowrap" title="Nilai per Detail Kawasan">Nilai</th>
                <th className="px-4 py-3.5 font-semibold text-center whitespace-nowrap" title="Total Nilai per Kawasan">Total Nilai (Kawasan)</th>
                <th className="px-4 py-3.5 font-semibold text-center whitespace-nowrap" title="Total Nilai ÷ Total Nilai Maksimal (semua uraian OK) pada Detail Kawasan ini × 100">Persentase Kepatuhan (Detail Kawasan)</th>
                <th className="px-4 py-3.5 font-semibold text-center whitespace-nowrap" title="1 jika uraian ini memiliki Issue, 0 jika tidak">Temuan (Uraian)</th>
                <th className="px-4 py-3.5 font-semibold text-center whitespace-nowrap" title="Seluruh foto bukti temuan awal pada uraian">Visual Temuan Awal</th>
                <th className="px-4 py-3.5 font-semibold text-center whitespace-nowrap" title="Seluruh foto bukti perbaikan dengan tipe FollowUp">Visual Follow-Up</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap" title="Keterangan pada setiap foto FollowUp">Keterangan Follow-Up</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap" title="Keterangan setiap foto bukti temuan awal dari Issue Photo">Keterangan Foto Temuan Awal</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap" title="Follow Up Datetime per Detail Kawasan">Follow Up Datetime</th>
                <th className="px-4 py-3.5 font-semibold whitespace-nowrap" title="Due Date per Detail Kawasan">Due Date</th>
                <th className="px-4 py-3.5 font-semibold text-center whitespace-nowrap" title="Selisih hari kalender: tanggal follow-up dikurangi due date">Gap Follow-Up</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/50 [&_td]:align-middle">
              {isPreviewLoading ? (
                <tr>
                  <td colSpan={17} className="px-6 py-12 text-center text-muted-foreground">
                    <Loader2 className="h-8 w-8 animate-spin mx-auto mb-3 text-primary/50" />
                    Memuat data inspeksi GMP...
                  </td>
                </tr>
              ) : filteredData.length === 0 ? (
                <tr>
                  <td colSpan={17} className="p-6 sm:p-10">
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
                filteredData.map((row, idx: number) => {
                  const span = rowSpans[idx];
                  return (
                  <tr key={idx} className="hover:bg-muted/30 transition-colors">
                    {/* ID Inspeksi & Info Header — merged (rowSpan) across
                        every row belonging to the same inspection. */}
                    {span.inspeksi > 0 && (
                      <td className="px-4 py-3 font-medium text-primary whitespace-nowrap align-middle" rowSpan={span.inspeksi}>
                        <div className="flex flex-col">
                          <span className="font-semibold">{row.inspection_id}</span>
                          <span className="text-[11px] text-muted-foreground flex items-center gap-1 mt-0.5">
                            <MapPin className="h-3 w-3 inline" /> {row.area || "-"}
                          </span>
                        </div>
                      </td>
                    )}

                    {/* Kawasan — same span as ID Inspeksi because it comes
                        from Inspection_Header. */}
                    {span.inspeksi > 0 && (
                      <td className="px-4 py-3 whitespace-nowrap align-middle" rowSpan={span.inspeksi}>
                        <span className="font-medium text-foreground">{row.kawasan || "-"}</span>
                      </td>
                    )}

                    {/* Detail Kawasan — rendered as its own field. */}
                    {span.inspeksi > 0 && (
                      <td className="px-4 py-3 whitespace-nowrap text-muted-foreground align-middle" rowSpan={span.inspeksi}>
                        {row.detail_kawasan || "-"}
                      </td>
                    )}

                    {/* Aspek — merged within each inspection. */}
                    {span.aspek > 0 && (
                      <td className="px-4 py-3 font-medium whitespace-nowrap align-middle" rowSpan={span.aspek}>{row.aspek || "-"}</td>
                    )}

                    {/* Detail Aspek — merged within each Aspek. */}
                    {span.detailAspek > 0 && (
                      <td className="px-4 py-3 whitespace-nowrap text-muted-foreground align-middle" rowSpan={span.detailAspek}>{row.detail || "-"}</td>
                    )}

                    {/* Uraian — merged within each Detail Aspek. */}
                    {span.uraian > 0 && (
                      <td className="max-w-sm px-4 py-3 text-xs leading-relaxed text-muted-foreground whitespace-normal align-middle" rowSpan={span.uraian}>
                        {row.uraian || "-"}
                      </td>
                    )}

                    {/* Nilai */}
                    <td className="px-4 py-3 text-center whitespace-nowrap">
                      <span className={cn(
                        "px-2.5 py-1 rounded-full text-xs font-bold border",
                        (row.nilai ?? 0) >= 80 ? "bg-green-500/10 text-green-600 border-green-500/20" :
                        (row.nilai ?? 0) >= 60 ? "bg-yellow-500/10 text-yellow-600 border-yellow-500/20" :
                        "bg-red-500/10 text-red-600 border-red-500/20"
                      )}>
                        {row.nilai ?? "-"}
                      </span>
                    </td>

                    {/* Total Nilai */}
                    {span.totalNilaiKawasan > 0 && (
                      <td className="px-4 py-3 text-center whitespace-nowrap font-semibold align-middle" rowSpan={span.totalNilaiKawasan}>
                        <span className="px-2 py-0.5 rounded bg-muted/60 text-xs">
                          {row.total_nilai ?? "-"}
                        </span>
                      </td>
                    )}

                    {/* Persentase Kepatuhan (Detail Kawasan) */}
                    {span.complianceDetailKawasan > 0 && (
                      <td className="px-4 py-3 text-center whitespace-nowrap align-middle" rowSpan={span.complianceDetailKawasan}>
                        {(() => {
                          const pct = row.persentase_kepatuhan_detail_kawasan;
                          if (pct === undefined || pct === null) return <span className="text-muted-foreground">-</span>;
                          return (
                            <span className={cn(
                              "px-2.5 py-1 rounded-full text-xs font-bold border",
                              pct >= 80 ? "bg-green-500/10 text-green-600 border-green-500/20" :
                              pct >= 60 ? "bg-yellow-500/10 text-yellow-600 border-yellow-500/20" :
                              "bg-red-500/10 text-red-600 border-red-500/20"
                            )}>
                              {pct.toFixed(1)}%
                            </span>
                          );
                        })()}
                      </td>
                    )}

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

                    {/* Visual Temuan Awal */}
                    <td className="px-4 py-3 text-center">
                      <GmpEvidenceImages
                        imageUrl={row.image_url}
                        imageUrls={row.image_urls}
                        label="Foto temuan awal"
                        onPreview={setPreviewImage}
                      />
                    </td>

                    {/* Visual Follow-Up */}
                    <td className="px-4 py-3 text-center">
                      <GmpEvidenceImages
                        imageUrl={row.follow_up_image_url}
                        imageUrls={row.follow_up_image_urls}
                        evidence={row.follow_up_evidence}
                        label="Foto follow-up"
                        variant="follow-up"
                        onPreview={setPreviewImage}
                      />
                    </td>

                    {/* Keterangan per Foto Follow-Up */}
                    <td className="px-4 py-3 align-middle">
                      <GmpFollowUpDescriptions evidence={row.follow_up_evidence} />
                    </td>

                    {/* Keterangan Temuan */}
                    <td className="max-w-sm px-4 py-3 align-middle" title={row.keterangan}>
                      {row.keterangan ? (
                        <span className="block whitespace-pre-line break-words text-xs leading-relaxed">{row.keterangan}</span>
                      ) : (
                        <span className="text-muted-foreground italic">-</span>
                      )}
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

                    {/* Gap Follow-Up */}
                    <td className="px-4 py-3 text-center whitespace-nowrap">
                      {row.follow_up_gap_days === undefined || row.follow_up_gap_days === null ? (
                        <span className="text-xs italic text-muted-foreground">-</span>
                      ) : row.follow_up_gap_days > 0 ? (
                        <span className="rounded-full border border-red-500/25 bg-red-500/10 px-2.5 py-1 text-xs font-bold text-red-500">
                          +{row.follow_up_gap_days} hari terlambat
                        </span>
                      ) : row.follow_up_gap_days < 0 ? (
                        <span className="rounded-full border border-emerald-500/25 bg-emerald-500/10 px-2.5 py-1 text-xs font-bold text-emerald-600">
                          {Math.abs(row.follow_up_gap_days)} hari lebih cepat
                        </span>
                      ) : (
                        <span className="rounded-full border border-blue-500/25 bg-blue-500/10 px-2.5 py-1 text-xs font-bold text-blue-500">
                          Tepat waktu
                        </span>
                      )}
                    </td>
                  </tr>
                  );
                })
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
