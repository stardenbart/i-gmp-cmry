import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { EmailTemplatesPanel } from "./EmailTemplatesPanel";

const settings = [
  { setting_key: "EMAIL_TEMPLATE_ISSUE_ASSIGNMENT_ENABLED", setting_value: "false", plant_id: "" },
  { setting_key: "EMAIL_TEMPLATE_DEADLINE_REMINDER_ENABLED", setting_value: "true", plant_id: "" },
] as never;

function renderPanel(onToggle = vi.fn()) {
  render(<EmailTemplatesPanel isLoading={false} plantId="GLOBAL" settings={settings} onEdit={vi.fn()} onToggle={onToggle} />);
  return onToggle;
}

describe("EmailTemplatesPanel", () => {
  it("shows an on/off switch per optional template reflecting its _ENABLED setting", () => {
    renderPanel();
    const issue = screen.getByRole("switch", { name: /Penugasan Temuan/ });
    expect(issue).toHaveAttribute("aria-checked", "false");
    expect(screen.getByRole("switch", { name: /Pengingat Tenggat/ })).toHaveAttribute("aria-checked", "true");
    // Missing setting means enabled.
    expect(screen.getByRole("switch", { name: /Konfirmasi Inspeksi/ })).toHaveAttribute("aria-checked", "true");
  });

  it("does not let the password-reset OTP email be switched off", () => {
    renderPanel();
    expect(screen.queryByRole("switch", { name: /OTP Lupa Password/ })).toBeNull();
    const card = screen.getByText("OTP Lupa Password").closest("article")!;
    expect(within(card).getByText("Selalu aktif")).toBeInTheDocument();
  });

  it("reports the new state of a switch", () => {
    const onToggle = renderPanel();
    fireEvent.click(screen.getByRole("switch", { name: /Penugasan Temuan/ }));
    expect(onToggle).toHaveBeenCalledWith("EMAIL_TEMPLATE_ISSUE_ASSIGNMENT_ENABLED", true);
  });
});
