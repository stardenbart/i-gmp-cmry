export type InspectionStep = { type: "open"; aspekIndex: number; detailIndex: number } | { type: "finish" };

/**
 * What "Berikutnya" does on the inspection form: the next detail of the
 * current aspek, then the first detail of the next aspek; after the last
 * aspek, back to the first unfinished one. Once every uraian has a score it
 * points to the "Selesaikan Audit" button instead.
 */
export function nextInspectionStep({
  detailsPerAspek,
  aspekIndex,
  detailIndex,
  allScored,
  firstIncompleteAspek,
}: {
  detailsPerAspek: number[];
  aspekIndex: number;
  detailIndex: number | "all";
  allScored: boolean;
  firstIncompleteAspek?: number;
}): InspectionStep {
  if (allScored) return { type: "finish" };

  const details = detailsPerAspek[aspekIndex] ?? 0;
  if (typeof detailIndex === "number" && detailIndex < details - 1) {
    return { type: "open", aspekIndex, detailIndex: detailIndex + 1 };
  }
  if (aspekIndex < detailsPerAspek.length - 1) {
    return { type: "open", aspekIndex: aspekIndex + 1, detailIndex: 0 };
  }
  return { type: "open", aspekIndex: firstIncompleteAspek ?? 0, detailIndex: 0 };
}
