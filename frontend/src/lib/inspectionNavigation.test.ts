import { describe, expect, it } from "vitest";

import { nextInspectionStep } from "./inspectionNavigation";

// detailsPerAspek[i] = number of details in aspek i; complete = every uraian scored.
const layout = { detailsPerAspek: [3, 2, 1] };

describe("nextInspectionStep", () => {
  it("moves to the next detail inside the current aspek", () => {
    expect(nextInspectionStep({ ...layout, aspekIndex: 0, detailIndex: 0, allScored: false })).toEqual({
      type: "open",
      aspekIndex: 0,
      detailIndex: 1,
    });
  });

  it("opens the next aspek after the last detail", () => {
    expect(nextInspectionStep({ ...layout, aspekIndex: 0, detailIndex: 2, allScored: false })).toEqual({
      type: "open",
      aspekIndex: 1,
      detailIndex: 0,
    });
  });

  it("opens the next aspek from the 'all details' view", () => {
    expect(nextInspectionStep({ ...layout, aspekIndex: 1, detailIndex: "all", allScored: false })).toEqual({
      type: "open",
      aspekIndex: 2,
      detailIndex: 0,
    });
  });

  it("goes back to the first unfinished aspek after the last one", () => {
    expect(
      nextInspectionStep({ ...layout, aspekIndex: 2, detailIndex: 0, allScored: false, firstIncompleteAspek: 1 }),
    ).toEqual({ type: "open", aspekIndex: 1, detailIndex: 0 });
  });

  it("points to the finish button once every uraian is scored", () => {
    expect(nextInspectionStep({ ...layout, aspekIndex: 0, detailIndex: 0, allScored: true })).toEqual({ type: "finish" });
  });
});
