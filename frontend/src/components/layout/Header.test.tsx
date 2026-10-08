import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { nav } from "@/test/shell-mocks";
import { Header } from "./Header";

vi.mock("@/components/notifications/NotificationBell", () => ({
  NotificationBell: ({ triggerClassName }: { triggerClassName?: string }) => (
    <button type="button" className={triggerClassName}>bell</button>
  ),
}));
vi.mock("./InstallAppButton", () => ({ InstallAppButton: () => null }));
vi.mock("@/components/ui/ThemeToggle", () => ({
  ThemeToggle: ({ className }: { className?: string }) => (
    <button type="button" className={className}>theme</button>
  ),
}));
vi.mock("@/lib/api/auth.api", () => ({ authApi: { logout: vi.fn() } }));

describe("Header", () => {
  it("is the fixed 56px brand gradient bar with logo, app name and company line", () => {
    render(<Header />);
    const header = screen.getByRole("banner");
    expect(header).toHaveClass("app-header", "fixed", "top-0");
    expect(screen.getByAltText("Cimory")).toBeInTheDocument();
    expect(screen.getByText("I-GMP")).toBeInTheDocument();
    expect(screen.getByText("PT CISARUA MOUNTAIN DAIRY TBK")).toBeInTheDocument();
  });

  it("greets the signed-in user", () => {
    render(<Header />);
    expect(screen.getByText("Halo, Budi Auditor")).toBeInTheDocument();
  });

  it("renders header actions with the translucent on-brand style", () => {
    render(<Header />);
    for (const name of ["bell", "theme"]) {
      expect(screen.getByRole("button", { name })).toHaveClass("bg-white/15", "text-white");
    }
    expect(screen.getByRole("button", { name: /Keluar/ })).toHaveClass("bg-white/15", "text-white");
  });

  it.each([
    ["/cimory/SNT/dashboard/USR-1/wowr", "Perintah Kerja"],
    ["/cimory/SNT/dashboard/USR-1/kpi", "Dashboard KPI"],
    ["/cimory/SNT/dashboard/USR-1/gmp-data", "Data Inspeksi (GMP)"],
    ["/cimory/SNT/dashboard/USR-1", "Dasbor Utama"],
  ])("titles %s as %s", (pathname, title) => {
    nav.pathname = pathname;
    render(<Header />);
    expect(screen.getByRole("heading", { level: 1, name: title })).toBeInTheDocument();
  });
});
