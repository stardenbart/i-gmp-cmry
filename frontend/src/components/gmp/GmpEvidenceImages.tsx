"use client";

import { Eye } from "lucide-react";

import { cn, formatImageUrl } from "@/lib/utils";
import type { GmpFollowUpEvidence } from "@/types/api/dashboard";

interface GmpEvidenceImagesProps {
  imageUrl?: string;
  imageUrls?: string[];
  evidence?: GmpFollowUpEvidence[];
  label: string;
  onPreview: (url: string) => void;
  variant?: "initial" | "follow-up";
}

export function GmpEvidenceImages({
  imageUrl,
  imageUrls,
  evidence,
  label,
  onPreview,
  variant = "initial",
}: GmpEvidenceImagesProps) {
  const fallbackImages = Array.from(new Set(
    (Array.isArray(imageUrls) && imageUrls.length > 0
      ? imageUrls
      : imageUrl ? [imageUrl] : []
    ).filter(Boolean)
  ));
  const items = evidence && evidence.length > 0
    ? evidence.map((item) => ({
        key: item.photo_id || item.image_url,
        imageUrl: item.image_url,
        description: item.keterangan,
      }))
    : fallbackImages.map((url) => ({ key: url, imageUrl: url, description: undefined }));

  if (items.length === 0) {
    return <span className="text-xs italic text-muted-foreground">-</span>;
  }

  return (
    <div className="flex min-w-max items-center justify-center gap-1.5">
      {items.map((item, index) => {
        const resolvedUrl = formatImageUrl(item.imageUrl);
        return (
          <button
            type="button"
            key={`${index}-${item.key}`}
            className={cn(
              "group relative h-12 w-16 shrink-0 cursor-pointer overflow-hidden rounded-lg border bg-muted shadow-sm",
              variant === "follow-up" ? "border-emerald-500/40" : "border-border/60"
            )}
            onClick={() => onPreview(resolvedUrl)}
            title={item.description
              ? `${label} ${index + 1}: ${item.description}`
              : `Buka ${label.toLowerCase()} ${index + 1} dari ${items.length}`}
          >
            {/* Dynamic authenticated upload URLs intentionally use a native image element. */}
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={resolvedUrl || "/placeholder.png"}
              alt={`${label} ${index + 1}`}
              loading="lazy"
              decoding="async"
              onError={(event) => { event.currentTarget.src = "/placeholder.png"; }}
              className="h-full w-full object-cover transition-transform group-hover:scale-110"
            />
            <span className="absolute left-1 top-1 rounded bg-black/65 px-1 text-[9px] font-semibold text-white">
              {index + 1}/{items.length}
            </span>
            <div className="absolute inset-0 flex items-center justify-center bg-black/40 text-white opacity-0 transition-opacity group-hover:opacity-100">
              <Eye className="h-3.5 w-3.5" />
            </div>
          </button>
        );
      })}
    </div>
  );
}

export function GmpFollowUpDescriptions({ evidence }: { evidence?: GmpFollowUpEvidence[] }) {
  if (!evidence || evidence.length === 0) {
    return <span className="text-xs italic text-muted-foreground">-</span>;
  }

  return (
    <div className="min-w-[220px] max-w-sm space-y-1.5">
      {evidence.map((item, index) => (
        <div key={item.photo_id || `${index}-${item.image_url}`} className="rounded-lg border border-emerald-500/20 bg-emerald-500/5 px-2.5 py-2">
          <span className="text-[10px] font-bold uppercase tracking-wide text-emerald-600">
            Follow-Up {index + 1}
          </span>
          <p className="mt-0.5 whitespace-pre-wrap text-xs leading-relaxed text-foreground">
            {item.keterangan?.trim() || "Tanpa keterangan"}
          </p>
        </div>
      ))}
    </div>
  );
}
