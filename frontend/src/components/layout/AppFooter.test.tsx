import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { AppFooter } from "./AppFooter";

describe("AppFooter", () => {
  it("shows the Plant Sentul copyright with the current year and the credit line", () => {
    render(<AppFooter />);
    const footer = screen.getByRole("contentinfo");
    const year = new Date().getFullYear();
    expect(footer).toHaveTextContent(`© ${year} PT Cisarua Mountain Dairy, Tbk — Plant Sentul`);
    expect(footer).toHaveTextContent("Powered by Digital Transformation");
    expect(footer).toHaveClass("bg-footer", "text-muted-foreground");
  });
});
