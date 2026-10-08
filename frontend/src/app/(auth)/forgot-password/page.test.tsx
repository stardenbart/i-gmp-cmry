import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import "@/test/shell-mocks";
import ForgotPasswordPage from "./page";

vi.mock("@/lib/api/axios", () => ({ api: { post: vi.fn() } }));

describe("ForgotPasswordPage", () => {
  it("shares the login card look", () => {
    const { container } = render(<ForgotPasswordPage />);
    expect(container.querySelector(".login-page .login-card")).not.toBeNull();
    expect(screen.getByAltText("Cimory")).toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 1, name: "Lupa Password" })).toBeInTheDocument();
    expect(screen.getByLabelText("Alamat Email")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Kirim Kode OTP/ })).toHaveClass("login-btn");
    expect(screen.getByText("Powered by Digital Transformation Plant Sentul")).toBeInTheDocument();
  });

  it("no longer hardcodes the old always-dark backdrop", () => {
    const { container } = render(<ForgotPasswordPage />);
    expect(container.innerHTML).not.toMatch(/#050505|purple-600|zinc-|indigo-/);
  });
});
