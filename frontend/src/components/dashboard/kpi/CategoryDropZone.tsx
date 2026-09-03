"use client";

import { useState, type DragEvent } from "react";
import { ChevronLeft, ChevronRight, GripVertical, X } from "lucide-react";
import type { CatalogField } from "@/lib/api/analytics.api";
import { KPI_CATEGORY_DRAG_MIME, readFieldDragPayload } from "./visualizationDnd";

interface CategoryDropZoneProps {
  items: CatalogField[];
  onChange: (ids: string[]) => void;
  matrixMode: boolean;
  maxItems: number;
}

function moveItem(ids: string[], sourceId: string, targetId: string): string[] {
  if (sourceId === targetId) return ids;
  const sourceIndex = ids.indexOf(sourceId);
  const targetIndex = ids.indexOf(targetId);
  if (sourceIndex < 0 || targetIndex < 0) return ids;
  const next = [...ids];
  next.splice(sourceIndex, 1);
  next.splice(targetIndex, 0, sourceId);
  return next;
}

export function CategoryDropZone({ items, onChange, matrixMode, maxItems }: CategoryDropZoneProps) {
  const [isOver, setIsOver] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const ids = items.map((item) => item.id);

  const addCategory = (id: string, beforeId?: string) => {
    if (ids.includes(id)) {
      setMessage("Kategori tersebut sudah ditambahkan.");
      return;
    }
    if (items.length >= maxItems) {
      setMessage(matrixMode ? "Heatmap/Sankey hanya mendukung tepat dua kategori." : `Maksimal ${maxItems} kategori.`);
      return;
    }
    const next = [...ids];
    const targetIndex = beforeId ? next.indexOf(beforeId) : -1;
    if (targetIndex >= 0) next.splice(targetIndex, 0, id);
    else next.push(id);
    setMessage(null);
    onChange(next);
  };

  const handleDrop = (event: DragEvent<HTMLDivElement>, beforeId?: string) => {
    event.preventDefault();
    event.stopPropagation();
    setIsOver(false);

    const movedCategoryId = event.dataTransfer.getData(KPI_CATEGORY_DRAG_MIME);
    if (movedCategoryId) {
      if (beforeId) onChange(moveItem(ids, movedCategoryId, beforeId));
      else {
        const withoutMoved = ids.filter((id) => id !== movedCategoryId);
        onChange([...withoutMoved, movedCategoryId]);
      }
      setMessage(null);
      return;
    }

    const payload = readFieldDragPayload(event.dataTransfer);
    if (!payload) return;
    if (payload.kind !== "dimension") {
      setMessage("Kotak Kategori hanya menerima field Kategori.");
      return;
    }
    addCategory(payload.id, beforeId);
  };

  const moveByOffset = (index: number, offset: -1 | 1) => {
    const targetIndex = index + offset;
    if (targetIndex < 0 || targetIndex >= ids.length) return;
    const next = [...ids];
    [next[index], next[targetIndex]] = [next[targetIndex], next[index]];
    setMessage(null);
    onChange(next);
  };

  return (
    <div>
      <div className="mb-1.5 flex items-center justify-between gap-2">
        <p className="text-xs font-semibold text-foreground">Kategori (maks. {maxItems})</p>
        {items.length > 1 && <span className="text-[10px] text-muted-foreground">Urutan menentukan hierarki</span>}
      </div>
      <div
        onDragOver={(event) => {
          event.preventDefault();
          setIsOver(true);
        }}
        onDragLeave={(event) => {
          if (!event.currentTarget.contains(event.relatedTarget as Node | null)) setIsOver(false);
        }}
        onDrop={(event) => handleDrop(event)}
        className={`min-h-[82px] rounded-xl border-2 border-dashed p-3 transition-colors ${
          isOver ? "border-primary bg-primary/5" : "border-border bg-muted/20"
        }`}
      >
        {items.length === 0 ? (
          <div className="flex min-h-14 items-center justify-center text-center">
            <span className="text-xs italic text-muted-foreground">Seret satu atau beberapa field Kategori ke sini</span>
          </div>
        ) : (
          <div className="flex flex-wrap gap-2">
            {items.map((item, index) => {
              const role = matrixMode ? (index === 0 ? "Sumbu 1" : "Sumbu 2") : index === 0 ? "Utama" : `Drill ${index}`;
              return (
                <div
                  key={item.id}
                  draggable
                  onDragStart={(event) => {
                    event.dataTransfer.setData(KPI_CATEGORY_DRAG_MIME, item.id);
                    event.dataTransfer.effectAllowed = "move";
                  }}
                  onDragOver={(event) => event.preventDefault()}
                  onDrop={(event) => handleDrop(event, item.id)}
                  className="group flex items-center gap-1 rounded-lg border border-primary/20 bg-primary/10 px-1.5 py-1 text-primary"
                >
                  <GripVertical className="h-3.5 w-3.5 cursor-grab text-primary/60 active:cursor-grabbing" />
                  <span className="rounded bg-background/70 px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wide">{index + 1} · {role}</span>
                  <span className="max-w-36 truncate text-xs font-semibold" title={item.label}>{item.label}</span>
                  <button
                    type="button"
                    onClick={() => moveByOffset(index, -1)}
                    disabled={index === 0}
                    aria-label={`Geser ${item.label} ke kiri`}
                    className="rounded p-0.5 hover:bg-primary/10 disabled:cursor-not-allowed disabled:opacity-30"
                  >
                    <ChevronLeft className="h-3 w-3" />
                  </button>
                  <button
                    type="button"
                    onClick={() => moveByOffset(index, 1)}
                    disabled={index === items.length - 1}
                    aria-label={`Geser ${item.label} ke kanan`}
                    className="rounded p-0.5 hover:bg-primary/10 disabled:cursor-not-allowed disabled:opacity-30"
                  >
                    <ChevronRight className="h-3 w-3" />
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setMessage(null);
                      onChange(ids.filter((id) => id !== item.id));
                    }}
                    aria-label={`Hapus ${item.label}`}
                    className="rounded p-0.5 hover:bg-red-500/10 hover:text-red-500"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </div>
              );
            })}
          </div>
        )}
      </div>
      <p className={`mt-1 text-[11px] ${message ? "font-medium text-red-500" : "text-muted-foreground"}`} role={message ? "alert" : undefined}>
        {message ?? (matrixMode
          ? "Dua kategori menjadi sumbu pertama dan kedua Heatmap/Sankey."
          : "Kategori pertama menjadi kategori utama; setiap level berikutnya menampilkan seluruh value kategori saat drill-down.")}
      </p>
    </div>
  );
}
