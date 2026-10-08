import { CheckCircle2, Clock, Loader2, XCircle } from "lucide-react";

import { cn } from "@/lib/utils";
import { WOWR_DISPLAY_LABEL, type WOWRDisplayStatus } from "@/lib/wowr-status";

const STYLE: Record<WOWRDisplayStatus, { icon: typeof Clock; className: string; spin?: boolean }> = {
  AwaitingEvidence: { icon: Clock, className: "bg-warning/10 text-warning border-warning/30" },
  PendingValidation: { icon: Loader2, className: "bg-info/10 text-info border-info/30", spin: true },
  Verified: { icon: CheckCircle2, className: "bg-success/10 text-success border-success/30" },
  Rejected: { icon: XCircle, className: "bg-destructive/10 text-destructive border-destructive/30" },
};

export function WOWRStatusBadge({ status }: { status: WOWRDisplayStatus }) {
  const { icon: Icon, className, spin } = STYLE[status];
  return (
    <span className={cn("inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[10px] font-bold", className)}>
      <Icon aria-hidden="true" className={cn("h-3 w-3", spin && "animate-spin")} />
      {WOWR_DISPLAY_LABEL[status]}
    </span>
  );
}
