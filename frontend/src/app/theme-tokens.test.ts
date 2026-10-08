import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";

const css = readFileSync(join(__dirname, "globals.css"), "utf8");

function block(selector: RegExp): string {
  const m = css.match(selector);
  if (!m) throw new Error(`block ${selector} not found`);
  return m[1];
}

function vars(body: string): Record<string, string> {
  const out: Record<string, string> = {};
  for (const m of body.matchAll(/--([\w-]+):\s*([^;]+);/g)) out[m[1]] = m[2].trim();
  return out;
}

const theme = vars(block(/@theme\s*\{([\s\S]*?)\n\}/));
const light = vars(block(/@layer base\s*\{[^}]*?:root\s*\{([\s\S]*?)\}/));
const dark = vars(block(/\n\s*\.dark\s*\{([\s\S]*?)\}/));

// Brand red (accent) is decoration only: white on #e63946 is 4.17:1, so
// text on red uses destructive (#c62833) instead.
// Semantic colors the components use (bg-*, text-*, ring-*). Each must be
// mapped in @theme or Tailwind v4 silently drops the utility.
const SEMANTIC = [
  "background", "foreground", "card", "card-foreground", "primary", "primary-foreground",
  "muted", "muted-foreground", "border", "ring", "destructive", "destructive-foreground",
  "accent", "success", "warning", "info",
  "sidebar", "sidebar-foreground", "sidebar-muted", "sidebar-active",
  "footer",
];

function luminance(hex: string): number {
  const h = hex.replace("#", "");
  const [r, g, b] = [0, 2, 4].map((i) => parseInt(h.slice(i, i + 2), 16) / 255).map((c) =>
    c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4,
  );
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrast(a: string, b: string): number {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (hi + 0.05) / (lo + 0.05);
}

describe("design tokens (globals.css)", () => {
  it.each(SEMANTIC)("maps --color-%s in @theme", (name) => {
    expect(theme[`color-${name}`]).toBe(`var(--${name})`);
  });

  it.each(SEMANTIC)("defines --%s for light and dark", (name) => {
    expect(light[name], `light --${name}`).toMatch(/^#[0-9a-f]{6}$/i);
    expect(dark[name], `dark --${name}`).toMatch(/^#[0-9a-f]{6}$/i);
  });

  it("uses the Cimory palette from the style guide in light mode", () => {
    expect(light).toMatchObject({
      background: "#f8fafc",
      foreground: "#1f2937",
      card: "#ffffff",
      border: "#e5e7eb",
      primary: "#1b3a6f",
      "muted-foreground": "#6b7280",
      accent: "#e63946",
      sidebar: "#1b3a6f",
    });
    expect(theme["color-brand-blue"]).toBe("#1b3a6f");
    expect(theme["color-brand-blue-dark"]).toBe("#122854");
    expect(theme["color-brand-red"]).toBe("#e63946");
  });

  it("uses the Segoe UI stack", () => {
    expect(theme["font-sans"]).toMatch(/^"Segoe UI"/);
  });

  // WCAG AA (4.5:1) for every text/background pair the shell relies on.
  const PAIRS: [string, string][] = [
    ["foreground", "background"],
    ["card-foreground", "card"],
    ["muted-foreground", "card"],
    ["muted-foreground", "background"],
    ["primary-foreground", "primary"],
    ["primary", "card"],
    ["destructive-foreground", "destructive"],
    ["destructive", "card"],
    ["success", "card"],
    ["warning", "card"],
    ["info", "card"],
    ["sidebar-foreground", "sidebar"],
    ["sidebar-muted", "sidebar"],
    ["sidebar-foreground", "sidebar-active"],
    ["muted-foreground", "footer"],
  ];

  for (const [mode, values] of [["light", light], ["dark", dark]] as const) {
    it.each(PAIRS)(`${mode}: %s on %s is at least 4.5:1`, (fg, bg) => {
      expect(contrast(values[fg], values[bg])).toBeGreaterThanOrEqual(4.5);
    });
  }
});
