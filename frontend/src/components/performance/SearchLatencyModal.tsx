"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { BarChart2, Play, RefreshCw, X, Zap } from "lucide-react";

import { Button } from "@/components/ui/button";
import { api } from "@/lib/api/axios";

interface SearchTarget {
  name: string;
  path: string;
  sampleQueries: string[];
}

interface SearchLatencyModalProps {
  pageName: string;
  apiPath: string;
  onClose: () => void;
}

const SEARCH_TARGETS: SearchTarget[] = [
  { name: "Inspeksi Digital", path: "/inspections", sampleQueries: ["Session", "AREA", "Completed", "Ongoing", "Draft"] },
  { name: "Temuan", path: "/issues", sampleQueries: ["Open", "Kebersihan", "Lantai", "Overdue", "GMP"] },
  { name: "GMP Data", path: "/dashboard/stats", sampleQueries: ["Area", "QC", "Kawasan", "Nilai", "2026"] },
  { name: "Activity Log", path: "/logs", sampleQueries: ["Login", "Update", "Create", "Delete", "Admin"] },
];

function percentile(sortedValues: number[], ratio: number) {
  if (sortedValues.length === 0) return 0;
  return sortedValues[Math.min(Math.floor(sortedValues.length * ratio), sortedValues.length - 1)];
}

export function SearchLatencyModal({ pageName, apiPath, onClose }: SearchLatencyModalProps) {
  const fallbackTarget = useMemo<SearchTarget>(() => ({ name: pageName, path: apiPath, sampleQueries: ["test"] }), [apiPath, pageName]);
  const [selectedTarget, setSelectedTarget] = useState(
    SEARCH_TARGETS.find((target) => target.name === pageName || target.path === apiPath) ?? fallbackTarget
  );
  const [isRunning, setIsRunning] = useState(false);
  const [results, setResults] = useState<number[]>([]);
  const [currentQuery, setCurrentQuery] = useState("");
  const abortRef = useRef<AbortController | null>(null);

  useEffect(() => {
    const previousOverflow = document.body.style.overflow;
    const handleEscape = (event: KeyboardEvent) => event.key === "Escape" && onClose();
    document.body.style.overflow = "hidden";
    document.addEventListener("keydown", handleEscape);
    return () => {
      abortRef.current?.abort();
      document.body.style.overflow = previousOverflow;
      document.removeEventListener("keydown", handleEscape);
    };
  }, [onClose]);

  const runTest = async () => {
    const controller = new AbortController();
    abortRef.current?.abort();
    abortRef.current = controller;
    setIsRunning(true);
    setResults([]);

    const samples: number[] = [];
    for (let index = 0; index < 15 && !controller.signal.aborted; index += 1) {
      if (document.hidden) break;
      const query = selectedTarget.sampleQueries[index % selectedTarget.sampleQueries.length];
      setCurrentQuery(query);
      const startedAt = performance.now();
      try {
        await api.get(selectedTarget.path, {
          params: { q: query, search: query, page: 1, limit: 10 },
          signal: controller.signal,
        });
      } catch {
        if (controller.signal.aborted) break;
      }
      samples.push(Math.round(performance.now() - startedAt));
      setResults([...samples]);
      await new Promise((resolve) => window.setTimeout(resolve, 80));
    }
    if (!controller.signal.aborted) setIsRunning(false);
  };

  const sortedResults = [...results].sort((a, b) => a - b);
  const average = results.length ? Math.round(results.reduce((total, value) => total + value, 0) / results.length) : 0;
  const progress = Math.round((results.length / 15) * 100);

  return createPortal(
    <div className="fixed inset-0 z-[9999] flex items-start justify-center overflow-y-auto bg-black/70 p-3 backdrop-blur-sm sm:items-center sm:p-6" role="dialog" aria-modal="true" aria-labelledby="search-latency-title" onClick={onClose}>
      <div className="my-auto flex max-h-[calc(100dvh-1.5rem)] w-full max-w-2xl flex-col overflow-y-auto rounded-2xl border border-border bg-card p-4 shadow-2xl sm:p-6" onClick={(event) => event.stopPropagation()}>
        <header className="flex items-start justify-between gap-4 border-b border-border pb-4">
          <div className="flex min-w-0 items-center gap-3">
            <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary"><Zap className="h-5 w-5" /></span>
            <div className="min-w-0">
              <h2 id="search-latency-title" className="font-bold">Pengujian Latensi Pencarian</h2>
              <p className="text-xs text-muted-foreground">15 request terukur ke endpoint pencarian</p>
            </div>
          </div>
          <button type="button" onClick={onClose} className="rounded-lg p-2 text-muted-foreground hover:bg-muted" aria-label="Tutup"><X className="h-5 w-5" /></button>
        </header>

        <section className="mt-4 space-y-2">
          <p className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground">Pilih modul</p>
          <div className="flex gap-2 overflow-x-auto pb-1">
            {SEARCH_TARGETS.map((target) => (
              <button key={target.path} type="button" disabled={isRunning} onClick={() => { setSelectedTarget(target); setResults([]); }} className={`whitespace-nowrap rounded-xl px-3 py-2 text-xs font-semibold ${selectedTarget.path === target.path ? "bg-primary text-primary-foreground" : "bg-muted text-foreground"}`}>
                {target.name}
              </button>
            ))}
          </div>
        </section>

        <section className="mt-4 rounded-xl border border-border bg-muted/30 p-4">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div><p className="text-xs text-muted-foreground">Endpoint target</p><code className="text-sm font-bold">{selectedTarget.path}</code></div>
            <Button type="button" size="sm" onClick={runTest} disabled={isRunning} className="rounded-xl">
              {isRunning ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : <Play className="mr-2 h-4 w-4" />}
              {isRunning ? `Menguji ${progress}%` : "Jalankan Pengujian"}
            </Button>
          </div>
          {isRunning && (
            <div className="mt-3 space-y-1.5">
              <div className="flex justify-between text-xs text-muted-foreground"><span>Kata kunci: <strong>{currentQuery}</strong></span><span>{progress}%</span></div>
              <div className="h-2 overflow-hidden rounded-full bg-muted"><div className="h-full rounded-full bg-primary transition-[width]" style={{ width: `${progress}%` }} /></div>
            </div>
          )}
        </section>

        {results.length > 0 && (
          <section className="mt-4 space-y-3">
            <h3 className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-muted-foreground"><BarChart2 className="h-4 w-4 text-primary" /> Hasil ({results.length} request)</h3>
            <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
              {[["Rata-rata", average], ["Median (P50)", percentile(sortedResults, 0.5)], ["P90", percentile(sortedResults, 0.9)], ["Maksimum", sortedResults.at(-1) ?? 0]].map(([label, value]) => (
                <div key={label} className="rounded-xl border border-border p-3 text-center"><p className="text-[10px] font-bold uppercase text-muted-foreground">{label}</p><p className="mt-1 text-lg font-bold">{value} <span className="text-xs font-normal">ms</span></p></div>
              ))}
            </div>
            <div className="flex max-h-24 flex-wrap gap-1.5 overflow-y-auto rounded-xl bg-muted/30 p-3">
              {results.map((value, index) => <span key={`${index}-${value}`} className="rounded-md border border-border bg-card px-2 py-1 font-mono text-[11px]">#{index + 1}: {value}ms</span>)}
            </div>
          </section>
        )}
      </div>
    </div>,
    document.body
  );
}
