"use client";

import { useState, useRef } from "react";
import { Camera, Trash2, Plus, Image as ImageIcon, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export interface PhotoItem {
  id: string;
  file?: File;
  previewUrl: string;
  keterangan: string;
  existingPhotoId?: string; // If photo was already uploaded to server
}

interface PhotoUploaderWithKeteranganProps {
  photos: PhotoItem[];
  onChange: (photos: PhotoItem[]) => void;
  disabled?: boolean;
  maxPhotos?: number;
}

export function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => resolve(reader.result as string);
    reader.onerror = (error) => reject(error);
  });
}

export function dataURLtoFile(dataurl: string, filename: string): File {
  try {
    const arr = dataurl.split(",");
    const mime = arr[0].match(/:(.*?);/)?.[1] || "image/jpeg";
    const bstr = atob(arr[1]);
    let n = bstr.length;
    const u8arr = new Uint8Array(n);
    while (n--) {
      u8arr[n] = bstr.charCodeAt(n);
    }
    return new File([u8arr], filename, { type: mime });
  } catch {
    return new File([], filename, { type: "image/jpeg" });
  }
}

export function PhotoUploaderWithKeterangan({
  photos,
  onChange,
  disabled = false,
  maxPhotos = 5,
}: PhotoUploaderWithKeteranganProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    if (files.length === 0) return;

    const newItems: PhotoItem[] = await Promise.all(
      files.map(async (file) => {
        let base64 = "";
        try {
          base64 = await fileToBase64(file);
        } catch {
          base64 = URL.createObjectURL(file);
        }
        return {
          id: `photo_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`,
          file,
          previewUrl: base64,
          keterangan: "",
        };
      })
    );

    const updated = [...photos, ...newItems].slice(0, maxPhotos);
    onChange(updated);

    // Reset file input
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  const handleKeteranganChange = (id: string, text: string) => {
    const updated = photos.map((item) =>
      item.id === id ? { ...item, keterangan: text } : item
    );
    onChange(updated);
  };

  const handleRemove = (id: string) => {
    const target = photos.find((p) => p.id === id);
    if (target?.previewUrl && target.file && target.previewUrl.startsWith("blob:")) {
      URL.revokeObjectURL(target.previewUrl);
    }
    const updated = photos.filter((p) => p.id !== id);
    onChange(updated);
  };

  return (
    <div className="space-y-3 mt-3 rounded-xl border border-destructive/20 bg-destructive/5 p-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm font-semibold text-destructive">
          <Camera className="h-4 w-4" />
          <span>Foto Bukti Temuan (NG) & Keterangan per Foto</span>
        </div>
        <span className="text-xs text-muted-foreground font-medium">
          {photos.length} / {maxPhotos} foto
        </span>
      </div>

      {/* List of uploaded photos with individual keterangan */}
      {photos.length > 0 && (
        <div className="space-y-3">
          {photos.map((item, index) => (
            <div
              key={item.id}
              className="flex flex-col sm:flex-row items-start gap-3 rounded-lg border border-border bg-card p-3 shadow-xs"
            >
              {/* Photo Thumbnail */}
              <div className="relative group shrink-0 w-full sm:w-28 h-28 rounded-md overflow-hidden bg-muted border border-border">
                <img
                  src={item.previewUrl}
                  alt={`Bukti temuan ${index + 1}`}
                  className="w-full h-full object-cover"
                  onError={(e) => {
                    (e.currentTarget as HTMLImageElement).style.display = "none";
                  }}
                />
                {!disabled && (
                  <button
                    type="button"
                    onClick={() => handleRemove(item.id)}
                    className="absolute top-1.5 right-1.5 flex h-7 w-7 items-center justify-center rounded-full bg-destructive text-white shadow-md hover:bg-destructive/90 transition-all"
                    title="Hapus foto ini"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                )}
                <span className="absolute bottom-1 left-1 rounded bg-black/60 px-1.5 py-0.5 text-[10px] font-medium text-white">
                  Foto #{index + 1}
                </span>
              </div>

              {/* Individual Keterangan Input */}
              <div className="flex-1 w-full space-y-1.5">
                <label htmlFor={`ket_${item.id}`} className="text-xs font-semibold flex items-center gap-1.5">
                  <ImageIcon className="h-3.5 w-3.5 text-primary" />
                  Keterangan Khusus Foto #{index + 1} <span className="text-destructive">*</span>
                </label>
                <Input
                  id={`ket_${item.id}`}
                  value={item.keterangan}
                  onChange={(e) => handleKeteranganChange(item.id, e.target.value)}
                  disabled={disabled}
                  placeholder={`Contoh: Kerusakan pada selang bagian kanan #${index + 1}`}
                  className={`text-xs ${
                    !item.keterangan.trim() ? "border-amber-500/50 focus:border-amber-500" : ""
                  }`}
                />
                {!item.keterangan.trim() && (
                  <p className="text-[11px] text-amber-600 dark:text-amber-400 flex items-center gap-1">
                    <AlertCircle className="h-3 w-3 shrink-0" />
                    Harap isi keterangan spesifik untuk foto bukti ini.
                  </p>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Add Photo Button */}
      {!disabled && photos.length < maxPhotos && (
        <div>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/png,image/jpeg,image/jpg,image/webp"
            multiple
            onChange={handleFileSelect}
            className="hidden"
            id="photo-upload-input"
          />
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => fileInputRef.current?.click()}
            className="w-full border-dashed border-destructive/40 text-destructive hover:bg-destructive/10 hover:border-destructive gap-2 text-xs py-2.5"
          >
            <Plus className="h-4 w-4" />
            Tambah Foto Bukti Temuan ({photos.length === 0 ? "Wajib min 1 foto" : "Tambah lagi"})
          </Button>
        </div>
      )}
    </div>
  );
}
