import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import "@/test/shell-mocks";
import LoginPage from "./page";

vi.mock("@/lib/api/axios", () => ({ api: { post: vi.fn() } }));

describe("LoginPage", () => {
  it("follows the style guide order: logo, title, company, form, credit", () => {
    const { container } = render(<LoginPage />);
    expect(container.querySelector(".login-page")).not.toBeNull();
    expect(container.querySelector(".login-card")).not.toBeNull();
    expect(screen.getByAltText("Cimory")).toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 1, name: "I-GMP" })).toBeInTheDocument();
    expect(screen.getByText("PT Cisarua Mountain Dairy, Tbk — Plant Sentul")).toBeInTheDocument();
    expect(screen.getByLabelText("Username atau Email")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
    // White text on the navy-to-red gradient in both themes (the theme's
    // primary-foreground is dark navy in dark mode).
    const submit = screen.getByRole("button", { name: /Masuk/ });
    expect(submit).toHaveClass("login-btn", "text-white");
    expect(submit).not.toHaveClass("text-primary-foreground");
    expect(screen.getByText("Powered by Digital Transformation Plant Sentul")).toBeInTheDocument();
  });

  it("no longer hardcodes the old always-dark backdrop", () => {
    const { container } = render(<LoginPage />);
    expect(container.innerHTML).not.toMatch(/#050505|purple-600|zinc-/);
  });
});
