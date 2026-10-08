export type WOWRDisplayStatus = "AwaitingEvidence" | "PendingValidation" | "Verified" | "Rejected";

/**
 * What a WO/WR item is actually waiting for. Saving the WO/WR number already
 * stores PendingValidation, but there is nothing for the auditor to validate
 * until the PIC uploads execution proof, so that case reads as waiting for
 * proof instead.
 */
export function getWOWRDisplayStatus(status: string | null | undefined, hasProof: boolean): WOWRDisplayStatus {
  if (status === "Verified") return "Verified";
  if (status === "Rejected") return "Rejected";
  if (status === "PendingValidation" && hasProof) return "PendingValidation";
  return "AwaitingEvidence";
}

export const WOWR_DISPLAY_LABEL: Record<WOWRDisplayStatus, string> = {
  AwaitingEvidence: "Menunggu Bukti Eksekusi WO/WR",
  PendingValidation: "Menunggu Validasi Auditor",
  Verified: "Terverifikasi",
  Rejected: "Ditolak",
};

interface WOWRStatusItem {
  wowr_status?: string | null;
  wowr_photos: Array<{ image_url?: string | null }>;
}

/** True once the PIC has uploaded at least one WO/WR execution proof image. */
export function hasWOWRProof(item: WOWRStatusItem): boolean {
  return item.wowr_photos.some((photo) => Boolean(photo.image_url));
}

export function displayStatusOf(item: WOWRStatusItem): WOWRDisplayStatus {
  return getWOWRDisplayStatus(item.wowr_status, hasWOWRProof(item));
}

export function countWOWRByDisplayStatus(items: WOWRStatusItem[]): Record<WOWRDisplayStatus, number> {
  const counts: Record<WOWRDisplayStatus, number> = { AwaitingEvidence: 0, PendingValidation: 0, Verified: 0, Rejected: 0 };
  for (const item of items) counts[displayStatusOf(item)] += 1;
  return counts;
}
