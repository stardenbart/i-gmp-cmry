"use client";

import { useState } from "react";
import {
  Activity,
  LogIn,
  Search,
  ChevronLeft,
  ChevronRight,
  ShieldAlert,
  Eye,
  X,
  FileCode,
  Calendar,
  User,
  Globe,
  Database
} from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/axios";
import { useMounted } from "@/lib/useMounted";
import { useAdminGuard } from "@/lib/useAdminGuard";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuthStore } from "@/stores/authStore";
import { SearchLatencyBadge } from "@/components/ui/SearchLatencyBadge";

const fetchActivityLogs = async (page = 1, search = "", userId = "") => {
  const params: Record<string, any> = { page, limit: 10 };
  if (search) params.search = search;
  if (userId && userId !== "ALL") params.user_id = userId;
  const res = await api.get("/logs/activity", { params });
  return res.data;
};

const fetchLoginLogs = async (page = 1, userId = "") => {
  const params: Record<string, any> = { page, limit: 10 };
  if (userId && userId !== "ALL") params.user_id = userId;
  const res = await api.get("/logs/login", { params });
  return res.data;
};

const fetchAllUsers = async () => {
  const res = await api.get("/users", { params: { limit: 1000 } });
  return res.data?.items || [];
};

export default function LogsPage() {
  const user = useAuthStore((state) => state.user);
  const { isAdmin, isLoading: isGuardLoading } = useAdminGuard();
  const mounted = useMounted();

  const [activeTab, setActiveTab] = useState<"activity" | "login">("activity");
  const [page, setPage] = useState(1);
  const [searchQuery, setSearchQuery] = useState("");
  const [userFilter, setUserFilter] = useState("ALL");
  const [selectedLog, setSelectedLog] = useState<any | null>(null);

  // Data Fetching
  const { data: usersList = [] } = useQuery({
    queryKey: ["users-all"],
    queryFn: fetchAllUsers,
    enabled: mounted && !!user && isAdmin,
  });

  const { data: activityRes, isLoading: isActivityLoading, isFetching: isActivityFetching } = useQuery({
    queryKey: ["logs-activity", page, searchQuery, userFilter],
    queryFn: () => fetchActivityLogs(page, searchQuery, userFilter),
    enabled: mounted && !!user && isAdmin && activeTab === "activity",
  });

  const { data: loginRes, isLoading: isLoginLoading, isFetching: isLoginFetching } = useQuery({
    queryKey: ["logs-login", page, userFilter],
    queryFn: () => fetchLoginLogs(page, userFilter),
    enabled: mounted && !!user && isAdmin && activeTab === "login",
  });

  const isLoading = activeTab === "activity" ? (isActivityLoading || isActivityFetching) : (isLoginLoading || isLoginFetching);
  
  const currentData = activeTab === "activity" ? activityRes?.data : loginRes?.data;
  const items = currentData?.items || [];
  const pagination = currentData?.pagination || { total: 0, page: 1, limit: 10, total_pages: 1 };

  const getUserName = (id: string) => {
    const found = usersList.find((u: any) => u.user_id === id);
    return found ? `${found.full_name} (${found.user_id})` : id;
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return "-";
    return new Date(dateString).toLocaleString('id-ID', {
      year: 'numeric', month: 'short', day: 'numeric',
      hour: '2-digit', minute: '2-digit', second: '2-digit'
    });
  };

  const formatJson = (jsonString: string) => {
    if (!jsonString) return null;
    try {
      const parsed = JSON.parse(jsonString);
      return JSON.stringify(parsed, null, 2);
    } catch {
      return jsonString;
    }
  };

  if (!isGuardLoading && !isAdmin) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] space-y-4">
        <ShieldAlert className="h-12 w-12 text-destructive" />
        <h2 className="text-xl font-semibold">Akses Ditolak</h2>
        <p className="text-sm text-muted-foreground text-center max-w-md">
          Anda tidak memiliki hak akses untuk membuka halaman Audit Logs. Silakan hubungi Administrator.
        </p>
        <Button onClick={() => window.history.back()}>Kembali</Button>
      </div>
    );
  }

  if (!mounted || isGuardLoading) return <div className="h-64 flex justify-center items-center"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div></div>;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Audit & System Logs</h1>
        <p className="text-muted-foreground">Pantau riwayat aktivitas, payload request API, dan log masuk pengguna secara realtime</p>
      </div>

      {/* Tabs */}
      <div className="border-b border-border">
        <div className="flex gap-4">
          <button
            onClick={() => { setActiveTab("activity"); setPage(1); }}
            className={`flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
              activeTab === "activity" ? "border-primary text-primary font-semibold" : "border-transparent text-muted-foreground hover:text-foreground"
            }`}
          >
            <Activity className="h-4 w-4" /> Riwayat Aktivitas & Permintaan
          </button>
          <button
            onClick={() => { setActiveTab("login"); setPage(1); }}
            className={`flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
              activeTab === "login" ? "border-primary text-primary font-semibold" : "border-transparent text-muted-foreground hover:text-foreground"
            }`}
          >
            <LogIn className="h-4 w-4" /> Riwayat Sesi Masuk
          </button>
        </div>
      </div>

      {/* Toolbar */}
      {activeTab === "activity" && (
        <div style={{ display: "flex", gap: "12px", alignItems: "center", width: "100%" }}>
          <div style={{ position: "relative", flex: "1", maxWidth: "512px" }}>
            <Search
              style={{
                position: "absolute",
                left: "12px",
                top: "50%",
                transform: "translateY(-50%)",
                width: "16px",
                height: "16px",
                color: "var(--muted-foreground)",
                pointerEvents: "none",
                zIndex: 1,
              }}
            />
            <input
              type="text"
              placeholder="Cari aksi, keterangan, tabel, atau ID catatan..."
              value={searchQuery}
              onChange={(e) => { setSearchQuery(e.target.value); setPage(1); }}
              style={{
                width: "100%",
                height: "44px",
                paddingLeft: "40px",
                paddingRight: "16px",
                borderRadius: "12px",
                border: "1px solid var(--border)",
                background: "var(--card)",
                color: "var(--foreground)",
                fontSize: "14px",
                outline: "none",
                boxShadow: "0 1px 3px 0 rgb(0 0 0 / 0.1)",
              }}
            />
          </div>
          <SearchLatencyBadge
            searchQuery={searchQuery}
            isFetching={isActivityLoading}
            pageName="Riwayat Log (Activity)"
            apiPath="/logs"
          />
        </div>
      )}

      {/* Table */}
      <div className="rounded-xl border border-border bg-card overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-sm text-left">
            <thead>
              <tr className="border-b border-border bg-muted/50 text-muted-foreground uppercase tracking-wider text-xs font-semibold">
                <th className="px-4 py-3.5">Waktu</th>
                <th className="px-4 py-3.5">Pengguna</th>
                {activeTab === "activity" ? (
                  <>
                    <th className="px-4 py-3.5">Aksi</th>
                    <th className="px-4 py-3.5">Tabel / Modul</th>
                    <th className="px-4 py-3.5">Deskripsi Aktivitas</th>
                    <th className="px-4 py-3.5 text-center">Detail Data</th>
                  </>
                ) : (
                  <>
                    <th className="px-4 py-3.5">Waktu Masuk</th>
                    <th className="px-4 py-3.5">Waktu Keluar</th>
                    <th className="px-4 py-3.5">Alamat IP</th>
                    <th className="px-4 py-3.5 text-center">Status</th>
                  </>
                )}
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {isLoading ? (
                <tr><td colSpan={7} className="p-12 text-center"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"></div><span className="text-xs text-muted-foreground">Memuat data log...</span></td></tr>
              ) : items.length === 0 ? (
                <tr><td colSpan={7} className="p-12 text-center text-muted-foreground">Tidak ada riwayat log ditemukan</td></tr>
              ) : (
                items.map((log: any, idx: number) => (
                  <tr key={log.activity_log_id || log.login_log_id || idx} className="hover:bg-muted/30 transition-colors">
                    {activeTab === "activity" ? (
                      <>
                        <td className="px-4 py-3 whitespace-nowrap text-xs text-muted-foreground font-mono">{formatDate(log.created_at)}</td>
                        <td className="px-4 py-3 font-medium">{getUserName(log.user_id)}</td>
                        <td className="px-4 py-3">
                          <span className={`px-2.5 py-0.5 rounded text-xs font-bold border ${
                            log.activity_action?.includes("CREATE") || log.activity_action?.includes("UPLOAD") ? "bg-green-500/10 text-green-600 border-green-500/20" :
                            log.activity_action?.includes("UPDATE") || log.activity_action?.includes("PUT") ? "bg-blue-500/10 text-blue-600 border-blue-500/20" :
                            log.activity_action?.includes("DELETE") ? "bg-red-500/10 text-red-600 border-red-500/20" :
                            "bg-primary/10 text-primary border-primary/20"
                          }`}>
                            {log.activity_action}
                          </span>
                        </td>
                        <td className="px-4 py-3 whitespace-nowrap">
                          <div className="flex flex-col">
                            <span className="font-semibold text-xs text-foreground">{log.table_affected || log.module_id || "-"}</span>
                            {log.record_id && <span className="text-[11px] font-mono text-muted-foreground">{log.record_id}</span>}
                          </div>
                        </td>
                        <td className="px-4 py-3 text-muted-foreground max-w-xs truncate" title={log.activity_description}>
                          {log.activity_description || "-"}
                        </td>
                        <td className="px-4 py-3 text-center whitespace-nowrap">
                          <Button
                            variant="outline"
                            size="sm"
                            className="h-8 gap-1.5 text-xs"
                            onClick={() => setSelectedLog(log)}
                          >
                            <Eye className="h-3.5 w-3.5" /> Detail
                          </Button>
                        </td>
                      </>
                    ) : (
                      <>
                        <td className="px-4 py-3 whitespace-nowrap text-xs text-muted-foreground font-mono">{formatDate(log.login_at)}</td>
                        <td className="px-4 py-3 font-medium">{getUserName(log.user_id)}</td>
                        <td className="px-4 py-3 text-xs whitespace-nowrap text-muted-foreground font-mono">{formatDate(log.login_at)}</td>
                        <td className="px-4 py-3 text-xs whitespace-nowrap text-muted-foreground font-mono">{log.logout_at ? formatDate(log.logout_at) : "Sesi Aktif"}</td>
                        <td className="px-4 py-3 text-xs font-mono">
                          <div className="flex flex-col">
                            <span>{log.ip_address || "-"}</span>
                            <span className="text-[11px] text-muted-foreground truncate max-w-[150px]" title={log.device_info}>{log.device_info || "-"}</span>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-center whitespace-nowrap">
                          {log.login_status === "Success" 
                            ? <span className="bg-green-500/10 text-green-600 px-2.5 py-1 rounded-full text-xs font-bold border border-green-500/20">Berhasil</span>
                            : <span className="bg-red-500/10 text-red-600 px-2.5 py-1 rounded-full text-xs font-bold border border-red-500/20">Gagal</span>
                          }
                        </td>
                      </>
                    )}
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {pagination.total_pages > 1 && (
          <div className="flex items-center justify-between border-t border-border px-4 py-3">
            <p className="text-sm text-muted-foreground">
              Hal {page} dari {pagination.total_pages} ({pagination.total} total log)
            </p>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1}><ChevronLeft className="h-4 w-4" /></Button>
              <Button variant="outline" size="sm" onClick={() => setPage(p => Math.min(pagination.total_pages, p + 1))} disabled={page === pagination.total_pages}><ChevronRight className="h-4 w-4" /></Button>
            </div>
          </div>
        )}
      </div>

      {/* Log Detail Modal */}
      {selectedLog && (
        <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-card border border-border rounded-xl shadow-2xl max-w-2xl w-full max-h-[90vh] flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-200">
            {/* Modal Header */}
            <div className="px-6 py-4 border-b border-border flex items-center justify-between bg-muted/30">
              <div className="flex items-center gap-2.5">
                <FileCode className="h-5 w-5 text-primary" />
                <h3 className="font-semibold text-lg text-foreground">Detail Request & Activity Log</h3>
              </div>
              <Button variant="ghost" size="icon" className="h-8 w-8 rounded-full" onClick={() => setSelectedLog(null)}>
                <X className="h-4 w-4" />
              </Button>
            </div>

            {/* Modal Body */}
            <div className="p-6 overflow-y-auto space-y-5">
              {/* Metadata Grid */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs bg-muted/40 p-4 rounded-lg border border-border/60">
                <div className="flex items-center gap-2">
                  <User className="h-4 w-4 text-muted-foreground shrink-0" />
                  <div>
                    <span className="text-muted-foreground block text-[10px] uppercase font-semibold">Pengguna</span>
                    <span className="font-medium text-foreground">{getUserName(selectedLog.user_id)}</span>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground shrink-0" />
                  <div>
                    <span className="text-muted-foreground block text-[10px] uppercase font-semibold">Waktu Kejadian</span>
                    <span className="font-medium text-foreground">{formatDate(selectedLog.created_at || selectedLog.login_at)}</span>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Globe className="h-4 w-4 text-muted-foreground shrink-0" />
                  <div>
                    <span className="text-muted-foreground block text-[10px] uppercase font-semibold">IP Address</span>
                    <span className="font-mono text-foreground">{selectedLog.ip_address || "127.0.0.1"}</span>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Database className="h-4 w-4 text-muted-foreground shrink-0" />
                  <div>
                    <span className="text-muted-foreground block text-[10px] uppercase font-semibold">Tabel / Record ID</span>
                    <span className="font-medium text-foreground">{selectedLog.table_affected || "-"} ({selectedLog.record_id || "-"})</span>
                  </div>
                </div>
              </div>

              {/* Action & Description */}
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Aktivitas / Action</label>
                <div className="flex items-center gap-2">
                  <span className="bg-primary/10 text-primary px-3 py-1 rounded text-xs font-bold border border-primary/20">
                    {selectedLog.activity_action}
                  </span>
                  <span className="text-sm font-medium text-foreground">{selectedLog.activity_description}</span>
                </div>
              </div>

              {/* Payload New Value */}
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Payload Request / New Value</label>
                {selectedLog.new_value ? (
                  <pre className="bg-muted/80 p-3.5 rounded-lg border border-border/60 text-xs font-mono overflow-x-auto text-foreground max-h-48 leading-relaxed">
                    {formatJson(selectedLog.new_value)}
                  </pre>
                ) : (
                  <div className="p-3 bg-amber-500/10 rounded-lg text-xs text-amber-600 dark:text-amber-400 italic border border-amber-500/20">
                    Log ini tercatat sebelum fitur penangkapan payload diaktifkan. Restart backend untuk mencatat payload pada aktivitas baru.
                  </div>
                )}
              </div>

              {/* Payload Old Value (if any) */}
              {selectedLog.old_value && (
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Previous State / Old Value</label>
                  <pre className="bg-muted/80 p-3.5 rounded-lg border border-border/60 text-xs font-mono overflow-x-auto text-foreground max-h-48 leading-relaxed">
                    {formatJson(selectedLog.old_value)}
                  </pre>
                </div>
              )}
            </div>

            {/* Modal Footer */}
            <div className="px-6 py-3 border-t border-border bg-muted/30 flex justify-end">
              <Button onClick={() => setSelectedLog(null)} size="sm">Tutup</Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
