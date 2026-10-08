import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Button } from "./button";
import { Input } from "./input";

describe("Button", () => {
  it("uses the 5px radius and a visible focus outline", () => {
    render(<Button>Simpan</Button>);
    const button = screen.getByRole("button", { name: "Simpan" });
    expect(button).toHaveClass("rounded-md", "bg-primary", "text-primary-foreground", "focus-visible:ring-2", "focus-visible:ring-ring");
    expect(button).not.toHaveClass("rounded-full");
  });

  it("renders destructive with its own readable foreground token", () => {
    render(<Button variant="destructive">Hapus</Button>);
    expect(screen.getByRole("button", { name: "Hapus" })).toHaveClass("bg-destructive", "text-destructive-foreground");
  });

  it("renders outline on the card surface", () => {
    render(<Button variant="outline">Batal</Button>);
    expect(screen.getByRole("button", { name: "Batal" })).toHaveClass("border-border", "bg-card", "text-foreground");
  });
});

describe("Input", () => {
  it("uses the 5px radius with a ring-colored focus border", () => {
    render(<Input aria-label="Nama" />);
    const input = screen.getByRole("textbox", { name: "Nama" });
    expect(input).toHaveClass("rounded-md", "border-border", "bg-card", "focus-visible:border-ring");
    expect(input).not.toHaveClass("rounded-2xl");
  });
});
