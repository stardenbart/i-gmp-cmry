import type { LucideIcon } from "lucide-react";
import {
  AlertTriangle,
  CircleDot,
  Loader2,
  XCircle,
  CheckCircle2,
  ShieldAlert,
  Clock,
} from "lucide-react";

export interface PillStyle {
  label: string;
  text: string;
  bg: string;
  border: string;
  icon?: LucideIcon;
}

export type PanduanRole = "all" | "admin" | "auditor" | "auditee";

export const ROLE_STYLE: Record<PanduanRole, PillStyle> = {
  all: { label: "Semua Peran", text: "text-muted-foreground", bg: "bg-muted", border: "border-border" },
  admin: { label: "Admin & Super Admin", text: "text-primary", bg: "bg-primary/10", border: "border-primary/30" },
  auditor: { label: "Auditor", text: "text-sky-400", bg: "bg-sky-500/10", border: "border-sky-500/30" },
  auditee: { label: "Auditee / PIC", text: "text-rose-400", bg: "bg-rose-500/10", border: "border-rose-500/30" },
};

/**
 * Sama persis dengan `statusConfig` di halaman Temuan Inspeksi
 * (.../issues/page.tsx) — dipertahankan identik supaya pill di Panduan
 * cocok dengan pill yang sungguhan dilihat pengguna di aplikasi.
 */
export const ISSUE_STATUS_STYLE: Record<string, PillStyle> = {
  Open: { label: "Open", text: "text-blue-400", bg: "bg-blue-500/10", border: "border-blue-500/30", icon: AlertTriangle },
  InProgress: { label: "In Progress", text: "text-amber-400", bg: "bg-amber-500/10", border: "border-amber-500/30", icon: CircleDot },
  PendingValidation: { label: "Pending Validation", text: "text-purple-400", bg: "bg-purple-500/10", border: "border-purple-500/30", icon: Loader2 },
  Closed: { label: "Closed", text: "text-zinc-400", bg: "bg-zinc-500/10", border: "border-zinc-500/30", icon: XCircle },
  Verified: { label: "Verified", text: "text-emerald-400", bg: "bg-emerald-500/10", border: "border-emerald-500/30", icon: CheckCircle2 },
  Overdue: { label: "Overdue", text: "text-red-400", bg: "bg-red-500/10", border: "border-red-500/30", icon: ShieldAlert },
  ClosedOverdue: { label: "Closed Overdue", text: "text-orange-400", bg: "bg-orange-500/10", border: "border-orange-500/30", icon: CheckCircle2 },
};

/** Sama persis dengan badge status WOWR di .../monitoring/wowr/page.tsx. */
export const WOWR_STATUS_STYLE: Record<string, PillStyle> = {
  PendingValidation: { label: "Menunggu Validasi", text: "text-purple-500", bg: "bg-purple-500/10", border: "border-purple-500/20", icon: Clock },
  Verified: { label: "Terverifikasi", text: "text-emerald-500", bg: "bg-emerald-500/10", border: "border-emerald-500/20", icon: CheckCircle2 },
  Rejected: { label: "Ditolak", text: "text-red-500", bg: "bg-red-500/10", border: "border-red-500/20", icon: XCircle },
  None: { label: "Menunggu Bukti", text: "text-muted-foreground italic", bg: "bg-transparent", border: "border-transparent" },
};
