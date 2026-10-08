import { render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { StackedBar } from "./StackedBar";

const segments = [
  { key: "done", label: "Selesai", value: 3, className: "bg-success" },
  { key: "running", label: "Berlangsung", value: 1, className: "bg-info" },
];

describe("StackedBar", () => {
  it("draws every part of one total in a single bar", () => {
    render(<StackedBar label="Aktivitas inspeksi" total={4} segments={segments} />);

    const bar = screen.getByRole("img", { name: "Aktivitas inspeksi: Selesai 3 dari 4, Berlangsung 1 dari 4" });
    expect(within(bar).getByTestId("segment-done")).toHaveStyle({ width: "75%" });
    expect(within(bar).getByTestId("segment-running")).toHaveStyle({ width: "25%" });
    expect(within(bar).getByTestId("segment-done")).toHaveClass("bg-success");
  });

  it("lists each part with its count and share of the total", () => {
    render(<StackedBar label="Aktivitas inspeksi" total={4} segments={segments} />);

    const legend = screen.getByRole("list");
    const items = within(legend).getAllByRole("listitem");
    expect(items[0]).toHaveTextContent("Selesai");
    expect(items[0]).toHaveTextContent("3");
    expect(items[0]).toHaveTextContent("75%");
    expect(items[1]).toHaveTextContent("Berlangsung");
    expect(items[1]).toHaveTextContent("25%");
    // No "x / total" pairs that read like two separate totals.
    expect(legend).not.toHaveTextContent("/");
  });

  it("leaves zero-value parts out of the bar but keeps them in the legend", () => {
    render(
      <StackedBar
        label="WO/WR"
        total={2}
        segments={[
          { key: "done", label: "Selesai", value: 2, className: "bg-success" },
          { key: "rejected", label: "Ditolak", value: 0, className: "bg-destructive" },
        ]}
      />,
    );
    expect(screen.queryByTestId("segment-rejected")).toBeNull();
    expect(screen.getByText("Ditolak").closest("li")).toHaveTextContent("0%");
  });

  it("shows an empty track when there is nothing to count", () => {
    render(<StackedBar label="Aktivitas inspeksi" total={0} segments={segments.map((s) => ({ ...s, value: 0 }))} />);
    expect(screen.getByRole("img", { name: "Aktivitas inspeksi: belum ada data" })).toBeInTheDocument();
    expect(screen.queryByTestId("segment-done")).toBeNull();
  });
});
