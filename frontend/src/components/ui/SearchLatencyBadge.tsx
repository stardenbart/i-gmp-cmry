"use client";

import { useState, useEffect, useRef } from "react";
import { createPortal } from "react-dom";
import { Zap, Play, BarChart2, X, RefreshCw } from "lucide-react";
import { Button } from "./button";
import { api } from "@/lib/api/axios";
import { useSettingsStore } from "@/stores/settingsStore";
import { usePermissions } from "@/lib/usePermissions";
import { useMounted } from "@/lib/useMounted";

interface SearchLatencyBadgeProps {
  searchQuery: string;
  isFetching: boolean;
  pageName: string;
  apiPath: string;
}

export function SearchLatencyBadge({
  searchQuery,
  isFetching,
  pageName,
  apiPath
}: SearchLatencyBadgeProps) {
  const [latencyMs, setLatencyMs] = useState<number | null>(null);
  const [isTesterOpen, setIsTesterOpen] = useState(false);
  const mounted = useMounted();
  const startTimeRef = useRef<number | null>(null);
  const showSearchLatencyButton = useSettingsStore((state) => state.showSearchLatencyButton);
  const { hasPermission } = usePermissions();

  // Track start time when searchQuery changes or fetching begins
  useEffect(() => {
    if (isFetching) {
      if (!startTimeRef.current) {
        startTimeRef.current = performance.now();
      }
    } else {
      if (startTimeRef.current) {
        const elapsed = performance.now() - startTimeRef.current;
        const roundedElapsed = Math.round(elapsed);
        requestAnimationFrame(() => setLatencyMs(roundedElapsed));
        startTimeRef.current = null;
      }
    }
  }, [isFetching, searchQuery]);

  const getLatencyColor = (ms: number) => {
    if (ms < 300) return "bg-emerald-500/10 text-emerald-600 border-emerald-500/30 dark:text-emerald-400";
    if (ms < 800) return "bg-amber-500/10 text-amber-600 border-amber-500/30 dark:text-amber-400";
    return "bg-rose-500/10 text-rose-600 border-rose-500/30 dark:text-rose-400";
  };

  const canViewBadge = hasPermission("PERM-LOG-R") || hasPermission("PERM-MSTR-R") || hasPermission("PERM-USR-R");

  // Completely hide badge and button if toggle is OFF, not mounted, or without management permissions
  if (!mounted || !showSearchLatencyButton || !canViewBadge) {
    return null;
  }

  return (
    <div className="inline-flex items-center gap-2 shrink-0">
      {/* Live Latency Badge */}
      {latencyMs !== null && (
        <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold border transition-all ${getLatencyColor(latencyMs)}`}>
          <Zap className="w-3.5 h-3.5 animate-pulse shrink-0" />
          <span>{latencyMs} ms</span>
        </span>
      )}

      {/* Latency Test Button */}
      <button
        type="button"
        onClick={() => setIsTesterOpen(true)}
        className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold bg-primary/10 text-primary hover:bg-primary/20 transition-all border border-primary/20 shrink-0 cursor-pointer"
        title={`Uji Latensi Pencarian ${pageName}`}
      >
        <Zap className="w-3.5 h-3.5" />
        <span>Uji Latensi Search</span>
      </button>

      {/* Interactive Modal rendered via Portal */}
      {isTesterOpen && (
        <SearchLatencyModalPortaled
          isOpen={isTesterOpen}
          onClose={() => setIsTesterOpen(false)}
          defaultPageName={pageName}
          defaultApiPath={apiPath}
        />
      )}
    </div>
  );
}

const TEST_PAGES = [
  { name: "Inspeksi Digital", path: "/inspections", sampleQueries: ["Session", "AREA", "Completed", "Ongoing", "Draft"] },
  { name: "Temuan (Issues)", path: "/issues", sampleQueries: ["Open", "Kebersihan", "Lantai", "Overdue", "GMP"] },
  { name: "Data Inspeksi (GMP)", path: "/dashboard/stats", sampleQueries: ["Area", "QC", "Kawasan", "Nilai", "2026"] },
  { name: "Riwayat Log (Activity)", path: "/logs", sampleQueries: ["Login", "Update", "Create", "Delete", "Admin"] },
  { name: "Manajemen Pengguna", path: "/users", sampleQueries: ["admin", "auditor", "pic", "qa", "manager"] }
];

function SearchLatencyModalPortaled({
  isOpen,
  onClose,
  defaultPageName,
  defaultApiPath
}: {
  isOpen: boolean;
  onClose: () => void;
  defaultPageName: string;
  defaultApiPath: string;
}) {
  const mounted = useMounted();
  const [selectedPage, setSelectedPage] = useState(
    TEST_PAGES.find(p => p.name === defaultPageName) || {
      name: defaultPageName,
      path: defaultApiPath,
      sampleQueries: [],
    }
  );
  const [isRunning, setIsRunning] = useState(false);
  const [results, setResults] = useState<number[]>([]);
  const [currentQuery, setCurrentQuery] = useState<string>("");
  const [progressPct, setProgressPct] = useState(0);

  const runTest = async () => {
    setIsRunning(true);
    setResults([]);
    setProgressPct(0);

    const latencies: number[] = [];
    const totalRuns = 15;

    for (let i = 0; i < totalRuns; i++) {
      const q = selectedPage.sampleQueries[i % selectedPage.sampleQueries.length];
      setCurrentQuery(q);

      const start = performance.now();
      try {
        await api.get(selectedPage.path, {
          params: { q, limit: 10, search: q, page: 1 }
        });
      } catch {
        // Continue benchmark even if empty
      }
      const elapsed = Math.round(performance.now() - start);
      latencies.push(elapsed);
      setResults([...latencies]);
      setProgressPct(Math.round(((i + 1) / totalRuns) * 100));

      await new Promise(r => setTimeout(r, 80));
    }

    setIsRunning(false);
  };

  if (!isOpen || !mounted) return null;

  const avg = results.length ? Math.round(results.reduce((a, b) => a + b, 0) / results.length) : 0;
  const sorted = [...results].sort((a, b) => a - b);
  const p50 = sorted.length ? sorted[Math.floor(sorted.length * 0.5)] : 0;
  const p90 = sorted.length ? sorted[Math.floor(sorted.length * 0.9)] : 0;
  const min = sorted.length ? sorted[0] : 0;
  const max = sorted.length ? sorted[sorted.length - 1] : 0;

  const modalContent = (
    <div
      style={{
        position: "fixed",
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        zIndex: 99999,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        backgroundColor: "rgba(0, 0, 0, 0.65)",
        backdropFilter: "blur(8px)",
        WebkitBackdropFilter: "blur(8px)",
        padding: "16px",
        boxSizing: "border-box",
      }}
    >
      <div
        style={{
          position: "relative",
          width: "92vw",
          maxWidth: "600px",
          minWidth: "320px",
          maxHeight: "90vh",
          overflowY: "auto",
          backgroundColor: "var(--card, #ffffff)",
          color: "var(--foreground, #0f172a)",
          border: "1px solid var(--border, #e2e8f0)",
          borderRadius: "24px",
          boxShadow: "0 25px 50px -12px rgba(0, 0, 0, 0.25)",
          padding: "24px",
          display: "flex",
          flexDirection: "column",
          gap: "20px",
          boxSizing: "border-box",
          flexShrink: 0,
        }}
      >
        {/* Modal Header */}
        <div style={{ display: "flex", alignItems: "flex-start", justifyContent: "space-between", borderBottom: "1px solid var(--border, #e2e8f0)", paddingBottom: "16px" }}>
          <div style={{ display: "flex", alignItems: "center", gap: "12px" }}>
            <div style={{ padding: "10px", background: "rgba(37, 99, 235, 0.1)", color: "#2563eb", borderRadius: "14px", display: "flex", alignItems: "center", justifyContent: "center" }}>
              <Zap style={{ width: "24px", height: "24px" }} />
            </div>
            <div>
              <h3 style={{ margin: 0, fontSize: "18px", fontWeight: "700", lineHeight: "1.3" }}>Pengujian Latensi Pencarian Data</h3>
              <p style={{ margin: "2px 0 0 0", fontSize: "13px", color: "var(--muted-foreground, #64748b)" }}>Uji kecepatan & latensi API pencarian secara langsung</p>
            </div>
          </div>
          <button
            onClick={onClose}
            style={{ border: "none", background: "transparent", cursor: "pointer", padding: "8px", color: "var(--muted-foreground, #64748b)", borderRadius: "10px", display: "flex", alignItems: "center", justifyContent: "center" }}
          >
            <X style={{ width: "20px", height: "20px" }} />
          </button>
        </div>

        {/* Page Selector Tabs */}
        <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
          <label style={{ fontSize: "11px", fontWeight: "700", textTransform: "uppercase", letterSpacing: "0.05em", color: "var(--muted-foreground, #64748b)" }}>
            Pilih Modul Pencarian
          </label>
          <div style={{ display: "flex", gap: "8px", overflowX: "auto", paddingBottom: "4px" }}>
            {TEST_PAGES.map(p => (
              <button
                key={p.name}
                onClick={() => { setSelectedPage(p); setResults([]); setProgressPct(0); }}
                disabled={isRunning}
                style={{
                  padding: "8px 14px",
                  borderRadius: "12px",
                  fontSize: "12px",
                  fontWeight: "600",
                  whiteSpace: "nowrap",
                  cursor: isRunning ? "not-allowed" : "pointer",
                  border: "none",
                  transition: "all 0.15s ease",
                  backgroundColor: selectedPage.name === p.name ? "var(--primary, #2563eb)" : "rgba(100, 116, 139, 0.12)",
                  color: selectedPage.name === p.name ? "#ffffff" : "var(--foreground, #0f172a)",
                }}
              >
                {p.name}
              </button>
            ))}
          </div>
        </div>

        {/* Control Box */}
        <div style={{ backgroundColor: "rgba(100, 116, 139, 0.08)", border: "1px solid var(--border, #e2e8f0)", borderRadius: "16px", padding: "16px", display: "flex", flexDirection: "column", gap: "12px" }}>
          <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", flexWrap: "wrap", gap: "12px" }}>
            <div>
              <span style={{ fontSize: "12px", color: "var(--muted-foreground, #64748b)", display: "block" }}>Endpoint Target</span>
              <span style={{ fontSize: "14px", fontWeight: "700", fontFamily: "monospace", color: "var(--foreground, #0f172a)" }}>{selectedPage.path}</span>
            </div>
            <Button onClick={runTest} disabled={isRunning} size="sm" className="rounded-xl">
              {isRunning ? <RefreshCw className="w-4 h-4 mr-2 animate-spin" /> : <Play className="w-4 h-4 mr-2" />}
              {isRunning ? `Menguji (${progressPct}%)...` : "Jalankan Pengujian"}
            </Button>
          </div>

          {/* Progress Bar */}
          {isRunning && (
            <div style={{ display: "flex", flexDirection: "column", gap: "6px", paddingTop: "4px" }}>
              <div style={{ display: "flex", justifyContent: "space-between", fontSize: "12px", color: "var(--muted-foreground, #64748b)" }}>
                <span>Menguji Kata Kunci: <strong>&quot;{currentQuery}&quot;</strong></span>
                <span style={{ fontFamily: "monospace" }}>{progressPct}%</span>
              </div>
              <div style={{ width: "100%", height: "8px", backgroundColor: "rgba(100, 116, 139, 0.2)", borderRadius: "9999px", overflow: "hidden" }}>
                <div
                  style={{
                    height: "100%",
                    backgroundColor: "var(--primary, #2563eb)",
                    borderRadius: "9999px",
                    width: `${progressPct}%`,
                    transition: "width 0.15s ease",
                  }}
                />
              </div>
            </div>
          )}
        </div>

        {/* Results Grid */}
        {results.length > 0 && (
          <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
            <h4 style={{ margin: 0, fontSize: "12px", fontWeight: "700", textTransform: "uppercase", letterSpacing: "0.05em", color: "var(--muted-foreground, #64748b)", display: "flex", alignItems: "center", gap: "6px" }}>
              <BarChart2 style={{ width: "16px", height: "16px", color: "var(--primary, #2563eb)" }} /> Hasil Pengujian ({results.length} Request)
            </h4>
            
            <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(110px, 1fr))", gap: "10px" }}>
              <div style={{ backgroundColor: "var(--card, #ffffff)", padding: "12px", borderRadius: "14px", border: "1px solid var(--border, #e2e8f0)", textAlign: "center" }}>
                <span style={{ fontSize: "10px", fontWeight: "700", textTransform: "uppercase", color: "var(--muted-foreground, #64748b)", display: "block" }}>Rata-Rata</span>
                <span style={{ fontSize: "18px", fontWeight: "700", display: "block", marginTop: "4px" }}>{avg} <span style={{ fontSize: "11px", fontWeight: "400" }}>ms</span></span>
              </div>
              <div style={{ backgroundColor: "var(--card, #ffffff)", padding: "12px", borderRadius: "14px", border: "1px solid var(--border, #e2e8f0)", textAlign: "center" }}>
                <span style={{ fontSize: "10px", fontWeight: "700", textTransform: "uppercase", color: "var(--muted-foreground, #64748b)", display: "block" }}>Median (P50)</span>
                <span style={{ fontSize: "18px", fontWeight: "700", color: "var(--primary, #2563eb)", display: "block", marginTop: "4px" }}>{p50} <span style={{ fontSize: "11px", fontWeight: "400" }}>ms</span></span>
              </div>
              <div style={{ backgroundColor: "var(--card, #ffffff)", padding: "12px", borderRadius: "14px", border: "1px solid var(--border, #e2e8f0)", textAlign: "center" }}>
                <span style={{ fontSize: "10px", fontWeight: "700", textTransform: "uppercase", color: "var(--muted-foreground, #64748b)", display: "block" }}>P90</span>
                <span style={{ fontSize: "18px", fontWeight: "700", color: "#d97706", display: "block", marginTop: "4px" }}>{p90} <span style={{ fontSize: "11px", fontWeight: "400" }}>ms</span></span>
              </div>
              <div style={{ backgroundColor: "var(--card, #ffffff)", padding: "12px", borderRadius: "14px", border: "1px solid var(--border, #e2e8f0)", textAlign: "center" }}>
                <span style={{ fontSize: "10px", fontWeight: "700", textTransform: "uppercase", color: "var(--muted-foreground, #64748b)", display: "block" }}>Min / Max</span>
                <span style={{ fontSize: "12px", fontWeight: "700", display: "block", marginTop: "6px" }}>
                  {min} ms <span style={{ fontWeight: "400", color: "#64748b" }}>/</span> {max} ms
                </span>
              </div>
            </div>

            {/* Individual Latency History Chips */}
            <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
              <span style={{ fontSize: "12px", fontWeight: "600", color: "var(--muted-foreground, #64748b)" }}>Sampel Waktu Respons (ms):</span>
              <div style={{ display: "flex", flexWrap: "wrap", gap: "6px", maxHeight: "100px", overflowY: "auto", padding: "10px", backgroundColor: "rgba(100, 116, 139, 0.08)", borderRadius: "14px", border: "1px solid var(--border, #e2e8f0)" }}>
                {results.map((ms, idx) => (
                  <span
                    key={idx}
                    style={{
                      padding: "4px 8px",
                      borderRadius: "8px",
                      fontSize: "11px",
                      fontFamily: "monospace",
                      fontWeight: "600",
                      backgroundColor: ms < 300 ? "rgba(16, 185, 129, 0.15)" : ms < 800 ? "rgba(245, 158, 11, 0.15)" : "rgba(244, 63, 94, 0.15)",
                      color: ms < 300 ? "#059669" : ms < 800 ? "#d97706" : "#e11d48",
                      border: `1px solid ${ms < 300 ? "rgba(16, 185, 129, 0.3)" : ms < 800 ? "rgba(245, 158, 11, 0.3)" : "rgba(244, 63, 94, 0.3)"}`,
                    }}
                  >
                    #{idx + 1}: {ms}ms
                  </span>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Modal Footer */}
        <div style={{ display: "flex", justifyContent: "flex-end", paddingTop: "12px", borderTop: "1px solid var(--border, #e2e8f0)" }}>
          <Button variant="outline" onClick={onClose} className="rounded-xl">
            Tutup
          </Button>
        </div>
      </div>
    </div>
  );

  if (!mounted || typeof window === "undefined") return null;
  return createPortal(modalContent, document.body);
}
