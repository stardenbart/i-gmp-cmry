import { describe, expect, it } from "vitest";

import { GMP_TABLE } from "./gmpTableLayout";

describe("GMP data table layout", () => {
  it("lets the table grow past the screen and scroll sideways instead of squeezing columns", () => {
    expect(GMP_TABLE.scroller).toMatch(/\boverflow-x-auto\b/);
    expect(GMP_TABLE.table).toMatch(/\bw-max\b/);
    expect(GMP_TABLE.table).toMatch(/\bmin-w-full\b/);
    expect(GMP_TABLE.table).not.toMatch(/min-w-\[\d+px\]/);
  });

  it("gives the long text columns a readable width", () => {
    for (const column of [GMP_TABLE.uraian, GMP_TABLE.keterangan, GMP_TABLE.followUpKeterangan]) {
      expect(column).toMatch(/\bmin-w-\[(3[2-9]\d|4\d\d)px\]/);
      expect(column).toMatch(/\bwhitespace-normal\b/);
      expect(column).not.toMatch(/\bmax-w-sm\b/);
    }
  });

  it("keeps the scrollbar visible so users can see the table scrolls", () => {
    expect(GMP_TABLE.scroller).toMatch(/scrollbar-width:thin/);
  });
});
