"use client";

import { ArrowUp, CornerRightDown, Layers3, RotateCcw } from "lucide-react";

interface DrillDownToolbarProps {
  currentDimensionLabel: string;
  nextDimensionLabel?: string;
  onDrillDown: () => void;
  onDrillUp: () => void;
  onReset: () => void;
  canDrillUp: boolean;
  isBusy?: boolean;
}

export function DrillDownToolbar({
  currentDimensionLabel,
  nextDimensionLabel,
  onDrillDown,
  onDrillUp,
  onReset,
  canDrillUp,
  isBusy = false,
}: DrillDownToolbarProps) {
  return (
    <div className="mb-3 rounded-lg border border-primary/20 bg-primary/5 p-2.5" aria-label="Kontrol drill-down">
      <div className="mb-2 flex flex-wrap items-center justify-between gap-1.5">
        <div className="flex min-w-0 items-center gap-1.5">
          <Layers3 className="h-3.5 w-3.5 shrink-0 text-primary" />
          <p className="truncate text-[11px] font-semibold text-foreground">
            {nextDimensionLabel ? `Drill-down: ${currentDimensionLabel} → ${nextDimensionLabel}` : `Level terakhir: ${currentDimensionLabel}`}
          </p>
        </div>
        {canDrillUp && (
          <div className="flex items-center gap-1">
            <button
              type="button"
              onClick={onDrillUp}
              disabled={isBusy}
              className="inline-flex items-center gap-1 rounded-md px-1.5 py-1 text-[10px] font-semibold text-primary hover:bg-primary/10 disabled:opacity-50"
              title="Kembali satu tingkat"
            >
              <ArrowUp className="h-3 w-3" /> Naik
            </button>
            <button
              type="button"
              onClick={onReset}
              disabled={isBusy}
              className="inline-flex items-center gap-1 rounded-md px-1.5 py-1 text-[10px] font-semibold text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
              title="Kembali ke kategori utama"
            >
              <RotateCcw className="h-3 w-3" /> Reset
            </button>
          </div>
        )}
      </div>

      {nextDimensionLabel && (
        <div className="flex flex-col gap-1.5 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-[10px] text-muted-foreground">
            Tambahkan seluruh {nextDimensionLabel} di bawah setiap {currentDimensionLabel}, tanpa memilih satu value.
          </p>
          <button
            type="button"
            onClick={onDrillDown}
            disabled={isBusy}
            className="inline-flex h-8 shrink-0 items-center justify-center gap-1.5 rounded-md bg-primary px-3 text-[11px] font-semibold text-primary-foreground shadow-sm transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
            title={`Tampilkan semua data berdasarkan ${nextDimensionLabel}`}
          >
            <CornerRightDown className="h-3.5 w-3.5" /> Tampilkan semua {nextDimensionLabel}
          </button>
        </div>
      )}
    </div>
  );
}
