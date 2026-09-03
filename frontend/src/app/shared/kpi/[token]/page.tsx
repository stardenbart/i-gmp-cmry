"use client";

import Image from "next/image";
import { useMemo } from "react";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { AlertTriangle, Building2, Clock3, Eye, Loader2, RefreshCw, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DashboardGrid } from "@/components/dashboard/DashboardGrid";
import { PublicKPIDashboardProvider } from "@/components/dashboard/kpi/KPIDashboardContext";
import { buildDynamicWidgetDefinition } from "@/components/dashboard/kpi/dynamicWidgetDefinition";
import { kpiShareApi } from "@/lib/api/kpi-share.api";
import { getApiErrorStatus } from "@/lib/api/error";
import type { WidgetDefinition } from "@/components/dashboard/types";

export default function PublicKPIPage() {
  const params = useParams<{ token: string }>();
  const token = params?.token || "";
  const bootstrapQuery = useQuery({
    queryKey: ["public-kpi-bootstrap", token],
    queryFn: () => kpiShareApi.bootstrap(token),
    enabled: !!token,
    staleTime: 60_000,
    retry: (count, error) => getApiErrorStatus(error) !== 404 && count < 1,
  });
  const bootstrap = bootstrapQuery.data;
  const registry = useMemo<WidgetDefinition[]>(() => {
    if (!bootstrap) return [];
    return bootstrap.layout.map((widget) => buildDynamicWidgetDefinition(widget)).filter((widget): widget is WidgetDefinition => !!widget);
  }, [bootstrap]);

  if (bootstrapQuery.isLoading) {
    return <main className="flex min-h-screen items-center justify-center bg-slate-50 text-slate-950"><div className="text-center"><Loader2 className="mx-auto h-8 w-8 animate-spin text-blue-600" /><p className="mt-3 text-sm text-slate-600">Memuat dashboard publik…</p></div></main>;
  }

  if (bootstrapQuery.isError || !bootstrap) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-50 p-5 text-slate-950">
        <section className="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-7 text-center shadow-lg">
          <Image src="/Logo_Cimory.png" alt="Cimory" width={150} height={60} className="mx-auto h-auto w-32 object-contain" priority />
          <AlertTriangle className="mx-auto mt-6 h-9 w-9 text-amber-500" />
          <h1 className="mt-3 text-xl font-bold">Dashboard tidak tersedia</h1>
          <p className="mt-2 text-sm text-slate-600">Link mungkin tidak valid, sudah kedaluwarsa, atau telah dicabut oleh pemiliknya.</p>
        </section>
      </main>
    );
  }

  return (
    <PublicKPIDashboardProvider token={token} bootstrap={bootstrap}>
      <main className="min-h-screen bg-white text-slate-950">
        <header className="border-b border-slate-200 bg-white shadow-sm">
          <div className="mx-auto flex max-w-[1600px] flex-col gap-4 px-4 py-4 sm:px-6 lg:flex-row lg:items-center lg:justify-between lg:px-8">
            <div className="flex min-w-0 items-center gap-4">
              <Image src="/Logo_Cimory.png" alt="Cimory" width={150} height={60} className="h-auto w-28 shrink-0 object-contain sm:w-32" priority />
              <div className="min-w-0 border-l border-slate-200 pl-4"><h1 className="truncate text-xl font-bold text-slate-950 sm:text-2xl">{bootstrap.public_title}</h1><p className="mt-1 flex items-center gap-1.5 text-xs text-slate-600"><Building2 className="h-3.5 w-3.5" /> {bootstrap.plant_name}</p></div>
            </div>
            <div className="flex flex-wrap items-center gap-2 text-xs">
              <span className="inline-flex items-center gap-1.5 rounded-full bg-blue-500/10 px-3 py-1.5 font-semibold text-blue-600"><Eye className="h-3.5 w-3.5" /> Tampilan publik · hanya baca</span>
              <span className="inline-flex items-center gap-1.5 rounded-full bg-slate-100 px-3 py-1.5 text-slate-600"><Clock3 className="h-3.5 w-3.5" /> Diperbarui {new Date(bootstrap.updated_at).toLocaleString("id-ID")}</span>
              <Button variant="outline" size="sm" onClick={() => window.location.reload()}><RefreshCw className="h-3.5 w-3.5" /> Refresh</Button>
            </div>
          </div>
        </header>
        <section className="mx-auto max-w-[1600px] px-4 py-6 sm:px-6 lg:px-8">
          {registry.length > 0 ? (
            <DashboardGrid registry={registry} enabled={false} providedLayout={bootstrap.layout} dashboardKey="kpi-public" editable={false} toolboxMode={false} />
          ) : (
            <div className="flex min-h-64 flex-col items-center justify-center rounded-2xl border border-dashed border-slate-300 bg-white px-6 text-center">
              <Sparkles className="h-8 w-8 text-slate-500" />
              <h2 className="mt-3 font-semibold">Belum ada visualisasi yang dibagikan</h2>
              <p className="mt-1 text-sm text-slate-600">Pemilik dashboard belum menambahkan visualisasi kustom.</p>
            </div>
          )}
        </section>
        <footer className="border-t border-slate-200 bg-white px-4 py-5 text-center text-xs text-slate-600">Dashboard KPI Cimory · data operasional bersifat read-only</footer>
      </main>
    </PublicKPIDashboardProvider>
  );
}
