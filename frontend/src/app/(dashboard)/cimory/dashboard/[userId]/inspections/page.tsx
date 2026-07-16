"use client";

import { useQuery } from "@tanstack/react-query";
import { Plus, Search, Filter, ClipboardCheck } from "lucide-react";
import Link from "next/link";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { inspectionApi, InspectionHeader } from "@/lib/api/inspection.api";

export default function InspectionsPage() {
  const { data, isLoading } = useQuery({
    queryKey: ["inspections"],
    queryFn: () => inspectionApi.getAll(),
  });

  const inspections: InspectionHeader[] = data?.data || [];

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Inspeksi</h2>
          <p className="text-muted-foreground">
            Kelola daftar inspeksi dan audit harian.
          </p>
        </div>
        <Link href="./inspections/create">
          <Button>
            <Plus className="mr-2 h-4 w-4" /> Buat Inspeksi
          </Button>
        </Link>
      </div>

      <Card className="p-4 bg-card/60 backdrop-blur-md flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input placeholder="Cari inspeksi..." className="pl-9 bg-background/50 border-border/50" />
        </div>
        <Button variant="outline" className="sm:w-auto w-full bg-background/50 border-border/50">
          <Filter className="mr-2 h-4 w-4" /> Filter
        </Button>
      </Card>

      <div className="grid gap-4">
        {isLoading ? (
          <div className="flex justify-center p-8">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
          </div>
        ) : inspections.length === 0 ? (
          <Card className="p-12 flex flex-col items-center justify-center text-center bg-card/40 backdrop-blur-md border-dashed">
            <ClipboardCheck className="h-12 w-12 text-muted-foreground mb-4 opacity-50" />
            <h3 className="text-lg font-semibold">Belum ada inspeksi</h3>
            <p className="text-muted-foreground text-sm mt-1 max-w-sm">
              Mulai audit dengan membuat data inspeksi baru.
            </p>
            <Link href="./inspections/create" className="mt-4">
              <Button variant="outline">Buat Inspeksi Pertama</Button>
            </Link>
          </Card>
        ) : (
          inspections.map((inspection) => (
            <Card key={inspection.inspection_id} className="p-4 hover:border-primary/50 transition-colors bg-card/60 backdrop-blur-md">
              <div className="flex flex-col sm:flex-row gap-4 sm:items-center justify-between">
                <div className="flex items-center gap-4">
                  <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
                    <ClipboardCheck className="h-5 w-5 text-primary" />
                  </div>
                  <div>
                    <h4 className="font-semibold">{inspection.inspection_id}</h4>
                    <p className="text-sm text-muted-foreground">
                      {new Date(inspection.created_at).toLocaleDateString('id-ID', {
                        day: 'numeric',
                        month: 'long',
                        year: 'numeric'
                      })}
                    </p>
                  </div>
                </div>
                
                <div className="flex items-center gap-4">
                  <span className={`text-xs font-medium px-2.5 py-1 rounded-full ${
                    inspection.status === 'Completed' ? 'bg-green-500/10 text-green-500' :
                    inspection.status === 'Ongoing' ? 'bg-blue-500/10 text-blue-500' :
                    'bg-zinc-500/10 text-zinc-500'
                  }`}>
                    {inspection.status}
                  </span>
                  <Link href={`./inspections/${inspection.inspection_id}`}>
                    <Button variant="outline" size="sm">
                      Detail
                    </Button>
                  </Link>
                </div>
              </div>
            </Card>
          ))
        )}
      </div>
    </div>
  );
}
