import { describe, expect, it } from "vitest";

import { EMAIL_TEMPLATES, GENERAL_SETTINGS, SMTP_DEFAULTS, SMTP_SAVE_KEYS } from "./settingsConfig";

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

  it("saves the SMTP sender display name with the other SMTP fields", () => {
    expect(SMTP_DEFAULTS).toHaveProperty("SMTP_SENDER_NAME", "");
    expect(SMTP_SAVE_KEYS).toContain("SMTP_SENDER_NAME");
    expect(SMTP_SAVE_KEYS).toContain("SMTP_SENDER_EMAIL");
    expect(SMTP_SAVE_KEYS).not.toContain("SMTP_PASSWORD");
  });
});
