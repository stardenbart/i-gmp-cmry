import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { getWOWRDisplayStatus } from "@/lib/wowr-status";
import { WOWRStatusBadge } from "./WOWRStatusBadge";

describe("getWOWRDisplayStatus", () => {
  it("waits for execution proof until a WO/WR proof photo is uploaded", () => {
    expect(getWOWRDisplayStatus("None", false)).toBe("AwaitingEvidence");
    expect(getWOWRDisplayStatus(undefined, false)).toBe("AwaitingEvidence");
    // Saving the WO/WR number already sets PendingValidation in the backend,
    // but nothing can be validated before proof exists.
    expect(getWOWRDisplayStatus("PendingValidation", false)).toBe("AwaitingEvidence");
  });

  it("is pending validation only once proof is submitted", () => {
    expect(getWOWRDisplayStatus("PendingValidation", true)).toBe("PendingValidation");
  });

  it("keeps final decisions as they are", () => {
    expect(getWOWRDisplayStatus("Verified", true)).toBe("Verified");
    expect(getWOWRDisplayStatus("Rejected", false)).toBe("Rejected");
  });
});

describe("WOWRStatusBadge", () => {
  it.each([
    ["AwaitingEvidence", "Menunggu Bukti Eksekusi WO/WR"],
    ["PendingValidation", "Menunggu Validasi Auditor"],
    ["Verified", "Terverifikasi"],
    ["Rejected", "Ditolak"],
  ] as const)("labels %s as %s", (status, label) => {
    render(<WOWRStatusBadge status={status} />);
    expect(screen.getByText(label)).toBeInTheDocument();
  });
});
