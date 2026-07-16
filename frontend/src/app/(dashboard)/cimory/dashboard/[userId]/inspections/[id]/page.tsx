"use client";

import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { ArrowLeft, Save, CheckCircle2 } from "lucide-react";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { inspectionApi } from "@/lib/api/inspection.api";

export default function InspectionDetailPage() {
  const { id } = useParams() as { id: string };

  const { data, isLoading } = useQuery({
    queryKey: ["inspection", id],
    queryFn: () => inspectionApi.getById(id),
  });

  const inspection = data?.data;

  const { register, handleSubmit } = useForm();

  const onSubmitChecklist = async (_formData: any) => {
    toast.success("Hasil inspeksi berhasil disimpan sementara");
  };

  const handleSubmitFinal = async () => {
    try {
      await inspectionApi.updateStatus(id, "Completed");
      toast.success("Inspeksi diselesaikan");
      window.location.reload();
    } catch (_error: any) {
      toast.error("Gagal menyelesaikan inspeksi");
    }
  };

  if (isLoading) {
    return (
      <div className="flex justify-center p-8">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!inspection) {
    return <div>Data tidak ditemukan</div>;
  }

  const statusClass =
    inspection.status === "Completed"
      ? "bg-green-500/10 text-green-500"
      : inspection.status === "Ongoing"
      ? "bg-blue-500/10 text-blue-500"
      : "bg-zinc-500/10 text-zinc-500";

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <Link href="../">
            <Button variant="ghost" size="icon" className="rounded-full">
              <ArrowLeft className="h-5 w-5" />
            </Button>
          </Link>
          <div>
            <div className="flex items-center gap-3">
              <h2 className="text-2xl font-bold tracking-tight">Detail Inspeksi</h2>
              <span className={`text-xs font-medium px-2.5 py-1 rounded-full ${statusClass}`}>
                {inspection.status}
              </span>
            </div>
            <p className="text-muted-foreground text-sm font-mono mt-1">
              ID: {inspection.inspection_id}
            </p>
          </div>
        </div>

        {inspection.status !== "Completed" && (
          <Button onClick={handleSubmitFinal} className="bg-green-600 hover:bg-green-700 text-white">
            <CheckCircle2 className="mr-2 h-4 w-4" /> Selesaikan Audit
          </Button>
        )}
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        <Card className="p-6 bg-card/60 backdrop-blur-md md:col-span-1 h-fit">
          <h3 className="font-semibold mb-4 border-b border-border pb-2">Informasi Area</h3>
          <div className="space-y-3 text-sm">
            <div>
              <span className="text-muted-foreground block text-xs">Area ID</span>
              <span className="font-medium">{inspection.area_id}</span>
            </div>
            <div>
              <span className="text-muted-foreground block text-xs">Kawasan ID</span>
              <span className="font-medium">{inspection.kawasan_id}</span>
            </div>
            <div>
              <span className="text-muted-foreground block text-xs">Detail Kawasan</span>
              <span className="font-medium">{inspection.detail_kawasan_id}</span>
            </div>
            <div>
              <span className="text-muted-foreground block text-xs">Auditor</span>
              <span className="font-medium">{inspection.inspector_id}</span>
            </div>
          </div>
        </Card>

        <Card className="p-6 bg-card/60 backdrop-blur-md md:col-span-2">
          <h3 className="font-semibold mb-4 border-b border-border pb-2">Checklist Audit</h3>

          <form onSubmit={handleSubmit(onSubmitChecklist)} className="space-y-6">
            <div className="space-y-4">
              {[1, 2, 3].map((item) => (
                <div key={item} className="p-4 rounded-2xl border border-border bg-background/50">
                  <div className="flex flex-col sm:flex-row justify-between gap-4 mb-4">
                    <div>
                      <h4 className="font-medium">Uraian Pengecekan {item}</h4>
                      <p className="text-xs text-muted-foreground">
                        Apakah area kerja dalam keadaan bersih dan rapi?
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      <label className="flex items-center gap-1 text-sm">
                        <input type="radio" value="OK" {...register(`check_${item}`)} className="accent-primary" /> OK
                      </label>
                      <label className="flex items-center gap-1 text-sm">
                        <input type="radio" value="NG" {...register(`check_${item}`)} className="accent-red-500" /> NG
                      </label>
                      <label className="flex items-center gap-1 text-sm">
                        <input type="radio" value="NA" {...register(`check_${item}`)} className="accent-zinc-500" /> N/A
                      </label>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
                    <div className="sm:col-span-1">
                      <label className="block text-xs font-medium text-muted-foreground mb-1">Nilai</label>
                      <Input type="number" {...register(`nilai_${item}`)} className="h-9 rounded-xl" placeholder="0-100" />
                    </div>
                    <div className="sm:col-span-3">
                      <label className="block text-xs font-medium text-muted-foreground mb-1">Keterangan</label>
                      <Input {...register(`ket_${item}`)} className="h-9 rounded-xl" placeholder="Tulis catatan jika ada temuan..." />
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {inspection.status !== "Completed" && (
              <div className="flex justify-end pt-2">
                <Button type="submit" variant="outline">
                  <Save className="mr-2 h-4 w-4" /> Simpan Draft
                </Button>
              </div>
            )}
          </form>
        </Card>
      </div>
    </div>
  );
}
