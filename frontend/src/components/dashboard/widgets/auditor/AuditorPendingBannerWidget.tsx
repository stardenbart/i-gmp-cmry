"use client";

import Link from "next/link";
import { ListTodo } from "lucide-react";
import { useAuditorDashboard } from "@/components/dashboard/auditor/AuditorDashboardContext";

export function AuditorPendingBannerWidget() {
  const { user, pendingValidationCount, pendingValidationList } = useAuditorDashboard();

  if (pendingValidationCount === 0) return null;

  return (
    <section className="rounded-2xl border border-purple-500/30 bg-gradient-to-r from-purple-500/10 via-card to-purple-500/5 p-5 shadow-sm space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="h-10 w-10 rounded-xl bg-purple-500/20 flex items-center justify-center text-purple-500 shrink-0">
            <ListTodo className="h-5 w-5" />
          </div>
          <div>
            <h3 className="font-bold text-base text-foreground flex items-center gap-2">
              <span>Temuan Membutuhkan Validasi Anda</span>
              <span className="bg-purple-500 text-white text-xs font-extrabold px-2 py-0.5 rounded-full">{pendingValidationCount}</span>
            </h3>
            <p className="text-xs text-muted-foreground">
              Auditee telah mengajukan bukti perbaikan/follow-up. Mohon lakukan verifikasi untuk menyetujui atau meminta revisi.
            </p>
          </div>
        </div>
        <Link href={`/cimory/${user?.plant_id || "global"}/dashboard/${user?.id || "overview"}/issues?status=PendingValidation`}>
          <span className="text-xs font-semibold text-purple-600 hover:text-purple-700 dark:text-purple-400 hover:underline">
            Lihat Semua ({pendingValidationCount}) →
          </span>
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 pt-1">
        {pendingValidationList.slice(0, 3).map((issue) => (
          <Link key={issue.issue_id} href={`/cimory/${user?.plant_id || "global"}/dashboard/${user?.id || "overview"}/issues/${issue.issue_id}`}>
            <div className="p-3.5 rounded-xl border border-purple-500/20 bg-card hover:border-purple-500/50 hover:shadow-md transition-all cursor-pointer space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-[11px] font-mono text-purple-600 dark:text-purple-400 font-semibold">ID: {issue.issue_id}</span>
                <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-purple-500/15 text-purple-600 border border-purple-500/30">
                  Pending Validation
                </span>
              </div>
              <p className="text-xs font-semibold text-foreground line-clamp-1">
                {issue.area_name || "—"} · {issue.kawasan_name || "—"} · {issue.detail_kawasan_name || "—"}
              </p>
              <p className="text-xs text-muted-foreground line-clamp-1">{issue.keterangan || "Tanpa keterangan"}</p>
            </div>
          </Link>
        ))}
      </div>
    </section>
  );
}
