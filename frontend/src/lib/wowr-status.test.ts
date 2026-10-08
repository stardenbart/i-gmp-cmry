import { describe, expect, it } from "vitest";

import { countWOWRByDisplayStatus, hasWOWRProof } from "./wowr-status";

const item = (wowr_status: string, urls: string[]) => ({ wowr_status, wowr_photos: urls.map((image_url) => ({ image_url })) });

describe("hasWOWRProof", () => {
  it("needs at least one proof photo with an image", () => {
    expect(hasWOWRProof(item("PendingValidation", []))).toBe(false);
    expect(hasWOWRProof(item("PendingValidation", [""]))).toBe(false);
    expect(hasWOWRProof(item("PendingValidation", ["/img/a.jpg"]))).toBe(true);
  });
});

describe("countWOWRByDisplayStatus", () => {
  it("counts PendingValidation without proof as awaiting execution proof", () => {
    const counts = countWOWRByDisplayStatus([
      item("None", []),
      item("PendingValidation", []),
      item("PendingValidation", ["/img/a.jpg"]),
      item("Verified", ["/img/b.jpg"]),
      item("Rejected", []),
    ]);
    expect(counts).toEqual({ AwaitingEvidence: 2, PendingValidation: 1, Verified: 1, Rejected: 1 });
  });
});
