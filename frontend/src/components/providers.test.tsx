import { act, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { THEME_INIT_SCRIPT } from "@/lib/theme";
import { ThemeProvider, useTheme } from "./providers";

function Probe() {
  const { theme, resolvedTheme } = useTheme();
  return <span data-testid="probe">{`${theme}/${resolvedTheme}`}</span>;
}

async function flushFrame() {
  await act(async () => {
    await new Promise((resolve) => requestAnimationFrame(() => resolve(null)));
  });
}

describe("ThemeProvider", () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove("dark");
    vi.stubGlobal("matchMedia", (q: string) => ({
      matches: false,
      media: q,
      addEventListener() {},
      removeEventListener() {},
    }));
  });
  afterEach(() => vi.unstubAllGlobals());

  it("defaults to the light theme", async () => {
    render(<ThemeProvider><Probe /></ThemeProvider>);
    expect(screen.getByTestId("probe")).toHaveTextContent("light/light");
    await flushFrame();
    expect(screen.getByTestId("probe")).toHaveTextContent("light/light");
    expect(document.documentElement).not.toHaveClass("dark");
  });

  it("restores a saved dark preference", async () => {
    localStorage.setItem("theme", "dark");
    render(<ThemeProvider><Probe /></ThemeProvider>);
    await flushFrame();
    expect(screen.getByTestId("probe")).toHaveTextContent("dark/dark");
    expect(document.documentElement).toHaveClass("dark");
  });
});

describe("THEME_INIT_SCRIPT", () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove("dark");
  });

  it("leaves the page light when nothing is saved", () => {
    new Function(THEME_INIT_SCRIPT)();
    expect(document.documentElement).not.toHaveClass("dark");
  });

  it("applies a saved dark theme before hydration", () => {
    localStorage.setItem("theme", "dark");
    new Function(THEME_INIT_SCRIPT)();
    expect(document.documentElement).toHaveClass("dark");
  });
});
