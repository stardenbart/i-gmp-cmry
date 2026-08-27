"use client";

import Link from "next/link";
import { CheckCircle, Loader2, ArrowRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { format, isPast } from "date-fns";
import { useAuditeeDashboard } from "@/components/dashboard/auditee/AuditeeDashboardContext";

export function AuditeeTaskListWidget() {
  const { isLoading, user, activeTasks } = useAuditeeDashboard();

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold text-foreground">Daftar Temuan Prioritas</h2>
      <div className="bg-card rounded-xl border border-border shadow-sm overflow-hidden">
        {isLoading ? (
          <div className="w-full h-40 flex items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : activeTasks.length === 0 ? (
          <div className="w-full h-40 flex flex-col items-center justify-center text-muted-foreground">
            <CheckCircle className="h-8 w-8 mb-2 opacity-50" />
            <p className="text-sm">Bagus! Tidak ada tugas terbuka saat ini.</p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {activeTasks.map((task) => {
              const overdue = task.due_date ? isPast(new Date(task.due_date)) : false;
              return (
                <div key={task.issue_id} className="p-4 hover:bg-muted/50 transition-colors flex items-center justify-between gap-4">
                  <div className="flex flex-col">
                    <span className="font-semibold text-foreground line-clamp-1">{task.keterangan || "Issue tanpa keterangan"}</span>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground mt-1">
                      <span
                        className={cn(
                          "px-2 py-0.5 rounded-full font-medium text-[10px] uppercase",
                          task.issue_status === "PendingValidation" ? "bg-amber-500/10 text-amber-600" : "bg-red-500/10 text-red-600"
                        )}
                      >
                        {task.issue_status}
                      </span>
                      <span>•</span>
                      <span className={cn(overdue && "text-red-500 font-semibold")}>
                        Tenggat: {task.due_date ? format(new Date(task.due_date), "dd MMM yyyy") : "Tidak ditentukan"}
                      </span>
                    </div>
                  </div>
                  <Link
                    href={`/cimory/${user?.plant_id || "global"}/dashboard/${user?.id || "overview"}/issues/${task.issue_id}`}
                    className="p-2 bg-primary/10 text-primary rounded-lg hover:bg-primary/20 transition-colors shrink-0"
                  >
                    <ArrowRight className="h-4 w-4" />
                  </Link>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
