"use client";

import Link from "next/link";
import { cn } from "@/lib/utils";
import { useAdminDashboard } from "@/components/dashboard/admin/AdminDashboardContext";

export function StatusAuditeeWidget() {
  const { stats, user } = useAdminDashboard();

  return (
    <div className="bg-card rounded-xl p-5 border border-border shadow-sm flex flex-col justify-between h-full">
      <div>
        <div className="flex justify-between items-center mb-4 pb-3 border-b border-border">
          <h3 className="text-base font-semibold text-foreground">Status Lokasi Auditee</h3>
          <span className="bg-primary/10 text-primary text-[10px] font-bold px-2 py-0.5 rounded uppercase tracking-wider">
            Ringkasan Area
          </span>
        </div>
        <div className="space-y-3 max-h-[220px] overflow-y-auto pr-1">
          {!stats?.auditee_status || stats?.auditee_status?.length === 0 ? (
            <div className="p-6 text-center text-xs text-muted-foreground italic">Belum ada data status auditee</div>
          ) : (
            stats?.auditee_status?.map((item, idx: number) => (
              <div key={idx} className="flex items-center justify-between p-3 bg-muted/40 rounded-lg border border-border/70 hover:bg-muted/70 transition-colors">
                <div>
                  <p className="text-xs font-semibold text-foreground">{item.name}</p>
                  <p className="text-[11px] text-muted-foreground">{item.type}</p>
                </div>
                <div className="text-right">
                  <p className={cn("text-xs font-bold", item.compliance < 70 ? "text-red-500" : "text-green-600")}>
                    {item.compliance}%
                  </p>
                  <p className="text-[10px] text-muted-foreground">{item.open_issues} Temuan Terbuka</p>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
      <Link
        href={`/cimory/${user?.plant_id || "global"}/dashboard/${user?.id}/gmp-data`}
        className="mt-4 text-xs font-semibold text-center text-muted-foreground hover:text-primary transition-colors block"
      >
        Lihat Data Auditee GMP &rarr;
      </Link>
    </div>
  );
}
