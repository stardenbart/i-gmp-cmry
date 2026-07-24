"use client";

import { useState, useEffect } from "react";
import { useRouter, useParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { ArrowLeft, Save, Loader2 } from "lucide-react";
import Link from "next/link";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { inspectionApi } from "@/lib/api/inspection.api";
import { masterApi, Area, Kawasan, DetailKawasan } from "@/lib/api/master.api";
import { useAuditorGuard } from "@/lib/useAdminGuard";

interface CreateFormValues {
  area_id: string;
  kawasan_id: string;
  detail_kawasan_id: string;
}

export default function CreateInspectionPage() {
  const { isAuditor, isLoading: isGuardLoading } = useAuditorGuard();
  const router = useRouter();
  const { userId } = useParams() as { userId: string };
  const [isLoading, setIsLoading] = useState(false);
  const [isAreasLoading, setIsAreasLoading] = useState(true);
  const [isKawasansLoading, setIsKawasansLoading] = useState(false);
  const [isDetailKawasansLoading, setIsDetailKawasansLoading] = useState(false);

  const [areas, setAreas] = useState<Area[]>([]);
  const [kawasans, setKawasans] = useState<Kawasan[]>([]);
  const [detailKawasans, setDetailKawasans] = useState<DetailKawasan[]>([]);

  const {
    register,
    handleSubmit,
    watch,
    reset,
    formState: { errors },
  } = useForm<CreateFormValues>();

  const selectedAreaId = watch("area_id");
  const selectedKawasanId = watch("kawasan_id");
  const selectedDetailKawasanId = watch("detail_kawasan_id");

  // Fetch all areas on mount
  useEffect(() => {
    const fetchAreas = async () => {
      try {
        setIsAreasLoading(true);
        const data = await masterApi.getAreas({ limit: 100 });
        setAreas(data || []);
      } catch (error) {
        console.error("Failed to fetch areas:", error);
        toast.error("Gagal memuat data area");
      } finally {
        setIsAreasLoading(false);
      }
    };
    fetchAreas();
  }, []);

  // Fetch kawasans when area changes
  useEffect(() => {
    const fetchKawasans = async () => {
      if (!selectedAreaId) {
        setKawasans([]);
        setDetailKawasans([]);
        reset((form) => ({ ...form, kawasan_id: "", detail_kawasan_id: "" }));
        return;
      }

      try {
        setIsKawasansLoading(true);
        const data = await masterApi.getKawasans({ area_id: selectedAreaId, limit: 100 });
        setKawasans(data || []);
      } catch (error) {
        console.error("Failed to fetch kawasans:", error);
        toast.error("Gagal memuat data kawasan");
      } finally {
        setIsKawasansLoading(false);
      }
    };
    fetchKawasans();
    // Reset kawasan and detail kawasan when area changes
    setDetailKawasans([]);
    reset((form) => ({ ...form, kawasan_id: "", detail_kawasan_id: "" }));
  }, [selectedAreaId, reset]);

  // Fetch detail kawasans when kawasan changes
  useEffect(() => {
    const fetchDetailKawasans = async () => {
      if (!selectedKawasanId) {
        setDetailKawasans([]);
        reset((form) => ({ ...form, detail_kawasan_id: "" }));
        return;
      }

      try {
        setIsDetailKawasansLoading(true);
        const data = await masterApi.getDetailKawasans({ kawasan_id: selectedKawasanId, limit: 100 });
        setDetailKawasans(data || []);
      } catch (error) {
        console.error("Failed to fetch detail kawasans:", error);
        toast.error("Gagal memuat data detail kawasan");
      } finally {
        setIsDetailKawasansLoading(false);
      }
    };
    fetchDetailKawasans();
  }, [selectedKawasanId, reset]);

  const onSubmit = async (data: CreateFormValues) => {
    try {
      setIsLoading(true);
      const res = await inspectionApi.create(data);
      toast.success("Inspeksi berhasil dibuat");
      router.push(`/cimory/dashboard/${userId}/inspections/${res.data.inspection_id}`);
    } catch (error: any) {
      toast.error(error.response?.data?.message || "Gagal membuat inspeksi");
    } finally {
      setIsLoading(false);
    }
  };

  if (isGuardLoading) {
    return (
      <div className="flex justify-center p-8">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!isAuditor) return null;

  return (
    <div className="space-y-6 max-w-2xl mx-auto">
      <div className="flex items-center gap-4">
        <Link href="./">
          <Button variant="ghost" size="icon" className="rounded-full">
            <ArrowLeft className="h-5 w-5" />
          </Button>
        </Link>
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Buat Inspeksi Baru</h2>
          <p className="text-muted-foreground text-sm">
            Mulai form audit baru untuk area tertentu.
          </p>
        </div>
      </div>

      <Card className="p-6 bg-card/60 backdrop-blur-md">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          <div className="space-y-4">
            {/* Area Select */}
            <div>
              <label className="block text-xs font-medium text-muted-foreground ml-1 mb-2">
                Pilih Area
              </label>
              <div className="relative">
                <select
                  {...register("area_id", { required: "Area wajib dipilih" })}
                  className="flex h-12 w-full rounded-2xl border border-border bg-card px-4 py-2 text-sm text-foreground shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary appearance-none"
                  disabled={isAreasLoading}
                >
                  <option value="">-- Pilih Area --</option>
                  {areas.map((area) => (
                    <option key={area.area_id} value={area.area_id}>
                      {area.area_name}
                    </option>
                  ))}
                </select>
                {isAreasLoading && (
                  <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 animate-spin text-muted-foreground" />
                )}
              </div>
              {errors.area_id && (
                <p className="text-red-500 text-xs mt-1 ml-1">{errors.area_id.message}</p>
              )}
            </div>

            {/* Kawasan Select */}
            <div>
              <label className="block text-xs font-medium text-muted-foreground ml-1 mb-2">
                Pilih Kawasan
              </label>
              <div className="relative">
                <select
                  {...register("kawasan_id", { required: "Kawasan wajib dipilih" })}
                  className="flex h-12 w-full rounded-2xl border border-border bg-card px-4 py-2 text-sm text-foreground shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary appearance-none"
                  disabled={!selectedAreaId || isKawasansLoading}
                >
                  <option value="">-- Pilih Kawasan --</option>
                  {kawasans.map((kawasan) => (
                    <option key={kawasan.kawasan_id} value={kawasan.kawasan_id}>
                      {kawasan.kawasan_name}
                    </option>
                  ))}
                </select>
                {isKawasansLoading && (
                  <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 animate-spin text-muted-foreground" />
                )}
              </div>
              {!selectedAreaId && (
                <p className="text-muted-foreground text-xs mt-1 ml-1">Pilih area terlebih dahulu</p>
              )}
              {errors.kawasan_id && (
                <p className="text-red-500 text-xs mt-1 ml-1">{errors.kawasan_id.message}</p>
              )}
            </div>

            {/* Detail Kawasan Select */}
            <div>
              <label className="block text-xs font-medium text-muted-foreground ml-1 mb-2">
                Detail Kawasan
              </label>
              <div className="relative">
                <select
                  {...register("detail_kawasan_id", { required: "Detail Kawasan wajib dipilih" })}
                  className="flex h-12 w-full rounded-2xl border border-border bg-card px-4 py-2 text-sm text-foreground shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary appearance-none"
                  disabled={!selectedKawasanId || isDetailKawasansLoading}
                >
                  <option value="">-- Pilih Detail --</option>
                  {detailKawasans.map((dk) => (
                    <option key={dk.detail_kawasan_id} value={dk.detail_kawasan_id}>
                      {dk.detail_kawasan_name}
                    </option>
                  ))}
                </select>
                {isDetailKawasansLoading && (
                  <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 animate-spin text-muted-foreground" />
                )}
              </div>
              {!selectedKawasanId && selectedAreaId && (
                <p className="text-muted-foreground text-xs mt-1 ml-1">Pilih kawasan terlebih dahulu</p>
              )}
              {errors.detail_kawasan_id && (
                <p className="text-red-500 text-xs mt-1 ml-1">{errors.detail_kawasan_id.message}</p>
              )}
            </div>
          </div>

          <div className="pt-4 flex justify-end">
            <Button type="submit" isLoading={isLoading} disabled={!selectedAreaId || !selectedKawasanId || !selectedDetailKawasanId}>
              <Save className="mr-2 h-4 w-4" /> Simpan & Mulai Audit
            </Button>
          </div>
        </form>
      </Card>
    </div>
  );
}
