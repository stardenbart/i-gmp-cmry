"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { AlertTriangle, Search, Filter, Clock, CheckCircle2, XCircle, Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { issueApi, Issue, IssueStatus } from "@/lib/api/issue.api";
import { cn } from "@/lib/utils";

const STATUS_OPTIONS: { label: string; value: IssueStatus | "all" }[] = [
  { label: "Semua", value: "all" },
  { label: "Open", value: "Open" },
  { label: "In Progress", value: "InProgress" },
  { label: "Closed", value: "Closed" },
  { label: "Verified", value: "Verified" },
];

const statusConfig: Record<IssueStatus, { label: string; icon: React.ElementType; color: string; bg: string }> = {
  Open: { label: "Open", icon: AlertTriangle, color: "text-red-500", bg: "bg-red-500/10" },
  InProgress: { label: "In Progress", icon: Loader2, color: "text-blue-500", bg: "bg-blue-500/10" },
  Closed: { label: "Closed", icon: XCircle, color: "text-zinc-500", bg: "bg-zinc-500/10" },
  Verified: { label: "Verified", icon: CheckCircle2, color: "text-green-500", bg: "bg-green-500/10" },
};

export default function IssuesPage() {
  const [activeStatus, setActiveStatus] = useState<IssueStatus | "all">("all");

  const { data, isLoading } = useQuery({
    queryKey: ["issues", activeStatus],
    queryFn: () => issueApi.getAll({ status: activeStatus === "all" ? undefined : activeStatus }),
  });

  const issues: Issue[] = data?.data || [];

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Temuan (Issues)</h2>
        <p className="text-muted-foreground">
          Daftar semua temuan dari hasil inspeksi audit.
        </p>
      </div>

      {/* Status Filter Tabs */}
      <div className="flex gap-2 overflow-x-auto pb-1">
        {STATUS_OPTIONS.map((opt) => (
          <button
            key={opt.value}
            onClick={() => setActiveStatus(opt.value)}
            className={cn(
              "flex-shrink-0 rounded-full px-4 py-1.5 text-sm font-medium transition-all",
              activeStatus === opt.value
                ? "bg-primary text-primary-foreground"
                : "bg-muted text-muted-foreground hover:bg-muted/80"
            )}
          >
            {opt.label}
          </button>
        ))}
      </div>

      {/* Search Bar */}
      <Card className="p-3 bg-card/60 backdrop-blur-md">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Cari temuan berdasarkan keterangan..."
            className="pl-9 bg-background/50 border-border/50"
          />
        </div>
      </Card>

      {/* Issues List */}
      <div className="grid gap-4">
        {isLoading ? (
          <div className="flex justify-center p-10">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
          </div>
        ) : issues.length === 0 ? (
          <Card className="p-12 flex flex-col items-center justify-center text-center bg-card/40 backdrop-blur-md border-dashed">
            <AlertTriangle className="h-12 w-12 text-muted-foreground mb-4 opacity-50" />
            <h3 className="text-lg font-semibold">Tidak ada temuan</h3>
            <p className="text-muted-foreground text-sm mt-1">
              Belum ada temuan dengan status ini.
            </p>
          </Card>
        ) : (
          issues.map((issue) => {
            const config = statusConfig[issue.issue_status];
            const StatusIcon = config.icon;
            const dueDate = issue.due_date ? new Date(issue.due_date) : null;
            const isOverdue = dueDate && dueDate < new Date() && issue.issue_status !== "Closed" && issue.issue_status !== "Verified";

            return (
              <Link key={issue.issue_id} href={`./issues/${issue.issue_id}`}>
                <Card className="p-4 hover:border-primary/50 transition-all bg-card/60 backdrop-blur-md cursor-pointer hover:shadow-lg hover:shadow-primary/5">
                  <div className="flex flex-col sm:flex-row gap-4 sm:items-center justify-between">
                    <div className="flex items-start gap-4">
                      <div className={cn("mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-full", config.bg)}>
                        <StatusIcon className={cn("h-5 w-5", config.color)} />
                      </div>
                      <div className="min-w-0">
                        <h4 className="font-semibold truncate">{issue.keterangan || "Tanpa keterangan"}</h4>
                        <p className="text-sm text-muted-foreground mt-0.5">
                          ID: <span className="font-mono text-xs">{issue.issue_id}</span>
                        </p>
                        {dueDate && (
                          <div className={cn("flex items-center gap-1 mt-1 text-xs", isOverdue ? "text-red-500" : "text-muted-foreground")}>
                            <Clock className="h-3 w-3" />
                            <span>
                              {isOverdue ? "Overdue – " : "Due: "}
                              {dueDate.toLocaleDateString("id-ID", { day: "numeric", month: "long", year: "numeric" })}
                            </span>
                          </div>
                        )}
                      </div>
                    </div>

                    <div className="flex items-center gap-3 pl-13 sm:pl-0">
                      {issue.photos && issue.photos.length > 0 && (
                        <span className="text-xs text-muted-foreground">
                          📷 {issue.photos.length} foto
                        </span>
                      )}
                      <span className={cn("text-xs font-medium px-2.5 py-1 rounded-full shrink-0", config.bg, config.color)}>
                        {config.label}
                      </span>
                    </div>
                  </div>
                </Card>
              </Link>
            );
          })
        )}
      </div>
    </div>
  );
}
