import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Card } from "./card";

describe("Card", () => {
  it("is the white 8px card with a hairline border and thin shadow", () => {
    render(<Card data-testid="card" />);
    const card = screen.getByTestId("card");
    expect(card).toHaveClass("rounded-lg", "border-border", "bg-card", "shadow-[0_1px_4px_rgba(0,0,0,0.06)]");
    expect(card).not.toHaveClass("rounded-3xl");
  });
});
