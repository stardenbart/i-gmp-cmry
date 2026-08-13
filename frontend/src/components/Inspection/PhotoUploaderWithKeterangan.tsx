"use client";

import { useState, useRef } from "react";
import { Camera, Trash2, Plus, Image as ImageIcon, AlertCircle, Tag } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useMasterHEI, useMasterHEICategories } from "@/hooks/useMasterData";

export interface PhotoItem {
  id: string;
  file?: File;
  previewUrl: string;
  keterangan: string;
  existingPhotoId?: string; // If photo was already uploaded to server
  hei_id?: string;
  hei_category?: string;
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

export function formatPhotoUrl(url?: string, inspectionId?: string): string {
  if (!url) return "";
  let formatted = url.trim();
  if (!formatted) return "";

  // 1. Data URLs (base64) returned immediately
  if (formatted.startsWith("data:")) {
    return formatted;
  }

  // Detect current hostname and protocol when running in browser
  let currentHost = "localhost";
  let protocol = "http:";
  let cleanHost = "";

  if (typeof window !== "undefined") {
    currentHost = window.location.hostname;
    protocol = window.location.protocol;
    if (protocol === "https:" || currentHost.includes("ngrok") || window.location.port === "") {
      cleanHost = `${protocol}//${window.location.host}`;
    } else {
      cleanHost = `${protocol}//${currentHost}:8080`;
    }
  } else {
    const apiHost = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
    cleanHost = apiHost.replace(/\/api\/v1\/?$/, "");
  }

  // If expired blob URL (blob:http://... or blob:https://...)
  if (formatted.startsWith("blob:")) {
    const lastSlashIdx = formatted.lastIndexOf("/");
    if (lastSlashIdx !== -1) {
      formatted = formatted.substring(lastSlashIdx + 1);
    } else {
      formatted = formatted.replace(/^blob:/, "");
    }
  }

  // Handle MinIO bucket photos via Next.js proxy route (/monitoring-audit-bucket/)
  if (formatted.startsWith("/monitoring-audit-bucket/") || formatted.includes("/monitoring-audit-bucket/")) {
    const bucketIdx = formatted.indexOf("/monitoring-audit-bucket/");
    const cleanBucketPath = formatted.substring(bucketIdx);
    return typeof window !== "undefined"
      ? `${protocol}//${window.location.host}${cleanBucketPath}`
      : cleanBucketPath;
  }
  if (formatted.includes("issues/") || formatted.includes(":9000/")) {
    const issueIdx = formatted.indexOf("issues/");
    if (issueIdx !== -1) {
      const pathAfterIssue = formatted.substring(issueIdx);
      return typeof window !== "undefined"
        ? `${protocol}//${window.location.host}/monitoring-audit-bucket/${pathAfterIssue}`
        : `/monitoring-audit-bucket/${pathAfterIssue}`;
    }
  }

  // Local uploads folder
  if (formatted.startsWith("/uploads/") || formatted.startsWith("uploads/")) {
    const cleanP = formatted.startsWith("/") ? formatted : `/${formatted}`;
    return `${cleanHost}${cleanP}`;
  }

  // Upgrade HTTP to HTTPS if page is HTTPS to eliminate Mixed Content errors
  if (protocol === "https:" && formatted.startsWith("http://")) {
    formatted = formatted.replace(/^http:\/\//, "https://").replace(/:9000|:8080/g, "");
    return formatted;
  }

  // Replace hardcoded localhost or 127.0.0.1 with current connected server hostname
  if (formatted.includes("localhost") || formatted.includes("127.0.0.1")) {
    formatted = formatted.replace(/localhost|127\.0\.0\.1/g, currentHost);
  }

  if (
    formatted.startsWith("http://") ||
    formatted.startsWith("https://") ||
    formatted.startsWith("data:")
  ) {
    return formatted;
  }

  // Handle bare UUID / filename without slash (e.g. 647fdba7-d04c-46b0-a4de-284a290fbb76)
  if (!formatted.includes("/")) {
    let inspId = inspectionId;
    if (!inspId && typeof window !== "undefined") {
      const match = window.location.pathname.match(/\/(INSP-[A-Za-z0-9_-]+)/i);
      if (match) inspId = match[1];
    }
    if (inspId) {
      formatted = `uploads/${inspId}/${formatted}`;
    } else {
      formatted = `uploads/${formatted}`;
    }
  } else if (!formatted.startsWith("/uploads/") && !formatted.startsWith("uploads/")) {
    formatted = `uploads/${formatted.replace(/^\/+/, '')}`;
  }

  const cleanPath = formatted.startsWith("/") ? formatted : `/${formatted}`;
  return `${cleanHost}${cleanPath}`;
}

export function PhotoUploaderWithKeterangan({
  photos,
  onChange,
  disabled = false,
  // maxPhotos = 5,
}: PhotoUploaderWithKeteranganProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Fetch dynamic HEI items and categories
  const { data: heiMasterRes } = useMasterHEI("", 1000);
  const { data: categoriesRes } = useMasterHEICategories();

  const heiItems = heiMasterRes?.items || [];
  const categories = categoriesRes || ["Habit", "Equipment", "Infrastructure"];

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    if (files.length === 0) return;

    const newItems: PhotoItem[] = await Promise.all(
      files.map(async (file) => {
        const objectUrl = URL.createObjectURL(file);
        return {
          id: `photo_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`,
          file,
          previewUrl: objectUrl,
          keterangan: "",
          hei_id: "",
          hei_category: "",
        };
      })
    );

    const updated = [...photos, ...newItems].slice(0);
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

  const handleHEISelect = (id: string, selectedValue: string) => {
    const updated = photos.map((item) => {
      if (item.id === id) {
        const matchedItem = heiItems.find((h: any) => h.hei_id === selectedValue || h.category_name === selectedValue);
        if (matchedItem) {
          return {
            ...item,
            hei_id: matchedItem.hei_id,
            hei_category: matchedItem.category_name,
          };
        }
        return {
          ...item,
          hei_id: "",
          hei_category: selectedValue,
        };
      }
      return item;
    });
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
          <Camera className="h-6 w-6" />
          <span className="text-xs">Foto Bukti Temuan (NG) & Keterangan per Foto</span>
        </div>
        <span className="text-xs text-muted-foreground font-medium">
          {photos.length} foto
        </span>
      </div>

      {/* List of uploaded photos with individual keterangan & HEI */}
      {photos.length > 0 && (
        <div className="space-y-3">
          {photos.map((item, index) => {
            const activeCategory = item.hei_category || "";
            const rawUrl = item.previewUrl ||
              (item as any).file_url ||
              (item as any).file_path ||
              (item as any).photo_url ||
              (item as any).image_url ||
              (item as any).url ||
              "";
            // base64 dan blob URL dipakai langsung tanpa formatPhotoUrl
            const displayUrl = rawUrl.startsWith("data:") || rawUrl.startsWith("blob:")
              ? rawUrl
              : item.file
                ? URL.createObjectURL(item.file)
                : formatPhotoUrl(rawUrl);

            return (
              <div
                key={item.id}
                className="flex flex-col sm:flex-row items-start gap-3 rounded-lg border border-border bg-card p-3 shadow-xs"
              >
                {/* Photo Thumbnail */}
                <div className="relative group shrink-0 w-full sm:w-28 h-28 rounded-md overflow-hidden bg-muted border border-border">
                  {displayUrl ? (
                    <img
                      src={displayUrl}
                      alt={`Bukti temuan ${index + 1}`}
                      className="w-full h-full object-cover"
                      onError={(e) => {
                        const target = e.currentTarget as HTMLImageElement;
                        const currentSrc = target.src;
                        const extensions = [".png", ".jpeg", ".jpg", ".webp", ".heic", ".heif"];
                        
                        let tried: string[] = [];
                        try {
                          tried = JSON.parse(target.dataset.triedExts || "[]");
                        } catch {}

                        const baseUrl = currentSrc.replace(/\.(png|jpeg|jpg|webp|heic|heif)$/i, "");
                        const nextExt = extensions.find((ext) => !tried.includes(ext));

                        if (nextExt) {
                          target.dataset.triedExts = JSON.stringify([...tried, nextExt]);
                          target.src = baseUrl + nextExt;
                          return;
                        }

                        // Fallback to local file object blob if available
                        if (item.file) {
                          target.src = URL.createObjectURL(item.file);
                          return;
                        }

                        // If all format extensions fail, display broken badge UI
                        const parentNode = target.parentNode as HTMLElement;
                        if (parentNode && !parentNode.querySelector(".broken-badge")) {
                          target.style.display = "none";
                          const fallbackDiv = document.createElement("div");
                          fallbackDiv.className = "broken-badge w-full h-full flex flex-col items-center justify-center bg-red-500/10 text-red-500 p-2 text-center text-[10px] font-medium gap-1";
                          fallbackDiv.innerHTML = `<svg class="h-4 w-4 shrink-0 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/></svg><span>Berkas Hilang</span><span class="text-[9px] text-muted-foreground">Klik Sampah (Hapus)</span>`;
                          parentNode.appendChild(fallbackDiv);
                        }
                      }}
                    />
                  ) : (
                    <div className="w-full h-full flex flex-col items-center justify-center text-muted-foreground bg-muted text-[10px] p-2 text-center gap-1">
                      <ImageIcon className="h-4 w-4 opacity-50" />
                      <span>Foto bukti temuan #{index + 1}</span>
                    </div>
                  )}
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

                {/* Individual Keterangan & HEI Input */}
                <div className="flex-1 w-full space-y-2.5">
                  <div className="space-y-1">
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

                  {/* Dynamic HEI Selection (Mandatory Select per Photo) */}
                  <div className="pt-2 border-t border-border/60 space-y-1.5">
                    <div className="w-full sm:w-1/2 space-y-1">
                      <label htmlFor={`hei_cat_${item.id}`} className="text-[11px] font-semibold flex items-center gap-1">
                        <Tag className="h-3 w-3 text-primary" />
                        Kategori HEI (Habit, Equipment, Infrastructure) <span className="text-destructive">*</span>
                      </label>
                      <select
                        id={`hei_cat_${item.id}`}
                        aria-label={`Kategori HEI untuk foto #${index + 1}`}
                        value={item.hei_id || item.hei_category || ""}
                        onChange={(e) => handleHEISelect(item.id, e.target.value)}
                        disabled={disabled}
                        className={`flex h-8 w-full rounded-md border bg-background px-2 py-1 text-xs shadow-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary disabled:opacity-50 ${
                          !item.hei_id && !item.hei_category ? "border-amber-500/80 focus:border-amber-500" : "border-input"
                        }`}
                      >
                        <option value="">-- Pilih Kategori / Item HEI (Wajib) --</option>
                        {categories.map((cat) => {
                          const itemsInCat = heiItems.filter((h: any) => h.category_name?.toLowerCase() === cat.toLowerCase());
                          if (itemsInCat.length === 0) {
                            return (
                              <option key={cat} value={cat}>
                                {cat}
                              </option>
                            );
                          }
                          return (
                            <optgroup key={cat} label={`Kategori: ${cat}`}>
                              {itemsInCat.map((h: any) => (
                                <option key={h.hei_id} value={h.hei_id}>
                                  {h.hei_code ? `[${h.hei_code}] ` : ""}{h.hei_name}
                                </option>
                              ))}
                            </optgroup>
                          );
                        })}
                      </select>
                      {!item.hei_id && !item.hei_category && (
                        <p className="text-[11px] text-amber-600 dark:text-amber-400 flex items-center gap-1">
                          <AlertCircle className="h-3 w-3 shrink-0" />
                          Kategori / Item HEI wajib dipilih untuk foto temuan ini.
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Add Photo Button */}
      {!disabled && (
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


