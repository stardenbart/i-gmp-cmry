"use client";

import { useState } from "react";
import {
  Activity,
  LogIn,
  Search,
  Filter,
  ChevronLeft,
  ChevronRight,
  ShieldAlert
} from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/axios";
import { useMounted } from "@/lib/useMounted";
import { useAdminGuard } from "@/lib/useAdminGuard";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuthStore } from "@/stores/authStore";

const fetchActivityLogs = async (page = 1, search = "", userId = "") => {
  const params: Record<string, any> = { page, limit: 10 };
  if (search) params.action = search; // We'll map search to action for now, or just leave it
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
    return found ? found.full_name : id;
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return "-";
    return new Date(dateString).toLocaleString('id-ID', {
      year: 'numeric', month: 'short', day: 'numeric',
      hour: '2-digit', minute: '2-digit'
    });
  };

  if (!isGuardLoading && !isAdmin) {
    return (
      <div className="flex flex-col items-center justify-center min-h-100 space-y-4">
        <ShieldAlert className="h-12 w-12 text-destructive" />
        <h2 className="text-xl font-semibold">Akses Ditolak</h2>
        <Button onClick={() => window.history.back()}>Kembali</Button>
      </div>
    );
  }

  if (!mounted || isGuardLoading) return <div className="h-64 flex justify-center items-center"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div></div>;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Audit Logs</h1>
        <p className="text-muted-foreground">Pantau aktivitas dan riwayat masuk pengguna dalam sistem</p>
      </div>

      {/* Tabs */}
      <div className="border-b border-border">
        <div className="flex gap-4">
          <button
            onClick={() => { setActiveTab("activity"); setPage(1); }}
            className={`flex items-center gap-2 px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === "activity" ? "border-primary text-primary" : "border-transparent text-muted-foreground hover:text-foreground"
            }`}
          >
            <Activity className="h-4 w-4" /> Activity Logs
          </button>
          <button
            onClick={() => { setActiveTab("login"); setPage(1); }}
            className={`flex items-center gap-2 px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === "login" ? "border-primary text-primary" : "border-transparent text-muted-foreground hover:text-foreground"
            }`}
          >
            <LogIn className="h-4 w-4" /> Login Logs
          </button>
        </div>
      </div>

      {/* Toolbar */}
      <div className="flex flex-col md:flex-row gap-4 justify-between items-center bg-card p-4 rounded-xl border border-border shadow-sm">
        <div className="flex flex-col sm:flex-row gap-3 w-full md:w-auto flex-1">
          {activeTab === "activity" && (
            <div className="relative flex-1 max-w-xs">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Cari action..."
                value={searchQuery}
                onChange={(e) => { setSearchQuery(e.target.value); setPage(1); }}
                className="pl-9"
              />
            </div>
          )}
          <div className="flex gap-2 items-center flex-1 max-w-xs">
            <Filter className="h-4 w-4 text-muted-foreground" />
            <select
              className="w-full text-sm bg-background border border-border rounded-md px-3 py-2 outline-none"
              value={userFilter}
              onChange={(e) => { setUserFilter(e.target.value); setPage(1); }}
            >
              <option value="ALL">Semua Pengguna</option>
              {usersList.map((u: any) => (
                <option key={u.user_id} value={u.user_id}>{u.full_name}</option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="rounded-xl border border-border bg-card overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-sm text-left">
            <thead>
              <tr className="border-b border-border bg-muted/50 text-muted-foreground uppercase tracking-wider text-xs font-semibold">
                <th className="px-4 py-3">Waktu</th>
                <th className="px-4 py-3">Pengguna</th>
                {activeTab === "activity" ? (
                  <>
                    <th className="px-4 py-3">Module</th>
                    <th className="px-4 py-3">Action</th>
                    <th className="px-4 py-3">Keterangan</th>
                  </>
                ) : (
                  <>
                    <th className="px-4 py-3">Login</th>
                    <th className="px-4 py-3">Logout</th>
                    <th className="px-4 py-3">IP & Device</th>
                    <th className="px-4 py-3">Status</th>
                  </>
                )}
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {isLoading ? (
                <tr><td colSpan={6} className="p-8 text-center"><div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary mx-auto"></div></td></tr>
              ) : items.length === 0 ? (
                <tr><td colSpan={6} className="p-8 text-center text-muted-foreground">Tidak ada riwayat log ditemukan</td></tr>
              ) : (
                items.map((log: any) => (
                  <tr key={log.activity_log_id || log.login_log_id} className="hover:bg-muted/30 transition-colors">
                    {activeTab === "activity" ? (
                      <>
                        <td className="px-4 py-3 whitespace-nowrap text-xs">{formatDate(log.created_at)}</td>
                        <td className="px-4 py-3 font-medium">{getUserName(log.user_id)}</td>
                        <td className="px-4 py-3">{log.module_id}</td>
                        <td className="px-4 py-3"><span className="bg-primary/10 text-primary px-2 py-0.5 rounded text-xs">{log.activity_action}</span></td>
                        <td className="px-4 py-3 text-muted-foreground">{log.activity_description}</td>
                      </>
                    ) : (
                      <>
                        <td className="px-4 py-3 whitespace-nowrap text-xs">{formatDate(log.login_at)}</td>
                        <td className="px-4 py-3 font-medium">{getUserName(log.user_id)}</td>
                        <td className="px-4 py-3 text-xs whitespace-nowrap">{formatDate(log.login_at)}</td>
                        <td className="px-4 py-3 text-xs whitespace-nowrap">{formatDate(log.logout_at)}</td>
                        <td className="px-4 py-3 text-xs">
                          <div className="flex flex-col">
                            <span>{log.ip_address}</span>
                            <span className="text-muted-foreground">{log.device_info}</span>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          {log.login_status === "Success" 
                            ? <span className="bg-green-500/10 text-green-500 px-2 py-0.5 rounded text-xs font-semibold">Success</span>
                            : <span className="bg-red-500/10 text-red-500 px-2 py-0.5 rounded text-xs font-semibold">Failed</span>
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
              Hal {page} dari {pagination.total_pages} ({pagination.total} total)
            </p>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1}><ChevronLeft className="h-4 w-4" /></Button>
              <Button variant="outline" size="sm" onClick={() => setPage(p => Math.min(pagination.total_pages, p + 1))} disabled={page === pagination.total_pages}><ChevronRight className="h-4 w-4" /></Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
