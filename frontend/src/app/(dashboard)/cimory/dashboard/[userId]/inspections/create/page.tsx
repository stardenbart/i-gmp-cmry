"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { ArrowLeft, Save } from "lucide-react";
import Link from "next/link";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { inspectionApi } from "@/lib/api/inspection.api";

interface CreateFormValues {
  area_id: string;
  kawasan_id: string;
  detail_kawasan_id: string;
}

export default function CreateInspectionPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<CreateFormValues>();

  const onSubmit = async (data: CreateFormValues) => {
    try {
      setIsLoading(true);
      const res = await inspectionApi.create(data);
      toast.success("Inspeksi berhasil dibuat");
      router.push(`/inspections/${res.data.inspection_id}`);
    } catch (error: any) {
      toast.error(error.response?.data?.message || "Gagal membuat inspeksi");
    } finally {
      setIsLoading(false);
    }
  };

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
            <div>
              <label className="block text-xs font-medium text-muted-foreground ml-1 mb-2">
                Pilih Area
              </label>
              {/* Using a simple select for now until a full custom Select component is added */}
              <select
                {...register("area_id", { required: "Area wajib dipilih" })}
                className="flex h-12 w-full rounded-2xl border border-border bg-card px-4 py-2 text-sm text-foreground shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary appearance-none"
              >
                <option value="">-- Pilih Area --</option>
                <option value="AREA_1">Area Produksi 1</option>
                <option value="AREA_2">Gudang Penyimpanan</option>
              </select>
              {errors.area_id && <p className="text-red-500 text-xs mt-1 ml-1">{errors.area_id.message}</p>}
            </div>

            <div>
              <label className="block text-xs font-medium text-muted-foreground ml-1 mb-2">
                Pilih Kawasan
              </label>
              <select
                {...register("kawasan_id", { required: "Kawasan wajib dipilih" })}
                className="flex h-12 w-full rounded-2xl border border-border bg-card px-4 py-2 text-sm text-foreground shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary appearance-none"
              >
                <option value="">-- Pilih Kawasan --</option>
                <option value="KAW_1">Kawasan A</option>
                <option value="KAW_2">Kawasan B</option>
              </select>
              {errors.kawasan_id && <p className="text-red-500 text-xs mt-1 ml-1">{errors.kawasan_id.message}</p>}
            </div>

            <div>
              <label className="block text-xs font-medium text-muted-foreground ml-1 mb-2">
                Detail Kawasan
              </label>
              <select
                {...register("detail_kawasan_id", { required: "Detail Kawasan wajib dipilih" })}
                className="flex h-12 w-full rounded-2xl border border-border bg-card px-4 py-2 text-sm text-foreground shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary appearance-none"
              >
                <option value="">-- Pilih Detail --</option>
                <option value="DET_1">Sektor 1A</option>
                <option value="DET_2">Sektor 1B</option>
              </select>
              {errors.detail_kawasan_id && <p className="text-red-500 text-xs mt-1 ml-1">{errors.detail_kawasan_id.message}</p>}
            </div>
          </div>

          <div className="pt-4 flex justify-end">
            <Button type="submit" isLoading={isLoading}>
              <Save className="mr-2 h-4 w-4" /> Simpan & Mulai Audit
            </Button>
          </div>
        </form>
      </Card>
    </div>
  );
}
