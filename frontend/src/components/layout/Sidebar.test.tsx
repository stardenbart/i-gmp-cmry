import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";

import { nav } from "@/test/shell-mocks";
import { Sidebar } from "./Sidebar";

describe("Sidebar", () => {
  beforeEach(() => {
    nav.pathname = "/cimory/SNT/dashboard/USR-1/issues";
  });

  it("sits on the brand sidebar surface below the 56px header", () => {
    render(<Sidebar />);
    const aside = screen.getByRole("complementary");
    expect(aside).toHaveClass("bg-sidebar", "text-sidebar-foreground", "top-14", "w-[220px]");
  });

  it("marks only the current page as active with the red accent bar", () => {
    render(<Sidebar />);
    const active = screen.getByRole("link", { name: /Temuan Inspeksi/ });
    expect(active).toHaveAttribute("aria-current", "page");
    expect(active).toHaveClass("bg-sidebar-active", "border-accent", "font-semibold");

    const other = screen.getByRole("link", { name: /Inspections/ });
    expect(other).not.toHaveAttribute("aria-current");
    expect(other).toHaveClass("border-transparent");
    const current = screen.getAllByRole("link").filter((a) => a.getAttribute("aria-current") === "page");
    expect(current).toHaveLength(1);
  });

  it("matches the dashboard home link exactly", () => {
    nav.pathname = "/cimory/SNT/dashboard/USR-1";
    render(<Sidebar />);
    expect(screen.getByRole("link", { name: /^Dasbor$/ })).toHaveAttribute("aria-current", "page");
  });
});
