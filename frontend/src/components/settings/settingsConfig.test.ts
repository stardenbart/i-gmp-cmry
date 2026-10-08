import { describe, expect, it } from "vitest";

import { EMAIL_TEMPLATES, GENERAL_SETTINGS } from "./settingsConfig";

describe("settings config", () => {
  it("lets admins edit the CC list of the Kawasan report email", () => {
    const group = GENERAL_SETTINGS.find((g) => g.group === "Pengingat & Notifikasi");
    const cc = group?.keys.find((k) => k.key === "EMAIL_CC_KAWASAN_REPORT");
    expect(cc).toMatchObject({ type: "text" });
    expect(cc?.label).toMatch(/CC/);
  });

  it("no longer offers the old Kawasan template, the report layout is fixed", () => {
    expect(EMAIL_TEMPLATES.map((t) => t.key)).not.toContain("EMAIL_TEMPLATE_KAWASAN_CONFIRMED");
  });
});
