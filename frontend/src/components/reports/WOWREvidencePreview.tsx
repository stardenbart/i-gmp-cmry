import { Image as ImageIcon } from "lucide-react";

import type { WOWRReportEvidence } from "@/lib/api/issue.api";
import { formatImageUrl } from "@/lib/utils";

interface EvidenceGroupProps {
  label: string;
  evidence: WOWRReportEvidence[];
  tone: "amber" | "emerald";
}

function EvidenceGroup({ label, evidence, tone }: EvidenceGroupProps) {
  const borderClass = tone === "emerald" ? "border-emerald-500/30" : "border-amber-500/30";
  const badgeClass = tone === "emerald"
    ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
    : "bg-amber-500/10 text-amber-600 dark:text-amber-400";

  return (
    <div className="min-w-[106px] space-y-1.5">
      <div className="flex items-center justify-between gap-2">
        <span className="text-[9px] font-bold uppercase tracking-wide text-muted-foreground">{label}</span>
        <span className={`rounded-full px-1.5 py-0.5 text-[9px] font-bold ${badgeClass}`}>{evidence.length}</span>
      </div>
      {evidence.length === 0 ? (
        <div className="flex h-12 items-center justify-center rounded-lg border border-dashed text-muted-foreground">
          <ImageIcon className="h-4 w-4 opacity-50" aria-hidden="true" />
        </div>
      ) : (
        <div className="flex gap-1.5">
          {evidence.slice(0, 2).map((photo, index) => (
            <a
              key={photo.photo_id}
              href={formatImageUrl(photo.image_url)}
              target="_blank"
              rel="noreferrer"
              className={`relative block h-12 w-12 overflow-hidden rounded-lg border bg-muted ${borderClass}`}
              title={`${label} ${index + 1}${photo.description ? ` — ${photo.description}` : ""}`}
            >
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={formatImageUrl(photo.image_url) || "/placeholder.png"}
                alt={`${label} ${index + 1}`}
                className="h-full w-full object-cover transition-transform hover:scale-110"
                onError={(event) => { event.currentTarget.src = "/placeholder.png"; }}
              />
              {index === 1 && evidence.length > 2 && (
                <span className="absolute inset-0 flex items-center justify-center bg-black/60 text-[10px] font-bold text-white">
                  +{evidence.length - 2}
                </span>
              )}
            </a>
          ))}
        </div>
      )}
    </div>
  );
}

interface WOWREvidencePreviewProps {
  initial: WOWRReportEvidence[];
  completion: WOWRReportEvidence[];
}

export function WOWREvidencePreview({ initial, completion }: WOWREvidencePreviewProps) {
  return (
    <div className="flex min-w-[224px] gap-3">
      <EvidenceGroup label="Temuan" evidence={initial} tone="amber" />
      <EvidenceGroup label="Penyelesaian" evidence={completion} tone="emerald" />
    </div>
  );
}
