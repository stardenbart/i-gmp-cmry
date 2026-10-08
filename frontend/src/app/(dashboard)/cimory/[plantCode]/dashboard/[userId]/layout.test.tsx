import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import DashboardLayout from "./layout";

vi.mock("@/components/layout/Sidebar", () => ({ Sidebar: () => <aside>sidebar</aside> }));
vi.mock("@/components/layout/Header", () => ({ Header: () => <header>header</header> }));
vi.mock("@/components/layout/BottomNav", () => ({ BottomNav: () => <nav>bottom-nav</nav> }));
vi.mock("@/components/layout/ScopeGuard", () => ({
  ScopeGuard: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

describe("DashboardLayout", () => {
  it("keeps the shell order header, sidebar, content, footer, bottom nav", () => {
    render(<DashboardLayout><p>page body</p></DashboardLayout>);

    const main = screen.getByRole("main");
    expect(main).toHaveTextContent("page body");
    expect(screen.getByRole("contentinfo")).toHaveTextContent("Plant Sentul");
    expect(screen.getByText("bottom-nav")).toBeInTheDocument();

    // Content column clears the 220px sidebar and the fixed header.
    const column = main.parentElement!;
    expect(column).toHaveClass("md:pl-[220px]");
    expect(column.firstElementChild).toHaveClass("pwa-header-safe");
    expect(main.nextElementSibling).toBe(screen.getByRole("contentinfo"));
  });
});
